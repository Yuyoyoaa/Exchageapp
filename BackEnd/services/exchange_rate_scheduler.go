package services

import (
	"context"
	"encoding/json" // 假设您有 config 包，否则请使用环境变量或常量
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

const (
	// 建议将 Key 放入 config.yaml 或环境变量
	ExchangeAPIKey       = "acbe932332b25dafdab979b5"
	ExchangeBaseURL      = "https://v6.exchangerate-api.com/v6/%s/latest/USD"
	ExchangeRateRedisKey = "rates:usd_base" // 使用 Hash 存储所有汇率
)

// FetchLatestRates 拉取 USD 基准汇率
func FetchLatestRates() (map[string]float64, error) {
	// 使用 config 包中的配置，如果没有则回退到常量
	// apiKey := config.AppConfig.ExchangeAPIKey
	apiKey := ExchangeAPIKey

	url := fmt.Sprintf(ExchangeBaseURL, apiKey)

	// 设置 HTTP Client 超时，防止请求挂起
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Result          string             `json:"result"`
		ConversionRates map[string]float64 `json:"conversion_rates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("json decode failed: %w", err)
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("API error: %s", result.Result)
	}

	return result.ConversionRates, nil
}

// StartExchangeRateScheduler 启动汇率定时任务
// 建议在 main.go 中传入 context 以便优雅关闭
func StartExchangeRateScheduler(ctx context.Context) {
	// 立即执行一次
	log.Println("Starting initial exchange rate update...")
	if err := updateRates(); err != nil {
		log.Printf("Initial rate update failed: %v\n", err)
	}

	// 定时器：每 24 小时执行一次
	ticker := time.NewTicker(24 * time.Hour)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Scheduled update triggered.")
				if err := updateRates(); err != nil {
					log.Printf("Scheduled rate update failed: %v\n", err)
				}
			case <-ctx.Done():
				log.Println("Stopping exchange rate scheduler...")
				return
			}
		}
	}()
}

// updateRates 核心逻辑：获取 -> 存DB -> 存Redis
func updateRates() error {
	ratesMap, err := FetchLatestRates()
	if err != nil {
		return err
	}

	now := time.Now()
	var rates []models.ExchangeRate

	// 1. 转换为 DB 模型
	// 优化策略：只存储 USD -> Any 的汇率 (约 160 条)，而不是 Any -> Any (25600 条)
	// 前端或其他服务计算 A -> B 时，公式为: (USD->B) / (USD->A)
	for code, rate := range ratesMap {
		rates = append(rates, models.ExchangeRate{
			FromCurrency: "USD",
			ToCurrency:   code,
			Rate:         rate,
			Date:         now,
		})
	}

	// 2. 数据库事务更新
	err = global.Db.Transaction(func(tx *gorm.DB) error {
		// 清空旧汇率表 (如果是只存最新汇率的策略)
		// 或者您可以选择保留历史记录，那样就不需要 Delete，只需 Insert
		if err := tx.Exec("DELETE FROM exchange_rates").Error; err != nil {
			return err
		}

		// 批量插入 (GORM v2 支持批量插入，性能很好)
		if len(rates) > 0 {
			// 分批次插入，防止 SQL 语句过长（虽然 160 条一次插入没问题）
			if err := tx.CreateInBatches(rates, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("db update failed: %w", err)
	}

	// 3. Redis 缓存优化
	// 使用 HSET 一次性写入所有汇率到 Hash 表中，避免成千上万个 Key
	// Key: "rates:usd_base", Field: "CNY", Value: "7.25"
	pipe := global.RedisDB.Pipeline()
	pipe.Del(context.Background(), ExchangeRateRedisKey) // 清除旧 Hash

	// 将 map[string]float64 转换为 map[string]interface{} 以适配 Redis HDel/HSet
	fields := make(map[string]interface{})
	for code, rate := range ratesMap {
		fields[code] = rate
	}

	pipe.HMSet(context.Background(), ExchangeRateRedisKey, fields)
	pipe.Expire(context.Background(), ExchangeRateRedisKey, 25*time.Hour) // 设置过期时间略大于更新间隔

	if _, err := pipe.Exec(context.Background()); err != nil {
		return fmt.Errorf("redis pipeline failed: %w", err)
	}

	log.Printf("Exchange rates updated successfully. Total currencies: %d\n", len(rates))
	return nil
}

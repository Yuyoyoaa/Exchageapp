package services

import (
	"context"
	"encoding/json"
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// FetchLatestRates 拉取 USD 基准汇率
func FetchLatestRates() (map[string]float64, error) {
	apiKey := "acbe932332b25dafdab979b5" // 替换成你的 API Key
	baseURL := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/USD", apiKey)

	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, fmt.Errorf("请求第三方API失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("第三方API返回状态码: %d", resp.StatusCode)
	}

	var result struct {
		Result          string             `json:"result"`
		ConversionRates map[string]float64 `json:"conversion_rates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("API返回失败: %v", result.Result)
	}

	return result.ConversionRates, nil
}

// StartExchangeRateScheduler 定时更新汇率
func StartExchangeRateScheduler() {
	updateRates := func() {
		fmt.Println("开始更新汇率:", time.Now())
		conversionRates, err := FetchLatestRates()
		if err != nil {
			fmt.Println("获取汇率失败:", err)
			return
		}

		now := time.Now()
		var rates []models.ExchangeRate

		// 生成任意货币对汇率
		for from, rateFrom := range conversionRates {
			for to, rateTo := range conversionRates {
				rateValue := 1.0
				if from != to {
					rateValue = rateTo / rateFrom
				}
				rates = append(rates, models.ExchangeRate{
					FromCurrency: from,
					ToCurrency:   to,
					Rate:         rateValue,
					Date:         now,
				})
			}
		}

		// 使用事务：先删除旧记录，再批量插入
		err = global.Db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("DELETE FROM exchange_rates").Error; err != nil {
				return err
			}
			if len(rates) > 0 {
				if err := tx.Create(&rates).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			fmt.Println("数据库更新失败:", err)
			return
		}

		// 缓存到 Redis
		for _, rate := range rates {
			cacheKey := fmt.Sprintf("exchangeRate:%s:%s", rate.FromCurrency, rate.ToCurrency)
			data, _ := json.Marshal(rate)
			global.RedisDB.Set(context.Background(), cacheKey, data, 24*time.Hour)
		}

		fmt.Println("汇率已更新")
	}

	// 启动时立即更新一次
	updateRates()

	// 每 24 小时更新一次
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			updateRates()
		}
	}()
}

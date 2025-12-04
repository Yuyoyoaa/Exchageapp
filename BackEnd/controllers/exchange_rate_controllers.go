package controllers

import (
	"context"
	"exchangeapp/global"
	"exchangeapp/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// 对应 Scheduler 中的 Key
	ExchangeRateRedisKey = "rates:usd_base"
)

// CreateExchangeRate 手动创建汇率 (仅用于测试或特殊补录)
func CreateExchangeRate(ctx *gin.Context) {
	var exchangeRate models.ExchangeRate
	if err := ctx.ShouldBindJSON(&exchangeRate); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exchangeRate.Date = time.Now()

	if err := global.Db.Create(&exchangeRate).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 注意：手动创建的汇率现在不会自动更新到 Redis 基准 Hash 中，
	// 除非它是 USD -> X 的汇率。这里仅做简单的单条缓存兼容。
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "汇率记录已创建 (注意：可能不影响实时计算逻辑)",
		"data":    exchangeRate,
	})
}

// GetLatestRate 获取最新汇率 (核心计算逻辑)
// 逻辑：利用 Redis 中的 USD 基准汇率进行实时换算
func GetLatestRate(ctx *gin.Context) {
	from := ctx.Query("from")
	to := ctx.Query("to")

	if from == "" || to == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数 from 和 to 必填"})
		return
	}

	// 1. 同币种直接返回
	if from == to {
		ctx.JSON(http.StatusOK, models.ExchangeRate{
			FromCurrency: from, ToCurrency: to, Rate: 1, Date: time.Now(),
		})
		return
	}

	// 2. 从 Redis Hash 中获取所有基准汇率
	// 这是一个 O(1) 操作，虽然取回了所有 160 个字段，但数据量很小 (几KB)，速度极快
	// 如果追求极致，可以使用 HMGet 只取 from 和 to 两个字段
	allRates, err := global.RedisDB.HGetAll(context.Background(), ExchangeRateRedisKey).Result()

	// 检查 Redis 是否有数据
	if err != nil || len(allRates) == 0 {
		// Redis 挂了或没数据，降级查 DB (尝试直接查库)
		var rate models.ExchangeRate
		if err := global.Db.Where("from_currency = ? AND to_currency = ?", from, to).Order("date desc").First(&rate).Error; err != nil {
			// 如果 DB 也只有 USD 基准，这里需要复杂的 SQL 自连接查询，或者直接报错
			// 为了简单，这里尝试找 USD 中转
			// 这是一个兜底策略，实际生产环境 Redis 不应该空
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "实时汇率暂时不可用"})
			return
		}
		ctx.JSON(http.StatusOK, rate)
		return
	}

	// 3. 解析汇率值
	// 我们需要: Rate(USD->From) 和 Rate(USD->To)
	// 如果 from 是 USD，Rate(USD->From) = 1
	rateUSDToFrom := 1.0
	rateUSDToTo := 1.0

	if from != "USD" {
		valStr, ok := allRates[from]
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "不支持的源货币: " + from})
			return
		}
		rateUSDToFrom, _ = strconv.ParseFloat(valStr, 64)
	}

	if to != "USD" {
		valStr, ok := allRates[to]
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "不支持的目标货币: " + to})
			return
		}
		rateUSDToTo, _ = strconv.ParseFloat(valStr, 64)
	}

	// 4. 计算交叉汇率
	// 公式: Rate(From -> To) = Rate(USD -> To) / Rate(USD -> From)
	if rateUSDToFrom == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "汇率数据异常"})
		return
	}
	finalRate := rateUSDToTo / rateUSDToFrom

	// 5. 构造返回结果
	response := models.ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         finalRate,
		Date:         time.Now(), // 这里其实应该取 Redis 里的更新时间，为了简化直接用 Now
	}

	ctx.JSON(http.StatusOK, response)
}

// 加一个专门返回所有 USD 基准汇率的接口，供前端初始化列表使用
func GetBaseRates(ctx *gin.Context) {
	// 1. 优先从 Redis 获取
	allRates, err := global.RedisDB.HGetAll(context.Background(), ExchangeRateRedisKey).Result()

	if err == nil && len(allRates) > 0 {
		var ratesList []models.ExchangeRate
		for currency, rateStr := range allRates {
			rate, _ := strconv.ParseFloat(rateStr, 64)
			ratesList = append(ratesList, models.ExchangeRate{
				FromCurrency: "USD",
				ToCurrency:   currency,
				Rate:         rate,
				Date:         time.Now(),
			})
		}
		ctx.JSON(http.StatusOK, ratesList)
		return
	}

	// 2. Redis 没数据，查数据库 (查所有 USD 开头的)
	var rates []models.ExchangeRate
	if err := global.Db.Where("from_currency = ?", "USD").Find(&rates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取汇率列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, rates)
}

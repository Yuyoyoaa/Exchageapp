package controllers

import (
	"context"
	"encoding/json"
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateExchangeRate 手动创建汇率
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

	cacheKey := fmt.Sprintf("exchangeRate:%s:%s", exchangeRate.FromCurrency, exchangeRate.ToCurrency)
	data, _ := json.Marshal(exchangeRate)
	global.RedisDB.Set(context.Background(), cacheKey, data, 24*time.Hour)

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "汇率创建成功",
		"data":    exchangeRate,
	})
}

// GetExchangeRates 查询汇率历史
func GetExchangeRates(ctx *gin.Context) {
	from := ctx.Query("from")
	to := ctx.Query("to")
	start := ctx.Query("start")
	end := ctx.Query("end")

	db := global.Db
	if from != "" && to != "" {
		db = db.Where("from_currency = ? AND to_currency = ?", from, to)
	}

	if start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			db = db.Where("date >= ?", t)
		}
	}
	if end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			db = db.Where("date <= ?", t)
		}
	}

	var rates []models.ExchangeRate
	if err := db.Order("date desc").Find(&rates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rates)
}

// GetLatestRate 获取最新汇率
func GetLatestRate(ctx *gin.Context) {
	from := ctx.Query("from")
	to := ctx.Query("to")

	if from == "" || to == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数 from 和 to 必填"})
		return
	}

	// 自对自汇率直接返回 1
	if from == to {
		ctx.JSON(http.StatusOK, models.ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         1,
			Date:         time.Now(),
		})
		return
	}

	cacheKey := fmt.Sprintf("exchangeRate:%s:%s", from, to)
	val, err := global.RedisDB.Get(context.Background(), cacheKey).Result()
	if err == nil && val != "" {
		var rate models.ExchangeRate
		json.Unmarshal([]byte(val), &rate)
		ctx.JSON(http.StatusOK, rate)
		return
	}

	var rate models.ExchangeRate
	if err := global.Db.Where("from_currency = ? AND to_currency = ?", from, to).Order("date desc").First(&rate).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "汇率未找到"})
		return
	}

	data, _ := json.Marshal(rate)
	global.RedisDB.Set(context.Background(), cacheKey, data, 24*time.Hour)

	ctx.JSON(http.StatusOK, rate)
}

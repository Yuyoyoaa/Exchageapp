package controllers

import (
	"encoding/json"
	"errors"
	"exchangeapp/global"
	"exchangeapp/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const CategoryCacheKey = "categories:list"

// ===== 公共接口 =====
func GetCategories(ctx *gin.Context) {
	// 尝试从缓存获取
	cached, err := global.RedisDB.Get(ctxRedis, CategoryCacheKey).Result()
	if err == nil {
		var categories []models.Category
		if json.Unmarshal([]byte(cached), &categories) == nil {
			ctx.JSON(http.StatusOK, categories)
			return
		}
	}

	// 缓存未命中，查询数据库
	var categories []models.Category
	if err := global.Db.Order("id ASC").Find(&categories).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 写入缓存
	data, _ := json.Marshal(categories)
	global.RedisDB.Set(ctxRedis, CategoryCacheKey, data, CacheExpire)

	ctx.JSON(http.StatusOK, categories)
}

// ===== 管理员接口 =====
func CreateCategory(ctx *gin.Context) {
	var category models.Category
	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := global.Db.Create(&category).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 清除缓存，下次读取刷新
	global.RedisDB.Del(ctxRedis, CategoryCacheKey)

	ctx.JSON(http.StatusCreated, category)
}

func DeleteCategory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	var category models.Category
	if err := global.Db.First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := global.Db.Delete(&category).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	// 清除缓存
	global.RedisDB.Del(ctxRedis, CategoryCacheKey)

	ctx.JSON(http.StatusOK, gin.H{"message": "分类已删除"})
}

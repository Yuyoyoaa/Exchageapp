package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var ctxRedis = context.Background()

const (
	ArticleListCachePrefix = "articles:list:"   // 分页+分类
	ArticleSingleCache     = "articles:single:" // 单篇文章缓存
	ArticleHotCache        = "articles:hot"     // 热门文章缓存
	CacheExpire            = 10 * time.Minute
)

// ======== 管理员文章操作 ========

func CreateArticle(ctx *gin.Context) {
	var article models.Article
	if err := ctx.ShouldBindJSON(&article); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	article.AuthorID = ctx.GetUint("userID")

	if err := global.Db.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clearArticleCache()
	ctx.JSON(http.StatusCreated, article)
}

func UpdateArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")
	var req models.Article
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var article models.Article
	if err := global.Db.First(&article, articleID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	article.Title = req.Title
	article.Content = req.Content
	article.Preview = req.Preview
	article.Cover = req.Cover
	article.CategoryID = req.CategoryID
	article.Status = req.Status

	if err := global.Db.Save(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clearArticleCacheByID(articleID)
	ctx.JSON(http.StatusOK, gin.H{"message": "文章已更新", "data": article})
}

func DeleteArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")
	if err := global.Db.Delete(&models.Article{}, articleID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	clearArticleCacheByID(articleID)
	ctx.JSON(http.StatusOK, gin.H{"message": "文章已删除"})
}

// ======== 文章列表（分页 + 分类 + 缓存） ========

func GetArticles(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")
	category := ctx.Query("category")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	cacheKey := fmt.Sprintf("%s%s:page_%d:limit_%d", ArticleListCachePrefix, category, page, limit)

	cached, err := global.RedisDB.Get(ctxRedis, cacheKey).Result()
	if err == nil {
		var articles []models.Article
		if json.Unmarshal([]byte(cached), &articles) == nil {
			ctx.JSON(http.StatusOK, articles)
			return
		}
	}

	var articles []models.Article
	db := global.Db
	if category != "" {
		db = db.Where("category_id = ?", category)
	}
	db = db.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC")
	if err := db.Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, _ := json.Marshal(articles)
	global.RedisDB.Set(ctxRedis, cacheKey, data, CacheExpire)
	ctx.JSON(http.StatusOK, articles)
}

// ======== 单篇文章缓存 + 浏览量 ========

func GetArticleByID(ctx *gin.Context) {
	id := ctx.Param("id")
	cacheKey := ArticleSingleCache + id

	cached, err := global.RedisDB.Get(ctxRedis, cacheKey).Result()
	if err == nil {
		var article models.Article
		if json.Unmarshal([]byte(cached), &article) == nil {
			global.Db.Model(&article).UpdateColumn("views_count", gorm.Expr("views_count + ?", 1))
			ctx.JSON(http.StatusOK, article)
			return
		}
	}

	var article models.Article
	if err := global.Db.First(&article, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	global.Db.Model(&article).UpdateColumn("views_count", gorm.Expr("views_count + ?", 1))

	data, _ := json.Marshal(article)
	global.RedisDB.Set(ctxRedis, cacheKey, data, CacheExpire)
	ctx.JSON(http.StatusOK, article)
}

// ======== 热门文章（按 views_count 排序前 10） ========

func GetHotArticles(ctx *gin.Context) {
	cached, err := global.RedisDB.Get(ctxRedis, ArticleHotCache).Result()
	if err == nil {
		var articles []models.Article
		if json.Unmarshal([]byte(cached), &articles) == nil {
			ctx.JSON(http.StatusOK, articles)
			return
		}
	}

	var articles []models.Article
	if err := global.Db.Order("views_count DESC").Limit(10).Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, _ := json.Marshal(articles)
	global.RedisDB.Set(ctxRedis, ArticleHotCache, data, CacheExpire)
	ctx.JSON(http.StatusOK, articles)
}

// ======== 收藏 ========

func ToggleFavorite(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	articleIDStr := ctx.Param("id")
	articleID, _ := strconv.Atoi(articleIDStr)

	var fav models.Favorite
	err := global.Db.Where("user_id = ? AND article_id = ?", userID, articleID).First(&fav).Error
	if err == nil {
		// 已收藏，取消
		global.Db.Delete(&fav)
		ctx.JSON(http.StatusOK, gin.H{"message": "取消收藏"})
		return
	}

	// 创建收藏
	fav = models.Favorite{UserID: userID, ArticleID: uint(articleID)}
	if err := global.Db.Create(&fav).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "收藏成功"})
}

// ======== 辅助函数：缓存清理 ========

func clearArticleCache() {
	keys := global.RedisDB.Keys(ctxRedis, ArticleListCachePrefix+"*").Val()
	for _, k := range keys {
		global.RedisDB.Del(ctxRedis, k)
	}
	global.RedisDB.Del(ctxRedis, ArticleHotCache)
}

func clearArticleCacheByID(id string) {
	global.RedisDB.Del(ctxRedis, ArticleSingleCache+id)
	clearArticleCache()
}

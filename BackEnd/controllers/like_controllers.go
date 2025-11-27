package controllers

import (
	"context"
	"encoding/json"
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// Redis key 前缀：用于存储文章点赞集合
	ArticleLikeSetPrefix = "article:likes:" // 存储 user_id 集合，key 为 article:likes:{articleID}
	// Redis key 前缀：用于存储点赞数计数
	ArticleLikeCountPrefix = "article:likes:count:" // 存储点赞数，key 为 article:likes:count:{articleID}
	// 缓存过期时间
	LikesCacheExpire = 24 * time.Hour
)

var ctxLike = context.Background()

// ===================== 点赞文章 =====================
// LikeArticle 点赞或取消点赞文章
// 第一次点赞：将用户ID加入Redis集合，点赞数+1
// 第二次点赞（取消）：将用户ID从Redis集合移除，点赞数-1
func LikeArticle(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	articleIDStr := ctx.Param("id")

	// 参数校验
	if articleIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "article id is required"})
		return
	}

	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	// 检查文章是否存在
	var article models.Article
	if err := global.Db.First(&article, articleID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	// Redis keys
	likeSetKey := fmt.Sprintf("%s%d", ArticleLikeSetPrefix, articleID)
	likeCountKey := fmt.Sprintf("%s%d", ArticleLikeCountPrefix, articleID)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	// 检查用户是否已经点赞过该文章
	isMember, err := global.RedisDB.SIsMember(ctxLike, likeSetKey, userIDStr).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check like status"})
		return
	}

	var message string
	var action string

	if isMember {
		// 用户已点赞，现在取消点赞
		// 从 Redis Set 中删除用户ID
		if err := global.RedisDB.SRem(ctxLike, likeSetKey, userIDStr).Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlike article"})
			return
		}

		// 点赞数 -1（如果存在的话）
		global.RedisDB.Decr(ctxLike, likeCountKey)

		// 更新数据库：点赞数 -1
		global.Db.Model(&article).Update("likes_count", article.LikesCount-1)

		// 同步删除点赞记录（可选，用于数据持久化）
		global.Db.Where("user_id = ? AND article_id = ?", userID, articleID).
			Delete(&models.ArticleLike{})

		message = "article unliked successfully"
		action = "unlike"
	} else {
		// 用户未点赞，现在点赞
		// 添加用户ID到 Redis Set
		if err := global.RedisDB.SAdd(ctxLike, likeSetKey, userIDStr).Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to like article"})
			return
		}

		// 设置 Redis key 过期时间（可选）
		global.RedisDB.Expire(ctxLike, likeSetKey, LikesCacheExpire)

		// 点赞数 +1
		global.RedisDB.Incr(ctxLike, likeCountKey)
		global.RedisDB.Expire(ctxLike, likeCountKey, LikesCacheExpire)

		// 更新数据库：点赞数 +1
		global.Db.Model(&article).Update("likes_count", article.LikesCount+1)

		// 同步添加点赞记录到数据库（用于数据持久化）
		articleLike := models.ArticleLike{
			UserID:    userID,
			ArticleID: uint(articleID),
		}
		if err := global.Db.Create(&articleLike).Error; err != nil {
			// 如果数据库操作失败，从 Redis 回滚
			global.RedisDB.SRem(ctxLike, likeSetKey, userIDStr)
			global.RedisDB.Decr(ctxLike, likeCountKey)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save like record"})
			return
		}

		message = "article liked successfully"
		action = "like"
	}

	// 同步文章缓存（立即更新单篇文章和列表缓存）
	syncArticleCacheAfterLike(uint(articleID))

	// 获取当前点赞数
	likesCount, _ := global.RedisDB.Get(ctxLike, likeCountKey).Int64()

	ctx.JSON(http.StatusOK, gin.H{
		"message":     message,
		"action":      action,
		"article_id":  articleID,
		"user_id":     userID,
		"likes_count": likesCount,
	})
}

// ===================== 获取文章点赞数 =====================
// GetArticleLikes 获取文章的总点赞数
func GetArticleLikes(ctx *gin.Context) {
	articleIDStr := ctx.Param("id")

	// 参数校验
	if articleIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "article id is required"})
		return
	}

	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	// 检查文章是否存在
	var article models.Article
	if err := global.Db.First(&article, articleID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	// Redis key
	likeCountKey := fmt.Sprintf("%s%d", ArticleLikeCountPrefix, articleID)

	// 从 Redis 获取点赞数
	likesCount, err := global.RedisDB.Get(ctxLike, likeCountKey).Int64()
	if err != nil {
		// 如果 Redis 中没有该 key，从数据库获取
		likesCount = article.LikesCount
		// 同步到 Redis
		global.RedisDB.Set(ctxLike, likeCountKey, likesCount, LikesCacheExpire)
	}

	// 获取当前用户的点赞状态（如果已认证）
	var userLiked bool
	if userID, exists := ctx.Get("userID"); exists {
		likeSetKey := fmt.Sprintf("%s%d", ArticleLikeSetPrefix, articleID)
		userIDStr := strconv.FormatUint(uint64(userID.(uint)), 10)
		isMember, err := global.RedisDB.SIsMember(ctxLike, likeSetKey, userIDStr).Result()
		if err == nil {
			userLiked = isMember
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"article_id":  articleID,
		"likes_count": likesCount,
		"user_liked":  userLiked,
	})
}

// ===================== 辅助函数 =====================

// GetUserLikeStatus 获取用户对某篇文章的点赞状态（辅助函数）
func GetUserLikeStatus(userID uint, articleID uint) (bool, error) {
	likeSetKey := fmt.Sprintf("%s%d", ArticleLikeSetPrefix, articleID)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	isMember, err := global.RedisDB.SIsMember(ctxLike, likeSetKey, userIDStr).Result()
	return isMember, err
}

// SyncLikesFromDB 从数据库同步点赞数到 Redis（用于数据恢复）
func SyncLikesFromDB(articleID uint) error {
	var article models.Article
	if err := global.Db.First(&article, articleID).Error; err != nil {
		return err
	}

	likeCountKey := fmt.Sprintf("%s%d", ArticleLikeCountPrefix, articleID)
	return global.RedisDB.Set(ctxLike, likeCountKey, article.LikesCount, LikesCacheExpire).Err()
}

func syncArticleCacheAfterLike(articleID uint) {
	// 更新单篇文章缓存
	cacheKey := fmt.Sprintf("articles:single:%d", articleID)
	var article models.Article
	if err := global.Db.First(&article, articleID).Error; err == nil {
		data, _ := json.Marshal(article)
		global.RedisDB.Set(ctxLike, cacheKey, data, 10*time.Minute)
	}

	// 清理分页列表缓存，让列表接口显示最新点赞数
	keys := global.RedisDB.Keys(ctxLike, "articles:list:*").Val()
	for _, k := range keys {
		global.RedisDB.Del(ctxLike, k)
	}
}

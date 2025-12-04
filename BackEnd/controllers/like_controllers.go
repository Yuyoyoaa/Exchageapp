package controllers

import (
	"context"
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

const (
	// Redis key 前缀
	ArticleLikeSetPrefix   = "article:likes:"       // Set: 存储 user_id
	ArticleLikeCountPrefix = "article:likes:count:" // String: 存储点赞数
	LikesCacheExpire       = 24 * time.Hour
)

var ctxLike = context.Background()

// ===================== 点赞文章 =====================
// LikeArticle 点赞或取消点赞文章
func LikeArticle(ctx *gin.Context) {
	// 1. 解析参数
	userID := ctx.GetUint("userID")
	articleIDStr := ctx.Param("id")
	if articleIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Article ID is required"})
		return
	}
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Article ID"})
		return
	}
	uid := uint(articleID)

	// Redis Keys
	likeSetKey := fmt.Sprintf("%s%d", ArticleLikeSetPrefix, uid)
	likeCountKey := fmt.Sprintf("%s%d", ArticleLikeCountPrefix, uid)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	// 2. 检查点赞状态 (优先查 Redis，减少 DB 压力)
	// 注意：如果 Redis 里的数据因过期丢失，这里可能会误判为未点赞。
	// 为严谨起见，如果 Redis 查不到，应该去 DB 查一次。
	_, err = global.RedisDB.SIsMember(ctxLike, likeSetKey, userIDStr).Result()
	if err != nil {
		// Redis 出错时降级或报错，这里选择报错
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "System error: cache check failed"})
		return
	}

	// 3. 执行逻辑
	var action string
	var finalCount int64

	// 开启 DB 事务，确保数据一致性
	err = global.Db.Transaction(func(tx *gorm.DB) error {
		// 检查文章是否存在 (加锁读取防止并发删除，可选)
		// 只要 ID 存在即可，不需要查出所有字段
		var exists int64
		if err := tx.Model(&models.Article{}).Where("id = ?", uid).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return errors.New("article not found")
		}

		// 这里再次确认 DB 中的点赞状态，防止 Redis 数据不一致
		var likeRecord models.ArticleLike
		result := tx.Where("user_id = ? AND article_id = ?", userID, uid).First(&likeRecord)
		isLikedInDB := result.Error == nil

		if isLikedInDB {
			// === 取消点赞 ===
			action = "unlike"

			// 1. 删除 DB 记录
			if err := tx.Where("user_id = ? AND article_id = ?", userID, uid).Delete(&models.ArticleLike{}).Error; err != nil {
				return err
			}

			// 2. 原子减少 DB 点赞数
			if err := tx.Model(&models.Article{}).Where("id = ?", uid).
				UpdateColumn("likes_count", gorm.Expr("likes_count - 1")).Error; err != nil {
				return err
			}

			// 3. 更新 Redis (事务提交前执行，或者 defer 执行)
			// 为了简单，这里直接执行，如果事务失败 Redis 会有短暂不一致，但可以通过 TTL 自动修复
			global.RedisDB.SRem(ctxLike, likeSetKey, userIDStr)
			finalCount = global.RedisDB.Decr(ctxLike, likeCountKey).Val()

		} else {
			// === 点赞 ===
			action = "like"

			// 1. 创建 DB 记录
			newLike := models.ArticleLike{UserID: userID, ArticleID: uid}
			if err := tx.Create(&newLike).Error; err != nil {
				return err
			}

			// 2. 原子增加 DB 点赞数
			if err := tx.Model(&models.Article{}).Where("id = ?", uid).
				UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
				return err
			}

			// 3. 更新 Redis
			global.RedisDB.SAdd(ctxLike, likeSetKey, userIDStr)
			// 刷新 Set 过期时间，保证热点数据常驻
			global.RedisDB.Expire(ctxLike, likeSetKey, LikesCacheExpire)

			finalCount = global.RedisDB.Incr(ctxLike, likeCountKey).Val()
			global.RedisDB.Expire(ctxLike, likeCountKey, LikesCacheExpire)
		}

		return nil
	})

	if err != nil {
		if err.Error() == "article not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update like status"})
		}
		return
	}

	// 4. 返回结果
	// 注意：这里不再暴力清理所有列表缓存。
	// 前端列表页的点赞数可以容忍短暂不一致，或者前端手动 +1/-1。
	// 只有文章详情缓存建议清理。
	go func() {
		// 仅清除单篇文章缓存，让下次请求重新加载最新数据
		global.RedisDB.Del(ctxLike, fmt.Sprintf("articles:single:%d", uid))
	}()

	ctx.JSON(http.StatusOK, gin.H{
		"message":     fmt.Sprintf("Article %sd successfully", action),
		"action":      action,
		"article_id":  uid,
		"user_id":     userID,
		"likes_count": finalCount,
	})
}

// ===================== 获取文章点赞数 =====================
func GetArticleLikes(ctx *gin.Context) {
	articleIDStr := ctx.Param("id")
	if articleIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Article ID is required"})
		return
	}
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Article ID"})
		return
	}
	uid := uint(articleID)

	likeCountKey := fmt.Sprintf("%s%d", ArticleLikeCountPrefix, uid)

	// 1. 尝试从 Redis 获取点赞数
	val, err := global.RedisDB.Get(ctxLike, likeCountKey).Result()
	var likesCount int64

	if err == nil {
		likesCount, _ = strconv.ParseInt(val, 10, 64)
	} else {
		// Redis 未命中，查 DB 并回填
		var article models.Article
		if err := global.Db.Select("likes_count").First(&article, uid).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		likesCount = article.LikesCount
		// 回填 Redis
		global.RedisDB.Set(ctxLike, likeCountKey, likesCount, LikesCacheExpire)
	}

	// 2. 获取当前用户是否点赞
	var userLiked bool
	if userIDVal, exists := ctx.Get("userID"); exists {
		userID := userIDVal.(uint)
		likeSetKey := fmt.Sprintf("%s%d", ArticleLikeSetPrefix, uid)
		userIDStr := strconv.FormatUint(uint64(userID), 10)

		// 查 Redis Set
		isMember, err := global.RedisDB.SIsMember(ctxLike, likeSetKey, userIDStr).Result()
		if err == nil {
			userLiked = isMember
		} else {
			// 如果 Redis Set 也不存在（可能过期了），查 DB 兜底
			var count int64
			global.Db.Model(&models.ArticleLike{}).Where("user_id = ? AND article_id = ?", userID, uid).Count(&count)
			userLiked = count > 0
			// 如果查到了，顺便回填 Redis Set (Lazy load)
			if userLiked {
				global.RedisDB.SAdd(ctxLike, likeSetKey, userIDStr)
				global.RedisDB.Expire(ctxLike, likeSetKey, LikesCacheExpire)
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"article_id":  uid,
		"likes_count": likesCount,
		"user_liked":  userLiked,
	})
}

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
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func LikeArticle(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	articleID := ctx.Param("id")

	// Redis key
	likeCountKey := fmt.Sprintf("article:%s:likes_count", articleID)
	userSetKey := fmt.Sprintf("article:%s:liked_users", articleID)

	var article models.Article
	if err := global.Db.First(&article, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否已经点赞
	exists, err := global.RedisDB.SIsMember(ctx.Request.Context(), userSetKey, userID).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "你已点赞过"})
		return
	}

	// 将用户加入已点赞集合
	if err := global.RedisDB.SAdd(ctx.Request.Context(), userSetKey, userID).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 设置集合过期时间，防止无限增长
	global.RedisDB.Expire(ctx.Request.Context(), userSetKey, 24*time.Hour*30)

	// 点赞计数 +1
	newCount, err := global.RedisDB.Incr(ctx.Request.Context(), likeCountKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 设置过期时间，确保同步数据库后仍能缓存热点数据
	global.RedisDB.Expire(ctx.Request.Context(), likeCountKey, 24*time.Hour*30)

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "点赞成功",
		"likesCount": newCount,
	})
}

// 后台定时同步 Redis 点赞到数据库
func SyncLikesToDB() {
	for {
		time.Sleep(5 * time.Minute) // 每5分钟同步一次
		var articles []models.Article
		global.Db.Find(&articles)

		for _, a := range articles {
			likeKey := fmt.Sprintf("article:%d:likes_count", a.ID)
			ctx := context.Background() // ✅ 这里创建 context

			countStr, err := global.RedisDB.Get(ctx, likeKey).Result()
			if err == nil {
				count, _ := strconv.ParseInt(countStr, 10, 64)
				global.Db.Model(&a).UpdateColumn("likes_count", count)
			}
		}
	}
}

func GetArticleLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")
	likeKey := fmt.Sprintf("article:%s:likes_count", articleID)

	countStr, err := global.RedisDB.Get(ctx.Request.Context(), likeKey).Result()
	if err == redis.Nil {
		// 缓存不存在，读取数据库
		var article models.Article
		if err := global.Db.Select("likes_count").First(&article, articleID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		countStr = strconv.FormatInt(article.LikesCount, 10)
		// 写入 Redis 缓存
		global.RedisDB.Set(ctx.Request.Context(), likeKey, countStr, 24*time.Hour*30)
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"likesCount": countStr})
}

package controllers

import (
	"exchangeapp/global"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":likes"

	// 获取点赞后的数量
	likes, err := global.RedisDB.Incr(ctx.Request.Context(), likeKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Successfully Liked",
		"likes":   likes, // 返回点赞数量保持一致性
	})
}

func GetArticleLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":likes"

	likes, err := global.RedisDB.Get(ctx.Request.Context(), likeKey).Result()

	if err == redis.Nil { // key不存在
		likes = "0"
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"likes": likes,
	})
}

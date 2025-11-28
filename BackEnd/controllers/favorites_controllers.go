package controllers

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ======== 收藏 / 取消收藏 ========
func ToggleFavorite(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	articleIDStr := ctx.Param("id")

	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文章 ID 无效"})
		return
	}

	var fav models.Favorite

	// 查是否已收藏
	err = global.Db.Where("user_id = ? AND article_id = ?", userID, articleID).First(&fav).Error
	if err == nil {
		// 已收藏 → 取消
		global.Db.Delete(&fav)
		ctx.JSON(http.StatusOK, gin.H{"message": "取消收藏"})
		return
	}

	// 创建收藏
	fav = models.Favorite{
		UserID:    userID,
		ArticleID: uint(articleID),
	}

	if err := global.Db.Create(&fav).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "收藏成功"})
}

// ======== 获取用户收藏的文章 ========
func GetUserFavorites(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	var favorites []models.Favorite

	// 用 Preload("Article") 自动加载文章内容
	if err := global.Db.
		Preload("Article").
		Where("user_id = ?", userID).
		Find(&favorites).Error; err != nil {

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, favorites)
}

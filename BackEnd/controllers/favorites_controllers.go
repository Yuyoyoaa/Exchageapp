package controllers

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

// Add this to BackEnd/controllers/article_controllers.go

// ======== 获取用户收藏的文章 ========
func GetUserFavorites(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	var favorites []models.Favorite
	if err := global.Db.Where("user_id = ?", userID).Find(&favorites).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(favorites) == 0 {
		ctx.JSON(http.StatusOK, []models.Article{})
		return
	}

	var articleIDs []uint
	for _, f := range favorites {
		articleIDs = append(articleIDs, f.ArticleID)
	}

	var articles []models.Article
	if err := global.Db.Where("id IN ?", articleIDs).Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, articles)
}

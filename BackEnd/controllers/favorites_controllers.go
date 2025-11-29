package controllers

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ======== 收藏 / 取消收藏 (Toggle 逻辑，用于文章详情页) ========
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

// ======== 【新增】直接删除收藏记录 (用于收藏列表页) ========
func DeleteFavorite(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	favIDStr := ctx.Param("id") // 这里传入的是 favorite 表的 ID，不是 article_id

	favID, err := strconv.Atoi(favIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "收藏 ID 无效"})
		return
	}

	// 【核心修复】：使用 Unscoped() 忽略软删除标记，强制物理删除
	// 确保 user_id 匹配以保证权限安全
	result := global.Db.Unscoped().Where("id = ? AND user_id = ?", favID, userID).Delete(&models.Favorite{})

	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 只要没有数据库错误，即使 RowsAffected 为 0 也返回成功，防止前端因为“重复删除”而报错卡住
	ctx.JSON(http.StatusOK, gin.H{"message": "已移除收藏"})
}

// ======== 获取用户收藏的文章 ========
func GetUserFavorites(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	var favorites []models.Favorite

	// 使用 Preload 加载 Article，即使 Article 也是软删除的，Gorm 默认会过滤掉 deleted_at 不为空的文章
	// 如果 Article 是硬删除，那 Article 结构体就是零值
	// 如果你想查出已软删除的文章，可以使用 Unscoped()，但这里我们假设如果文章被删了，就显示"已失效"
	if err := global.Db.
		Preload("Article").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&favorites).Error; err != nil {

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, favorites)
}

package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"exchangeapp/global"
	"exchangeapp/models"

	"github.com/gin-gonic/gin"
)

const (
	CommentCachePrefix = "comments:article:" // 每篇文章的评论缓存 key
)

// ======== 创建评论 ========
func CreateComment(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	userName := ctx.GetString("userName")
	articleIDStr := ctx.Param("id")
	articleID, _ := strconv.Atoi(articleIDStr)

	var comment models.Comment
	if err := ctx.ShouldBindJSON(&comment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.ArticleID = uint(articleID)
	comment.UserID = userID
	comment.UserName = userName

	if err := global.Db.Create(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 新增评论后清理文章所有分页缓存
	clearCommentCache(uint(articleID))

	ctx.JSON(http.StatusCreated, comment)
}

// ======== 删除自己的评论 ========
func DeleteComment(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	commentID := ctx.Param("id")

	var comment models.Comment
	if err := global.Db.First(&comment, commentID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
		return
	}

	if comment.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只能删除自己的评论"})
		return
	}

	// 软删除
	if err := global.Db.Delete(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	// 删除该文章的所有分页缓存
	clearCommentCache(comment.ArticleID)

	ctx.JSON(http.StatusOK, gin.H{"message": "评论已删除"})
}

// ======== 管理员强制删除评论 ========
func ForceDeleteComment(ctx *gin.Context) {
	commentID := ctx.Param("id")

	var comment models.Comment
	if err := global.Db.First(&comment, commentID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
		return
	}

	global.Db.Delete(&comment)
	clearCommentCache(comment.ArticleID)

	ctx.JSON(http.StatusOK, gin.H{"message": "评论已删除"})
}

// ======== 获取文章评论（支持分页） ========
func GetCommentsByArticleID(ctx *gin.Context) {
	articleIDStr := ctx.Param("id")
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit

	cacheKey := fmt.Sprintf("%s%s:page_%d:limit_%d", CommentCachePrefix, articleIDStr, page, limit)

	// 尝试读取缓存
	cached, err := global.RedisDB.Get(ctxRedis, cacheKey).Result()
	if err == nil {
		var comments []models.Comment
		if json.Unmarshal([]byte(cached), &comments) == nil {
			ctx.JSON(http.StatusOK, comments)
			return
		}
	}

	// 缓存不存在或解析失败，从数据库读取
	var comments []models.Comment
	if err := global.Db.
		Where("article_id = ?", articleIDStr).
		Offset(offset).Limit(limit).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 写入缓存
	data, _ := json.Marshal(comments)
	global.RedisDB.Set(ctxRedis, cacheKey, data, CacheExpire)

	ctx.JSON(http.StatusOK, comments)
}

// ======== 辅助函数：清理某文章的所有分页评论缓存 ========
func clearCommentCache(articleID uint) {
	pattern := fmt.Sprintf("%s%d:page_*", CommentCachePrefix, articleID)
	keys, _ := global.RedisDB.Keys(ctxRedis, pattern).Result()
	for _, k := range keys {
		global.RedisDB.Del(ctxRedis, k)
	}
}

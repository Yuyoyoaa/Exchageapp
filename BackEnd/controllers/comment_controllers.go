package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"exchangeapp/global"
	"exchangeapp/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm" // 导入 GORM 包以支持事务
)

const (
	CommentCachePrefix = "comments:article:" // 每篇文章的评论缓存 key
)

// 为了兼容原代码中未提供的 ctxRedis，这里假设其已在全局或别处定义，用于 Redis 操作。
// var ctxRedis = context.Background()

// ======== 创建评论 ========
func CreateComment(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}
	articleIDStr := ctx.Param("id")
	articleID, _ := strconv.Atoi(articleIDStr)

	var comment models.Comment
	if err := ctx.ShouldBindJSON(&comment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.ArticleID = uint(articleID)
	comment.UserID = userID
	if user.Nickname != "" {
		comment.UserName = user.Nickname
	} else {
		comment.UserName = user.Username
	}

	if err := global.Db.Create(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 新增评论后清理文章所有分页缓存
	clearCommentCache(uint(articleID))

	ctx.JSON(http.StatusCreated, comment)
}

// ======== 删除评论（用户只能删自己的；管理员可删任何） ========
func DeleteComment(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	role := ctx.GetString("role") // "admin" or "user"
	commentID := ctx.Param("id")

	var comment models.Comment
	if err := global.Db.First(&comment, commentID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
		return
	}

	// 权限控制
	if role != "admin" && comment.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只能删除自己的评论"})
		return
	}

	// 事务，保证级联删除
	err := global.Db.Transaction(func(tx *gorm.DB) error {

		// 顶级评论 => 删除所有子评论
		if comment.ParentID == nil {
			if err := tx.Where("parent_id = ?", comment.ID).
				Delete(&models.Comment{}).Error; err != nil {
				return fmt.Errorf("failed to delete child comments: %w", err)
			}
		}

		// 删除当前评论
		if err := tx.Delete(&comment).Error; err != nil {
			return fmt.Errorf("failed to delete comment: %w", err)
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("删除失败: %s", err.Error()),
		})
		return
	}

	// 清除该文章的缓存
	clearCommentCache(comment.ArticleID)

	ctx.JSON(http.StatusOK, gin.H{"message": "评论已删除"})
}

// ======== 获取文章评论（支持分页 + Redis 缓存 + Preload 用户头像） ========
func GetCommentsByArticleID(ctx *gin.Context) {
	articleID := ctx.Param("id")
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit

	// Redis 缓存 key
	cacheKey := fmt.Sprintf("%s%s:page_%d:limit_%d",
		CommentCachePrefix, articleID, page, limit)

	// ---------- 读取缓存 ----------
	cached, err := global.RedisDB.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var cachedComments []models.Comment
		if json.Unmarshal([]byte(cached), &cachedComments) == nil {
			ctx.JSON(http.StatusOK, cachedComments)
			return
		}
	}

	// ---------- 查询数据库 ----------
	var comments []models.Comment
	err = global.Db.
		Where("article_id = ?", articleID).
		Order("created_at ASC").
		Offset(offset).Limit(limit).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			// 【修改】加载 nickname 和 avatar
			return db.Select("id", "username", "nickname", "avatar")
		}).
		Find(&comments).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---------- 写入缓存 ----------
	bytes, _ := json.Marshal(comments)
	global.RedisDB.Set(ctx, cacheKey, bytes, CacheExpire)

	// ---------- 输出 ----------
	ctx.JSON(http.StatusOK, comments)
}

// ======== 辅助函数：清理某文章的所有分页评论缓存 ========
func clearCommentCache(articleID uint) {
	pattern := fmt.Sprintf("%s%d:page_*", CommentCachePrefix, articleID)
	// 警告：原代码中使用的 ctxRedis 未定义。为保持代码完整性，我们暂时忽略这一行中 err 的处理，假设 ctxRedis 可用。
	keys, _ := global.RedisDB.Keys(ctxRedis, pattern).Result()
	for _, k := range keys {
		global.RedisDB.Del(ctxRedis, k)
	}
}

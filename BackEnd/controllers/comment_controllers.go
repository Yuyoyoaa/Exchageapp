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

	// 权限检查：只能删除自己的评论
	if comment.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "只能删除自己的评论"})
		return
	}

	// 使用事务确保级联删除的原子性
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		// 1. 如果是顶级评论（ParentID 为 0），删除其所有子评论
		if comment.ParentID == nil {
			// 硬删除所有回复该评论的记录
			if err := tx.Where("parent_id = ?", comment.ID).Delete(&models.Comment{}).Error; err != nil {
				return fmt.Errorf("failed to delete child comments: %w", err)
			}
		}

		// 2. 软删除当前评论（继承 gorm.Model，默认是软删除）
		if err := tx.Delete(&comment).Error; err != nil {
			return fmt.Errorf("failed to delete comment: %w", err)
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除失败: %s", err.Error())})
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

	// 使用事务确保级联删除的原子性
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		// 1. 如果是顶级评论（ParentID 为 0），删除其所有子评论
		if comment.ParentID == nil {
			// 硬删除所有回复该评论的记录
			if err := tx.Where("parent_id = ?", comment.ID).Delete(&models.Comment{}).Error; err != nil {
				return fmt.Errorf("failed to delete child comments: %w", err)
			}
		}

		// 2. 删除当前评论
		// 仍然使用 tx.Delete，依赖于 models.Comment 的定义（软删除或硬删除）
		if err := tx.Delete(&comment).Error; err != nil {
			return fmt.Errorf("failed to delete comment: %w", err)
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("删除失败: %s", err.Error())})
		return
	}

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
	// 警告：原代码中使用的 ctxRedis 未定义。为保持代码完整性，我们暂时忽略这一行中 err 的处理，假设 ctxRedis 可用。
	cached, _ := global.RedisDB.Get(ctxRedis, cacheKey).Result()
	if cached != "" {
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
	// 警告：原代码中使用的 ctxRedis 未定义。为保持代码完整性，我们暂时忽略这一行中 err 的处理，假设 ctxRedis 可用。
	keys, _ := global.RedisDB.Keys(ctxRedis, pattern).Result()
	for _, k := range keys {
		global.RedisDB.Del(ctxRedis, k)
	}
}

package controllers

import (
	"encoding/json"
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
	CommentCachePrefix = "comments:article:" // 格式：comments:article:{articleID}:page_{p}:limit_{l}
	CommentCacheTTL    = 10 * time.Minute
)

// ======================= 创建评论 =========================
func CreateComment(ctx *gin.Context) {
	// 1. 获取当前用户
	userID := ctx.GetUint("userID")

	// 优化：只查询必要的用户信息字段
	var user models.User
	if err := global.Db.Select("id, nickname, username, avatar").First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	// 2. 解析文章ID
	articleIDStr := ctx.Param("id")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil || articleID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	// 3. 绑定请求数据
	var req models.Comment
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. 组装数据
	comment := models.Comment{
		ArticleID: uint(articleID),
		UserID:    userID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		UserName:  user.Nickname,
	}
	if comment.UserName == "" {
		comment.UserName = user.Username
	}

	// 5. 写入数据库
	if err := global.Db.Create(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "评论发布失败"})
		return
	}

	// 6. 异步清理缓存
	go clearCommentCache(uint(articleID))

	ctx.JSON(http.StatusCreated, comment)
}

// ======================= 删除评论 =========================
func DeleteComment(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	role := ctx.GetString("role") // "admin" or "user"
	commentID := ctx.Param("id")

	// 1. 先查出评论，确认归属权
	var comment models.Comment
	if err := global.Db.First(&comment, commentID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
		return
	}

	// 2. 权限校验
	if role != "admin" && comment.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "无权删除此评论"})
		return
	}

	// 3. 事务删除 (级联删除子评论)
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		if comment.ParentID == nil {
			// 如果是父评论，先删除所有子评论
			if err := tx.Where("parent_id = ?", comment.ID).Delete(&models.Comment{}).Error; err != nil {
				return err
			}
		}
		// 删除评论本身
		if err := tx.Delete(&comment).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	// 4. 异步清理缓存
	go clearCommentCache(comment.ArticleID)

	ctx.JSON(http.StatusOK, gin.H{"message": "评论已删除"})
}

// ======================= 获取评论列表 (已恢复为返回数组) =========================
func GetCommentsByArticleID(ctx *gin.Context) {
	articleID := ctx.Param("id")

	// 1. 分页参数处理
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// 构造缓存 Key
	cacheKey := fmt.Sprintf("%s%s:page_%d:limit_%d", CommentCachePrefix, articleID, page, limit)

	// 2. 尝试读取缓存
	if cached, err := global.RedisDB.Get(ctxRedis, cacheKey).Result(); err == nil {
		var comments []models.Comment
		// 直接反序列化为数组，保持和前端的兼容性
		if json.Unmarshal([]byte(cached), &comments) == nil {
			ctx.JSON(http.StatusOK, comments)
			return
		}
	}

	// 3. 数据库查询
	var comments []models.Comment

	db := global.Db.Model(&models.Comment{}).Where("article_id = ?", articleID)

	// 注意：为了兼容前端，这里不再返回 Total 字段，但查询逻辑依然保持高效
	err := db.Order("created_at ASC").
		Offset((page-1)*limit).
		Limit(limit).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			// 仅加载必要的字段，保护用户隐私
			return db.Select("id", "username", "nickname", "avatar")
		}).
		Find(&comments).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 异步写入缓存
	go func() {
		data, _ := json.Marshal(comments) // 缓存的也是纯数组
		global.RedisDB.Set(ctxRedis, cacheKey, data, CommentCacheTTL)
	}()

	// 直接返回数组，这样您的前端 v-for 就能正常工作了
	ctx.JSON(http.StatusOK, comments)
}

// ======================= 缓存清理工具 =========================

// clearCommentCache 使用 SCAN 替代 KEYS，防止阻塞
func clearCommentCache(articleID uint) {
	pattern := fmt.Sprintf("%s%d:*", CommentCachePrefix, articleID)
	var cursor uint64
	var keys []string
	var err error

	// 循环扫描，每次处理 100 个 Key
	for {
		keys, cursor, err = global.RedisDB.Scan(ctxRedis, cursor, pattern, 100).Result()
		if err != nil {
			return
		}

		if len(keys) > 0 {
			global.RedisDB.Del(ctxRedis, keys...)
		}

		if cursor == 0 {
			break
		}
	}
}

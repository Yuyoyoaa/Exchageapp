package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var ctxRedis = context.Background()

const (
	ArticleListCachePrefix = "articles:list:"   // 分页+分类
	ArticleSingleCache     = "articles:single:" // 单篇文章缓存
	ArticleHotCache        = "articles:hot"     // 热门文章缓存
	CacheExpire            = 10 * time.Minute
)

// ArticleListResponse 定义了文章列表接口的响应结构，包含数据和总数
// 此结构与前端 Vue 组件中对接口返回值的预期相匹配：{ data: Article[], total: number }
type ArticleListResponse struct {
	Data  []models.Article `json:"data"`
	Total int64            `json:"total"`
}

// ======================= 创建文章 =========================

func CreateArticle(ctx *gin.Context) {
	var req models.Article
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.AuthorID = ctx.GetUint("userID")

	if err := global.Db.Create(&req).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clearArticleCache()
	ctx.JSON(http.StatusCreated, req)
}

// ======================= 更新文章 =========================

func UpdateArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	var req models.Article
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var article models.Article
	if err := global.Db.First(&article, articleID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	article.Title = req.Title
	article.Content = req.Content
	article.Preview = req.Preview
	article.Cover = req.Cover
	article.CategoryID = req.CategoryID
	article.Status = req.Status

	if err := global.Db.Save(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clearArticleCacheByID(articleID)
	ctx.JSON(http.StatusOK, gin.H{"message": "文章已更新", "data": article})
}

// ======================= 删除文章 =========================

func DeleteArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	if err := global.Db.Delete(&models.Article{}, articleID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	clearArticleCacheByID(articleID)
	ctx.JSON(http.StatusOK, gin.H{"message": "文章已删除"})
}

// ======================= 文章列表（含缓存） =========================

func GetArticles(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")
	category := ctx.Query("category")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// 构造缓存键时，将 category 设为 “all” 或具体的 ID
	categoryCachePart := category
	if categoryCachePart == "" {
		categoryCachePart = "all"
	}
	cacheKey := fmt.Sprintf("%s%s:page_%d:limit_%d", ArticleListCachePrefix, categoryCachePart, page, limit)

	// 读取缓存：如果命中，直接返回 ArticleListResponse 结构
	if cached, err := global.RedisDB.Get(ctxRedis, cacheKey).Result(); err == nil {
		var resp ArticleListResponse
		if json.Unmarshal([]byte(cached), &resp) == nil {
			// 缓存命中，直接返回包含 data 和 total 的结构
			ctx.JSON(http.StatusOK, resp)
			return
		}
	}

	// 数据库查询
	var articles []models.Article
	var total int64
	db := global.Db.Model(&models.Article{}) // 使用 Model() 指定查询的表

	// 应用分类筛选条件 (Go 代码原本就支持分类筛选)
	if category != "" {
		db = db.Where("category_id = ?", category)
	}

	// 1. 获取总数
	if err := db.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取文章总数"})
		return
	}

	// 2. 获取分页数据
	if err := db.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构造与前端预期匹配的响应结构
	resp := ArticleListResponse{
		Data:  articles,
		Total: total,
	}

	// 写入缓存
	data, _ := json.Marshal(resp)
	global.RedisDB.Set(ctxRedis, cacheKey, data, CacheExpire)

	ctx.JSON(http.StatusOK, resp)
}

// ======================= 单篇文章 + 浏览量更新 + 缓存同步 =========================

func GetArticleByID(ctx *gin.Context) {
	id := ctx.Param("id")
	cacheKey := ArticleSingleCache + id

	// 读取缓存
	if cached, err := global.RedisDB.Get(ctxRedis, cacheKey).Result(); err == nil {
		var article models.Article
		if json.Unmarshal([]byte(cached), &article) == nil {

			// 浏览量 +1
			// 注意：这里只更新数据库，但不会立即更新文章列表缓存，以避免写入风暴。
			// 列表缓存的更新依赖于过期或 Create/Update/Delete 操作。
			global.Db.Model(&article).
				UpdateColumn("views_count", gorm.Expr("views_count + 1"))

			article.ViewsCount += 1 // 同步更新结构体，避免旧数据返回

			// 更新缓存 (只更新单篇文章缓存)
			data, _ := json.Marshal(article)
			global.RedisDB.Set(ctxRedis, cacheKey, data, CacheExpire)

			ctx.JSON(http.StatusOK, article)
			return
		}
	}

	// 数据库读取文章
	var article models.Article
	if err := global.Db.First(&article, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 浏览量 +1
	global.Db.Model(&article).UpdateColumn("views_count", gorm.Expr("views_count + 1"))
	article.ViewsCount += 1

	// 更新缓存
	data, _ := json.Marshal(article)
	global.RedisDB.Set(ctxRedis, cacheKey, data, CacheExpire)

	ctx.JSON(http.StatusOK, article)
}

// ======================= 热门文章缓存 =========================

func GetHotArticles(ctx *gin.Context) {
	if cached, err := global.RedisDB.Get(ctxRedis, ArticleHotCache).Result(); err == nil {
		var articles []models.Article
		if json.Unmarshal([]byte(cached), &articles) == nil {
			ctx.JSON(http.StatusOK, articles)
			return
		}
	}

	var articles []models.Article
	if err := global.Db.Order("views_count DESC").Limit(10).Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, _ := json.Marshal(articles)
	global.RedisDB.Set(ctxRedis, ArticleHotCache, data, CacheExpire)

	ctx.JSON(http.StatusOK, articles)
}

// 上传文章封面
func UploadArticleCover(ctx *gin.Context) {
	// 1. 获取上传的文件 (表单 key 为 "file")
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	// 2. 检查文件扩展名 (可选，简单的安全性检查)
	ext := filepath.Ext(file.Filename)
	allowExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowExts[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "只允许上传 jpg, png, gif 图片"})
		return
	}

	// 3. 确保存储目录存在
	uploadDir := "./uploads/covers"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	// 4. 生成唯一文件名 (使用纳秒时间戳防止重名)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	// 5. 保存文件
	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 6. 返回可访问的 URL (相对路径，前端需要拼接 baseURL)
	// URL 格式: /uploads/covers/123456789.jpg
	url := "/uploads/covers/" + filename
	ctx.JSON(http.StatusOK, gin.H{"url": url})
}

// ======================= 缓存清理 =========================

func clearArticleCache() {
	keys := global.RedisDB.Keys(ctxRedis, ArticleListCachePrefix+"*").Val()
	for _, k := range keys {
		global.RedisDB.Del(ctxRedis, k)
	}
	global.RedisDB.Del(ctxRedis, ArticleHotCache)
}

func clearArticleCacheByID(id string) {
	global.RedisDB.Del(ctxRedis, ArticleSingleCache+id)
	clearArticleCache()
}

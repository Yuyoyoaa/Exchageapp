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
	ArticleSingleCache     = "articles:single:" // 单篇文章详情缓存
	ArticleHotCache        = "articles:hot"     // 热门文章缓存
	ArticleViewKey         = "articles:views:"  // 浏览量缓冲池 (Hash or String)
	CacheExpire            = 10 * time.Minute
	UploadDir              = "./uploads/covers"
)

// 允许的图片扩展名
var allowExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// ArticleListResponse 定义了文章列表接口的响应结构
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

	// 清理相关缓存
	go clearArticleCache() // 异步清理，不阻塞主请求
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

	// 更新字段
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

	go clearArticleCacheByID(articleID)
	ctx.JSON(http.StatusOK, gin.H{"message": "文章已更新", "data": article})
}

// ======================= 删除文章 =========================

func DeleteArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	// 开启事务
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		// 1. 删除该文章的所有评论
		if err := tx.Where("article_id = ?", articleID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}

		// 2. 删除该文章的所有点赞
		if err := tx.Where("article_id = ?", articleID).Delete(&models.ArticleLike{}).Error; err != nil {
			return err
		}

		// 3. 删除该文章的所有收藏
		if err := tx.Where("article_id = ?", articleID).Delete(&models.Favorite{}).Error; err != nil {
			return err
		}

		// 4. 删除文章本身
		if err := tx.Delete(&models.Article{}, articleID).Error; err != nil {
			return err
		}

		return nil // 提交事务
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}

	// 清理缓存
	go clearArticleCacheByID(articleID)
	ctx.JSON(http.StatusOK, gin.H{"message": "文章及其关联数据已删除"})
}

// ======================= 文章列表（含缓存） =========================

func GetArticles(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1") // 获取名为 "page" 的查询参数，如果不存在则返回默认值 "1"
	limitStr := ctx.DefaultQuery("limit", "10")
	category := ctx.Query("category")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// 构造缓存键
	categoryCachePart := category
	if categoryCachePart == "" {
		categoryCachePart = "all"
	}
	cacheKey := fmt.Sprintf("%s%s:page_%d:limit_%d", ArticleListCachePrefix, categoryCachePart, page, limit)

	// 1. 尝试读取缓存
	if cached, err := global.RedisDB.Get(ctxRedis, cacheKey).Result(); err == nil {
		var resp ArticleListResponse
		// 将 JSON 数据转换为 Go 结构体
		if json.Unmarshal([]byte(cached), &resp) == nil {
			ctx.JSON(http.StatusOK, resp)
			return
		}
	}

	// 2. 数据库查询
	var articles []models.Article
	var total int64
	// 明确指定查询的目标数据表Article
	db := global.Db.Model(&models.Article{})

	if category != "" {
		db = db.Where("category_id = ?", category)
	}

	// 计算符合当前条件的记录总数（分页功能需要知道总记录数，才能计算总页数）
	if err := db.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取文章总数"})
		return
	}

	// 按 created_at字段降序排列
	if err := db.Order("created_at DESC").
		Offset((page - 1) * limit). // 设置查询的偏移量，用于分页
		Limit(limit).               // 限制返回的记录数量
		Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := ArticleListResponse{
		Data:  articles,
		Total: total,
	}

	// 3. 写入缓存 (异步写入，减少接口延迟)
	go func() {
		data, _ := json.Marshal(resp)
		global.RedisDB.Set(ctxRedis, cacheKey, data, CacheExpire)
	}()

	ctx.JSON(http.StatusOK, resp)
}

// ======================= 单篇文章 + 浏览量优化 =========================

func GetArticleByID(ctx *gin.Context) {
	id := ctx.Param("id")
	cacheKey := ArticleSingleCache + id

	// 尝试读取缓存
	var article models.Article
	cached, err := global.RedisDB.Get(ctxRedis, cacheKey).Result()
	cacheHit := err == nil

	if cacheHit {
		json.Unmarshal([]byte(cached), &article)
	} else {
		// 缓存未命中，查库
		if err := global.Db.First(&article, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
	}

	// 异步更新浏览量（Fire and Forget）
	// 这样用户获取文章详情的速度不受 DB 写入速度影响
	go func(aid string) {
		// 1. DB 浏览量 +1
		if err := global.Db.Model(&models.Article{}).Where("id = ?", aid).
			UpdateColumn("views_count", gorm.Expr("views_count + 1")).Error; err == nil {

			// 2. 如果缓存存在，更新缓存中的 ViewsCount，或者直接删除缓存让下次重建
			// 为了数据实时性，这里选择删除单条缓存（简单粗暴但有效）或者重新 Set
			// 鉴于浏览量是高频变动，仅仅为了浏览量失效整个缓存代价太大。
			// 更好的做法是：列表页/详情页的浏览量不走强一致缓存，或者前端单独调接口获取实时浏览量。
			// 这里保持原逻辑：仅更新 DB，不强制刷新缓存（因为缓存几分钟后过期自然会刷新）
		}
	}(id)

	// 如果是缓存命中，views_count 是旧的。
	// 为了前端展示体验，可以在返回前手动 +1
	if cacheHit {
		article.ViewsCount++
	}

	// 只有在缓存未命中时，才需要回填缓存
	if !cacheHit {
		data, _ := json.Marshal(article)
		global.RedisDB.Set(ctxRedis, cacheKey, data, CacheExpire)
	}

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

	// 异步写入缓存
	go func() {
		data, _ := json.Marshal(articles)
		global.RedisDB.Set(ctxRedis, ArticleHotCache, data, CacheExpire)
	}()

	ctx.JSON(http.StatusOK, articles)
}

// 上传文章封面
func UploadArticleCover(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if !allowExts[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式"})
		return
	}

	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误: 无法创建目录"})
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(UploadDir, filename)

	if err := ctx.SaveUploadedFile(file, dst); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 统一 URL 路径处理
	url := "/uploads/covers/" + filename
	ctx.JSON(http.StatusOK, gin.H{"url": url})
}

// ======================= 缓存清理工具 =========================

// clearArticleCache 使用 SCAN 命令替代 KEYS，避免阻塞 Redis
func clearArticleCache() {
	var cursor uint64
	var keys []string
	var err error

	// 使用 Scan 迭代查找匹配的 key
	// 注意：生产环境如果有数百万 key，Scan 仍然需要一定时间，但不会阻塞
	for {
		keys, cursor, err = global.RedisDB.Scan(ctxRedis, cursor, ArticleListCachePrefix+"*", 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			global.RedisDB.Del(ctxRedis, keys...)
		}
		if cursor == 0 {
			break
		}
	}

	// 同时清除热门文章缓存
	global.RedisDB.Del(ctxRedis, ArticleHotCache)
}

func clearArticleCacheByID(id string) {
	// 1. 删除详情缓存
	global.RedisDB.Del(ctxRedis, ArticleSingleCache+id)
	// 2. 列表缓存也需要清理（因为标题、封面等可能变了）
	clearArticleCache()
}

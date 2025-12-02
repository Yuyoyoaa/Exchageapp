package router

import (
	"exchangeapp/controllers"
	"exchangeapp/middlewares"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 从环境读取上传目录，默认 ./uploads
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	// 转为绝对路径，便于调试
	absUploadDir, _ := filepath.Abs(uploadDir)
	// 静态资源映射
	r.Static("/uploads", absUploadDir)

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ===== 用户注册登录 =====
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/register", controllers.Register)
	}

	// ===== 公共 API =====
	api := r.Group("/api")
	{
		api.GET("/exchangeRates", controllers.GetExchangeRates)
		api.GET("/exchangeRates/latest", controllers.GetLatestRate)

		// 文章公共接口（无需登录）
		api.GET("/articles", controllers.GetArticles) // 分页 + 分类
		api.GET("/articles/hot", controllers.GetHotArticles)
		api.GET("/articles/:id", controllers.GetArticleByID) // 文章详情（自动 views++）
		api.GET("/articles/:id/comments", controllers.GetCommentsByArticleID)
		api.GET("/categories", controllers.GetCategories)
		api.GET("/articles/:id/like", controllers.GetArticleLikes)
	}

	// ===== 需要认证的接口 =====
	authAPI := api.Group("")
	authAPI.Use(middlewares.AuthMiddleWare())
	{
		// 普通用户操作（点赞、评论）
		authAPI.POST("/articles/:id/like", controllers.LikeArticle)        // 每用户只能点赞一次
		authAPI.POST("/articles/:id/favorite", controllers.ToggleFavorite) // 收藏/取消收藏(在新闻页面操作)
		authAPI.POST("/articles/:id/comments", controllers.CreateComment)  // 创建评论
		authAPI.DELETE("/comments/:id", controllers.DeleteComment)         // 删除自己评论
		authAPI.DELETE("/favorites/:id", controllers.DeleteFavorite)       // 通过收藏ID删除收藏（在个人收藏区操作）

		// 用户信息
		user := authAPI.Group("/user")
		{
			user.GET("/profile", controllers.GetProfile)
			user.PUT("/profile", controllers.UpdateProfile)
			user.GET("/favorites", controllers.GetUserFavorites)
			user.POST("/upload/avatar", controllers.UploadAvatar)
		}

		// ===== 管理员 API =====
		admin := authAPI.Group("/admin")
		admin.Use(middlewares.AdminMiddleware())
		{
			// 管理员 - 用户管理
			admin.GET("/users", controllers.GetUserList)
			admin.PATCH("/users/:id/role", controllers.UpdateUserRole)
			admin.DELETE("/users/:id", controllers.DeleteUser)

			// 管理员 - 文章管理
			admin.POST("/articles", controllers.CreateArticle)       // 只有管理员能发文章
			admin.PUT("/articles/:id", controllers.UpdateArticle)    // 管理员编辑文章
			admin.DELETE("/articles/:id", controllers.DeleteArticle) // 管理员删除文章
			admin.POST("/articles/upload/cover", controllers.UploadArticleCover)

			// 分类管理
			admin.POST("/categories", controllers.CreateCategory)
			admin.DELETE("/categories/:id", controllers.DeleteCategory)
		}
	}

	return r
}

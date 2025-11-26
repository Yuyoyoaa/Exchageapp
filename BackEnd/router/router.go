package router

import (
	"exchangeapp/controllers"
	"exchangeapp/middlewares"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 允许跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, // 添加 PATCH
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 用户注册与登录
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/register", controllers.Register)
	}

	api := r.Group("/api")
	{
		// 不需要认证的公共接口
		api.GET("/exchangeRates", controllers.GetExchangeRates)

		// 需要认证的接口组
		authenticated := api.Group("")
		authenticated.Use(middlewares.AuthMiddleWare())
		{
			authenticated.POST("/exchangeRates", controllers.CreateExchangeRate)
			authenticated.POST("/articles", controllers.CreateArticle)
			authenticated.GET("/articles", controllers.GetArticles)
			authenticated.GET("/articles/:id", controllers.GetArticleByID)
			authenticated.POST("/articles/:id/like", controllers.LikeArticle)
			authenticated.GET("/articles/:id/like", controllers.GetArticleLikes)

			// 用户信息接口
			user := authenticated.Group("/user")
			{
				user.GET("/profile", controllers.GetProfile)
				user.PUT("/profile", controllers.UpdateProfile)
			}

			// 管理员专用接口组
			adminGroup := authenticated.Group("/admin")
			adminGroup.Use(middlewares.AdminMiddleware())
			{
				adminGroup.GET("/users", controllers.GetUserList)
				adminGroup.PATCH("/users/:id/role", controllers.UpdateUserRole)
				adminGroup.DELETE("/users/:id", controllers.DeleteUser)
			}
		}
	}

	return r
}

package middlewares

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"exchangeapp/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 验证JWT是否有效

func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			ctx.Abort()
			return
		}

		username, role, err := utils.ParseJWT(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			ctx.Abort()
			return
		}

		// 查询用户ID
		var user models.User
		if err := global.Db.Where("username = ?", username).First(&user).Error; err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			ctx.Abort()
			return
		}

		// 设置到 context
		ctx.Set("userID", user.ID)
		ctx.Set("username", username)
		ctx.Set("role", role)
		ctx.Next()
	}
}

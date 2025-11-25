package middlewares

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 检查角色
func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username, exists := ctx.Get("username")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}
		var user models.User
		if err := global.Db.Where("username = ?", username.(string)).First(&user).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			ctx.Abort()
			return
		}
		if user.Role != "admin" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

package middlewares

import (
	"exchangeapp/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 验证JWT是否有效

func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			ctx.Abort() // 阻止后续的中间件和处理函数执行,但不会中断当前函数中 ctx.Abort()之后的代码执行
			return
		}
		username, err := utils.ParseJWT(token)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token: " + err.Error(),
			})
			ctx.Abort()
			return
		}

		ctx.Set("username", username) // 用于在不同中间件传递kv数据
		ctx.Next()

	}
}

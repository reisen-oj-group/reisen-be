package middleware

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *service.AuthService, required bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		// 没有 Authorization 头的情况
		if authHeader == "" {
			if required {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
				return
			}
			ctx.Set("user", (*model.User)(nil))
			ctx.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Bearer token 格式不正确的情况
		if tokenString == authHeader {
			if required {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
				return
			}
			ctx.Set("user", (*model.User)(nil))
			ctx.Next()
			return
		}

		// 解析 token 失败的情况
		user, err := authService.ParseToken(tokenString)
		if err != nil {
			if required {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			ctx.Set("user", (*model.User)(nil))
			ctx.Next()
			return
		}

		// 成功解析 token
		ctx.Set("user", user)
		ctx.Next()
	}
}

func RoleRequired(minRole model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*model.User)

		if user.Role < minRole {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		}
		c.Next()
	}
}

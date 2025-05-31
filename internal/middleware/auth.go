package middleware

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		user, err := authService.ParseToken(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

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

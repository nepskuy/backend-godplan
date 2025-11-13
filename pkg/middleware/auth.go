package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// JWT util instance - PASTIKAN ADA DI SINI
var jwtUtil = utils.NewJWTUtil("your-secret-key-change-in-production")

// GinAuthMiddleware untuk framework Gin
func GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for public routes
		if c.Request.URL.Path == "/api/v1/auth/login" ||
			c.Request.URL.Path == "/api/v1/auth/register" ||
			c.Request.URL.Path == "/health" ||
			c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token and get claims
		claims, err := jwtUtil.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set userID in Gin context
		c.Set("userID", int(claims.UserID))

		// Token valid, continue to next handler
		c.Next()
	}
}

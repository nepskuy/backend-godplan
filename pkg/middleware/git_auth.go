package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

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
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token and get claims - pakai jwtUtil dari auth.go
		claims, err := jwtUtil.ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set userID in Gin context
		c.Set("userID", claims.UserID)

		// Token valid, continue to next handler
		c.Next()
	}
}

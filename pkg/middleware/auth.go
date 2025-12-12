package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// getEnv helper to get env with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// JWT util instance - uses JWT_SECRET from environment for cross-backend compatibility
var jwtUtil = utils.NewJWTUtil(getEnv("JWT_SECRET", "dev-secret-key-change-in-production"))

// GinAuthMiddleware untuk framework Gin
func GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		c.Set("userID", claims.UserID)
		c.Set("tenant_id", claims.TenantID.String())

		// Token valid, continue to next handler
		c.Next()
	}
}


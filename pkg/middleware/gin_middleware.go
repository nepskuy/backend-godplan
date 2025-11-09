package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// GinCORS middleware untuk CORS
func GinCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// GinLogging middleware untuk logging
func GinLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[GIN] %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// GinDatabaseCheck middleware untuk cek koneksi database
func GinDatabaseCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := database.HealthCheck(); err != nil {
			log.Printf("⚠️ Database connection issue: %v", err)
			utils.GinErrorResponse(c, http.StatusServiceUnavailable, "Database temporarily unavailable")
			c.Abort()
			return
		}
		c.Next()
	}
}

// GinRecovery middleware untuk recovery dari panic
func GinRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("❌ Panic recovered: %s", err)
			utils.GinErrorResponse(c, http.StatusInternalServerError, "Internal Server Error")
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// GinAuthMiddleware middleware untuk authentication
func GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			utils.GinErrorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Validasi token JWT di sini
		// userID, err := utils.ValidateJWT(token)
		// if err != nil {
		//     utils.GinErrorResponse(c, http.StatusUnauthorized, "Invalid token")
		//     c.Abort()
		//     return
		// }

		// c.Set("userID", userID)
		c.Next()
	}
}

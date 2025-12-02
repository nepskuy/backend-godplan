package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
)

func GinDatabaseCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip database check for specific routes if needed
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}

		// Check database connection sebelum memproses request
		if err := database.HealthCheck(); err != nil {
			log.Printf("❌ Database connection lost: %v", err)

			// Try to reconnect
			cfg := config.Load()
			if reconnectErr := database.InitDB(cfg); reconnectErr != nil {
				log.Printf("❌ Failed to reconnect to database: %v", reconnectErr)
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Database connection lost",
				})
				c.Abort()
				return
			}

			log.Printf("✅ Database reconnected successfully")
		}

		c.Next()
	}
}

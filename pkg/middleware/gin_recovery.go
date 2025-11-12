package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinRecovery middleware untuk Gin
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("🚨 PANIC recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

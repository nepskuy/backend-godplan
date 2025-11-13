package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogging middleware untuk Gin
func GinLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log after request completes
		duration := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		log.Printf("📍 %s %s - %d - %v", method, path, status, duration)
	}
}

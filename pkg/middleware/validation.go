package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// InputValidationMiddleware validates incoming requests
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Limit request body size to 10MB
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10*1024*1024)

		// Validate content type for POST/PUT/PATCH
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			
			// Allow JSON and multipart form data
			if !strings.Contains(contentType, "application/json") &&
				!strings.Contains(contentType, "multipart/form-data") &&
				!strings.Contains(contentType, "application/x-www-form-urlencoded") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content type"})
				c.Abort()
				return
			}
		}

		// Validate User-Agent header exists (prevent bot attacks)
		userAgent := c.GetHeader("User-Agent")
		if userAgent == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User-Agent header required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

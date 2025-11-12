package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
)

// GinCORS middleware untuk Gin
func GinCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Daftar allowed origins
		allowedOrigins := []string{
			"https://godplan.godjahstudio.com",
			"https://be-godplan.godjahstudio.com",
			"https://fe-godplan.vercel.app",
		}

		// Jika running di development, tambahkan localhost
		if os.Getenv("ENVIRONMENT") != "production" {
			allowedOrigins = append(allowedOrigins,
				"http://localhost:3000",
				"http://127.0.0.1:3000",
				"https://localhost:3000",
				"http://localhost:8080",
			)
		}

		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// Jika origin tidak match, dan development, pakai localhost default
		if !allowed && os.Getenv("ENVIRONMENT") != "production" && len(allowedOrigins) > 0 {
			c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")

		// Tangani preflight OPTIONS
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func GinCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		env := os.Getenv("ENVIRONMENT")

		allowedOrigins := []string{
			"https://godplan.godjahstudio.com",
			"https://be-godplan.godjahstudio.com",
			"https://fe-godplan.vercel.app",
			"http://localhost:3000",
			"http://localhost:3001",
		}

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				break
			}
		}

		// Allow all Vercel deployments (preview & production)
		if !allowed && strings.HasSuffix(origin, ".vercel.app") {
			allowed = true
		}

		// Fallback for development/localhost if not in the allowed list
		if !allowed && (env != "production") {
			if origin != "" && (strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1") ||
				strings.HasPrefix(origin, "https://localhost")) {
				allowed = true
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
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

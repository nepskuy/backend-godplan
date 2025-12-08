package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds common security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protect against clickjacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		
		// Protect against MIME sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		
		// Enable XSS filtering in browser
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Content Security Policy (Basic)
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; img-src * data:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self' 'unsafe-inline'")
		
		// Referrer Policy
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		c.Next()
	}
}

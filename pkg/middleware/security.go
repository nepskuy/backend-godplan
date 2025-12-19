package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds comprehensive security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protect against clickjacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		
		// Protect against MIME sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		
		// Enable XSS filtering in browser
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Strict Transport Security (HTTPS only) - 1 year
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		
		// Content Security Policy (Enhanced)
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com data:; " +
			"img-src 'self' data: https: blob:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Writer.Header().Set("Content-Security-Policy", csp)
		
		// Referrer Policy
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy (restrict dangerous features)
		c.Writer.Header().Set("Permissions-Policy", "geolocation=(self), microphone=(), camera=(), payment=()")
		
		// Remove server information
		c.Writer.Header().Set("Server", "")
		
		c.Next()
	}
}

package middleware

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/nepskuy/be-godplan/pkg/database"
)

func GinDatabaseCheck() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip database check for specific routes OR OPTIONS method
        if c.Request.Method == "OPTIONS" || c.Request.URL.Path == "/health" || c.Request.URL.Path == "/api/v1/health" {
            c.Next()
            return
        }

        // Check database connection sebelum memproses request
        if err := database.HealthCheck(); err != nil {
            log.Printf("⚠️ Database connection check failed: %v", err)
            // Don't try to reconnect here - it creates new connections!
            // Let the connection pool handle recovery naturally
            c.JSON(http.StatusServiceUnavailable, gin.H{
                "error":   "Database temporarily unavailable",
                "message": "Please try again in a moment",
                "details": "Connection pool may be exhausted or database unreachable",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

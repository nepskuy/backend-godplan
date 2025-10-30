package middleware

import (
	"log"
	"net/http"

	"github.com/nepskuy/be-godplan/pkg/database"
)

func DatabaseCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip database check for specific routes if needed
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Check database connection sebelum memproses request
		if err := database.HealthCheck(); err != nil {
			log.Printf("❌ Database connection lost: %v", err)

			// Try to reconnect
			if reconnectErr := database.InitDB(); reconnectErr != nil {
				log.Printf("❌ Failed to reconnect to database: %v", reconnectErr)
				http.Error(w, "Database connection lost", http.StatusServiceUnavailable)
				return
			}

			log.Printf("✅ Database reconnected successfully")
		}

		next.ServeHTTP(w, r)
	})
}

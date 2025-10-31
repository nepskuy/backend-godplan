package middleware

import (
	"log"
	"net/http"
	"os"
)

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

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
			)
		}

		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// Jika origin tidak match, dan development, pakai localhost default
		if !allowed && os.Getenv("ENVIRONMENT") != "production" && len(allowedOrigins) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")

		// Tangani preflight OPTIONS
		if r.Method == "OPTIONS" {
			log.Printf("✅ CORS Preflight handled for: %s", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

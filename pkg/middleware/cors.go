package middleware

import (
	"log"
	"net/http"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log every request to verify middleware is working
		log.Printf("🌐 CORS Middleware TRIGGERED - Method: %s, Path: %s, Origin: %s",
			r.Method, r.URL.Path, r.Header.Get("Origin"))

		// Allow multiple origins
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"https://localhost:3000",
			"https://fe-godplan.vercel.app",
			"https://godplan.godjahstudio.com",
			"http://godplan.godjahstudio.com",
		}

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// If no specific origin matched, use the first one as default for development
		if !allowed && len(allowedOrigins) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			log.Printf("✅ CORS Preflight handled for: %s", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Printf("➡️ CORS passing to next handler: %s", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

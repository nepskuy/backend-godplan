package api

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/handlers"
	"github.com/nepskuy/be-godplan/pkg/middleware"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

var router *mux.Router

func init() {
	log.Printf("🚀 Initializing GodPlan API for Vercel...")

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Printf("❌ Database connection failed: %v", err)
	} else {
		log.Printf("✅ Database connected successfully")
	}

	router = mux.NewRouter()
	setupRoutes()
}

// CORS middleware langsung di file ini
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log CORS request
		log.Printf("🌐 CORS Middleware - Origin: %s, Method: %s", r.Header.Get("Origin"), r.Method)

		// Allow multiple origins
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:3000",
			"https://localhost:3000",
			"https://fe-godplan.vercel.app",
			"https://godplan.godjahstudio.com",
			"https://be-godplan.godjahstudio.com",
		}

		// Check if origin is allowed
		for _, o := range allowedOrigins {
			if origin == o {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// Fallback untuk development
		if w.Header().Get("Access-Control-Allow-Origin") == "" {
			w.Header().Set("Access-Control-Allow-Origin", "https://godplan.godjahstudio.com")
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

		next.ServeHTTP(w, r)
	})
}

func setupRoutes() {
	router.StrictSlash(true)

	log.Printf("🟡 Registering middleware...")

	// Gunakan CORS middleware langsung di sini
	router.Use(corsMiddleware)
	router.Use(middleware.Logging)
	router.Use(middleware.DatabaseCheck)
	router.Use(middleware.Recovery)

	// Global OPTIONS handler
	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("⚙️ Global OPTIONS handler triggered for %s", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("🟢 Middleware registered: CORS, Logging, DatabaseCheck, Recovery")

	// Health check endpoint (public)
	router.HandleFunc("/health", healthCheck).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/health", healthCheck).Methods("GET", "OPTIONS")

	// Swagger
	router.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	}).Methods("GET")

	swaggerHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>GodPlan API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/swagger.json',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>
		`))
	}

	router.HandleFunc("/swagger", swaggerHandler).Methods("GET")
	router.HandleFunc("/swagger/", swaggerHandler).Methods("GET")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusFound)
	}).Methods("GET")

	// Public API routes
	publicAPI := router.PathPrefix("/api/v1").Subrouter()
	publicAPI.HandleFunc("/auth/register", handlers.Register).Methods("POST", "OPTIONS")
	publicAPI.HandleFunc("/auth/login", handlers.Login).Methods("POST", "OPTIONS")

	// Protected API routes
	protectedAPI := router.PathPrefix("/api/v1").Subrouter()
	protectedAPI.Use(middleware.AuthMiddleware)

	protectedAPI.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	protectedAPI.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	protectedAPI.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")

	// Task handlers (gunakan placeholder untuk yang belum ada)
	protectedAPI.HandleFunc("/tasks", notImplementedHandler).Methods("GET")
	protectedAPI.HandleFunc("/tasks", notImplementedHandler).Methods("POST")
	protectedAPI.HandleFunc("/tasks/{id}", notImplementedHandler).Methods("GET")
	protectedAPI.HandleFunc("/tasks/{id}", notImplementedHandler).Methods("PUT")
	protectedAPI.HandleFunc("/tasks/{id}", notImplementedHandler).Methods("DELETE")

	// Gunakan HTTP handlers untuk attendance
	protectedAPI.HandleFunc("/attendance/clock-in", handlers.ClockInHTTP).Methods("POST")
	protectedAPI.HandleFunc("/attendance/clock-out", handlers.ClockOutHTTP).Methods("POST")
	protectedAPI.HandleFunc("/attendance/check-location", handlers.CheckLocationHTTP).Methods("POST")
	protectedAPI.HandleFunc("/attendance", handlers.GetAttendanceHTTP).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.ErrorResponse(w, http.StatusNotFound, "Route not found: "+r.URL.Path)
	})

	log.Printf("✅ GodPlan API initialized successfully for Vercel")
}

func notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	utils.ErrorResponse(w, http.StatusNotImplemented, "Handler not implemented")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if err := database.HealthCheck(); err != nil {
		dbStatus = "disconnected"
		log.Printf("❌ Database health check failed: %v", err)
	}

	platform := "vercel"
	if os.Getenv("VERCEL") == "" {
		platform = "local"
	}

	utils.SuccessResponse(w, http.StatusOK, "Server is healthy", map[string]interface{}{
		"status":    "ok",
		"service":   "godplan-backend",
		"database":  dbStatus,
		"timestamp": time.Now().Format(time.RFC3339),
		"platform":  platform,
	})
}

// Handler function untuk Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 Incoming request: %s %s", r.Method, r.URL.Path)
	router.ServeHTTP(w, r)
}

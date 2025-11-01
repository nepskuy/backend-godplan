package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/handlers"
	"github.com/nepskuy/be-godplan/pkg/middleware"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// @title GodPlan API
// @version 1.0
// @description Backend API for GodPlan application
// @host localhost:8080
// @BasePath /api/v1
func main() {
	log.Println("🚀 Starting GodPlan Backend Server...")

	// Load .env file hanya di development
	if config.IsDevelopment() {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ No .env file found, using environment variables")
		} else {
			log.Println("✅ .env file loaded")
		}
	}

	cfg := config.Load()

	log.Println("🔌 Connecting to database...")

	// Debug info untuk DATABASE_URL
	if cfg.DatabaseURL != "" {
		log.Println("✅ DATABASE_URL is available")
		maskedURL := maskPassword(cfg.DatabaseURL)
		log.Printf("📝 Using DATABASE_URL: %s", maskedURL)
	} else {
		log.Println("⚠️ DATABASE_URL not found, using individual DB config")
		log.Printf("📝 DB Host: %s, Port: %s, User: %s, Name: %s",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName)
	}

	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	if err := database.HealthCheck(); err != nil {
		log.Printf("⚠️ Database health check warning: %v", err)
	} else {
		log.Println("✅ Database health check passed")
	}

	router := setupRouter()

	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)

	// Log environment info
	env := "development"
	if config.IsProduction() {
		env = "production"
	}

	log.Printf("🌐 Server starting in %s mode", env)
	log.Printf("📍 Listening on http://localhost%s", addr)
	log.Printf("📚 Swagger UI available at http://localhost%s/swagger", addr)
	log.Printf("🏥 Health check at http://localhost%s/health", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("✨ Server is ready to accept connections!")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.StrictSlash(true)

	log.Println("🔧 Setting up middleware...")

	router.Use(middleware.CORS)
	router.Use(middleware.Logging)
	router.Use(middleware.DatabaseCheck)
	router.Use(middleware.Recovery)

	log.Println("✅ Middleware registered")

	// Global OPTIONS handler
	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("🔧 Setting up routes...")

	router.HandleFunc("/health", healthCheck).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/health", healthCheck).Methods("GET", "OPTIONS")

	// Serve Swagger JSON
	router.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	}).Methods("GET")

	// Serve Swagger YAML (alternatif)
	router.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.yaml")
	}).Methods("GET")

	// Swagger UI handler
	router.HandleFunc("/swagger", swaggerHandler).Methods("GET")
	router.HandleFunc("/swagger/", swaggerHandler).Methods("GET")

	// Root redirect to Swagger
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusFound)
	}).Methods("GET")

	// Static files untuk Swagger UI assets (jika diperlukan)
	router.PathPrefix("/swagger/static/").Handler(http.StripPrefix("/swagger/static/",
		http.FileServer(http.Dir("./docs/"))))

	publicAPI := router.PathPrefix("/api/v1").Subrouter()
	publicAPI.HandleFunc("/auth/register", handlers.Register).Methods("POST", "OPTIONS")
	publicAPI.HandleFunc("/auth/login", handlers.Login).Methods("POST", "OPTIONS")

	protectedAPI := router.PathPrefix("/api/v1").Subrouter()
	protectedAPI.Use(middleware.AuthMiddleware)

	protectedAPI.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	protectedAPI.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	protectedAPI.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")

	protectedAPI.HandleFunc("/tasks", handlers.GetTasks).Methods("GET")
	protectedAPI.HandleFunc("/tasks", handlers.CreateTask).Methods("POST")
	protectedAPI.HandleFunc("/tasks/{id}", handlers.GetTask).Methods("GET")
	protectedAPI.HandleFunc("/tasks/{id}", handlers.UpdateTask).Methods("PUT")
	protectedAPI.HandleFunc("/tasks/{id}", handlers.DeleteTask).Methods("DELETE")

	protectedAPI.HandleFunc("/attendance/clock-in", handlers.ClockIn).Methods("POST")
	protectedAPI.HandleFunc("/attendance/clock-out", handlers.ClockOut).Methods("POST")
	protectedAPI.HandleFunc("/attendance/check-location", handlers.CheckLocation).Methods("POST")
	protectedAPI.HandleFunc("/attendance", handlers.GetAttendance).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.ErrorResponse(w, http.StatusNotFound, "Route not found: "+r.URL.Path)
	})

	log.Println("✅ Routes registered successfully")

	return router
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if err := database.HealthCheck(); err != nil {
		dbStatus = "disconnected"
		log.Printf("❌ Database health check failed: %v", err)
	}

	platform := "local"
	if os.Getenv("VERCEL") != "" {
		platform = "vercel"
	}

	cfg := config.Load()

	utils.SuccessResponse(w, http.StatusOK, "Server is healthy", map[string]interface{}{
		"status":       "ok",
		"service":      "godplan-backend",
		"database":     dbStatus,
		"environment":  platform,
		"timestamp":    time.Now().Format(time.RFC3339),
		"version":      "1.0.0",
		"using_db_url": cfg.DatabaseURL != "",
	})
}

func swaggerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>GodPlan API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.9.0/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.9.0/favicon-16x16.png" sizes="16x16" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *,
        *:before,
        *:after {
            box-sizing: inherit;
        }
        body {
            margin: 0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #2c3e50;
            padding: 10px 0;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                defaultModelsExpandDepth: -1,
                operationsSorter: "alpha",
                tagsSorter: "alpha",
                docExpansion: "none"
            });
            
            // Handle jika swagger.json tidak ditemukan
            fetch('/swagger.json')
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Swagger JSON not found');
                    }
                    return response.json();
                })
                .catch(error => {
                    document.getElementById('swagger-ui').innerHTML = 
                        '<div style="padding: 20px; text-align: center; color: red;">' +
                        '<h2>Error Loading Swagger Documentation</h2>' +
                        '<p>' + error.message + '</p>' +
                        '<p>Please check if swagger.json exists and is properly generated.</p>' +
                        '</div>';
                });
        }
    </script>
</body>
</html>
	`
	w.Write([]byte(html))
}

func maskPassword(connStr string) string {
	// Mask password dalam connection string
	for _, prefix := range []string{"password=", "Password="} {
		if idx := findIndex(connStr, prefix); idx != -1 {
			end := findNextSeparator(connStr, idx+len(prefix))
			return connStr[:idx+len(prefix)] + "****" + connStr[end:]
		}
	}

	// Mask password dalam URL format (postgres://user:pass@host)
	if idx := findIndex(connStr, "://"); idx != -1 {
		if idx2 := findIndex(connStr[idx+3:], "@"); idx2 != -1 {
			start := idx + 3
			end := start + idx2
			if colonIdx := findIndex(connStr[start:end], ":"); colonIdx != -1 {
				return connStr[:start+colonIdx+1] + "****" + connStr[end:]
			}
		}
	}
	return connStr
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func findNextSeparator(s string, start int) int {
	for i := start; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '&' || s[i] == '?' || s[i] == '#' {
			return i
		}
	}
	return len(s)
}

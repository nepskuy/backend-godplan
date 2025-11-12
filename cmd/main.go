package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/handlers"
	"github.com/nepskuy/be-godplan/pkg/middleware"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

// @title GodPlan API
// @version 1.0
// @description Backend API for GodPlan application
// @host localhost:8080
// @BasePath /api/v1
func main() {
	log.Println("🚀 Starting GodPlan Backend Server with GIN...")

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

	// Setup repository
	db := database.GetDB()
	userRepo := repository.NewUserRepository(db)

	// Setup Gin router
	router := setupGinRouter(userRepo)

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

func setupGinRouter(userRepo *repository.UserRepository) *gin.Engine {
	// Set Gin mode
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	log.Println("🔧 Setting up Gin middleware...")

	// Apply middleware
	router.Use(middleware.GinCORS())
	router.Use(middleware.GinLogging())
	router.Use(middleware.GinRecovery())

	log.Println("✅ Gin middleware registered")

	log.Println("🔧 Setting up routes...")

	// Health check routes
	router.GET("/health", ginHealthCheck)
	router.GET("/api/v1/health", ginHealthCheck)

	// Swagger routes
	router.GET("/swagger", ginSwaggerHandler)
	router.GET("/swagger.json", func(c *gin.Context) {
		c.File("./docs/swagger.json")
	})
	router.GET("/swagger.yaml", func(c *gin.Context) {
		c.File("./docs/swagger.yaml")
	})

	// Root redirect to Swagger
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger")
	})

	// API Routes
	api := router.Group("/api/v1")
	{
		// Public routes - No authentication required
		public := api.Group("/auth")
		{
			public.POST("/register", ginAuthWrapper(handlers.Register))
			public.POST("/login", ginAuthWrapper(handlers.Login))
		}

		// Protected routes - Authentication required
		protected := api.Group("")
		protected.Use(middleware.GinAuthMiddleware()) // ← PAKAI GIN AUTH MIDDLEWARE
		{
			// User routes
			protected.GET("/users", ginAuthWrapper(handlers.GetUsers))
			protected.POST("/users", ginAuthWrapper(handlers.CreateUser))
			protected.GET("/users/:id", ginAuthWrapper(handlers.GetUser))

			// Profile routes - NEW
			protected.GET("/profile", handlers.GinGetProfile(userRepo))

			// Task routes
			protected.GET("/tasks", ginAuthWrapper(handlers.GetTasks))
			protected.POST("/tasks", ginAuthWrapper(handlers.CreateTask))
			protected.GET("/tasks/:id", ginAuthWrapper(handlers.GetTask))
			protected.PUT("/tasks/:id", ginAuthWrapper(handlers.UpdateTask))
			protected.DELETE("/tasks/:id", ginAuthWrapper(handlers.DeleteTask))

			// Attendance routes
			protected.POST("/attendance/clock-in", ginAuthWrapper(handlers.ClockInHTTP))
			protected.POST("/attendance/clock-out", ginAuthWrapper(handlers.ClockOutHTTP))
			protected.POST("/attendance/check-location", ginAuthWrapper(handlers.CheckLocationHTTP))
			protected.GET("/attendance", ginAuthWrapper(handlers.GetAttendanceHTTP))
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found: " + c.Request.URL.Path,
		})
	})

	log.Println("✅ Gin routes registered successfully")
	return router
}

// ginAuthWrapper converts existing HTTP handlers to Gin handlers
func ginAuthWrapper(handler http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Convert Gin context to HTTP request
		handler(c.Writer, c.Request)
	}
}

func ginHealthCheck(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"service":      "godplan-backend",
		"database":     dbStatus,
		"environment":  platform,
		"timestamp":    time.Now().Format(time.RFC3339),
		"version":      "1.0.0",
		"using_db_url": cfg.DatabaseURL != "",
	})
}

func ginSwaggerHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `<!DOCTYPE html>
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
</html>`)
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

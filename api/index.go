package api

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/handlers"
	"github.com/nepskuy/be-godplan/pkg/middleware"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

var router *gin.Engine
var userRepo *repository.UserRepository

func init() {
	log.Printf("🚀 Initializing GodPlan API for Vercel with GIN...")

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Printf("❌ Database connection failed: %v", err)
	} else {
		log.Printf("✅ Database connected successfully")
	}

	// Setup repository
	db := database.GetDB()
	userRepo = repository.NewUserRepository(db)

	// Setup Gin
	setupGin()
}

func setupGin() {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router = gin.New()

	log.Printf("🟡 Registering Gin middleware...")

	// Apply middleware - HANYA middleware yang tersedia
	router.Use(gin.Recovery())
	router.Use(middleware.GinCORS())
	router.Use(middleware.GinLogging())
	// DatabaseCheck dihapus karena tidak ada di package middleware

	log.Printf("🟢 Gin middleware registered")

	// Health check endpoints
	router.GET("/health", ginHealthCheck)
	router.GET("/api/v1/health", ginHealthCheck)

	// Swagger routes
	router.GET("/swagger", ginSwaggerHandler)
	router.GET("/swagger.json", func(c *gin.Context) {
		// Fallback jika file tidak ada
		c.JSON(200, gin.H{
			"info": map[string]interface{}{
				"title":   "GodPlan API",
				"version": "1.0",
			},
			"openapi": "3.0.0",
		})
	})
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger")
	})

	// API Routes
	api := router.Group("/api/v1")
	{
		// Public routes - No authentication required
		public := api.Group("/auth")
		{
			public.POST("/register", ginHandlerWrapper(handlers.Register))
			public.POST("/login", ginHandlerWrapper(handlers.Login))
		}

		// Protected routes - Authentication required
		protected := api.Group("")
		protected.Use(middleware.GinAuthMiddleware())
		{
			// User routes
			protected.GET("/users", ginHandlerWrapper(handlers.GetUsers))
			protected.POST("/users", ginHandlerWrapper(handlers.CreateUser))
			protected.GET("/users/:id", ginHandlerWrapper(handlers.GetUser))

			// Profile routes - NEWLY ADDED
			protected.GET("/profile", handlers.GinGetProfile(userRepo))

			// Task routes
			protected.GET("/tasks", ginHandlerWrapper(handlers.GetTasks))
			protected.POST("/tasks", ginHandlerWrapper(handlers.CreateTask))
			protected.GET("/tasks/:id", ginHandlerWrapper(handlers.GetTask))
			protected.PUT("/tasks/:id", ginHandlerWrapper(handlers.UpdateTask))
			protected.DELETE("/tasks/:id", ginHandlerWrapper(handlers.DeleteTask))

			// Attendance routes
			protected.POST("/attendance/clock-in", ginHandlerWrapper(handlers.ClockInHTTP))
			protected.POST("/attendance/clock-out", ginHandlerWrapper(handlers.ClockOutHTTP))
			protected.POST("/attendance/check-location", ginHandlerWrapper(handlers.CheckLocationHTTP))
			protected.GET("/attendance", ginHandlerWrapper(handlers.GetAttendanceHTTP))
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Route not found: " + c.Request.URL.Path,
		})
	})

	log.Printf("✅ GodPlan API with GIN initialized successfully for Vercel")
}

// ginHandlerWrapper converts existing HTTP handlers to Gin handlers
func ginHandlerWrapper(handler http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c.Writer, c.Request)
	}
}

func ginHealthCheck(c *gin.Context) {
	dbStatus := "connected"
	if err := database.HealthCheck(); err != nil {
		dbStatus = "disconnected"
		log.Printf("❌ Database health check failed: %v", err)
	}

	platform := "vercel"
	if os.Getenv("VERCEL") == "" {
		platform = "local"
	}

	response := map[string]interface{}{
		"status":    "ok",
		"service":   "godplan-backend",
		"database":  dbStatus,
		"timestamp": time.Now().Format(time.RFC3339),
		"platform":  platform,
		"framework": "gin",
	}

	c.JSON(http.StatusOK, response)
}

func ginSwaggerHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `<!DOCTYPE html>
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
</html>`)
}

// Handler function untuk Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 Incoming request: %s %s", r.Method, r.URL.Path)
	router.ServeHTTP(w, r)
}

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

	// Swagger routes - FIXED
	router.GET("/swagger", ginSwaggerHandler)
	router.GET("/swagger/*any", ginSwaggerRedirectHandler)
	router.GET("/swagger.json", ginSwaggerJSONHandler)
	router.GET("/swagger.yaml", ginSwaggerYAMLHandler)
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

// ginSwaggerJSONHandler handles swagger.json request
func ginSwaggerJSONHandler(c *gin.Context) {
	// Try to read the swagger.json file first
	data, err := os.ReadFile("./docs/swagger.json")
	if err != nil {
		log.Printf("❌ Failed to read swagger.json: %v", err)
		// Fallback to embedded swagger spec
		embeddedSwagger := map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]interface{}{
				"title":       "GodPlan API",
				"version":     "1.0",
				"description": "Backend API for GodPlan application",
			},
			"servers": []map[string]interface{}{
				{
					"url":         "https://be-godplan.godjahstudio.com",
					"description": "Production server",
				},
			},
			"paths": map[string]interface{}{
				"/api/v1/health": map[string]interface{}{
					"get": map[string]interface{}{
						"summary":     "Health Check",
						"description": "Check API health status",
						"responses": map[string]interface{}{
							"200": map[string]interface{}{
								"description": "OK",
								"content": map[string]interface{}{
									"application/json": map[string]interface{}{
										"schema": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"status":   map[string]interface{}{"type": "string"},
												"service":  map[string]interface{}{"type": "string"},
												"database": map[string]interface{}{"type": "string"},
											},
										},
									},
								},
							},
						},
					},
				},
				"/api/v1/auth/register": map[string]interface{}{
					"post": map[string]interface{}{
						"summary": "Register new user",
						"requestBody": map[string]interface{}{
							"required": true,
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"email":    map[string]interface{}{"type": "string"},
											"password": map[string]interface{}{"type": "string"},
											"name":     map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
						"responses": map[string]interface{}{
							"200": map[string]interface{}{
								"description": "User registered successfully",
							},
						},
					},
				},
				"/api/v1/auth/login": map[string]interface{}{
					"post": map[string]interface{}{
						"summary": "Login user",
						"requestBody": map[string]interface{}{
							"required": true,
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"email":    map[string]interface{}{"type": "string"},
											"password": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
						"responses": map[string]interface{}{
							"200": map[string]interface{}{
								"description": "Login successful",
							},
						},
					},
				},
			},
			"components": map[string]interface{}{
				"securitySchemes": map[string]interface{}{
					"bearerAuth": map[string]interface{}{
						"type":         "http",
						"scheme":       "bearer",
						"bearerFormat": "JWT",
					},
				},
			},
			"security": []map[string]interface{}{
				{
					"bearerAuth": []string{},
				},
			},
		}
		c.JSON(200, embeddedSwagger)
		return
	}

	c.Data(200, "application/json", data)
}

// ginSwaggerYAMLHandler handles swagger.yaml request
func ginSwaggerYAMLHandler(c *gin.Context) {
	// Try to read the swagger.yaml file first
	data, err := os.ReadFile("./docs/swagger.yaml")
	if err != nil {
		log.Printf("❌ Failed to read swagger.yaml: %v", err)
		c.JSON(404, gin.H{
			"error":   true,
			"message": "swagger.yaml not found",
		})
		return
	}

	c.Data(200, "application/yaml", data)
}

// ginSwaggerRedirectHandler handles other Swagger UI routes
func ginSwaggerRedirectHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/swagger")
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
        .swagger-ui .info hgroup.main {
            text-align: center;
        }
        .loading {
            padding: 20px;
            text-align: center;
            font-family: Arial, sans-serif;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            // Show loading message
            document.getElementById('swagger-ui').innerHTML = '<div class="loading"><h3>Loading GodPlan API Documentation...</h3></div>';
            
            // Initialize Swagger UI
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
                defaultModelsExpandDepth: 1,
                operationsSorter: "alpha",
                tagsSorter: "alpha",
                docExpansion: "none",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true
            });
            
            // Error handling for Swagger JSON
            fetch('/swagger.json')
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Failed to load Swagger JSON: ' + response.status);
                    }
                    return response.json();
                })
                .then(data => {
                    console.log('Swagger JSON loaded successfully', data);
                })
                .catch(error => {
                    console.error('Error loading Swagger JSON:', error);
                    document.getElementById('swagger-ui').innerHTML = 
                        '<div style="padding: 20px; text-align: center; font-family: Arial, sans-serif;">' +
                        '<h2>GodPlan API Documentation</h2>' +
                        '<p>Basic API documentation is loaded. For full Swagger documentation, generate swagger.json file.</p>' +
                        '<p style="color: red;"><strong>Error:</strong> ' + error.message + '</p>' +
                        '<p><a href="/health">Check API Health</a></p>' +
                        '</div>';
                });
        }
    </script>
</body>
</html>`)
}

// Handler function untuk Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 Incoming request: %s %s", r.Method, r.URL.Path)
	router.ServeHTTP(w, r)
}

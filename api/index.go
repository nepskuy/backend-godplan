package api

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/handlers"
	"github.com/nepskuy/be-godplan/pkg/middleware"
	"github.com/nepskuy/be-godplan/pkg/repository"

	// Swagger
	docs "github.com/nepskuy/be-godplan/docs"
	"github.com/swaggo/swag"
)

var router *gin.Engine
var userRepo *repository.UserRepository

func init() {
	log.Printf("🚀 Initializing GodPlan API for Vercel with GIN...")

	// Load config
	cfg := config.Load()

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

	// Initialize database
	if err := database.InitDB(cfg); err != nil {
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

	// Apply middleware
	router.Use(gin.Recovery())
	router.Use(middleware.GinCORS())
	router.Use(middleware.GinLogging())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.GinDatabaseCheck())

	log.Printf("🟢 Gin middleware registered")

	// Health check endpoints
	router.GET("/health", ginHealthCheck)
	router.GET("/api/v1/health", ginHealthCheck)

	// Swagger routes
	router.GET("/swagger", ginSwaggerHandler)
	router.GET("/swagger/*any", ginSwaggerRedirectHandler)
	router.GET("/swagger.json", ginSwaggerJSONHandler)
	router.GET("/swagger.yaml", ginSwaggerYAMLHandler)
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger")
	})

	// API Routes - UPDATE: LENGKAPI DENGAN ENDPOINT HOME
	api := router.Group("/api/v1")
	{
		// Public routes - No authentication required
		public := api.Group("/auth")
		{
			public.POST("/register", handlers.Register)
			public.POST("/login", handlers.Login)
			public.POST("/refresh", handlers.RefreshToken)
		}


		// Protected routes - Authentication required
		protected := api.Group("")
		protected.Use(middleware.GinAuthMiddleware())
		{
			// Dashboard routes - DITAMBAHKAN ENDPOINT HOME
			protected.GET("/home", handlers.GetHomeDashboard)
			protected.GET("/dashboard/stats", handlers.GetDashboardStats)
			protected.GET("/teams", handlers.GetTeamMembers)

			// User routes
			protected.GET("/users", handlers.GetUsers)
			protected.POST("/users", handlers.CreateUser)
			protected.GET("/users/:id", handlers.GetUser)

			// Profile routes
			protected.GET("/profile", handlers.GinGetProfile(userRepo))

			// Task routes - LENGKAP
			protected.GET("/tasks", handlers.GetTasks)
			protected.POST("/tasks", handlers.CreateTask)
			protected.GET("/tasks/:id", handlers.GetTask)
			protected.PUT("/tasks/:id", handlers.UpdateTask)
			protected.DELETE("/tasks/:id", handlers.DeleteTask)
			protected.GET("/tasks/upcoming", handlers.GetUpcomingTasks)
			protected.PATCH("/tasks/:id/progress", handlers.UpdateTaskProgress)
			protected.PATCH("/tasks/:id/complete", handlers.CompleteTask)
			protected.GET("/tasks/statistics", handlers.GetTaskStatistics)

			// Attendance routes
			protected.POST("/attendance/clock-in", handlers.ClockIn)
			protected.POST("/attendance/clock-out", handlers.ClockOut)
			protected.POST("/attendance/check-location", handlers.CheckLocation)
			protected.GET("/attendance", handlers.GetAttendance)
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
	log.Printf("📍 Available endpoints:")
	log.Printf("   - GET  /health")
	log.Printf("   - GET  /swagger")
	log.Printf("   - POST /api/v1/auth/register")
	log.Printf("   - POST /api/v1/auth/login")
	log.Printf("   - GET  /api/v1/home") // DITAMBAHKAN
	log.Printf("   - GET  /api/v1/dashboard/stats")
	log.Printf("   - GET  /api/v1/teams")
	log.Printf("   - GET  /api/v1/profile")
	log.Printf("   - GET  /api/v1/tasks")
	log.Printf("   - POST /api/v1/tasks")
	log.Printf("   - GET  /api/v1/tasks/:id")
	log.Printf("   - PUT  /api/v1/tasks/:id")
	log.Printf("   - DELETE /api/v1/tasks/:id")
	log.Printf("   - GET  /api/v1/tasks/upcoming")
	log.Printf("   - PATCH /api/v1/tasks/:id/progress")
	log.Printf("   - PATCH /api/v1/tasks/:id/complete")
	log.Printf("   - GET  /api/v1/tasks/statistics")
	log.Printf("   - POST /api/v1/attendance/clock-in")
	log.Printf("   - POST /api/v1/attendance/clock-out")
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

	cfg := config.Load()

	response := map[string]interface{}{
		"status":       "ok",
		"service":      "godplan-backend",
		"database":     dbStatus,
		"timestamp":    time.Now().Format(time.RFC3339),
		"platform":     platform,
		"framework":    "gin",
		"version":      "1.0.0",
		"using_db_url": cfg.DatabaseURL != "",
	}

	c.JSON(http.StatusOK, response)
}

// ginSwaggerJSONHandler handles swagger.json request
func ginSwaggerJSONHandler(c *gin.Context) {
	// Coba baca file swagger.json dari root directory
	data, err := os.ReadFile("./docs/swagger.json")
	if err != nil {
		// Fallback: coba baca dari path relative
		data, err = os.ReadFile("docs/swagger.json")
		if err != nil {
			// Fallback: try generated docs
			doc, err := swag.ReadDoc(docs.SwaggerInfo.InstanceName())
			if err == nil {
				c.Data(200, "application/json", []byte(doc))
				return
			}

			log.Printf("❌ Failed to read swagger.json: %v", err)
			// Fallback ke embedded swagger spec yang LENGKAP
			embeddedSwagger := createEmbeddedSwaggerSpec()
			c.JSON(200, embeddedSwagger)
			return
		}
	}

	c.Data(200, "application/json", data)
}

// ginSwaggerYAMLHandler handles swagger.yaml request
func ginSwaggerYAMLHandler(c *gin.Context) {
	// Coba baca file swagger.yaml dari root directory
	data, err := os.ReadFile("./docs/swagger.yaml")
	if err != nil {
		// Fallback: coba baca dari path relative
		data, err = os.ReadFile("docs/swagger.yaml")
		if err != nil {
			log.Printf("❌ Failed to read swagger.yaml: %v", err)
			c.JSON(404, gin.H{
				"error":   true,
				"message": "swagger.yaml not found",
			})
			return
		}
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

// createEmbeddedSwaggerSpec creates a complete swagger spec when file is missing
func createEmbeddedSwaggerSpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "GodPlan API",
			"version":     "1.0",
			"description": "Backend API for GodPlan application - HR Management System",
		},
		"servers": []map[string]interface{}{
			{
				"url":         "https://be-godplan.godjahstudio.com",
				"description": "Production server",
			},
			{
				"url":         "/",
				"description": "Current server",
			},
		},
		"paths": map[string]interface{}{
			"/api/v1/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health Check",
					"description": "Check API health status",
					"tags":        []string{"health"},
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
					"summary":     "Register new user",
					"description": "Register a new user account with complete profile data",
					"tags":        []string{"authentication"},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"username", "name", "email", "password"},
									"properties": map[string]interface{}{
										"username": map[string]interface{}{"type": "string", "example": "johndoe"},
										"name":     map[string]interface{}{"type": "string", "example": "John Doe"},
										"email":    map[string]interface{}{"type": "string", "example": "john@example.com"},
										"password": map[string]interface{}{"type": "string", "example": "password123"},
										"phone":    map[string]interface{}{"type": "string", "example": "+628123456789"},
										"role":     map[string]interface{}{"type": "string", "example": "employee"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "User registered successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{"type": "boolean"},
											"message": map[string]interface{}{"type": "string"},
											"data": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"token": map[string]interface{}{"type": "string"},
													"user": map[string]interface{}{
														"type": "object",
														"properties": map[string]interface{}{
															"id":        map[string]interface{}{"type": "integer"},
															"username":  map[string]interface{}{"type": "string"},
															"email":     map[string]interface{}{"type": "string"},
															"role":      map[string]interface{}{"type": "string"},
															"name":      map[string]interface{}{"type": "string"},
															"phone":     map[string]interface{}{"type": "string"},
															"is_active": map[string]interface{}{"type": "boolean"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Bad Request",
						},
						"500": map[string]interface{}{
							"description": "Internal Server Error",
						},
					},
				},
			},
			"/api/v1/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Login user",
					"description": "Authenticate user and return JWT token",
					"tags":        []string{"authentication"},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"email", "password"},
									"properties": map[string]interface{}{
										"email":    map[string]interface{}{"type": "string", "example": "admin@godplan.com"},
										"password": map[string]interface{}{"type": "string", "example": "password"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Login successful",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{"type": "boolean"},
											"message": map[string]interface{}{"type": "string"},
											"data": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"token": map[string]interface{}{"type": "string"},
													"user": map[string]interface{}{
														"type": "object",
														"properties": map[string]interface{}{
															"id":        map[string]interface{}{"type": "integer"},
															"username":  map[string]interface{}{"type": "string"},
															"email":     map[string]interface{}{"type": "string"},
															"role":      map[string]interface{}{"type": "string"},
															"name":      map[string]interface{}{"type": "string"},
															"phone":     map[string]interface{}{"type": "string"},
															"is_active": map[string]interface{}{"type": "boolean"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Bad Request",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			// DITAMBAHKAN endpoint home
			"/api/v1/home": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get home dashboard data",
					"description": "Get complete data for home dashboard including stats, team members, and user profile",
					"tags":        []string{"dashboard"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Home dashboard data retrieved successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{"type": "boolean"},
											"message": map[string]interface{}{"type": "string"},
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/HomeDashboardResponse",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/dashboard/stats": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get dashboard statistics",
					"description": "Get overview statistics for home dashboard",
					"tags":        []string{"dashboard"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Dashboard stats retrieved successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{"type": "boolean"},
											"message": map[string]interface{}{"type": "string"},
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/DashboardStats",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/teams": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get team members",
					"description": "Get list of team members",
					"tags":        []string{"dashboard"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Team members retrieved successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{"type": "boolean"},
											"message": map[string]interface{}{"type": "string"},
											"data": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"$ref": "#/components/schemas/TeamMember",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/profile": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get user profile",
					"description": "Get current user profile information",
					"tags":        []string{"profile"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Profile retrieved successfully",
						},
					},
				},
			},
			"/api/v1/users": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get all users",
					"description": "Get list of all users (admin only)",
					"tags":        []string{"users"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Users retrieved successfully",
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create new user",
					"description": "Create a new user (admin only)",
					"tags":        []string{"users"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"username", "name", "email", "password"},
									"properties": map[string]interface{}{
										"username": map[string]interface{}{"type": "string", "example": "johndoe"},
										"name":     map[string]interface{}{"type": "string", "example": "John Doe"},
										"email":    map[string]interface{}{"type": "string", "example": "john@example.com"},
										"password": map[string]interface{}{"type": "string", "example": "password123"},
										"phone":    map[string]interface{}{"type": "string", "example": "+628123456789"},
										"role":     map[string]interface{}{"type": "string", "example": "employee"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "User created successfully",
						},
					},
				},
			},
			"/api/v1/tasks": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get all tasks",
					"description": "Get list of tasks for the current user",
					"tags":        []string{"tasks"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Tasks retrieved successfully",
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create new task",
					"description": "Create a new task for the current user",
					"tags":        []string{"tasks"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"title"},
									"properties": map[string]interface{}{
										"title":           map[string]interface{}{"type": "string"},
										"description":     map[string]interface{}{"type": "string"},
										"due_date":        map[string]interface{}{"type": "string", "format": "date"},
										"project_id":      map[string]interface{}{"type": "string"},
										"assignee_id":     map[string]interface{}{"type": "string"},
										"estimated_hours": map[string]interface{}{"type": "number"},
										"status":          map[string]interface{}{"type": "string"},
										"priority":        map[string]interface{}{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Task created successfully",
						},
					},
				},
			},
			"/api/v1/tasks/upcoming": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get upcoming tasks",
					"description": "Get upcoming tasks for dashboard (limit 3)",
					"tags":        []string{"tasks"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Upcoming tasks retrieved successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{"type": "boolean"},
											"message": map[string]interface{}{"type": "string"},
											"data": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"$ref": "#/components/schemas/UpcomingTask",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/attendance/clock-in": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Clock in",
					"description": "Record employee clock-in with location check",
					"tags":        []string{"attendance"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"latitude":  map[string]interface{}{"type": "number", "format": "float"},
										"longitude": map[string]interface{}{"type": "number", "format": "float"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Clock-in recorded successfully",
						},
					},
				},
			},
			"/api/v1/attendance/clock-out": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Clock out",
					"description": "Record employee clock-out",
					"tags":        []string{"attendance"},
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Clock-out recorded successfully",
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
			"schemas": map[string]interface{}{
				"DashboardStats": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"active_projects":   map[string]interface{}{"type": "integer"},
						"pending_tasks":     map[string]interface{}{"type": "integer"},
						"attendance_status": map[string]interface{}{"type": "string"},
						"completion_rate":   map[string]interface{}{"type": "integer"},
					},
				},
				"TeamMember": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":         map[string]interface{}{"type": "integer"},
						"name":       map[string]interface{}{"type": "string"},
						"avatar_url": map[string]interface{}{"type": "string"},
						"position":   map[string]interface{}{"type": "string"},
					},
				},
				"UpcomingTask": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":       map[string]interface{}{"type": "integer"},
						"title":    map[string]interface{}{"type": "string"},
						"due_date": map[string]interface{}{"type": "string"},
						"priority": map[string]interface{}{"type": "string"},
					},
				},
				"HomeDashboardResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"stats": map[string]interface{}{
							"$ref": "#/components/schemas/DashboardStats",
						},
						"team_members": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"$ref": "#/components/schemas/TeamMember",
							},
						},
						"greeting":    map[string]interface{}{"type": "string"},
						"user_name":   map[string]interface{}{"type": "string"},
						"user_avatar": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
		"security": []map[string]interface{}{
			{
				"bearerAuth": []string{},
			},
		},
	}
}

// Helper functions untuk mask password
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

// Handler function untuk Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 Incoming request: %s %s", r.Method, r.URL.Path)
	router.ServeHTTP(w, r)
}

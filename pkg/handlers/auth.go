package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// getEnv helper to get env with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// JWT util instance - uses JWT_SECRET from environment for consistency with middleware
var jwtUtil = utils.NewJWTUtil(getEnv("JWT_SECRET", "dev-secret-key-change-in-production"))

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with complete profile data
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.UserRegistrationRequest true "User registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/register [post]  //
// Register godoc
// @Summary Register a new user
// @Description Create a new user account with complete profile data
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.UserRegistrationRequest true "User registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/register [post]  //
// Register godoc
// @Summary Register a new user
// @Description Create a new user account with complete profile data
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.UserRegistrationRequest true "User registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/register [post]
func Register(c *gin.Context) {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		c.JSON(503, gin.H{"success": false, "error": "Service temporarily unavailable"})
		return
	}

	var req models.UserRegistrationRequest
	// 1. Validation using struct tags
	if err := c.ShouldBindJSON(&req); err != nil {
		// Nice error message formatting
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Validation failed: " + utils.FormatValidationError(err),
		})
		return
	}

	// 2. Extra custom validation
	if len(req.Password) < 8 {
		c.JSON(400, gin.H{"success": false, "error": "Password must be at least 8 characters"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Internal server error"})
		return
	}

	// Set default role
	if req.Role == "" {
		req.Role = "employee"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	defaultTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	var userID uuid.UUID

	// 3. Database Insertion with Safe Error Handling
	err = database.DB.QueryRowContext(ctx,
		`INSERT INTO godplan.users 
			(tenant_id, username, email, password, role, full_name, phone, avatar_url, is_active, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		 RETURNING id`,
		defaultTenantID, req.Username, req.Email, string(hashedPassword), req.Role, req.FullName, req.Phone, "", true, time.Now(), time.Now(),
	).Scan(&userID)

	if err != nil {
		// Log the real error internally
		if config.IsDevelopment() {
			fmt.Printf("❌ Register DB Error: %v\n", err)
		}
		
		// Return safe error to client
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(409, gin.H{"success": false, "error": "Username or email already exists"})
		} else {
			c.JSON(500, gin.H{"success": false, "error": "Failed to create account. Please try again later."})
		}
		return
	}

	// Auto-create employee record (logic preserved)
	if req.Role == "employee" {
		employeeID := fmt.Sprintf("EMP-%s", userID.String()[:8])
		database.DB.ExecContext(ctx,
			`INSERT INTO godplan.employees (tenant_id, user_id, employee_id, join_date, created_at, updated_at) 
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			defaultTenantID, userID, employeeID, time.Now(), time.Now(), time.Now(),
		)
	}

	// Fetch created user for response
	var createdUser models.User
	var phone, avatarURL sql.NullString
	database.DB.QueryRowContext(ctx, `SELECT id, tenant_id, username, email, role, full_name, phone, avatar_url, is_active, created_at, updated_at FROM godplan.users WHERE id = $1`, userID).
		Scan(&createdUser.ID, &createdUser.TenantID, &createdUser.Username, &createdUser.Email, &createdUser.Role, &createdUser.FullName, &phone, &avatarURL, &createdUser.IsActive, &createdUser.CreatedAt, &createdUser.UpdatedAt)
	
	createdUser.Phone = phone.String
	createdUser.AvatarURL = avatarURL.String

	// 4. Generate Tokens (Access + Refresh)
	accessToken, _ := jwtUtil.GenerateToken(createdUser.ID, createdUser.Email, createdUser.Role, createdUser.TenantID)
	refreshToken, _ := jwtUtil.GenerateRefreshToken(createdUser.ID, createdUser.TenantID)

	c.JSON(201, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data": map[string]interface{}{
			"token":         accessToken,
			"refresh_token": refreshToken,
			"user":          createdUser,
		},
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	if err := database.HealthCheck(); err != nil {
		c.JSON(503, gin.H{"success": false, "error": "Service temporarily unavailable"})
		return
	}

	var credentials models.LoginRequest
	// 1. Validation
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "Validation failed: " + utils.FormatValidationError(err)})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var user models.User
	var phone, avatarURL sql.NullString
	
	// 2. Safe Query
	err := database.DB.QueryRowContext(ctx,
		`SELECT id, tenant_id, username, email, password, role, full_name, phone, avatar_url, is_active, created_at, updated_at
		 FROM godplan.users WHERE email = $1 AND is_active = true`,
		credentials.Email,
	).Scan(&user.ID, &user.TenantID, &user.Username, &user.Email, &user.Password, &user.Role, &user.FullName, &phone, &avatarURL, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	user.Phone = phone.String
	user.AvatarURL = avatarURL.String

	// Generic error for securities
	invalidCredentialsMsg := "Invalid email or password"

	if err != nil {
		c.JSON(401, gin.H{"success": false, "error": invalidCredentialsMsg})
		return
	}

	// 3. Password Check
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		c.JSON(401, gin.H{"success": false, "error": invalidCredentialsMsg})
		return
	}

	// 4. Generate Tokens
	accessToken, err := jwtUtil.GenerateToken(user.ID, user.Email, user.Role, user.TenantID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to generate session"})
		return
	}
	
	refreshToken, err := jwtUtil.GenerateRefreshToken(user.ID, user.TenantID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to generate session"})
		return
	}

	user.Password = "" // Hide password

	c.JSON(200, gin.H{
		"success": true,
		"message": "Login successful",
		"data": map[string]interface{}{
			"token":         accessToken,
			"refresh_token": refreshToken,
			"user":          user,
		},
	})
}

// RefreshToken godoc
// @Summary Refresh Access Token
// @Description Get a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh Token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "refresh_token is required"})
		return
	}

	// Validate Refresh Token
	claims, err := jwtUtil.ValidateRefreshToken(body.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"success": false, "error": "Invalid or expired refresh token"})
		return
	}

	// Get fresh user data (to ensure role/email hasn't changed)
	var user models.User
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = database.DB.QueryRowContext(ctx, 
		"SELECT email, role FROM godplan.users WHERE id = $1 AND is_active = true", 
		claims.UserID).Scan(&user.Email, &user.Role)
	
	if err != nil {
		c.JSON(401, gin.H{"success": false, "error": "User no longer active"})
		return
	}

	// Generate NEW Access Token
	newAccessToken, err := jwtUtil.GenerateToken(claims.UserID, user.Email, user.Role, claims.TenantID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": map[string]string{
			"token": newAccessToken,
		},
	})
}


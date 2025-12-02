package handlers

import (
	"context"
	"fmt"
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

var jwtUtil = utils.NewJWTUtil("your-secret-key-change-in-production")

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
func Register(c *gin.Context) {
	// Check database connection first
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in Register: %v\n", err)
		}
		c.JSON(503, gin.H{
			"success": false,
			"error":   "Database connection lost",
		})
		return
	}

	var req models.UserRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid JSON format",
		})
		return
	}

	// Required fields validation
	if req.Username == "" || req.Name == "" || req.Email == "" || req.Password == "" {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Required fields: username, name, email, password",
		})
		return
	}

	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid email format",
		})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Password must be at least 6 characters",
		})
		return
	}

	// DEBUG: Print received data
	if config.IsDevelopment() {
		fmt.Printf("📥 REGISTER ATTEMPT - Username: %s, Email: %s, Name: %s, Role: %s\n",
			req.Username, req.Email, req.Name, req.Role)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to hash password: %v\n", err)
		}
		c.JSON(500, gin.H{
			"success": false,
			"error":   "Failed to process request",
		})
		return
	}

	// Set default role
	if req.Role == "" {
		req.Role = "employee"
	}

	// Gunakan context dengan timeout untuk query database
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Default Tenant ID
	defaultTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	var userID uuid.UUID
	err = database.DB.QueryRowContext(ctx,
		`INSERT INTO godplan.users 
			(tenant_id, username, email, password, role, name, phone, avatar_url, is_active, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		 RETURNING id`,
		defaultTenantID,
		req.Username,
		req.Email,
		string(hashedPassword),
		req.Role,
		req.Name,
		req.Phone,
		"",   // avatar_url kosong
		true, // is_active
		time.Now(),
		time.Now(),
	).Scan(&userID)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to create user: %v\n", err)

			// Check if user already exists
			var usernameCount, emailCount int
			database.DB.QueryRow("SELECT COUNT(*) FROM godplan.users WHERE username = $1 AND tenant_id = $2", req.Username, defaultTenantID).Scan(&usernameCount)
			database.DB.QueryRow("SELECT COUNT(*) FROM godplan.users WHERE email = $1 AND tenant_id = $2", req.Email, defaultTenantID).Scan(&emailCount)

			fmt.Printf("🔍 Username '%s' exists: %d, Email '%s' exists: %d\n",
				req.Username, usernameCount, req.Email, emailCount)
		}

		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(409, gin.H{
				"success": false,
				"error":   "User with this email or username already exists",
			})
		} else {
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Failed to create user - " + err.Error(),
			})
		}
		return
	}

	// ✅ AUTO-CREATE EMPLOYEE RECORD FOR EMPLOYEE ROLE
	if req.Role == "employee" {
		// Use UUID for employee ID logic if needed, but here we just use a string format
		// Since userID is UUID, we can't use %04d. We'll use a random suffix or part of UUID.
		employeeID := fmt.Sprintf("EMP-%s", userID.String()[:8])
		if config.IsDevelopment() {
			fmt.Printf("👨‍💼 Creating employee record - UserID: %s, EmployeeID: %s\n", userID, employeeID)
		}

		_, err = database.DB.ExecContext(ctx,
			`INSERT INTO godplan.employees 
			 (tenant_id, user_id, employee_id, position_id, department_id, join_date, created_at, updated_at) 
			 VALUES ($1, $2, $3, NULL, NULL, $4, $5, $6)`,
			defaultTenantID,
			userID,
			employeeID,
			time.Now(), // Join date
			time.Now(),
			time.Now(),
		)

		if err != nil {
			if config.IsDevelopment() {
				fmt.Printf("⚠️ Failed to create employee record: %v\n", err)
			}
			// Jangan return error di sini, karena user sudah berhasil dibuat
			// Employee record bisa dibuat manual nanti
		} else {
			if config.IsDevelopment() {
				fmt.Printf("✅ Employee record created successfully\n")
			}
		}
	}

	// Get the created user to return complete data
	var createdUser models.User
	err = database.DB.QueryRowContext(ctx,
		`SELECT id, tenant_id, username, email, role, name, phone, avatar_url, is_active, created_at, updated_at 
		 FROM godplan.users WHERE id = $1`,
		userID,
	).Scan(
		&createdUser.ID,
		&createdUser.TenantID,
		&createdUser.Username,
		&createdUser.Email,
		&createdUser.Role,
		&createdUser.Name,
		&createdUser.Phone,
		&createdUser.AvatarURL,
		&createdUser.IsActive,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to fetch created user: %v\n", err)
		}
		c.JSON(500, gin.H{
			"success": false,
			"error":   "User created but failed to retrieve details",
		})
		return
	}

	// Generate JWT token
	token, err := jwtUtil.GenerateToken(createdUser.ID, createdUser.Email, createdUser.Role, createdUser.TenantID)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to generate token: %v\n", err)
		}
		c.JSON(500, gin.H{
			"success": false,
			"error":   "Failed to process request",
		})
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ User registered successfully - ID: %s, Email: %s, Role: %s\n",
			createdUser.ID, createdUser.Email, createdUser.Role)
	}

	// Return response with complete user data
	c.JSON(201, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data": map[string]interface{}{
			"token": token,
			"user":  createdUser,
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
	// Check database connection first
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in Login: %v\n", err)
		}
		c.JSON(503, gin.H{
			"success": false,
			"error":   "Database connection lost",
		})
		return
	}

	var credentials models.LoginRequest

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid JSON format",
		})
		return
	}

	if credentials.Email == "" || credentials.Password == "" {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Email and password are required",
		})
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("🔐 LOGIN ATTEMPT - Email: %s\n", credentials.Email)
	}

	// Gunakan context dengan timeout untuk query database
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Default Tenant ID
	defaultTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	var user models.User
	err := database.DB.QueryRowContext(ctx,
		`SELECT id, tenant_id, username, email, password, role, name, phone, 
			avatar_url, is_active, created_at, updated_at
		 FROM godplan.users WHERE email = $1 AND tenant_id = $2 AND is_active = true`,
		credentials.Email, defaultTenantID,
	).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Name,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ User not found or DB error: %v\n", err)
		}
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Invalid credentials",
		})
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ User found - ID: %s, Email: %s, Role: %s\n", user.ID, user.Email, user.Role)
		fmt.Printf("🔐 PASSWORD DEBUG - Stored hash: %s\n", user.Password)
		fmt.Printf("🔐 PASSWORD DEBUG - Input password: %s\n", credentials.Password)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ PASSWORD MISMATCH: %v\n", err)

			// Test hash the input password to debug
			testHash, hashErr := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
			if hashErr == nil {
				fmt.Printf("🔐 DEBUG - New hash of input: %s\n", string(testHash))
			}
		}
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Invalid credentials",
		})
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ Password verified successfully\n")
	}

	// Generate JWT token
	token, err := jwtUtil.GenerateToken(user.ID, user.Email, user.Role, user.TenantID)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Token generation failed: %v\n", err)
		}
		c.JSON(500, gin.H{
			"success": false,
			"error":   "Failed to process request",
		})
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ Login successful - User ID: %s, Role: %s\n", user.ID, user.Role)
	}

	// Clear password before returning user data
	user.Password = ""

	// Return response dengan data user lengkap
	c.JSON(200, gin.H{
		"success": true,
		"message": "Login successful",
		"data": map[string]interface{}{
			"token": token,
			"user":  user,
		},
	})
}

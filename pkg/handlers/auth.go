package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
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
// @Router /auth/register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	// Check database connection first
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in Register: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Database connection lost")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		utils.ErrorResponse(w, http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	var req models.UserRegistrationRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Required fields validation - SESUAI DENGAN MODEL BARU
	if req.Username == "" || req.FullName == "" || req.Email == "" || req.Password == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Required fields: username, full_name, email, password")
		return
	}

	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	if len(req.Password) < 6 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Failed to hash password: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	// Set default role
	if req.Role == "" {
		req.Role = "employee"
	}

	// Gunakan context dengan timeout untuk query database
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var userID int64
	err = database.DB.QueryRowContext(ctx,
		`INSERT INTO godplan.users 
			(username, email, password, role, full_name, phone, avatar_url, is_active, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		 RETURNING id`,
		req.Username,
		req.Email,
		string(hashedPassword),
		req.Role,
		req.FullName,
		req.Phone,
		"",   // avatar_url kosong
		true, // is_active
		time.Now(),
		time.Now(),
	).Scan(&userID)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Failed to create user: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create user - user may already exist")
		return
	}

	// Get the created user to return complete data
	var createdUser models.User
	err = database.DB.QueryRowContext(ctx,
		`SELECT id, username, email, role, full_name, phone, avatar_url, is_active, created_at, updated_at 
		 FROM godplan.users WHERE id = $1`,
		userID,
	).Scan(
		&createdUser.ID,
		&createdUser.Username,
		&createdUser.Email,
		&createdUser.Role,
		&createdUser.FullName,
		&createdUser.Phone,
		&createdUser.AvatarURL,
		&createdUser.IsActive,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Failed to fetch created user: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "User created but failed to retrieve details")
		return
	}

	// Generate JWT token
	token, err := jwtUtil.GenerateToken(int(createdUser.ID), createdUser.Email, createdUser.Role)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Failed to generate token: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("User registered successfully - ID: %d, Email: %s\n", createdUser.ID, createdUser.Email)
	}

	// Return response with complete user data
	utils.SuccessResponse(w, http.StatusCreated, "User registered successfully", map[string]interface{}{
		"token": token,
		"user":  createdUser,
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
// @Router /auth/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	// Check database connection first
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in Login: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Database connection lost")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		utils.ErrorResponse(w, http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	var credentials models.LoginRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&credentials); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if credentials.Email == "" || credentials.Password == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("Login attempt - Email: %s\n", credentials.Email)
	}

	// Gunakan context dengan timeout untuk query database
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var user models.User
	err := database.DB.QueryRowContext(ctx,
		`SELECT id, username, email, password, role, full_name, phone, 
			avatar_url, is_active, created_at, updated_at
		 FROM godplan.users WHERE email = $1 AND is_active = true`,
		credentials.Email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.FullName,
		&user.Phone,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("User not found or DB error: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("User found - ID: %d, Email: %s, Role: %s\n", user.ID, user.Email, user.Role)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Password mismatch for user: %s\n", credentials.Email)
		}
		utils.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("Password verified successfully\n")
	}

	// Generate JWT token
	token, err := jwtUtil.GenerateToken(int(user.ID), user.Email, user.Role)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Token generation failed: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("Login successful - User ID: %d, Role: %s\n", user.ID, user.Role)
	}

	// Clear password before returning user data
	user.Password = ""

	// Return response dengan data user lengkap
	utils.SuccessResponse(w, http.StatusOK, "Login successful", map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

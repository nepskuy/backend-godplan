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
	"github.com/nepskuy/be-godplan/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

var jwtUtil = utils.NewJWTUtil("your-secret-key-change-in-production")

type RegisterRequest struct {
	Username string `json:"username" example:"johndoe"`
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"admin@godplan.com"`
	Password string `json:"password" example:"password"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
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

	var req RegisterRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if req.Username == "" || req.Name == "" || req.Email == "" || req.Password == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "All fields are required: username, name, email, password")
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Failed to hash password: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	// Gunakan context dengan timeout untuk query database
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var userID int
	err = database.DB.QueryRowContext(ctx,
		"INSERT INTO users (username, name, email, password, role) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		req.Username, req.Name, req.Email, string(hashedPassword), "employee",
	).Scan(&userID)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Failed to create user: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	token, err := jwtUtil.GenerateToken(userID, req.Email, "employee")
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("Failed to generate token: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("User registered successfully - ID: %d, Email: %s\n", userID, req.Email)
	}

	utils.SuccessResponse(w, http.StatusCreated, "User registered successfully", map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       userID,
			"username": req.Username,
			"name":     req.Name,
			"email":    req.Email,
			"role":     "employee",
		},
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
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

	var credentials LoginRequest

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

	var user struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	// Gunakan context dengan timeout untuk query database
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := database.DB.QueryRowContext(ctx,
		"SELECT id, username, name, email, password, role FROM users WHERE email = $1",
		credentials.Email,
	).Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.Password, &user.Role)

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

	token, err := jwtUtil.GenerateToken(user.ID, user.Email, user.Role)
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

	utils.SuccessResponse(w, http.StatusOK, "Login successful", map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in GetUsers: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Database connection lost")
		return
	}

	// Query ke schema godplan
	rows, err := database.DB.Query(`
		SELECT id, username, email, role, full_name, phone, avatar_url, is_active, created_at, updated_at 
		FROM godplan.users 
		WHERE is_active = true
	`)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to fetch users: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
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
				fmt.Printf("❌ Failed to scan user: %v\n", err)
			}
			continue
		}
		users = append(users, user)
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ GetUsers successful: found %d users\n", len(users))
	}

	utils.SuccessResponse(w, http.StatusOK, "Users retrieved successfully", users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in CreateUser: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Database connection lost")
		return
	}

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if user.Username == "" || user.Email == "" || user.Password == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Set default values
	if user.Role == "" {
		user.Role = "employee"
	}
	if user.FullName == "" {
		user.FullName = user.Username
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to hash password: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	var id int64
	err = database.DB.QueryRow(
		`INSERT INTO godplan.users (
			username, email, password, role, full_name, phone, avatar_url, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		user.Username,
		user.Email,
		string(hashedPassword),
		user.Role,
		user.FullName,
		user.Phone,
		user.AvatarURL,
		true, // is_active
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to create user: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create user - user may already exist")
		return
	}

	// Get the created user to return complete data
	var createdUser models.User
	err = database.DB.QueryRow(
		`SELECT id, username, email, role, full_name, phone, avatar_url, is_active, created_at, updated_at 
		 FROM godplan.users WHERE id = $1`,
		id,
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
			fmt.Printf("❌ Failed to fetch created user: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "User created but failed to retrieve details")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ CreateUser successful: ID=%d, Username=%s, Email=%s\n", createdUser.ID, createdUser.Username, createdUser.Email)
	}

	utils.SuccessResponse(w, http.StatusCreated, "User created successfully", createdUser)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in GetUser: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusServiceUnavailable, "Database connection lost")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	err = database.DB.QueryRow(
		`SELECT id, username, email, role, full_name, phone, avatar_url, is_active, created_at, updated_at 
		 FROM godplan.users WHERE id = $1 AND is_active = true`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
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
			fmt.Printf("❌ User not found: ID=%d, error=%v\n", id, err)
		}
		utils.ErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ GetUser successful: ID=%d, Username=%s\n", user.ID, user.Username)
	}

	utils.SuccessResponse(w, http.StatusOK, "User retrieved successfully", user)
}

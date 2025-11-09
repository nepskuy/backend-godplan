package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

	rows, err := database.DB.Query("SELECT id, username, email, created_at FROM users")
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
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
		if err != nil {
			if config.IsDevelopment() {
				fmt.Printf("❌ Failed to scan user: %v\n", err)
			}
			continue // Skip invalid rows instead of failing entire request
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

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to hash password: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	var id int
	err = database.DB.QueryRow(
		"INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id",
		user.Username, string(hashedPassword), user.Email,
	).Scan(&id)

	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to create user: %v\n", err)
		}
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	user.ID = id
	user.Password = "" // Clear password from response

	if config.IsDevelopment() {
		fmt.Printf("✅ CreateUser successful: ID=%d, Username=%s\n", user.ID, user.Username)
	}

	utils.SuccessResponse(w, http.StatusCreated, "User created successfully", user)
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
		"SELECT id, username, email, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

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

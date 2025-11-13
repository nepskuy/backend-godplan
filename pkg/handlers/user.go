package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/nepskuy/be-godplan/pkg/config"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// GetUsers godoc
// @Summary Get all users
// @Description Get list of all active users
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.GinResponse
// @Failure 500 {object} utils.GinResponse
// @Router /users [get]
func GetUsers(c *gin.Context) {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in GetUsers: %v\n", err)
		}
		utils.GinErrorResponse(c, 503, "Database connection lost")
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
		utils.GinErrorResponse(c, 500, "Failed to fetch users")
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

	utils.GinSuccessResponse(c, 200, "Users retrieved successfully", users)
}

// CreateUser godoc
// @Summary Create new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.User true "User data"
// @Success 201 {object} utils.GinResponse
// @Failure 400 {object} utils.GinResponse
// @Failure 500 {object} utils.GinResponse
// @Router /users [post]
func CreateUser(c *gin.Context) {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in CreateUser: %v\n", err)
		}
		utils.GinErrorResponse(c, 503, "Database connection lost")
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid request body")
		return
	}

	// Validate required fields
	if user.Username == "" || user.Email == "" || user.Password == "" {
		utils.GinErrorResponse(c, 400, "Username, email, and password are required")
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
		utils.GinErrorResponse(c, 500, "Failed to process request")
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
		utils.GinErrorResponse(c, 500, "Failed to create user - user may already exist")
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
		utils.GinErrorResponse(c, 500, "User created but failed to retrieve details")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ CreateUser successful: ID=%d, Username=%s, Email=%s\n", createdUser.ID, createdUser.Username, createdUser.Email)
	}

	utils.GinSuccessResponse(c, 201, "User created successfully", createdUser)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a specific user by ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} utils.GinResponse
// @Failure 400 {object} utils.GinResponse
// @Failure 404 {object} utils.GinResponse
// @Failure 500 {object} utils.GinResponse
// @Router /users/{id} [get]
func GetUser(c *gin.Context) {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Database connection error in GetUser: %v\n", err)
		}
		utils.GinErrorResponse(c, 503, "Database connection lost")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid user ID")
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
		utils.GinErrorResponse(c, 404, "User not found")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ GetUser successful: ID=%d, Username=%s\n", user.ID, user.Username)
	}

	utils.GinSuccessResponse(c, 200, "User retrieved successfully", user)
}

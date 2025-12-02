package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 401, "Invalid tenant ID")
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, tenant_id, username, email, role, name, phone, avatar_url, is_active, created_at, updated_at 
		FROM godplan.users 
		WHERE is_active = true AND tenant_id = $1
	`, tenantID)
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
			&user.TenantID,
			&user.Username,
			&user.Email,
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

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 401, "Invalid tenant ID")
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
	if user.Name == "" {
		user.Name = user.Username
	}
	user.TenantID = tenantID

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		if config.IsDevelopment() {
			fmt.Printf("❌ Failed to hash password: %v\n", err)
		}
		utils.GinErrorResponse(c, 500, "Failed to process request")
		return
	}

	var id uuid.UUID
	err = database.DB.QueryRow(
		`INSERT INTO godplan.users (
			tenant_id, username, email, password, role, name, phone, avatar_url, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		user.TenantID,
		user.Username,
		user.Email,
		string(hashedPassword),
		user.Role,
		user.Name,
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
		`SELECT id, tenant_id, username, email, role, name, phone, avatar_url, is_active, created_at, updated_at 
		 FROM godplan.users WHERE id = $1`,
		id,
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
		utils.GinErrorResponse(c, 500, "User created but failed to retrieve details")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ CreateUser successful: ID=%s, Username=%s, Email=%s\n", createdUser.ID, createdUser.Username, createdUser.Email)
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
// @Param id path string true "User ID"
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

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 401, "Invalid tenant ID")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid user ID")
		return
	}

	var user models.User
	err = database.DB.QueryRow(
		`SELECT id, tenant_id, username, email, role, name, phone, avatar_url, is_active, created_at, updated_at 
		 FROM godplan.users WHERE id = $1 AND tenant_id = $2 AND is_active = true`,
		id, tenantID,
	).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
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
			fmt.Printf("❌ User not found: ID=%s, error=%v\n", id, err)
		}
		utils.GinErrorResponse(c, 404, "User not found")
		return
	}

	if config.IsDevelopment() {
		fmt.Printf("✅ GetUser successful: ID=%s, Username=%s\n", user.ID, user.Username)
	}

	utils.GinSuccessResponse(c, 200, "User retrieved successfully", user)
}

package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents user data structure
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	FullName  string    `json:"full_name" db:"full_name"` // Changed from Name to FullName
	Role      string    `json:"role,omitempty" db:"role"`
	Phone     string    `json:"phone,omitempty" db:"phone"`
	AvatarURL string    `json:"avatar_url,omitempty" db:"avatar_url"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	// Employee fields
	EmployeeID     string `json:"employee_id,omitempty" db:"employee_id"`
	Department     string `json:"department,omitempty" db:"department"`
	Position       string `json:"position,omitempty" db:"position"`
	Status         string `json:"status,omitempty" db:"status"`
	EmploymentType string `json:"employment_type,omitempty" db:"employment_type"`
	JoinDate       string `json:"join_date,omitempty" db:"join_date"`
	WorkSchedule   string `json:"work_schedule,omitempty" db:"work_schedule"`
}

// UserRegistrationRequest for register endpoint
type UserRegistrationRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name" binding:"required"` // Changed from Name to FullName
	Phone    string `json:"phone,omitempty"`
	Role     string `json:"role,omitempty"`
}

// LoginRequest for login endpoint
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Employee represents employee data structure
type Employee struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	TenantID       uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	EmployeeID     string     `json:"employee_id" db:"employee_id"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty" db:"department_id"`
	PositionID     *uuid.UUID `json:"position_id,omitempty" db:"position_id"`
	BaseSalary     float64    `json:"base_salary" db:"base_salary"`
	JoinDate       string     `json:"join_date" db:"join_date"`
	EmploymentType string     `json:"employment_type" db:"employment_type"`
	WorkSchedule   string     `json:"work_schedule" db:"work_schedule"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

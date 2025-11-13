package models

import (
	"time"
)

// User represents user data structure
type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Name      string    `json:"full_name,omitempty" db:"name"`
	Role      string    `json:"role,omitempty" db:"role"`
	Phone     string    `json:"phone,omitempty" db:"phone"`
	AvatarURL string    `json:"avatar_url,omitempty" db:"avatar_url"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRegistrationRequest for register endpoint
type UserRegistrationRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"full_name" binding:"required"`
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
	ID             int64     `json:"id" db:"id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	EmployeeID     string    `json:"employee_id" db:"employee_id"`
	DepartmentID   *int64    `json:"department_id,omitempty" db:"department_id"`
	PositionID     *int64    `json:"position_id,omitempty" db:"position_id"`
	BaseSalary     float64   `json:"base_salary" db:"base_salary"`
	JoinDate       string    `json:"join_date" db:"join_date"`
	EmploymentType string    `json:"employment_type" db:"employment_type"`
	WorkSchedule   string    `json:"work_schedule" db:"work_schedule"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

package models

import (
	"time"
)

// User represents user data structure
// @Description User information
type User struct {
	ID        int64     `json:"id" db:"id" example:"1"`
	Username  string    `json:"username" db:"username" example:"johndoe"`
	Email     string    `json:"email" db:"email" example:"john@example.com"`
	Password  string    `json:"-" db:"password"`
	Role      string    `json:"role" db:"role" example:"employee"`
	FullName  string    `json:"full_name" db:"full_name" example:"John Doe"`
	Phone     string    `json:"phone,omitempty" db:"phone" example:"+628123456789"`
	AvatarURL string    `json:"avatar_url,omitempty" db:"avatar_url" example:"https://example.com/avatar.jpg"`
	IsActive  bool      `json:"is_active" db:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2023-10-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" example:"2023-10-01T00:00:00Z"`
}

// Employee represents employee data structure
// @Description Employee information
type Employee struct {
	ID             string    `json:"id" db:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         string    `json:"user_id" db:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	EmployeeID     string    `json:"employee_id" db:"employee_id" example:"EMP001"`
	DepartmentID   string    `json:"department_id" db:"department_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	PositionID     string    `json:"position_id" db:"position_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	BaseSalary     float64   `json:"base_salary" db:"base_salary" example:"5000000.00"`
	JoinDate       string    `json:"join_date" db:"join_date" example:"2023-01-15"`
	EmploymentType string    `json:"employment_type" db:"employment_type" example:"full_time"`
	WorkSchedule   string    `json:"work_schedule" db:"work_schedule" example:"9-to-5"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" example:"2023-10-01T00:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at" example:"2023-10-01T00:00:00Z"`
}

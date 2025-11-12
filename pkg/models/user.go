package models

import (
	"time"
)

type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Role      string    `json:"role" db:"role"`
	FullName  string    `json:"full_name" db:"full_name"`
	Phone     string    `json:"phone" db:"phone"`
	AvatarURL string    `json:"avatar_url" db:"avatar_url"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Employee struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	EmployeeID     string    `json:"employee_id" db:"employee_id"`
	DepartmentID   string    `json:"department_id" db:"department_id"`
	PositionID     string    `json:"position_id" db:"position_id"`
	BaseSalary     float64   `json:"base_salary" db:"base_salary"`
	JoinDate       string    `json:"join_date" db:"join_date"`
	EmploymentType string    `json:"employment_type" db:"employment_type"`
	WorkSchedule   string    `json:"work_schedule" db:"work_schedule"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

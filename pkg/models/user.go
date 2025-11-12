package models

import (
	"time"
)

type User struct {
	ID         int64     `json:"id" db:"id"`
	Username   string    `json:"username" db:"username"`
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	Password   string    `json:"-" db:"password"`
	EmployeeID string    `json:"employee_id,omitempty" db:"employee_id"`
	NISN       string    `json:"nisn,omitempty" db:"nisn"`
	Department string    `json:"department,omitempty" db:"department"`
	Position   string    `json:"position,omitempty" db:"position"`
	Status     string    `json:"status,omitempty" db:"status"`
	Phone      string    `json:"phone,omitempty" db:"phone"`
	Role       string    `json:"role" db:"role"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

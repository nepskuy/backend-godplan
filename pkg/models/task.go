package models

import "time"

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      int       `json:"user_id"`
	AssignedTo  int       `json:"assigned_to"`
	Status      string    `json:"status"`
	Completed   bool      `json:"completed"`
	Deadline    time.Time `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
}

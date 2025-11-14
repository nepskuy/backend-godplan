package models

import "time"

type Task struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	AssigneeID     string    `json:"assignee_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	DueDate        string    `json:"due_date"`
	EstimatedHours float64   `json:"estimated_hours"`
	ActualHours    float64   `json:"actual_hours"`
	Progress       int       `json:"progress"`
	Status         string    `json:"status"`
	Priority       string    `json:"priority"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TaskRequest struct {
	ProjectID      string  `json:"project_id"`
	AssigneeID     string  `json:"assignee_id"`
	Title          string  `json:"title" binding:"required"`
	Description    string  `json:"description"`
	DueDate        string  `json:"due_date"`
	EstimatedHours float64 `json:"estimated_hours"`
	ActualHours    float64 `json:"actual_hours"`
	Progress       int     `json:"progress"`
	Status         string  `json:"status"`
	Priority       string  `json:"priority"`
}

type UpcomingTask struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	DueDate  string `json:"due_date"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

type TaskStatistics struct {
	TotalTasks     int `json:"total_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	PendingTasks   int `json:"pending_tasks"`
	CompletionRate int `json:"completion_rate"`
}

package models

import "time"

// CRMProject represents a CRM deal / project in the sales pipeline
// This is designed to be parallel with the mobile CRM UI and the projects table used in dashboard stats.
type CRMProject struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Client        string    `json:"client"`
	Value         float64   `json:"value"`
	Stage         string    `json:"stage"`          // e.g. 'new', 'qualified', 'proposal', 'won'
	Urgency       string    `json:"urgency"`       // e.g. 'low', 'medium', 'high'
	Deadline      string    `json:"deadline"`      // ISO date string
	ContactPerson string    `json:"contact_person"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`      // e.g. 'godjah', 'godtive', 'godweb'
	Status        string    `json:"status"`
	ManagerID     string    `json:"manager_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CRMProjectRequest is used for create/update operations
// ID, ManagerID, CreatedAt, UpdatedAt are managed by the backend.
type CRMProjectRequest struct {
	Title         string  `json:"title" binding:"required"`
	Client        string  `json:"client" binding:"required"`
	Value         float64 `json:"value"`
	Stage         string  `json:"stage"`
	Urgency       string  `json:"urgency"`
	Deadline      string  `json:"deadline"`
	ContactPerson string  `json:"contact_person"`
	Description   string  `json:"description"`
	Category      string  `json:"category"`
	Status        string  `json:"status"`
}

// CRMProjectStatistics can be used later for aggregated CRM analytics if needed.
type CRMProjectStatistics struct {
	TotalProjects int     `json:"total_projects"`
	WonProjects   int     `json:"won_projects"`
	TotalValue    float64 `json:"total_value"`
	WinRate       int     `json:"win_rate"`
}

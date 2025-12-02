package models

import "github.com/google/uuid"

type DashboardStats struct {
	ActiveProjects   int    `json:"active_projects"`
	PendingTasks     int    `json:"pending_tasks"`
	AttendanceStatus string `json:"attendance_status"`
	CompletionRate   int    `json:"completion_rate"`
}

type TeamMember struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Position  string    `json:"position"`
}

type HomeDashboardResponse struct {
	Stats       DashboardStats `json:"stats"`
	TeamMembers []TeamMember   `json:"team_members"`
	Greeting    string         `json:"greeting"`
	UserName    string         `json:"user_name"`
	UserAvatar  string         `json:"user_avatar"`
}

package models

type DashboardStats struct {
	ActiveProjects   int    `json:"active_projects"`
	PendingTasks     int    `json:"pending_tasks"`
	AttendanceStatus string `json:"attendance_status"`
	CompletionRate   int    `json:"completion_rate"`
}

type TeamMember struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Position  string `json:"position"`
}

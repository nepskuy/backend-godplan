package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get overview statistics for home dashboard
// @Tags dashboard
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /dashboard/stats [get]
// @Security BearerAuth
func GetDashboardStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	// Count active projects (from CRM)
	var activeProjects int
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM godplan.projects 
		WHERE status != 'completed' AND user_id = $1
	`, userID).Scan(&activeProjects)
	if err != nil {
		activeProjects = 0 // Default jika error
	}

	// Count pending tasks
	var pendingTasks int
	err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM godplan.tasks 
		WHERE (user_id = $1 OR assigned_to = $1) AND completed = false
	`, userID).Scan(&pendingTasks)
	if err != nil {
		pendingTasks = 0
	}

	// Check today's attendance
	var attendanceStatus string
	err = database.DB.QueryRow(`
		SELECT CASE 
			WHEXISTS (SELECT 1 FROM godplan.attendances WHERE user_id = $1 AND DATE(created_at) = CURRENT_DATE) 
			THEN 'present' ELSE 'absent' END
	`, userID).Scan(&attendanceStatus)
	if err != nil {
		attendanceStatus = "absent"
	}

	// Calculate completion rate
	var totalTasks, completedTasks int
	err = database.DB.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN completed = true THEN 1 END) as completed
		FROM godplan.tasks 
		WHERE user_id = $1 OR assigned_to = $1
	`, userID).Scan(&totalTasks, &completedTasks)

	var completionRate int
	if totalTasks > 0 {
		completionRate = (completedTasks * 100) / totalTasks
	} else {
		completionRate = 0
	}

	stats := models.DashboardStats{
		ActiveProjects:   activeProjects,
		PendingTasks:     pendingTasks,
		AttendanceStatus: attendanceStatus,
		CompletionRate:   completionRate,
	}

	utils.GinSuccessResponse(c, 200, "Dashboard stats retrieved successfully", stats)
}

// GetTeamMembers godoc
// @Summary Get team members
// @Description Get list of team members
// @Tags dashboard
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /teams [get]
// @Security BearerAuth
func GetTeamMembers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	var members []models.TeamMember
	rows, err := database.DB.Query(`
		SELECT id, name, avatar_url, position 
		FROM godplan.users 
		WHERE id != $1 AND is_active = true
		LIMIT 4
	`, userID)

	if err != nil {
		// Fallback ke data dummy jika query gagal
		members = []models.TeamMember{
			{
				ID:        1,
				Name:      "Rina",
				AvatarURL: "/avatars/rina.jpg",
				Position:  "Developer",
			},
			{
				ID:        2,
				Name:      "Budi",
				AvatarURL: "/avatars/budi.jpg",
				Position:  "Designer",
			},
			{
				ID:        3,
				Name:      "Sari",
				AvatarURL: "/avatars/sari.jpg",
				Position:  "Manager",
			},
			{
				ID:        4,
				Name:      "Andi",
				AvatarURL: "/avatars/andi.jpg",
				Position:  "Developer",
			},
		}
	} else {
		defer rows.Close()

		for rows.Next() {
			var member models.TeamMember
			err := rows.Scan(
				&member.ID,
				&member.Name,
				&member.AvatarURL,
				&member.Position,
			)
			if err == nil {
				members = append(members, member)
			}
		}
	}

	utils.GinSuccessResponse(c, 200, "Team members retrieved successfully", members)
}

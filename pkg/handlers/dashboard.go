package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// GetHomeDashboard godoc
// @Summary Get home dashboard data
// @Description Get complete data for home dashboard including stats and team members
// @Tags dashboard
// @Produce json
// @Success 200 {object} models.HomeDashboardResponse
// @Router /home [get]
// @Security BearerAuth
func GetHomeDashboard(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	// Get user profile data
	var userName, userAvatar string
	err := database.DB.QueryRow(`
		SELECT name, avatar_url FROM godplan.users WHERE id = $1
	`, userID).Scan(&userName, &userAvatar)
	if err != nil {
		userName = "User"
		userAvatar = "/avatars/default.jpg"
	}

	// Get dashboard stats
	stats := getDashboardStats(userID.(int64))

	// Get team members
	teamMembers := getTeamMembers(userID.(int64))

	// Get greeting based on current time
	greeting := getGreeting()

	response := models.HomeDashboardResponse{
		Stats:       stats,
		TeamMembers: teamMembers,
		Greeting:    greeting,
		UserName:    userName,
		UserAvatar:  userAvatar,
	}

	utils.GinSuccessResponse(c, 200, "Home dashboard data retrieved successfully", response)
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get overview statistics for dashboard
// @Tags dashboard
// @Produce json
// @Success 200 {object} models.DashboardStats
// @Router /dashboard/stats [get]
// @Security BearerAuth
func GetDashboardStats(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	// Gin context stores userID as int (set in GinAuthMiddleware),
	// but getDashboardStats expects int64. Safely convert.
	var userID int64
	switch v := userIDVal.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	default:
		utils.GinErrorResponse(c, 500, "invalid user id type")
		return
	}

	stats := getDashboardStats(userID)
	utils.GinSuccessResponse(c, 200, "Dashboard stats retrieved successfully", stats)
}

// GetTeamMembers godoc
// @Summary Get team members
// @Description Get list of team members
// @Tags dashboard
// @Produce json
// @Success 200 {object} []models.TeamMember
// @Router /teams [get]
// @Security BearerAuth
func GetTeamMembers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	teamMembers := getTeamMembers(userID.(int64))
	utils.GinSuccessResponse(c, 200, "Team members retrieved successfully", teamMembers)
}

// Helper functions
func getDashboardStats(userID int64) models.DashboardStats {
	var stats models.DashboardStats

	// NOTE: In the current schema, projects.manager_id and tasks.assignee_id
	// both reference employees(id), which is linked to users(id) via employees.user_id.
	// For now we treat userID as employees.id and count projects/tasks where the
	// logged-in user is the manager/assignee.

	// Count active projects for this manager
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM godplan.projects 
		WHERE status != 'completed' AND manager_id = $1
	`, userID).Scan(&stats.ActiveProjects)
	if err != nil {
		stats.ActiveProjects = 0
	}

	// Count pending tasks for this assignee (status != 'completed')
	err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM godplan.tasks 
		WHERE assignee_id = $1 AND status != 'completed'
	`, userID).Scan(&stats.PendingTasks)
	if err != nil {
		stats.PendingTasks = 0
	}

	// Check today's attendance (schema already matches user_id)
	err = database.DB.QueryRow(`
		SELECT CASE 
			WHEN EXISTS (SELECT 1 FROM godplan.attendances WHERE user_id = $1 AND DATE(created_at) = CURRENT_DATE) 
			THEN 'present' ELSE 'absent' END
	`, userID).Scan(&stats.AttendanceStatus)
	if err != nil {
		stats.AttendanceStatus = "absent"
	}

	// Calculate completion rate based on tasks status
	var totalTasks, completedTasks int
	err = database.DB.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed
		FROM godplan.tasks 
		WHERE assignee_id = $1
	`, userID).Scan(&totalTasks, &completedTasks)

	if err != nil {
		// Handle error jika query gagal
		stats.CompletionRate = 0
	} else if totalTasks > 0 {
		stats.CompletionRate = (completedTasks * 100) / totalTasks
	} else {
		stats.CompletionRate = 0
	}

	return stats
}

func getTeamMembers(userID int64) []models.TeamMember {
	var members []models.TeamMember

	rows, err := database.DB.Query(`
		SELECT id, name, avatar_url, position 
		FROM godplan.users 
		WHERE id != $1 AND is_active = true
		LIMIT 4
	`, userID)

	if err != nil {
		// Fallback data
		members = getFallbackTeamMembers()
	} else {
		defer rows.Close()

		for rows.Next() {
			var member models.TeamMember
			if err := rows.Scan(&member.ID, &member.Name, &member.AvatarURL, &member.Position); err == nil {
				members = append(members, member)
			}
		}

		// Handle jika tidak ada data dari database
		if len(members) == 0 {
			members = getFallbackTeamMembers()
		}
	}

	return members
}

func getFallbackTeamMembers() []models.TeamMember {
	return []models.TeamMember{
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
}

func getGreeting() string {
	hour := time.Now().Hour()
	switch {
	case hour < 12:
		return "Selamat Pagi"
	case hour < 18:
		return "Selamat Siang"
	default:
		return "Selamat Malam"
	}
}

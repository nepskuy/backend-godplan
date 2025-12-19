package handlers

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	userIDVal, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 401, "Invalid tenant ID")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			utils.GinErrorResponse(c, 500, "Invalid user ID format")
			return
		}
	default:
		utils.GinErrorResponse(c, 500, "Invalid user ID type")
		return
	}

	// Get user profile data
	var userName, userAvatar string
	err = database.DB.QueryRow(`
		SELECT COALESCE(full_name, username) as name, COALESCE(avatar_url, '') as avatar_url 
		FROM godplan.users WHERE id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&userName, &userAvatar)
	if err != nil {
		userName = "User"
		userAvatar = "/avatars/default.jpg"
	}

	// Get dashboard stats
	stats := getDashboardStats(tenantID, userID)

	// Get team members
	teamMembers := getTeamMembers(tenantID, userID)

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

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 401, "Invalid tenant ID")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			utils.GinErrorResponse(c, 500, "Invalid user ID format")
			return
		}
	default:
		utils.GinErrorResponse(c, 500, "Invalid user ID type")
		return
	}

	stats := getDashboardStats(tenantID, userID)
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
	userIDVal, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 401, "Invalid tenant ID")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			utils.GinErrorResponse(c, 500, "Invalid user ID format")
			return
		}
	default:
		utils.GinErrorResponse(c, 500, "Invalid user ID type")
		return
	}

	teamMembers := getTeamMembers(tenantID, userID)
	utils.GinSuccessResponse(c, 200, "Team members retrieved successfully", teamMembers)
}

// Helper functions
func getDashboardStats(tenantID uuid.UUID, userID uuid.UUID) models.DashboardStats {
	var stats models.DashboardStats

	// Cari employee_id berdasarkan user_id (this must be done first)
	var employeeID uuid.UUID
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		// Jika tidak ada employee record, return empty stats
		return stats
	}

	// Execute remaining queries in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(4)

	// 1. Count active projects for this manager
	go func() {
		defer wg.Done()
		var activeProjects int
		err := database.DB.QueryRow(`
			SELECT COUNT(*) FROM godplan.projects 
			WHERE status != 'completed' AND manager_id = $1 AND tenant_id = $2
		`, employeeID, tenantID).Scan(&activeProjects)
		if err != nil {
			activeProjects = 0
		}
		mu.Lock()
		stats.ActiveProjects = activeProjects
		mu.Unlock()
	}()

	// 2. Count pending tasks for this assignee
	go func() {
		defer wg.Done()
		var pendingTasks int
		err := database.DB.QueryRow(`
			SELECT COUNT(*) FROM godplan.tasks 
			WHERE assignee_id = $1 AND status != 'completed' AND tenant_id = $2
		`, employeeID, tenantID).Scan(&pendingTasks)
		if err != nil {
			pendingTasks = 0
		}
		mu.Lock()
		stats.PendingTasks = pendingTasks
		mu.Unlock()
	}()

	// 3. Check today's attendance
	go func() {
		defer wg.Done()
		var attendanceStatus string
		err := database.DB.QueryRow(`
			SELECT CASE 
				WHEN EXISTS (SELECT 1 FROM godplan.attendances WHERE user_id = $1 AND tenant_id = $2 AND attendance_date = CURRENT_DATE) 
				THEN 'present' ELSE 'absent' END
		`, userID, tenantID).Scan(&attendanceStatus)
		if err != nil {
			attendanceStatus = "absent"
		}
		mu.Lock()
		stats.AttendanceStatus = attendanceStatus
		mu.Unlock()
	}()

	// 4. Calculate completion rate based on tasks status
	go func() {
		defer wg.Done()
		var totalTasks, completedTasks int
		err := database.DB.QueryRow(`
			SELECT 
				COUNT(*) as total,
				COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed
			FROM godplan.tasks 
			WHERE assignee_id = $1 AND tenant_id = $2
		`, employeeID, tenantID).Scan(&totalTasks, &completedTasks)

		var completionRate int
		if err != nil {
			completionRate = 0
		} else if totalTasks > 0 {
			completionRate = (completedTasks * 100) / totalTasks
		} else {
			completionRate = 0
		}
		mu.Lock()
		stats.CompletionRate = completionRate
		mu.Unlock()
	}()

	// Wait for all parallel queries to complete
	wg.Wait()

	return stats
}

func getTeamMembers(tenantID uuid.UUID, userID uuid.UUID) []models.TeamMember {
	var members []models.TeamMember

	rows, err := database.DB.Query(`
		SELECT 
			u.id, 
			COALESCE(u.full_name, u.username) as name, 
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(p.name, 'Employee') as position 
		FROM godplan.users u
		LEFT JOIN godplan.employees e ON e.user_id = u.id AND e.tenant_id = u.tenant_id
		LEFT JOIN godplan.positions p ON e.position_id = p.id
		WHERE u.id != $1 AND u.tenant_id = $2 AND u.is_active = true
		LIMIT 4
	`, userID, tenantID)

	if err != nil {
		// Return empty array instead of fallback
		return []models.TeamMember{}
	}
	
	defer rows.Close()

	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.ID, &member.Name, &member.AvatarURL, &member.Position); err == nil {
			members = append(members, member)
		}
	}

	// Return whatever we got from database (could be empty)
	return members
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

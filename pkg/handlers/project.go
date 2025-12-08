package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// ProjectResponse represents a project for mobile API
type ProjectResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	Status      string  `json:"status"`
	Progress    int     `json:"progress"`
	ManagerID   string  `json:"manager_id"`
	ManagerName string  `json:"manager_name"`
	PhaseName   string  `json:"phase_name,omitempty"`
}

// GetProjects godoc
// @Summary Get projects for current user
// @Description Get list of operational projects where user is manager
// @Tags projects
// @Produce json
// @Success 200 {object} []ProjectResponse
// @Router /projects [get]
// @Security BearerAuth
func GetProjects(c *gin.Context) {
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

	// Get employee ID for this user
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		utils.GinSuccessResponse(c, 200, "Projects retrieved successfully", []ProjectResponse{})
		return
	}

	// Get projects where user is manager
	rows, err := database.DB.Query(`
		SELECT 
			p.id, 
			p.name, 
			p.description, 
			p.start_date, 
			p.end_date, 
			p.status, 
			p.progress,
			p.manager_id,
			COALESCE(u.name, '') as manager_name,
			COALESCE(pp.name, '') as phase_name
		FROM godplan.projects p
		LEFT JOIN godplan.employees e ON p.manager_id = e.id
		LEFT JOIN godplan.users u ON e.user_id = u.id
		LEFT JOIN godplan.project_phases pp ON p.current_phase_id = pp.id
		WHERE p.tenant_id = $1 
		AND (p.manager_id = $2 OR p.assigned_to = $2)
		ORDER BY p.updated_at DESC
	`, tenantID, employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to fetch projects")
		return
	}
	defer rows.Close()

	var projects []ProjectResponse
	for rows.Next() {
		var p ProjectResponse
		var startDate, endDate interface{}
		
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&startDate,
			&endDate,
			&p.Status,
			&p.Progress,
			&p.ManagerID,
			&p.ManagerName,
			&p.PhaseName,
		)
		if err != nil {
			continue
		}

		// Format dates
		if t, ok := startDate.(interface{ String() string }); ok {
			p.StartDate = t.String()
		}
		if t, ok := endDate.(interface{ String() string }); ok {
			p.EndDate = t.String()
		}

		projects = append(projects, p)
	}

	if projects == nil {
		projects = []ProjectResponse{}
	}

	utils.GinSuccessResponse(c, 200, "Projects retrieved successfully", projects)
}

// GetProject godoc
// @Summary Get single project details
// @Description Get details of a specific project
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} ProjectResponse
// @Router /projects/{id} [get]
// @Security BearerAuth
func GetProject(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		utils.GinErrorResponse(c, 400, "Project ID is required")
		return
	}

	tenantIDStr := c.GetString("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 401, "Invalid tenant ID")
		return
	}

	var p ProjectResponse
	var startDate, endDate interface{}

	err = database.DB.QueryRow(`
		SELECT 
			p.id, 
			p.name, 
			p.description, 
			p.start_date, 
			p.end_date, 
			p.status, 
			p.progress,
			p.manager_id,
			COALESCE(u.name, '') as manager_name,
			COALESCE(pp.name, '') as phase_name
		FROM godplan.projects p
		LEFT JOIN godplan.employees e ON p.manager_id = e.id
		LEFT JOIN godplan.users u ON e.user_id = u.id
		LEFT JOIN godplan.project_phases pp ON p.current_phase_id = pp.id
		WHERE p.id = $1 AND p.tenant_id = $2
	`, projectID, tenantID).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&startDate,
		&endDate,
		&p.Status,
		&p.Progress,
		&p.ManagerID,
		&p.ManagerName,
		&p.PhaseName,
	)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Project not found")
		return
	}

	// Format dates
	if t, ok := startDate.(interface{ String() string }); ok {
		p.StartDate = t.String()
	}
	if t, ok := endDate.(interface{ String() string }); ok {
		p.EndDate = t.String()
	}

	utils.GinSuccessResponse(c, 200, "Project retrieved successfully", p)
}

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
	"github.com/nepskuy/be-godplan/pkg/service"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

var (
	crmRepo    repository.CRMRepository = repository.NewCRMRepository(database.GetDB())
	crmService service.CRMService       = service.NewCRMService(crmRepo)
)

// GetCRMProjects godoc
// @Summary Get CRM projects for current user
// @Description Get list of CRM projects (pipeline) for the current user as manager
// @Tags crm
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.GinResponse
// @Router /crm/projects [get]
func GetCRMProjects(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	// Map user_id -> employee_id (same pattern as tasks)
	var managerID string
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&managerID)

	if err != nil {
		// If no employee record, return empty list
		utils.GinSuccessResponse(c, 200, "CRM projects retrieved successfully", []models.CRMProject{})
		return
	}

	projects, err := crmService.GetProjectsByManager(managerID)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to fetch CRM projects")
		return
	}

	utils.GinSuccessResponse(c, 200, "CRM projects retrieved successfully", projects)
}

// CreateCRMProject godoc
// @Summary Create new CRM project
// @Description Create a new CRM project (used by dashboard, PWA should remain read-only)
// @Tags crm
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CRMProjectRequest true "CRM project data"
// @Success 201 {object} utils.GinResponse
// @Router /crm/projects [post]
func CreateCRMProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	var managerID string
	if err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&managerID); err != nil {
		utils.GinErrorResponse(c, 400, "User doesn't have employee record")
		return
	}

	var req models.CRMProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid request data")
		return
	}

	project := &models.CRMProject{
		Title:         req.Title,
		Client:        req.Client,
		Value:         req.Value,
		Stage:         req.Stage,
		Urgency:       req.Urgency,
		Deadline:      req.Deadline,
		ContactPerson: req.ContactPerson,
		Description:   req.Description,
		Category:      req.Category,
		Status:        req.Status,
		ManagerID:     managerID,
	}

	if err := crmService.CreateProject(project); err != nil {
		utils.GinErrorResponse(c, 500, "Failed to create CRM project")
		return
	}

	utils.GinSuccessResponse(c, 201, "CRM project created successfully", project)
}

// GetCRMProject godoc
// @Summary Get CRM project by ID
// @Description Get a specific CRM project by ID
// @Tags crm
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "CRM Project ID"
// @Success 200 {object} utils.GinResponse
// @Router /crm/projects/{id} [get]
func GetCRMProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var managerID string
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	hasAccess, err := crmService.ValidateProjectAccess(projectID, managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "CRM project not found")
		return
	}
	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this CRM project")
		return
	}

	project, err := crmService.GetProjectByID(projectID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "CRM project not found")
		return
	}

	utils.GinSuccessResponse(c, 200, "CRM project retrieved successfully", project)
}

// UpdateCRMProject godoc
// @Summary Update CRM project
// @Description Update an existing CRM project
// @Tags crm
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "CRM Project ID"
// @Param request body models.CRMProjectRequest true "CRM project data"
// @Success 200 {object} utils.GinResponse
// @Router /crm/projects/{id} [put]
func UpdateCRMProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var managerID string
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	hasAccess, err := crmService.ValidateProjectAccess(projectID, managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "CRM project not found")
		return
	}
	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this CRM project")
		return
	}

	var req models.CRMProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid request data")
		return
	}

	project, err := crmService.GetProjectByID(projectID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "CRM project not found")
		return
	}

	project.Title = req.Title
	project.Client = req.Client
	project.Value = req.Value
	project.Stage = req.Stage
	project.Urgency = req.Urgency
	project.Deadline = req.Deadline
	project.ContactPerson = req.ContactPerson
	project.Description = req.Description
	project.Category = req.Category
	project.Status = req.Status

	if err := crmService.UpdateProject(project); err != nil {
		utils.GinErrorResponse(c, 500, "Failed to update CRM project")
		return
	}

	utils.GinSuccessResponse(c, 200, "CRM project updated successfully", project)
}

// DeleteCRMProject godoc
// @Summary Delete CRM project
// @Description Delete a CRM project by ID
// @Tags crm
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "CRM Project ID"
// @Success 200 {object} utils.GinResponse
// @Router /crm/projects/{id} [delete]
func DeleteCRMProject(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	projectID := c.Param("id")

	var managerID string
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	hasAccess, err := crmService.ValidateProjectAccess(projectID, managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "CRM project not found")
		return
	}
	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this CRM project")
		return
	}

	if err := crmService.DeleteProject(projectID); err != nil {
		utils.GinErrorResponse(c, 500, "Failed to delete CRM project")
		return
	}

	utils.GinSuccessResponse(c, 200, "CRM project deleted successfully", nil)
}

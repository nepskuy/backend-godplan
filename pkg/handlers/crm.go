package handlers

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
	"github.com/nepskuy/be-godplan/pkg/service"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

var (
	crmRepo    repository.CRMRepository
	crmService service.CRMService
	crmOnce    sync.Once
)

// getCRMService returns lazily initialized CRM service
// This prevents nil pointer panic when database is not yet connected at package init time
func getCRMService() service.CRMService {
	crmOnce.Do(func() {
		crmRepo = repository.NewCRMRepository(database.GetDB())
		crmService = service.NewCRMService(crmRepo)
	})
	return crmService
}

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

	// Map user_id -> employee_id (same pattern as tasks)
	var managerID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&managerID)

	if err != nil {
		// If no employee record, return empty list
		utils.GinSuccessResponse(c, 200, "CRM projects retrieved successfully", []models.CRMProject{})
		return
	}

	projects, err := getCRMService().GetProjectsByManager(tenantID, managerID)
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

	var managerID uuid.UUID
	if err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&managerID); err != nil {
		utils.GinErrorResponse(c, 400, "User doesn't have employee record")
		return
	}

	var req models.CRMProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid request data")
		return
	}

	project := &models.CRMProject{
		TenantID:      tenantID,
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

	if err := getCRMService().CreateProject(project); err != nil {
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

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid project ID")
		return
	}

	var managerID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	hasAccess, err := getCRMService().ValidateProjectAccess(tenantID, projectID, managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "CRM project not found")
		return
	}
	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this CRM project")
		return
	}

	project, err := getCRMService().GetProjectByID(tenantID, projectID)
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

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid project ID")
		return
	}

	var managerID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	hasAccess, err := getCRMService().ValidateProjectAccess(tenantID, projectID, managerID)
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

	project, err := getCRMService().GetProjectByID(tenantID, projectID)
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

	if err := getCRMService().UpdateProject(project); err != nil {
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

	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid project ID")
		return
	}

	var managerID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	hasAccess, err := getCRMService().ValidateProjectAccess(tenantID, projectID, managerID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "CRM project not found")
		return
	}
	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this CRM project")
		return
	}

	if err := getCRMService().DeleteProject(tenantID, projectID); err != nil {
		utils.GinErrorResponse(c, 500, "Failed to delete CRM project")
		return
	}

	utils.GinSuccessResponse(c, 200, "CRM project deleted successfully", nil)
}

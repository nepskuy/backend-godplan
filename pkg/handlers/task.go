package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/database"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
	"github.com/nepskuy/be-godplan/pkg/service"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

var (
	taskRepo    repository.TaskRepository = repository.NewTaskRepository(database.GetDB())
	taskService service.TaskService       = service.NewTaskService(taskRepo)
)

// GetTasks godoc
// @Summary Get all tasks for current user
// @Description Get list of tasks assigned to the current user
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.GinResponse
// @Router /tasks [get]
func GetTasks(c *gin.Context) {
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

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		// Jika tidak ada employee record, return empty array
		utils.GinSuccessResponse(c, 200, "Tasks retrieved successfully", []models.Task{})
		return
	}

	tasks, err := taskService.GetTasksByAssignee(tenantID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to fetch tasks")
		return
	}

	utils.GinSuccessResponse(c, 200, "Tasks retrieved successfully", tasks)
}

// CreateTask godoc
// @Summary Create new task
// @Description Create a new task for the current user
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.TaskRequest true "Task data"
// @Success 201 {object} utils.GinResponse
// @Router /tasks [post]
func CreateTask(c *gin.Context) {
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

	var taskReq models.TaskRequest
	if err := c.ShouldBindJSON(&taskReq); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid request data")
		return
	}

	// Validate required fields
	if taskReq.Title == "" {
		utils.GinErrorResponse(c, 400, "Task title is required")
		return
	}

	var assigneeID uuid.UUID
	// Jika assignee_id kosong, gunakan employee_id dari user yang login
	if taskReq.AssigneeID == "" {
		err := database.DB.QueryRow(`
			SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
		`, userID, tenantID).Scan(&assigneeID)

		if err != nil {
			utils.GinErrorResponse(c, 400, "User doesn't have employee record")
			return
		}
	} else {
		var err error
		assigneeID, err = uuid.Parse(taskReq.AssigneeID)
		if err != nil {
			utils.GinErrorResponse(c, 400, "Invalid assignee ID")
			return
		}
	}

	var projectID uuid.UUID
	if taskReq.ProjectID != "" {
		var err error
		projectID, err = uuid.Parse(taskReq.ProjectID)
		if err != nil {
			utils.GinErrorResponse(c, 400, "Invalid project ID")
			return
		}
	}

	task := &models.Task{
		TenantID:       tenantID,
		ProjectID:      projectID,
		AssigneeID:     assigneeID,
		Title:          taskReq.Title,
		Description:    taskReq.Description,
		Completed:      taskReq.Completed,
		Priority:       taskReq.Priority,
		DueDate:        taskReq.DueDate,
		Category:       taskReq.Category,
		EstimatedHours: taskReq.EstimatedHours,
		ActualHours:    taskReq.ActualHours,
		Progress:       taskReq.Progress,
		Status:         taskReq.Status,
	}

	err = taskService.CreateTask(task)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to create task")
		return
	}

	utils.GinSuccessResponse(c, 201, "Task created successfully", task)
}

// GetTask godoc
// @Summary Get task by ID
// @Description Get a specific task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id} [get]
func GetTask(c *gin.Context) {
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

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid task ID")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(tenantID, taskID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this task")
		return
	}

	task, err := taskService.GetTaskByID(tenantID, taskID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	utils.GinSuccessResponse(c, 200, "Task retrieved successfully", task)
}

// UpdateTask godoc
// @Summary Update task
// @Description Update an existing task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Param request body models.TaskRequest true "Task data"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id} [put]
func UpdateTask(c *gin.Context) {
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

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid task ID")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(tenantID, taskID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this task")
		return
	}

	var taskReq models.TaskRequest
	if err := c.ShouldBindJSON(&taskReq); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid request data")
		return
	}

	// Get existing task
	existingTask, err := taskService.GetTaskByID(tenantID, taskID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	var projectID uuid.UUID
	if taskReq.ProjectID != "" {
		projectID, err = uuid.Parse(taskReq.ProjectID)
		if err != nil {
			utils.GinErrorResponse(c, 400, "Invalid project ID")
			return
		}
	}

	var assigneeID uuid.UUID
	if taskReq.AssigneeID != "" {
		assigneeID, err = uuid.Parse(taskReq.AssigneeID)
		if err != nil {
			utils.GinErrorResponse(c, 400, "Invalid assignee ID")
			return
		}
	} else {
		assigneeID = existingTask.AssigneeID
	}

	// Update task fields
	existingTask.ProjectID = projectID
	existingTask.AssigneeID = assigneeID
	existingTask.Title = taskReq.Title
	existingTask.Description = taskReq.Description
	existingTask.Completed = taskReq.Completed
	existingTask.Priority = taskReq.Priority
	existingTask.DueDate = taskReq.DueDate
	existingTask.Category = taskReq.Category
	existingTask.EstimatedHours = taskReq.EstimatedHours
	existingTask.ActualHours = taskReq.ActualHours
	existingTask.Progress = taskReq.Progress
	existingTask.Status = taskReq.Status

	err = taskService.UpdateTask(existingTask)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to update task")
		return
	}

	utils.GinSuccessResponse(c, 200, "Task updated successfully", existingTask)
}

// DeleteTask godoc
// @Summary Delete task
// @Description Delete a task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id} [delete]
func DeleteTask(c *gin.Context) {
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

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid task ID")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(tenantID, taskID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this task")
		return
	}

	err = taskService.DeleteTask(tenantID, taskID)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			utils.GinErrorResponse(c, 404, "Task not found")
		} else {
			utils.GinErrorResponse(c, 500, "Failed to delete task")
		}
		return
	}

	utils.GinSuccessResponse(c, 200, "Task deleted successfully", nil)
}

// GetUpcomingTasks godoc
// @Summary Get upcoming tasks
// @Description Get upcoming tasks for dashboard (limit 3)
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.GinResponse
// @Router /tasks/upcoming [get]
func GetUpcomingTasks(c *gin.Context) {
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

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		// Jika tidak ada employee record, return empty array
		utils.GinSuccessResponse(c, 200, "Upcoming tasks retrieved successfully", []models.UpcomingTask{})
		return
	}

	tasks, err := taskService.GetUpcomingTasks(tenantID, employeeID, 3)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to fetch upcoming tasks")
		return
	}

	utils.GinSuccessResponse(c, 200, "Upcoming tasks retrieved successfully", tasks)
}

// ToggleTaskCompletion godoc
// @Summary Toggle task completion status
// @Description Toggle the completed status of a task (true/false)
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Param request body models.ToggleTaskRequest true "Completion status"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id}/toggle [patch]
func ToggleTaskCompletion(c *gin.Context) {
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

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid task ID")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(tenantID, taskID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this task")
		return
	}

	var toggleReq models.ToggleTaskRequest
	if err := c.ShouldBindJSON(&toggleReq); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid request data")
		return
	}

	err = taskService.ToggleTaskCompletion(tenantID, taskID, toggleReq.Completed)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to toggle task completion")
		return
	}

	task, err := taskService.GetTaskByID(tenantID, taskID)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Task updated but failed to retrieve")
		return
	}

	utils.GinSuccessResponse(c, 200, "Task completion status updated successfully", task)
}

// UpdateTaskCategory godoc
// @Summary Update task category
// @Description Update category of a specific task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Param request body models.UpdateTaskCategoryRequest true "Category data"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id}/category [patch]
func UpdateTaskCategory(c *gin.Context) {
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

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid task ID")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(tenantID, taskID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this task")
		return
	}

	var categoryReq models.UpdateTaskCategoryRequest
	if err := c.ShouldBindJSON(&categoryReq); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid request data")
		return
	}

	err = taskService.UpdateTaskCategory(tenantID, taskID, categoryReq.Category)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to update task category")
		return
	}

	task, err := taskService.GetTaskByID(tenantID, taskID)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Category updated but failed to retrieve task")
		return
	}

	utils.GinSuccessResponse(c, 200, "Task category updated successfully", task)
}

// UpdateTaskProgress godoc
// @Summary Update task progress
// @Description Update progress of a specific task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Param request body map[string]interface{} true "Progress data"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id}/progress [patch]
func UpdateTaskProgress(c *gin.Context) {
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

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid task ID")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(tenantID, taskID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this task")
		return
	}

	var progressReq struct {
		Progress int `json:"progress" binding:"required,min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&progressReq); err != nil {
		utils.GinErrorResponse(c, 400, "Invalid progress value")
		return
	}

	err = taskService.UpdateTaskProgress(tenantID, taskID, progressReq.Progress)
	if err != nil {
		if err == repository.ErrInvalidProgress {
			utils.GinErrorResponse(c, 400, "Progress must be between 0 and 100")
		} else {
			utils.GinErrorResponse(c, 500, "Failed to update task progress")
		}
		return
	}

	task, err := taskService.GetTaskByID(tenantID, taskID)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Progress updated but failed to retrieve task")
		return
	}

	utils.GinSuccessResponse(c, 200, "Task progress updated successfully", task)
}

// CompleteTask godoc
// @Summary Complete task
// @Description Mark a task as completed
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id}/complete [patch]
func CompleteTask(c *gin.Context) {
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

	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		utils.GinErrorResponse(c, 400, "Invalid task ID")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(tenantID, taskID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this task")
		return
	}

	err = taskService.CompleteTask(tenantID, taskID)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to complete task")
		return
	}

	task, err := taskService.GetTaskByID(tenantID, taskID)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Task completed but failed to retrieve")
		return
	}

	utils.GinSuccessResponse(c, 200, "Task completed successfully", task)
}

// GetTaskStatistics godoc
// @Summary Get task statistics
// @Description Get task statistics for current user
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.GinResponse
// @Router /tasks/statistics [get]
func GetTaskStatistics(c *gin.Context) {
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

	// Cari employee_id berdasarkan user_id
	var employeeID uuid.UUID
	err = database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(&employeeID)

	if err != nil {
		// Jika tidak ada employee record, return empty statistics
		utils.GinSuccessResponse(c, 200, "Task statistics retrieved successfully", models.TaskStatistics{})
		return
	}

	statistics, err := taskService.GetTaskStatistics(tenantID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to fetch task statistics")
		return
	}

	utils.GinSuccessResponse(c, 200, "Task statistics retrieved successfully", statistics)
}

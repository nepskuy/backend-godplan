// Di pkg/handlers/task.go - update untuk pakai service
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
	taskRepo    = repository.NewTaskRepository(database.GetDB())
	taskService = service.NewTaskService(taskRepo)
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
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID string
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&employeeID)

	if err != nil {
		// Jika tidak ada employee record, return empty array
		utils.GinSuccessResponse(c, 200, "Tasks retrieved successfully", []models.Task{})
		return
	}

	tasks, err := taskService.GetTasksByAssignee(employeeID)
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
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
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

	// Jika assignee_id kosong, gunakan employee_id dari user yang login
	if taskReq.AssigneeID == "" {
		err := database.DB.QueryRow(`
			SELECT id FROM godplan.employees WHERE user_id = $1
		`, userID).Scan(&taskReq.AssigneeID)

		if err != nil {
			utils.GinErrorResponse(c, 400, "User doesn't have employee record")
			return
		}
	}

	task := &models.Task{
		ProjectID:      taskReq.ProjectID,
		AssigneeID:     taskReq.AssigneeID,
		Title:          taskReq.Title,
		Description:    taskReq.Description,
		DueDate:        taskReq.DueDate,
		EstimatedHours: taskReq.EstimatedHours,
		ActualHours:    taskReq.ActualHours,
		Progress:       taskReq.Progress,
		Status:         taskReq.Status,
		Priority:       taskReq.Priority,
	}

	err := taskService.CreateTask(task)
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
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	taskID := c.Param("id")

	// Cari employee_id berdasarkan user_id
	var employeeID string
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(taskID, employeeID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	if !hasAccess {
		utils.GinErrorResponse(c, 403, "Access denied to this task")
		return
	}

	task, err := taskService.GetTaskByID(taskID)
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
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	taskID := c.Param("id")

	// Cari employee_id berdasarkan user_id
	var employeeID string
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&employeeID)

	if err != nil {
		utils.GinErrorResponse(c, 404, "Employee record not found")
		return
	}

	// Validate task access
	hasAccess, err := taskService.ValidateTaskAccess(taskID, employeeID)
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
	existingTask, err := taskService.GetTaskByID(taskID)
	if err != nil {
		utils.GinErrorResponse(c, 404, "Task not found")
		return
	}

	// Update task fields
	existingTask.ProjectID = taskReq.ProjectID
	existingTask.AssigneeID = taskReq.AssigneeID
	existingTask.Title = taskReq.Title
	existingTask.Description = taskReq.Description
	existingTask.DueDate = taskReq.DueDate
	existingTask.EstimatedHours = taskReq.EstimatedHours
	existingTask.ActualHours = taskReq.ActualHours
	existingTask.Progress = taskReq.Progress
	existingTask.Status = taskReq.Status
	existingTask.Priority = taskReq.Priority

	err = taskService.UpdateTask(existingTask)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to update task")
		return
	}

	utils.GinSuccessResponse(c, 200, "Task updated successfully", existingTask)
}

// GetUpcomingTasks godoc
// @Summary Get upcoming tasks
// @Description Get upcoming tasks for dashboard (limit 5)
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.GinResponse
// @Router /tasks/upcoming [get]
func GetUpcomingTasks(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.GinErrorResponse(c, 401, "Unauthorized")
		return
	}

	// Cari employee_id berdasarkan user_id
	var employeeID string
	err := database.DB.QueryRow(`
		SELECT id FROM godplan.employees WHERE user_id = $1
	`, userID).Scan(&employeeID)

	if err != nil {
		// Jika tidak ada employee record, return empty array
		utils.GinSuccessResponse(c, 200, "Upcoming tasks retrieved successfully", []models.UpcomingTask{})
		return
	}

	tasks, err := taskService.GetUpcomingTasks(employeeID, 5)
	if err != nil {
		utils.GinErrorResponse(c, 500, "Failed to fetch upcoming tasks")
		return
	}

	utils.GinSuccessResponse(c, 200, "Upcoming tasks retrieved successfully", tasks)
}

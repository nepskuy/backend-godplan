package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// GetTasks godoc
// @Summary Get all tasks
// @Description Get list of tasks for the current user
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.GinResponse
// @Router /tasks [get]
func GetTasks(c *gin.Context) {
	utils.GinSuccessResponse(c, 200, "Get tasks - not implemented", nil)
}

// CreateTask godoc
// @Summary Create new task
// @Description Create a new task for the current user
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Task data"
// @Success 200 {object} utils.GinResponse
// @Router /tasks [post]
func CreateTask(c *gin.Context) {
	utils.GinSuccessResponse(c, 200, "Create task - not implemented", nil)
}

// GetTask godoc
// @Summary Get task by ID
// @Description Get a specific task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id} [get]
func GetTask(c *gin.Context) {
	utils.GinSuccessResponse(c, 200, "Get task - not implemented", nil)
}

// UpdateTask godoc
// @Summary Update task
// @Description Update an existing task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Param request body map[string]interface{} true "Task data"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id} [put]
func UpdateTask(c *gin.Context) {
	utils.GinSuccessResponse(c, 200, "Update task - not implemented", nil)
}

// DeleteTask godoc
// @Summary Delete task
// @Description Delete a task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 200 {object} utils.GinResponse
// @Router /tasks/{id} [delete]
func DeleteTask(c *gin.Context) {
	utils.GinSuccessResponse(c, 200, "Delete task - not implemented", nil)
}

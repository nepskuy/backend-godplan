// pkg/handlers/task.go
package handlers

import (
	"net/http"

	"github.com/nepskuy/be-godplan/pkg/utils"
)

func GetTasks(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, "Get tasks - not implemented", nil)
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, "Create task - not implemented", nil)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, "Get task - not implemented", nil)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, "Update task - not implemented", nil)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, http.StatusOK, "Delete task - not implemented", nil)
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"

	"github.com/gorilla/mux"
)

func GetTasks(w http.ResponseWriter, r *http.Request) {
    // Sample tasks dengan field yang sesuai model
    tasks := []models.Task{
        {
            ID:          1,
            Title:       "Sample Task 1",
            Description: "This is a sample task",
            UserID:      1,
            AssignedTo:  1,
            Status:      "pending",
            Completed:   false,
        },
        {
            ID:          2,
            Title:       "Sample Task 2",
            Description: "Another sample task",
            UserID:      1,
            AssignedTo:  1,
            Status:      "in_progress",
            Completed:   false,
        },
    }

    utils.SuccessResponse(w, http.StatusOK, "Tasks retrieved successfully", tasks)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        utils.ErrorResponse(w, http.StatusBadRequest, "Invalid task ID")
        return
    }

    task := models.Task{
        ID:          id,
        Title:       "Sample Task",
        Description: "This is a sample task",
        UserID:      1,
        AssignedTo:  1,
        Status:      "pending",
        Completed:   false,
    }

    utils.SuccessResponse(w, http.StatusOK, "Task retrieved successfully", task)
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
    var task struct {
        Title       string `json:"title"`
        Description string `json:"description"`
    }

    if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    // FIX: Validasi required field - title tidak boleh kosong
    if task.Title == "" {
        utils.ErrorResponse(w, http.StatusBadRequest, "Title is required")
        return
    }

    // Create task object dengan default values
    newTask := models.Task{
        ID:          3,
        Title:       task.Title,
        Description: task.Description,
        UserID:      1,
        AssignedTo:  1,
        Status:      "pending",
        Completed:   false,
    }

    utils.SuccessResponse(w, http.StatusCreated, "Task created successfully", newTask)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        utils.ErrorResponse(w, http.StatusBadRequest, "Invalid task ID")
        return
    }

    var task models.Task
    if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    task.ID = id

    utils.SuccessResponse(w, http.StatusOK, "Task updated successfully", task)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        utils.ErrorResponse(w, http.StatusBadRequest, "Invalid task ID")
        return
    }

    utils.SuccessResponse(w, http.StatusOK, "Task deleted successfully", map[string]int{"deleted_id": id})
}

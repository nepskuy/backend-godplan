package repository

import (
	"database/sql"
	"errors"

	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// Define custom errors
var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrInvalidProgress = errors.New("progress must be between 0 and 100")
	ErrAccessDenied    = errors.New("access denied to task")
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(task *models.Task) error {
	query := `INSERT INTO godplan.tasks 
		(project_id, assignee_id, title, description, due_date, 
		 estimated_hours, actual_hours, progress, status, priority) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		task.ProjectID,
		task.AssigneeID,
		task.Title,
		task.Description,
		task.DueDate,
		task.EstimatedHours,
		task.ActualHours,
		task.Progress,
		task.Status,
		task.Priority,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *TaskRepository) GetTasks() ([]models.Task, error) {
	query := `SELECT id, project_id, assignee_id, title, description, due_date, 
		 estimated_hours, actual_hours, progress, status, priority, 
		 created_at, updated_at 
		 FROM godplan.tasks ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.EstimatedHours,
			&task.ActualHours,
			&task.Progress,
			&task.Status,
			&task.Priority,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, utils.ErrInternalServer
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepository) GetTasksByAssignee(assigneeID string) ([]models.Task, error) {
	query := `SELECT id, project_id, assignee_id, title, description, due_date, 
		 estimated_hours, actual_hours, progress, status, priority, 
		 created_at, updated_at 
		 FROM godplan.tasks 
		 WHERE assignee_id = $1 
		 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, assigneeID)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.EstimatedHours,
			&task.ActualHours,
			&task.Progress,
			&task.Status,
			&task.Priority,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, utils.ErrInternalServer
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepository) GetTaskByID(id string) (*models.Task, error) {
	task := &models.Task{}
	query := `SELECT id, project_id, assignee_id, title, description, due_date, 
		 estimated_hours, actual_hours, progress, status, priority, 
		 created_at, updated_at 
		 FROM godplan.tasks WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.ProjectID,
		&task.AssigneeID,
		&task.Title,
		&task.Description,
		&task.DueDate,
		&task.EstimatedHours,
		&task.ActualHours,
		&task.Progress,
		&task.Status,
		&task.Priority,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	return task, nil
}

func (r *TaskRepository) UpdateTask(task *models.Task) error {
	query := `UPDATE godplan.tasks 
		SET project_id = $1, assignee_id = $2, title = $3, description = $4, 
		    due_date = $5, estimated_hours = $6, actual_hours = $7, 
		    progress = $8, status = $9, priority = $10, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $11`

	_, err := r.db.Exec(query,
		task.ProjectID,
		task.AssigneeID,
		task.Title,
		task.Description,
		task.DueDate,
		task.EstimatedHours,
		task.ActualHours,
		task.Progress,
		task.Status,
		task.Priority,
		task.ID,
	)

	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *TaskRepository) DeleteTask(id string) error {
	query := `DELETE FROM godplan.tasks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *TaskRepository) GetUpcomingTasks(assigneeID string, limit int) ([]models.UpcomingTask, error) {
	query := `SELECT id, title, due_date, status, priority
		FROM godplan.tasks 
		WHERE assignee_id = $1 
		AND due_date >= CURRENT_DATE
		AND status NOT IN ('completed', 'cancelled')
		ORDER BY due_date ASC
		LIMIT $2`

	rows, err := r.db.Query(query, assigneeID, limit)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.UpcomingTask
	for rows.Next() {
		var task models.UpcomingTask
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.DueDate,
			&task.Status,
			&task.Priority,
		)
		if err != nil {
			return nil, utils.ErrInternalServer
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepository) GetTaskCountByAssignee(assigneeID string) (int, int, error) {
	var totalTasks, completedTasks int
	query := `SELECT 
		COUNT(*) as total,
		COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed
		FROM godplan.tasks 
		WHERE assignee_id = $1`

	err := r.db.QueryRow(query, assigneeID).Scan(&totalTasks, &completedTasks)
	if err != nil {
		return 0, 0, utils.ErrInternalServer
	}
	return totalTasks, completedTasks, nil
}

func (r *TaskRepository) GetPendingTasksCount(assigneeID string) (int, error) {
	var pendingTasks int
	query := `SELECT COUNT(*) 
		FROM godplan.tasks 
		WHERE assignee_id = $1 
		AND status NOT IN ('completed', 'cancelled')`

	err := r.db.QueryRow(query, assigneeID).Scan(&pendingTasks)
	if err != nil {
		return 0, utils.ErrInternalServer
	}
	return pendingTasks, nil
}

// ValidateTaskAccess - Check if user has access to this task
func (r *TaskRepository) ValidateTaskAccess(taskID string, assigneeID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM godplan.tasks WHERE id = $1 AND assignee_id = $2`

	err := r.db.QueryRow(query, taskID, assigneeID).Scan(&count)
	if err != nil {
		return false, utils.ErrInternalServer
	}

	return count > 0, nil
}

// UpdateTaskProgress - Update only task progress
func (r *TaskRepository) UpdateTaskProgress(taskID string, progress int) error {
	query := `UPDATE godplan.tasks 
		SET progress = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`

	_, err := r.db.Exec(query, progress, taskID)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

// CompleteTask - Mark task as completed
func (r *TaskRepository) CompleteTask(taskID string) error {
	query := `UPDATE godplan.tasks 
		SET status = 'completed', progress = 100, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1`

	_, err := r.db.Exec(query, taskID)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

// GetTaskStatistics - Get task statistics for dashboard
func (r *TaskRepository) GetTaskStatistics(assigneeID string) (*models.TaskStatistics, error) {
	totalTasks, completedTasks, err := r.GetTaskCountByAssignee(assigneeID)
	if err != nil {
		return nil, err
	}

	pendingTasks, err := r.GetPendingTasksCount(assigneeID)
	if err != nil {
		return nil, err
	}

	var completionRate int
	if totalTasks > 0 {
		completionRate = (completedTasks * 100) / totalTasks
	}

	return &models.TaskStatistics{
		TotalTasks:     totalTasks,
		CompletedTasks: completedTasks,
		PendingTasks:   pendingTasks,
		CompletionRate: completionRate,
	}, nil
}

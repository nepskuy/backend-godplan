package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

// Define custom errors
var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrInvalidProgress = errors.New("progress must be between 0 and 100")
	ErrAccessDenied    = errors.New("access denied to task")
)

// TaskRepository interface
type TaskRepository interface {
	CreateTask(task *models.Task) error
	GetTasks(tenantID uuid.UUID) ([]models.Task, error)
	GetTaskByID(tenantID uuid.UUID, id uuid.UUID) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(tenantID uuid.UUID, id uuid.UUID) error
	GetTasksByAssignee(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error)
	GetUpcomingTasks(tenantID uuid.UUID, assigneeID uuid.UUID, limit int) ([]models.UpcomingTask, error)
	GetTaskCountByAssignee(tenantID uuid.UUID, assigneeID uuid.UUID) (int, int, error)
	GetPendingTasksCount(tenantID uuid.UUID, assigneeID uuid.UUID) (int, error)
	ValidateTaskAccess(tenantID uuid.UUID, taskID, assigneeID uuid.UUID) (bool, error)
	UpdateTaskProgress(tenantID uuid.UUID, taskID uuid.UUID, progress int) error
	CompleteTask(tenantID uuid.UUID, taskID uuid.UUID) error
	UpdateTaskCompletion(tenantID uuid.UUID, taskID uuid.UUID, completed bool) error
	UpdateTaskCategory(tenantID uuid.UUID, taskID uuid.UUID, category string) error
	GetTaskStatistics(tenantID uuid.UUID, assigneeID uuid.UUID) (*models.TaskStatistics, error)
	GetTasksByCategory(tenantID uuid.UUID, assigneeID uuid.UUID, category string) ([]models.Task, error)
	GetCompletedTasks(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error)
	GetActiveTasks(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error)
}

// taskRepositoryImpl implementasi konkret
type taskRepositoryImpl struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepositoryImpl{db: db}
}

func (r *taskRepositoryImpl) CreateTask(task *models.Task) error {
	query := `INSERT INTO godplan.tasks 
		(tenant_id, project_id, assignee_id, title, description, completed, priority, due_date, category,
		 estimated_hours, actual_hours, progress, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		task.TenantID,
		task.ProjectID,
		task.AssigneeID,
		task.Title,
		task.Description,
		task.Completed,
		task.Priority,
		task.DueDate,
		task.Category,
		task.EstimatedHours,
		task.ActualHours,
		task.Progress,
		task.Status,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *taskRepositoryImpl) GetTasks(tenantID uuid.UUID) ([]models.Task, error) {
	query := `SELECT id, tenant_id, project_id, assignee_id, title, description, completed, priority, due_date, category,
		 estimated_hours, actual_hours, progress, status, 
		 created_at, updated_at 
		 FROM godplan.tasks WHERE tenant_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.Priority,
			&task.DueDate,
			&task.Category,
			&task.EstimatedHours,
			&task.ActualHours,
			&task.Progress,
			&task.Status,
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

func (r *taskRepositoryImpl) GetTasksByAssignee(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error) {
	query := `SELECT id, tenant_id, project_id, assignee_id, title, description, completed, priority, due_date, category,
		 estimated_hours, actual_hours, progress, status, 
		 created_at, updated_at 
		 FROM godplan.tasks 
		 WHERE assignee_id = $1 AND tenant_id = $2
		 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, assigneeID, tenantID)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.Priority,
			&task.DueDate,
			&task.Category,
			&task.EstimatedHours,
			&task.ActualHours,
			&task.Progress,
			&task.Status,
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

func (r *taskRepositoryImpl) GetTaskByID(tenantID uuid.UUID, id uuid.UUID) (*models.Task, error) {
	task := &models.Task{}
	query := `SELECT id, tenant_id, project_id, assignee_id, title, description, completed, priority, due_date, category,
		 estimated_hours, actual_hours, progress, status, 
		 created_at, updated_at 
		 FROM godplan.tasks WHERE id = $1 AND tenant_id = $2`

	err := r.db.QueryRow(query, id, tenantID).Scan(
		&task.ID,
		&task.TenantID,
		&task.ProjectID,
		&task.AssigneeID,
		&task.Title,
		&task.Description,
		&task.Completed,
		&task.Priority,
		&task.DueDate,
		&task.Category,
		&task.EstimatedHours,
		&task.ActualHours,
		&task.Progress,
		&task.Status,
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

func (r *taskRepositoryImpl) UpdateTask(task *models.Task) error {
	query := `UPDATE godplan.tasks 
		SET project_id = $1, assignee_id = $2, title = $3, description = $4, 
		    completed = $5, priority = $6, due_date = $7, category = $8,
		    estimated_hours = $9, actual_hours = $10, 
		    progress = $11, status = $12, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $13 AND tenant_id = $14`

	_, err := r.db.Exec(query,
		task.ProjectID,
		task.AssigneeID,
		task.Title,
		task.Description,
		task.Completed,
		task.Priority,
		task.DueDate,
		task.Category,
		task.EstimatedHours,
		task.ActualHours,
		task.Progress,
		task.Status,
		task.ID,
		task.TenantID,
	)

	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *taskRepositoryImpl) DeleteTask(tenantID uuid.UUID, id uuid.UUID) error {
	query := `DELETE FROM godplan.tasks WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(query, id, tenantID)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *taskRepositoryImpl) GetUpcomingTasks(tenantID uuid.UUID, assigneeID uuid.UUID, limit int) ([]models.UpcomingTask, error) {
	query := `SELECT id, title, due_date, status, priority
		FROM godplan.tasks 
		WHERE assignee_id = $1 AND tenant_id = $2
		AND due_date >= CURRENT_DATE
		AND status NOT IN ('completed', 'cancelled')
		AND completed = false
		ORDER BY due_date ASC
		LIMIT $3`

	rows, err := r.db.Query(query, assigneeID, tenantID, limit)
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

func (r *taskRepositoryImpl) GetTaskCountByAssignee(tenantID uuid.UUID, assigneeID uuid.UUID) (int, int, error) {
	var totalTasks, completedTasks int
	query := `SELECT 
		COUNT(*) as total,
		COUNT(CASE WHEN completed = true OR status = 'completed' THEN 1 END) as completed
		FROM godplan.tasks 
		WHERE assignee_id = $1 AND tenant_id = $2`

	err := r.db.QueryRow(query, assigneeID, tenantID).Scan(&totalTasks, &completedTasks)
	if err != nil {
		return 0, 0, utils.ErrInternalServer
	}
	return totalTasks, completedTasks, nil
}

func (r *taskRepositoryImpl) GetPendingTasksCount(tenantID uuid.UUID, assigneeID uuid.UUID) (int, error) {
	var pendingTasks int
	query := `SELECT COUNT(*) 
		FROM godplan.tasks 
		WHERE assignee_id = $1 AND tenant_id = $2
		AND (completed = false AND status NOT IN ('completed', 'cancelled'))`

	err := r.db.QueryRow(query, assigneeID, tenantID).Scan(&pendingTasks)
	if err != nil {
		return 0, utils.ErrInternalServer
	}
	return pendingTasks, nil
}

// ValidateTaskAccess - Check if user has access to this task
func (r *taskRepositoryImpl) ValidateTaskAccess(tenantID uuid.UUID, taskID, assigneeID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM godplan.tasks WHERE id = $1 AND assignee_id = $2 AND tenant_id = $3`

	err := r.db.QueryRow(query, taskID, assigneeID, tenantID).Scan(&count)
	if err != nil {
		return false, utils.ErrInternalServer
	}

	return count > 0, nil
}

// UpdateTaskProgress - Update only task progress
func (r *taskRepositoryImpl) UpdateTaskProgress(tenantID uuid.UUID, taskID uuid.UUID, progress int) error {
	query := `UPDATE godplan.tasks 
		SET progress = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND tenant_id = $3`

	_, err := r.db.Exec(query, progress, taskID, tenantID)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

// CompleteTask - Mark task as completed
func (r *taskRepositoryImpl) CompleteTask(tenantID uuid.UUID, taskID uuid.UUID) error {
	query := `UPDATE godplan.tasks 
		SET status = 'completed', progress = 100, completed = true, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND tenant_id = $2`

	_, err := r.db.Exec(query, taskID, tenantID)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

// UpdateTaskCompletion - Update completed status
func (r *taskRepositoryImpl) UpdateTaskCompletion(tenantID uuid.UUID, taskID uuid.UUID, completed bool) error {
	query := `UPDATE godplan.tasks 
		SET completed = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND tenant_id = $3`

	result, err := r.db.Exec(query, completed, taskID, tenantID)
	if err != nil {
		return utils.ErrInternalServer
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrInternalServer
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

// UpdateTaskCategory - Update task category
func (r *taskRepositoryImpl) UpdateTaskCategory(tenantID uuid.UUID, taskID uuid.UUID, category string) error {
	query := `UPDATE godplan.tasks 
		SET category = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND tenant_id = $3`

	result, err := r.db.Exec(query, category, taskID, tenantID)
	if err != nil {
		return utils.ErrInternalServer
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrInternalServer
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

// GetTaskStatistics - Get task statistics for dashboard
func (r *taskRepositoryImpl) GetTaskStatistics(tenantID uuid.UUID, assigneeID uuid.UUID) (*models.TaskStatistics, error) {
	totalTasks, completedTasks, err := r.GetTaskCountByAssignee(tenantID, assigneeID)
	if err != nil {
		return nil, err
	}

	pendingTasks, err := r.GetPendingTasksCount(tenantID, assigneeID)
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

// GetTasksByCategory - Get tasks filtered by category
func (r *taskRepositoryImpl) GetTasksByCategory(tenantID uuid.UUID, assigneeID uuid.UUID, category string) ([]models.Task, error) {
	query := `SELECT id, tenant_id, project_id, assignee_id, title, description, completed, priority, due_date, category,
		 estimated_hours, actual_hours, progress, status, 
		 created_at, updated_at 
		 FROM godplan.tasks 
		 WHERE assignee_id = $1 AND category = $2 AND tenant_id = $3
		 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, assigneeID, category, tenantID)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.Priority,
			&task.DueDate,
			&task.Category,
			&task.EstimatedHours,
			&task.ActualHours,
			&task.Progress,
			&task.Status,
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

// GetCompletedTasks - Get completed tasks
func (r *taskRepositoryImpl) GetCompletedTasks(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error) {
	query := `SELECT id, tenant_id, project_id, assignee_id, title, description, completed, priority, due_date, category,
		 estimated_hours, actual_hours, progress, status, 
		 created_at, updated_at 
		 FROM godplan.tasks 
		 WHERE assignee_id = $1 AND (completed = true OR status = 'completed') AND tenant_id = $2
		 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, assigneeID, tenantID)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.Priority,
			&task.DueDate,
			&task.Category,
			&task.EstimatedHours,
			&task.ActualHours,
			&task.Progress,
			&task.Status,
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

// GetActiveTasks - Get active (not completed) tasks
func (r *taskRepositoryImpl) GetActiveTasks(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error) {
	query := `SELECT id, tenant_id, project_id, assignee_id, title, description, completed, priority, due_date, category,
		 estimated_hours, actual_hours, progress, status, 
		 created_at, updated_at 
		 FROM godplan.tasks 
		 WHERE assignee_id = $1 AND completed = false AND status != 'completed' AND tenant_id = $2
		 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, assigneeID, tenantID)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.Priority,
			&task.DueDate,
			&task.Category,
			&task.EstimatedHours,
			&task.ActualHours,
			&task.Progress,
			&task.Status,
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

package repository

import (
	"database/sql"

	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/utils"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(task *models.Task) error {
	query := `INSERT INTO tasks (title, description, user_id, assigned_to, status, completed, deadline) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`
	err := r.db.QueryRow(query, task.Title, task.Description, task.UserID, task.AssignedTo, task.Status, task.Completed, task.Deadline).Scan(&task.ID, &task.CreatedAt)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *TaskRepository) GetTasks() ([]models.Task, error) {
	query := `SELECT id, title, description, user_id, assigned_to, status, completed, deadline, created_at FROM tasks ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.UserID, &task.AssignedTo, &task.Status, &task.Completed, &task.Deadline, &task.CreatedAt)
		if err != nil {
			return nil, utils.ErrInternalServer
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepository) GetTaskByID(id int) (*models.Task, error) {
	task := &models.Task{}
	query := `SELECT id, title, description, user_id, assigned_to, status, completed, deadline, created_at FROM tasks WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&task.ID, &task.Title, &task.Description, &task.UserID, &task.AssignedTo, &task.Status, &task.Completed, &task.Deadline, &task.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, utils.ErrTaskNotFound
	}
	if err != nil {
		return nil, utils.ErrInternalServer
	}
	return task, nil
}

func (r *TaskRepository) UpdateTask(task *models.Task) error {
	query := `UPDATE tasks SET title = $1, description = $2, status = $3, completed = $4, deadline = $5 WHERE id = $6`
	_, err := r.db.Exec(query, task.Title, task.Description, task.Status, task.Completed, task.Deadline, task.ID)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

func (r *TaskRepository) DeleteTask(id int) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.ErrInternalServer
	}
	return nil
}

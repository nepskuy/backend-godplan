package service

import (
	"strings"

	"github.com/google/uuid"
	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

// TaskService interface
type TaskService interface {
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
	ToggleTaskCompletion(tenantID uuid.UUID, taskID uuid.UUID, completed bool) error
	UpdateTaskCategory(tenantID uuid.UUID, taskID uuid.UUID, category string) error
	GetTaskStatistics(tenantID uuid.UUID, assigneeID uuid.UUID) (*models.TaskStatistics, error)
	GetTasksByStatus(tenantID uuid.UUID, assigneeID uuid.UUID, status string) ([]models.Task, error)
	GetTasksByPriority(tenantID uuid.UUID, assigneeID uuid.UUID, priority string) ([]models.Task, error)
	GetTasksByCategory(tenantID uuid.UUID, assigneeID uuid.UUID, category string) ([]models.Task, error)
	GetCompletedTasks(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error)
	GetActiveTasks(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error)
	SearchTasks(tenantID uuid.UUID, assigneeID uuid.UUID, query string) ([]models.Task, error)
}

// taskServiceImpl implementasi konkret
type taskServiceImpl struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskServiceImpl{
		taskRepo: taskRepo,
	}
}

func (s *taskServiceImpl) CreateTask(task *models.Task) error {
	// Set default values
	if task.Status == "" {
		task.Status = "pending"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}
	if task.Progress == 0 {
		task.Progress = 0
	}
	if task.Category == "" {
		task.Category = "Personal"
	}
	if !task.Completed {
		task.Completed = false
	}

	return s.taskRepo.CreateTask(task)
}

func (s *taskServiceImpl) GetTasks(tenantID uuid.UUID) ([]models.Task, error) {
	return s.taskRepo.GetTasks(tenantID)
}

func (s *taskServiceImpl) GetTaskByID(tenantID uuid.UUID, id uuid.UUID) (*models.Task, error) {
	return s.taskRepo.GetTaskByID(tenantID, id)
}

func (s *taskServiceImpl) UpdateTask(task *models.Task) error {
	return s.taskRepo.UpdateTask(task)
}

func (s *taskServiceImpl) DeleteTask(tenantID uuid.UUID, id uuid.UUID) error {
	return s.taskRepo.DeleteTask(tenantID, id)
}

func (s *taskServiceImpl) GetTasksByAssignee(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error) {
	return s.taskRepo.GetTasksByAssignee(tenantID, assigneeID)
}

func (s *taskServiceImpl) GetUpcomingTasks(tenantID uuid.UUID, assigneeID uuid.UUID, limit int) ([]models.UpcomingTask, error) {
	return s.taskRepo.GetUpcomingTasks(tenantID, assigneeID, limit)
}

func (s *taskServiceImpl) GetTaskCountByAssignee(tenantID uuid.UUID, assigneeID uuid.UUID) (int, int, error) {
	return s.taskRepo.GetTaskCountByAssignee(tenantID, assigneeID)
}

func (s *taskServiceImpl) GetPendingTasksCount(tenantID uuid.UUID, assigneeID uuid.UUID) (int, error) {
	return s.taskRepo.GetPendingTasksCount(tenantID, assigneeID)
}

// ValidateTaskAccess - Check if user has access to this task
func (s *taskServiceImpl) ValidateTaskAccess(tenantID uuid.UUID, taskID, assigneeID uuid.UUID) (bool, error) {
	return s.taskRepo.ValidateTaskAccess(tenantID, taskID, assigneeID)
}

// UpdateTaskProgress - Update task progress with validation
func (s *taskServiceImpl) UpdateTaskProgress(tenantID uuid.UUID, taskID uuid.UUID, progress int) error {
	if progress < 0 || progress > 100 {
		return repository.ErrInvalidProgress
	}

	task, err := s.taskRepo.GetTaskByID(tenantID, taskID)
	if err != nil {
		return err
	}

	task.Progress = progress
	if progress == 100 {
		task.Status = "completed"
		task.Completed = true
	} else if progress > 0 {
		task.Status = "in_progress"
		task.Completed = false
	} else {
		task.Status = "pending"
		task.Completed = false
	}

	return s.taskRepo.UpdateTask(task)
}

// CompleteTask - Mark task as completed
func (s *taskServiceImpl) CompleteTask(tenantID uuid.UUID, taskID uuid.UUID) error {
	task, err := s.taskRepo.GetTaskByID(tenantID, taskID)
	if err != nil {
		return err
	}

	task.Status = "completed"
	task.Progress = 100
	task.Completed = true

	return s.taskRepo.UpdateTask(task)
}

// ToggleTaskCompletion - Toggle completed status
func (s *taskServiceImpl) ToggleTaskCompletion(tenantID uuid.UUID, taskID uuid.UUID, completed bool) error {
	task, err := s.taskRepo.GetTaskByID(tenantID, taskID)
	if err != nil {
		return err
	}

	task.Completed = completed
	if completed {
		task.Status = "completed"
		task.Progress = 100
	} else {
		task.Status = "pending"
		task.Progress = 0
	}

	return s.taskRepo.UpdateTask(task)
}

// UpdateTaskCategory - Update task category
func (s *taskServiceImpl) UpdateTaskCategory(tenantID uuid.UUID, taskID uuid.UUID, category string) error {
	task, err := s.taskRepo.GetTaskByID(tenantID, taskID)
	if err != nil {
		return err
	}

	task.Category = category
	return s.taskRepo.UpdateTask(task)
}

// GetTaskStatistics - Get task statistics for dashboard
func (s *taskServiceImpl) GetTaskStatistics(tenantID uuid.UUID, assigneeID uuid.UUID) (*models.TaskStatistics, error) {
	totalTasks, completedTasks, err := s.taskRepo.GetTaskCountByAssignee(tenantID, assigneeID)
	if err != nil {
		return nil, err
	}

	pendingTasks, err := s.taskRepo.GetPendingTasksCount(tenantID, assigneeID)
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

// GetTasksByStatus - Get tasks filtered by status
func (s *taskServiceImpl) GetTasksByStatus(tenantID uuid.UUID, assigneeID uuid.UUID, status string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(tenantID, assigneeID)
	if err != nil {
		return nil, err
	}

	var filteredTasks []models.Task
	for _, task := range allTasks {
		if task.Status == status {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks, nil
}

// GetTasksByPriority - Get tasks filtered by priority
func (s *taskServiceImpl) GetTasksByPriority(tenantID uuid.UUID, assigneeID uuid.UUID, priority string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(tenantID, assigneeID)
	if err != nil {
		return nil, err
	}

	var filteredTasks []models.Task
	for _, task := range allTasks {
		if task.Priority == priority {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks, nil
}

// GetTasksByCategory - Get tasks filtered by category
func (s *taskServiceImpl) GetTasksByCategory(tenantID uuid.UUID, assigneeID uuid.UUID, category string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(tenantID, assigneeID)
	if err != nil {
		return nil, err
	}

	var filteredTasks []models.Task
	for _, task := range allTasks {
		if task.Category == category {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks, nil
}

// GetCompletedTasks - Get completed tasks
func (s *taskServiceImpl) GetCompletedTasks(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(tenantID, assigneeID)
	if err != nil {
		return nil, err
	}

	var filteredTasks []models.Task
	for _, task := range allTasks {
		if task.Completed {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks, nil
}

// GetActiveTasks - Get active (not completed) tasks
func (s *taskServiceImpl) GetActiveTasks(tenantID uuid.UUID, assigneeID uuid.UUID) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(tenantID, assigneeID)
	if err != nil {
		return nil, err
	}

	var filteredTasks []models.Task
	for _, task := range allTasks {
		if !task.Completed {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks, nil
}

// SearchTasks - Search tasks by title or description
func (s *taskServiceImpl) SearchTasks(tenantID uuid.UUID, assigneeID uuid.UUID, query string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(tenantID, assigneeID)
	if err != nil {
		return nil, err
	}

	var filteredTasks []models.Task
	for _, task := range allTasks {
		if containsIgnoreCase(task.Title, query) || containsIgnoreCase(task.Description, query) {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks, nil
}

// Helper function for case-insensitive search
func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}

	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}

package service

import (
	"strings"

	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

// TaskService interface
type TaskService interface {
	CreateTask(task *models.Task) error
	GetTasks() ([]models.Task, error)
	GetTaskByID(id string) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
	GetTasksByAssignee(assigneeID string) ([]models.Task, error)
	GetUpcomingTasks(assigneeID string, limit int) ([]models.UpcomingTask, error)
	GetTaskCountByAssignee(assigneeID string) (int, int, error)
	GetPendingTasksCount(assigneeID string) (int, error)
	ValidateTaskAccess(taskID string, assigneeID string) (bool, error)
	UpdateTaskProgress(taskID string, progress int) error
	CompleteTask(taskID string) error
	ToggleTaskCompletion(taskID string, completed bool) error
	UpdateTaskCategory(taskID string, category string) error
	GetTaskStatistics(assigneeID string) (*models.TaskStatistics, error)
	GetTasksByStatus(assigneeID string, status string) ([]models.Task, error)
	GetTasksByPriority(assigneeID string, priority string) ([]models.Task, error)
	GetTasksByCategory(assigneeID string, category string) ([]models.Task, error)
	GetCompletedTasks(assigneeID string) ([]models.Task, error)
	GetActiveTasks(assigneeID string) ([]models.Task, error)
	SearchTasks(assigneeID string, query string) ([]models.Task, error)
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

func (s *taskServiceImpl) GetTasks() ([]models.Task, error) {
	return s.taskRepo.GetTasks()
}

func (s *taskServiceImpl) GetTaskByID(id string) (*models.Task, error) {
	return s.taskRepo.GetTaskByID(id)
}

func (s *taskServiceImpl) UpdateTask(task *models.Task) error {
	return s.taskRepo.UpdateTask(task)
}

func (s *taskServiceImpl) DeleteTask(id string) error {
	return s.taskRepo.DeleteTask(id)
}

func (s *taskServiceImpl) GetTasksByAssignee(assigneeID string) ([]models.Task, error) {
	return s.taskRepo.GetTasksByAssignee(assigneeID)
}

func (s *taskServiceImpl) GetUpcomingTasks(assigneeID string, limit int) ([]models.UpcomingTask, error) {
	return s.taskRepo.GetUpcomingTasks(assigneeID, limit)
}

func (s *taskServiceImpl) GetTaskCountByAssignee(assigneeID string) (int, int, error) {
	return s.taskRepo.GetTaskCountByAssignee(assigneeID)
}

func (s *taskServiceImpl) GetPendingTasksCount(assigneeID string) (int, error) {
	return s.taskRepo.GetPendingTasksCount(assigneeID)
}

// ValidateTaskAccess - Check if user has access to this task
func (s *taskServiceImpl) ValidateTaskAccess(taskID string, assigneeID string) (bool, error) {
	return s.taskRepo.ValidateTaskAccess(taskID, assigneeID)
}

// UpdateTaskProgress - Update task progress with validation
func (s *taskServiceImpl) UpdateTaskProgress(taskID string, progress int) error {
	if progress < 0 || progress > 100 {
		return repository.ErrInvalidProgress
	}

	task, err := s.taskRepo.GetTaskByID(taskID)
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
func (s *taskServiceImpl) CompleteTask(taskID string) error {
	task, err := s.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	task.Status = "completed"
	task.Progress = 100
	task.Completed = true

	return s.taskRepo.UpdateTask(task)
}

// ToggleTaskCompletion - Toggle completed status
func (s *taskServiceImpl) ToggleTaskCompletion(taskID string, completed bool) error {
	task, err := s.taskRepo.GetTaskByID(taskID)
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
func (s *taskServiceImpl) UpdateTaskCategory(taskID string, category string) error {
	task, err := s.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	task.Category = category
	return s.taskRepo.UpdateTask(task)
}

// GetTaskStatistics - Get task statistics for dashboard
func (s *taskServiceImpl) GetTaskStatistics(assigneeID string) (*models.TaskStatistics, error) {
	totalTasks, completedTasks, err := s.taskRepo.GetTaskCountByAssignee(assigneeID)
	if err != nil {
		return nil, err
	}

	pendingTasks, err := s.taskRepo.GetPendingTasksCount(assigneeID)
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
func (s *taskServiceImpl) GetTasksByStatus(assigneeID string, status string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(assigneeID)
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
func (s *taskServiceImpl) GetTasksByPriority(assigneeID string, priority string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(assigneeID)
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
func (s *taskServiceImpl) GetTasksByCategory(assigneeID string, category string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(assigneeID)
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
func (s *taskServiceImpl) GetCompletedTasks(assigneeID string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(assigneeID)
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
func (s *taskServiceImpl) GetActiveTasks(assigneeID string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(assigneeID)
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
func (s *taskServiceImpl) SearchTasks(assigneeID string, query string) ([]models.Task, error) {
	allTasks, err := s.taskRepo.GetTasksByAssignee(assigneeID)
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

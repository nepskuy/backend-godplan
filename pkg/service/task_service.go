package service

import (
	"strings"

	"github.com/nepskuy/be-godplan/pkg/models"
	"github.com/nepskuy/be-godplan/pkg/repository"
)

type TaskService struct {
	taskRepo *repository.TaskRepository
}

func NewTaskService(taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{
		taskRepo: taskRepo,
	}
}

func (s *TaskService) CreateTask(task *models.Task) error {
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

	return s.taskRepo.CreateTask(task)
}

func (s *TaskService) GetTasks() ([]models.Task, error) {
	return s.taskRepo.GetTasks()
}

func (s *TaskService) GetTaskByID(id string) (*models.Task, error) {
	return s.taskRepo.GetTaskByID(id)
}

func (s *TaskService) UpdateTask(task *models.Task) error {
	return s.taskRepo.UpdateTask(task)
}

func (s *TaskService) DeleteTask(id string) error {
	return s.taskRepo.DeleteTask(id)
}

func (s *TaskService) GetTasksByAssignee(assigneeID string) ([]models.Task, error) {
	return s.taskRepo.GetTasksByAssignee(assigneeID)
}

func (s *TaskService) GetUpcomingTasks(assigneeID string, limit int) ([]models.UpcomingTask, error) {
	return s.taskRepo.GetUpcomingTasks(assigneeID, limit)
}

func (s *TaskService) GetTaskCountByAssignee(assigneeID string) (int, int, error) {
	return s.taskRepo.GetTaskCountByAssignee(assigneeID)
}

func (s *TaskService) GetPendingTasksCount(assigneeID string) (int, error) {
	return s.taskRepo.GetPendingTasksCount(assigneeID)
}

// ValidateTaskAccess - Check if user has access to this task
func (s *TaskService) ValidateTaskAccess(taskID string, assigneeID string) (bool, error) {
	return s.taskRepo.ValidateTaskAccess(taskID, assigneeID)
}

// UpdateTaskProgress - Update task progress with validation
func (s *TaskService) UpdateTaskProgress(taskID string, progress int) error {
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
	} else if progress > 0 {
		task.Status = "in_progress"
	}

	return s.taskRepo.UpdateTask(task)
}

// CompleteTask - Mark task as completed
func (s *TaskService) CompleteTask(taskID string) error {
	task, err := s.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	task.Status = "completed"
	task.Progress = 100

	return s.taskRepo.UpdateTask(task)
}

// GetTaskStatistics - Get task statistics for dashboard
func (s *TaskService) GetTaskStatistics(assigneeID string) (*models.TaskStatistics, error) {
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
func (s *TaskService) GetTasksByStatus(assigneeID string, status string) ([]models.Task, error) {
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
func (s *TaskService) GetTasksByPriority(assigneeID string, priority string) ([]models.Task, error) {
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

// SearchTasks - Search tasks by title or description
func (s *TaskService) SearchTasks(assigneeID string, query string) ([]models.Task, error) {
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

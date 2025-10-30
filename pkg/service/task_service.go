package service

import (
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

	return s.taskRepo.CreateTask(task)
}

func (s *TaskService) GetTasks() ([]models.Task, error) {
	return s.taskRepo.GetTasks()
}

func (s *TaskService) GetTaskByID(id int) (*models.Task, error) {
	return s.taskRepo.GetTaskByID(id)
}

func (s *TaskService) UpdateTask(task *models.Task) error {
	return s.taskRepo.UpdateTask(task)
}

func (s *TaskService) DeleteTask(id int) error {
	return s.taskRepo.DeleteTask(id)
}

func (s *TaskService) GetUserTasks(userID int) ([]models.Task, error) {
	// TODO: Implement specific user tasks filtering
	allTasks, err := s.taskRepo.GetTasks()
	if err != nil {
		return nil, err
	}

	// Filter tasks by user ID
	var userTasks []models.Task
	for _, task := range allTasks {
		if task.UserID == userID || task.AssignedTo == userID {
			userTasks = append(userTasks, task)
		}
	}

	return userTasks, nil
}

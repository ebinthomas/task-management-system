package service

import (
	"context"
	"errors"

	"sample/task-management-system/pkg/models"
	"sample/task-management-system/pkg/repository"
)

// TaskService handles task-related business logic
type TaskService interface {
	CreateTask(ctx context.Context, task *models.TaskCreate) (*models.Task, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	UpdateTask(ctx context.Context, id string, task *models.TaskUpdate) (*models.Task, error)
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, status models.TaskStatus, page, limit int) ([]*models.Task, int, error)
}

type taskService struct {
	repo repository.TaskRepository
}

// NewTaskService creates a new task service
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, task *models.TaskCreate) (*models.Task, error) {
	if err := task.Validate(); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, task)
}

func (s *taskService) GetTask(ctx context.Context, id string) (*models.Task, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	return s.repo.GetByID(ctx, id)
}

func (s *taskService) UpdateTask(ctx context.Context, id string, task *models.TaskUpdate) (*models.Task, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	if err := task.Validate(); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, id, task)
}

func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	return s.repo.Delete(ctx, id)
}

func (s *taskService) ListTasks(ctx context.Context, status models.TaskStatus, page, limit int) ([]*models.Task, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	filter := repository.TaskFilter{
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	return s.repo.List(ctx, filter)
} 
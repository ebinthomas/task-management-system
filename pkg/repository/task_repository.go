package repository

import (
	"context"

	"sample/task-management-system/pkg/models"
)

// TaskFilter represents the filtering options for tasks
type TaskFilter struct {
	Status models.TaskStatus
	Page   int
	Limit  int
}

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	// Create creates a new task
	Create(ctx context.Context, task *models.TaskCreate) (*models.Task, error)

	// GetByID retrieves a task by its ID
	GetByID(ctx context.Context, id string) (*models.Task, error)

	// Update updates an existing task
	Update(ctx context.Context, id string, task *models.TaskUpdate) (*models.Task, error)

	// Delete removes a task by its ID
	Delete(ctx context.Context, id string) error

	// List retrieves tasks with pagination and filtering
	List(ctx context.Context, filter TaskFilter) ([]*models.Task, int, error)
} 
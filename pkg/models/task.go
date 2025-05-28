package models

import (
	"errors"
	"time"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted TaskStatus = "completed"
	StatusCancelled TaskStatus = "cancelled"
)

// Task represents a task in the system
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	DueDate     time.Time  `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TaskCreate represents the data required to create a new task
type TaskCreate struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	DueDate     time.Time  `json:"due_date"`
}

// TaskUpdate represents the data that can be updated for a task
type TaskUpdate struct {
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Status      *TaskStatus `json:"status,omitempty"`
	DueDate     *time.Time  `json:"due_date,omitempty"`
}

// Validate checks if the task create request is valid
func (t *TaskCreate) Validate() error {
	if t.Title == "" {
		return errors.New("title is required")
	}
	if t.Status == "" {
		t.Status = StatusPending
	}
	if !isValidStatus(t.Status) {
		return errors.New("invalid status")
	}
	if t.DueDate.IsZero() {
		return errors.New("due date is required")
	}
	if t.DueDate.Before(time.Now()) {
		return errors.New("due date must be in the future")
	}
	return nil
}

// Validate checks if the task update request is valid
func (t *TaskUpdate) Validate() error {
	if t.Status != nil && !isValidStatus(*t.Status) {
		return errors.New("invalid status")
	}
	if t.DueDate != nil && t.DueDate.Before(time.Now()) {
		return errors.New("due date must be in the future")
	}
	return nil
}

// isValidStatus checks if the given status is valid
func isValidStatus(status TaskStatus) bool {
	switch status {
	case StatusPending, StatusInProgress, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
} 
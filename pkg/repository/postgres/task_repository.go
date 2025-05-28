package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"sample/task-management-system/pkg/models"
	"sample/task-management-system/pkg/repository"
)

type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new PostgreSQL task repository
func NewTaskRepository(db *sql.DB) repository.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.TaskCreate) (*models.Task, error) {
	query := `
		INSERT INTO tasks (id, title, description, status, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, due_date, created_at, updated_at`

	now := time.Now()
	id := uuid.New().String()

	result := &models.Task{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
		task.Title,
		task.Description,
		task.Status,
		task.DueDate,
		now,
		now,
	).Scan(
		&result.ID,
		&result.Title,
		&result.Description,
		&result.Status,
		&result.DueDate,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, title, description, status, due_date, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	task := &models.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("task not found")
	}
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *taskRepository) Update(ctx context.Context, id string, task *models.TaskUpdate) (*models.Task, error) {
	query := `
		UPDATE tasks
		SET title = COALESCE($1, title),
			description = COALESCE($2, description),
			status = COALESCE($3, status),
			due_date = COALESCE($4, due_date),
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, due_date, created_at, updated_at`

	var title, description *string
	var status *models.TaskStatus
	var dueDate *time.Time

	if task.Title != nil {
		title = task.Title
	}
	if task.Description != nil {
		description = task.Description
	}
	if task.Status != nil {
		status = task.Status
	}
	if task.DueDate != nil {
		dueDate = task.DueDate
	}

	result := &models.Task{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		title,
		description,
		status,
		dueDate,
		time.Now(),
		id,
	).Scan(
		&result.ID,
		&result.Title,
		&result.Description,
		&result.Status,
		&result.DueDate,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("task not found")
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("task not found")
	}

	return nil
}

func (r *taskRepository) List(ctx context.Context, filter repository.TaskFilter) ([]*models.Task, int, error) {
	// First, get total count
	countQuery := `SELECT COUNT(*) FROM tasks`
	var params []interface{}
	var whereClause string

	paramCount := 1
	if filter.Status != "" {
		whereClause = fmt.Sprintf(" WHERE status = $%d", paramCount)
		params = append(params, filter.Status)
		paramCount++
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery+whereClause, params...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Then get paginated results
	query := `
		SELECT id, title, description, status, due_date, created_at, updated_at
		FROM tasks`

	if whereClause != "" {
		query += whereClause
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramCount, paramCount+1)
	params = append(params, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
} 
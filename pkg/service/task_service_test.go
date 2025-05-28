package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sample/task-management-system/pkg/models"
	"sample/task-management-system/pkg/repository"
)

// MockTaskRepository is a mock implementation of TaskRepository
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *models.TaskCreate) (*models.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, id string, task *models.TaskUpdate) (*models.Task, error) {
	args := m.Called(ctx, id, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) List(ctx context.Context, filter repository.TaskFilter) ([]*models.Task, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Task), args.Int(1), args.Error(2)
}

func TestCreateTask(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *models.TaskCreate
		mock    func()
		want    *models.Task
		wantErr bool
	}{
		{
			name: "successful creation",
			input: &models.TaskCreate{
				Title:       "Test Task",
				Description: "Test Description",
				Status:     models.StatusPending,
				DueDate:    time.Now().Add(24 * time.Hour),
			},
			mock: func() {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.TaskCreate")).
					Return(&models.Task{
						ID:          "test-id",
						Title:       "Test Task",
						Description: "Test Description",
						Status:     models.StatusPending,
					}, nil)
			},
			want: &models.Task{
				ID:          "test-id",
				Title:       "Test Task",
				Description: "Test Description",
				Status:     models.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "validation error - empty title",
			input: &models.TaskCreate{
				Title:    "",
				DueDate: time.Now().Add(24 * time.Hour),
			},
			mock:    func() {},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := service.CreateTask(ctx, tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.Status, got.Status)
		})
	}
} 
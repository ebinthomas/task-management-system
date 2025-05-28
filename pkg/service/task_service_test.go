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

func TestGetTask(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      string
		mock    func()
		want    *models.Task
		wantErr bool
	}{
		{
			name: "successful retrieval",
			id:   "test-id",
			mock: func() {
				mockRepo.On("GetByID", mock.Anything, "test-id").
					Return(&models.Task{
						ID:          "test-id",
						Title:       "Test Task",
						Description: "Test Description",
						Status:      models.StatusPending,
					}, nil)
			},
			want: &models.Task{
				ID:          "test-id",
				Title:       "Test Task",
				Description: "Test Description",
				Status:      models.StatusPending,
			},
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			mock:    func() {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "not found",
			id:   "non-existent",
			mock: func() {
				mockRepo.On("GetByID", mock.Anything, "non-existent").
					Return(nil, errors.New("task not found"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := service.GetTask(ctx, tt.id)
			
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

func TestUpdateTask(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	newTitle := "Updated Title"
	newStatus := models.StatusInProgress

	tests := []struct {
		name    string
		id      string
		update  *models.TaskUpdate
		mock    func()
		want    *models.Task
		wantErr bool
	}{
		{
			name: "successful update",
			id:   "test-id",
			update: &models.TaskUpdate{
				Title:  &newTitle,
				Status: &newStatus,
			},
			mock: func() {
				mockRepo.On("Update", mock.Anything, "test-id", mock.MatchedBy(func(update *models.TaskUpdate) bool {
					return *update.Title == newTitle && *update.Status == newStatus
				})).Return(&models.Task{
					ID:     "test-id",
					Title:  newTitle,
					Status: newStatus,
				}, nil)
			},
			want: &models.Task{
				ID:     "test-id",
				Title:  newTitle,
				Status: newStatus,
			},
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			update:  &models.TaskUpdate{},
			mock:    func() {},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := service.UpdateTask(ctx, tt.id, tt.update)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Status, got.Status)
		})
	}
}

func TestDeleteTask(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   "test-id",
			mock: func() {
				mockRepo.On("Delete", mock.Anything, "test-id").Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "not found",
			id:   "non-existent",
			mock: func() {
				mockRepo.On("Delete", mock.Anything, "non-existent").
					Return(errors.New("task not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := service.DeleteTask(ctx, tt.id)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListTasks(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name       string
		status     models.TaskStatus
		page       int
		limit      int
		mock       func()
		wantTasks  []*models.Task
		wantTotal  int
		wantErr    bool
	}{
		{
			name:   "successful list",
			status: models.StatusPending,
			page:   1,
			limit:  10,
			mock: func() {
				tasks := []*models.Task{
					{
						ID:     "task-1",
						Title:  "Task 1",
						Status: models.StatusPending,
					},
					{
						ID:     "task-2",
						Title:  "Task 2",
						Status: models.StatusPending,
					},
				}
				mockRepo.On("List", mock.Anything, repository.TaskFilter{
					Status: models.StatusPending,
					Page:   1,
					Limit:  10,
				}).Return(tasks, 2, nil)
			},
			wantTasks: []*models.Task{
				{
					ID:     "task-1",
					Title:  "Task 1",
					Status: models.StatusPending,
				},
				{
					ID:     "task-2",
					Title:  "Task 2",
					Status: models.StatusPending,
				},
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:   "invalid page",
			status: models.StatusPending,
			page:   0,
			limit:  10,
			mock: func() {
				tasks := []*models.Task{
					{
						ID:     "task-1",
						Title:  "Task 1",
						Status: models.StatusPending,
					},
					{
						ID:     "task-2",
						Title:  "Task 2",
						Status: models.StatusPending,
					},
				}
				mockRepo.On("List", mock.Anything, repository.TaskFilter{
					Status: models.StatusPending,
					Page:   1,
					Limit:  10,
				}).Return(tasks, 2, nil)
			},
			wantTasks: []*models.Task{
				{
					ID:     "task-1",
					Title:  "Task 1",
					Status: models.StatusPending,
				},
				{
					ID:     "task-2",
					Title:  "Task 2",
					Status: models.StatusPending,
				},
			},
			wantTotal: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			tasks, total, err := service.ListTasks(ctx, tt.status, tt.page, tt.limit)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.Equal(t, len(tt.wantTasks), len(tasks))
			assert.Equal(t, tt.wantTotal, total)
			
			for i := range tasks {
				if i < len(tt.wantTasks) {
					assert.Equal(t, tt.wantTasks[i].ID, tasks[i].ID)
					assert.Equal(t, tt.wantTasks[i].Title, tasks[i].Title)
					assert.Equal(t, tt.wantTasks[i].Status, tasks[i].Status)
				}
			}
		})
	}
} 
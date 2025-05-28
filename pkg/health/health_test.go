package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sample/task-management-system/pkg/monitoring"
)

// MockServiceMonitor is a mock implementation of the service monitor
type MockServiceMonitor struct {
	mock.Mock
}

func (m *MockServiceMonitor) UpdateServiceState(state monitoring.ServiceState) error {
	args := m.Called(state)
	return args.Error(0)
}

func TestHealthHandler_ServeHTTP(t *testing.T) {
	// Setup mock Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Setup mock service monitor
	mockMonitor := new(MockServiceMonitor)
	mockMonitor.On("UpdateServiceState", mock.AnythingOfType("monitoring.ServiceState")).Return(nil)

	tests := []struct {
		name           string
		db            *sql.DB
		cache         *redis.Client
		monitor       *MockServiceMonitor
		expectedCode  int
		expectedState Status
	}{
		{
			name:           "All Services Up",
			cache:         redisClient,
			monitor:       mockMonitor,
			expectedCode:  http.StatusOK,
			expectedState: StatusUp,
		},
		{
			name:           "No Redis",
			cache:         nil,
			monitor:       mockMonitor,
			expectedCode:  http.StatusServiceUnavailable,
			expectedState: StatusDown,
		},
		{
			name:           "No Monitor",
			cache:         redisClient,
			monitor:       nil,
			expectedCode:  http.StatusOK,
			expectedState: StatusUp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler("1.0.0", tt.db, tt.cache, tt.monitor)
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var response HealthResponse
			err := json.NewDecoder(rr.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedState, response.Status)
			assert.Equal(t, "1.0.0", response.Version)
			assert.NotEmpty(t, response.System.GoVersion)
			assert.NotZero(t, response.System.NumCPU)

			if tt.monitor != nil {
				tt.monitor.AssertExpectations(t)
			}
		})
	}
}

func TestHealthHandler_CheckComponents(t *testing.T) {
	// Setup mock Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Setup mock service monitor
	mockMonitor := new(MockServiceMonitor)
	mockMonitor.On("UpdateServiceState", mock.AnythingOfType("monitoring.ServiceState")).Return(nil)

	handler := NewHandler("1.0.0", nil, redisClient, mockMonitor)
	ctx := context.Background()

	// Test Redis check
	redisStatus := handler.checkRedis(ctx)
	assert.Equal(t, StatusUp, redisStatus.Status)
	assert.Contains(t, redisStatus.Message, "successful")

	// Test Redis failure
	mr.Close()
	redisStatus = handler.checkRedis(ctx)
	assert.Equal(t, StatusDown, redisStatus.Status)
	assert.Contains(t, redisStatus.Message, "Failed")

	// Test DB check with nil connection
	dbStatus := handler.checkDatabase(ctx)
	assert.Equal(t, StatusDown, dbStatus.Status)
	assert.Contains(t, dbStatus.Message, "not configured")

	mockMonitor.AssertExpectations(t)
}

func TestHealthResponse_SystemInfo(t *testing.T) {
	// Setup mock service monitor
	mockMonitor := new(MockServiceMonitor)
	mockMonitor.On("UpdateServiceState", mock.AnythingOfType("monitoring.ServiceState")).Return(nil)

	handler := NewHandler("1.0.0", nil, nil, mockMonitor)
	response := handler.checkHealth(context.Background())

	assert.NotEmpty(t, response.System.GoVersion)
	assert.NotZero(t, response.System.NumCPU)
	assert.NotZero(t, response.System.NumGoroutine)
	assert.NotZero(t, response.System.HeapInUse)
	assert.NotEmpty(t, response.Timestamp)

	mockMonitor.AssertExpectations(t)
} 
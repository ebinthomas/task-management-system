package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sample/task-management-system/pkg/monitoring"
)

// MockRedisCache is a mock implementation of RedisCache
type MockRedisCache struct {
	mock.Mock
}

func (m *MockRedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockRedisCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRedisCache) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRedisCache) Ping(ctx context.Context) error {
	args := m.Called()
	return args.Error(0)
}

// MockServiceMonitor is a mock implementation of ServiceMonitor
type MockServiceMonitor struct {
	mock.Mock
}

func (m *MockServiceMonitor) UpdateServiceState(state monitoring.ServiceState) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockServiceMonitor) IsAlarmsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockServiceMonitor) Start(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockServiceMonitor) Stop() {
	m.Called()
}

func (m *MockServiceMonitor) CreateServiceAlarm(ctx context.Context, serviceName, alarmName string, threshold float64, operator monitoring.ComparisonOperator) error {
	args := m.Called(ctx, serviceName, alarmName, threshold, operator)
	return args.Error(0)
}

// redisWrapper implements the Ping interface for testing
type redisWrapper struct {
	mr *miniredis.Miniredis
	closed bool
}

func (r *redisWrapper) Ping(ctx context.Context) error {
	if r.closed {
		return errors.New("redis is not connected")
	}
	return nil
}

func TestHealthHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		db            *sql.DB
		cache         *MockRedisCache
		monitor       *MockServiceMonitor
		expectedCode  int
		expectedState string
	}{
		{
			name:           "All Systems Operational",
			version:        "1.0.0",
			db:            nil, // No DB connection is OK for tests
			cache:         &MockRedisCache{},
			monitor:       &MockServiceMonitor{},
			expectedCode:  http.StatusOK,
			expectedState: "UP",
		},
		{
			name:           "Cache Down",
			version:        "1.0.0",
			db:            nil, // No DB connection is OK for tests
			cache:         &MockRedisCache{},
			monitor:       &MockServiceMonitor{},
			expectedCode:  http.StatusServiceUnavailable,
			expectedState: "DOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedState == "UP" {
				tt.cache.On("Ping").Return(nil)
			} else {
				tt.cache.On("Ping").Return(errors.New("connection failed"))
			}

			// Mock all possible service state updates
			tt.monitor.On("UpdateServiceState", mock.MatchedBy(func(state monitoring.ServiceState) bool {
				return true // Accept any service state update
			})).Return(nil)

			handler := NewHandler(tt.version, tt.db, tt.cache, tt.monitor)
			req := httptest.NewRequest("GET", "/health", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			var response map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedState, response["status"])
			assert.Equal(t, tt.version, response["version"])

			tt.cache.AssertExpectations(t)
			tt.monitor.AssertExpectations(t)
		})
	}
}

func TestHealthHandler_CacheTimeout(t *testing.T) {
	mockCache := &MockRedisCache{}
	mockMonitor := &MockServiceMonitor{}
	version := "1.0.0"

	// Create a channel to control when Ping returns
	done := make(chan struct{})
	defer close(done)

	// Simulate a cache timeout that blocks until we signal it
	mockCache.On("Ping", mock.Anything).Return(errors.New("timeout"))

	// Mock all possible service state updates
	mockMonitor.On("UpdateServiceState", mock.MatchedBy(func(state monitoring.ServiceState) bool {
		return true // Accept any service state update
	})).Return(nil)

	handler := NewHandler(version, nil, mockCache, mockMonitor)
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "DOWN", response["status"])
	assert.Equal(t, version, response["version"])

	mockCache.AssertExpectations(t)
	mockMonitor.AssertExpectations(t)
}

func TestHealthHandler_MonitoringDisabled(t *testing.T) {
	mockCache := &MockRedisCache{}
	mockMonitor := &MockServiceMonitor{}
	version := "1.0.0"

	mockCache.On("Ping", mock.Anything).Return(nil)

	// Mock all possible service state updates
	mockMonitor.On("UpdateServiceState", mock.MatchedBy(func(state monitoring.ServiceState) bool {
		return true // Accept any service state update
	})).Return(nil)

	handler := NewHandler(version, nil, mockCache, mockMonitor)
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "UP", response["status"])
	assert.Equal(t, version, response["version"])

	mockCache.AssertExpectations(t)
	mockMonitor.AssertExpectations(t)
}

func TestHealthHandler_CheckComponents(t *testing.T) {
	// Setup mock Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	redisClient := &redisWrapper{mr: mr}

	// Setup mock service monitor
	mockMonitor := new(MockServiceMonitor)
	mockMonitor.On("UpdateServiceState", mock.AnythingOfType("monitoring.ServiceState")).Return(nil).Times(6) // 2 calls to checkHealth, 3 updates each (cache, db, system)

	handler := NewHandler("1.0.0", nil, redisClient, mockMonitor)
	ctx := context.Background()

	// Test initial state
	response := handler.checkHealth(ctx)
	assert.Equal(t, StatusUp, response.Services["cache"].Status)
	assert.Contains(t, response.Services["cache"].Message, "successful")
	assert.Equal(t, StatusDown, response.Services["database"].Status)
	assert.Contains(t, response.Services["database"].Message, "not configured")

	// Test Redis failure
	mr.Close()
	redisClient.closed = true
	response = handler.checkHealth(ctx)
	assert.Equal(t, StatusDown, response.Services["cache"].Status)
	assert.Contains(t, response.Services["cache"].Message, "Failed")
	assert.Equal(t, StatusDown, response.Services["database"].Status)
	assert.Contains(t, response.Services["database"].Message, "not configured")

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
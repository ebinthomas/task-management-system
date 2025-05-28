package monitoring

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCloudWatchClient is a mock implementation of CloudWatchClient
type MockCloudWatchClient struct {
	mock.Mock
}

func (m *MockCloudWatchClient) PutMetricData(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudwatch.PutMetricDataOutput), args.Error(1)
}

// MockAlarmService is a mock implementation of AlarmService
type MockAlarmService struct {
	mock.Mock
}

func (m *MockAlarmService) CreateAlarm(ctx context.Context, alarm Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *MockAlarmService) UpdateAlarm(ctx context.Context, alarm Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *MockAlarmService) DeleteAlarm(ctx context.Context, alarmName string) error {
	args := m.Called(ctx, alarmName)
	return args.Error(0)
}

func (m *MockAlarmService) GetAlarmState(ctx context.Context, alarmName string) (AlarmState, error) {
	args := m.Called(ctx, alarmName)
	return args.Get(0).(AlarmState), args.Error(1)
}

func (m *MockAlarmService) IsAlarmsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestServiceMonitor_UpdateServiceState(t *testing.T) {
	mockClient := &MockCloudWatchClient{}
	mockAlarmService := &MockAlarmService{}
	monitor := NewServiceMonitor(mockClient, mockAlarmService, "TestNamespace", time.Minute)

	state := ServiceState{
		Name:      "TestService",
		Status:    "UP",
		Message:   "Service is healthy",
		Timestamp: time.Now(),
		Metrics: map[string]float64{
			"CPUUsage": 75.5,
			"Memory":   80.0,
		},
	}

	// Set environment variable for metrics
	os.Setenv("ENABLE_METRICS", "true")
	defer os.Unsetenv("ENABLE_METRICS")

	// Expect PutMetricData calls for status and each metric
	mockClient.On("PutMetricData", mock.Anything, mock.MatchedBy(func(input *cloudwatch.PutMetricDataInput) bool {
		return *input.Namespace == "TestNamespace"
	})).Return(&cloudwatch.PutMetricDataOutput{}, nil).Times(3) // Once for status, twice for metrics

	// We don't actually need IsAlarmsEnabled for UpdateServiceState
	mockAlarmService.On("IsAlarmsEnabled").Return(true).Maybe()

	err := monitor.UpdateServiceState(state)
	assert.NoError(t, err)

	// Verify the state was stored
	storedState, exists := monitor.states[state.Name]
	assert.True(t, exists)
	assert.Equal(t, state.Status, storedState.Status)
	assert.Equal(t, state.Message, storedState.Message)

	mockClient.AssertExpectations(t)
	mockAlarmService.AssertExpectations(t)
}

func TestServiceMonitor_CreateServiceAlarm(t *testing.T) {
	mockClient := &MockCloudWatchClient{}
	mockAlarmService := &MockAlarmService{}
	monitor := NewServiceMonitor(mockClient, mockAlarmService, "TestNamespace", time.Minute)

	// Expect IsAlarmsEnabled check
	mockAlarmService.On("IsAlarmsEnabled").Return(true)

	// Expect CreateAlarm call
	mockAlarmService.On("CreateAlarm", mock.Anything, mock.MatchedBy(func(alarm Alarm) bool {
		return alarm.Name == "TestAlarm" && alarm.Namespace == "TestNamespace"
	})).Return(nil)

	err := monitor.CreateServiceAlarm(context.Background(), "TestService", "TestAlarm", 0.5, LessThanThreshold)
	assert.NoError(t, err)

	mockAlarmService.AssertExpectations(t)
}

func TestServiceMonitor_GetStatusValue(t *testing.T) {
	mockClient := &MockCloudWatchClient{}
	mockAlarmService := &MockAlarmService{}
	monitor := NewServiceMonitor(mockClient, mockAlarmService, "TestNamespace", time.Minute)

	tests := []struct {
		status string
		want   float64
	}{
		{"UP", 1.0},
		{"DOWN", 0.0},
		{"DEGRADED", 0.5},
		{"UNKNOWN", -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := monitor.getStatusValue(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServiceMonitor_CheckAndUpdateStates(t *testing.T) {
	mockClient := &MockCloudWatchClient{}
	mockAlarmService := &MockAlarmService{}
	monitor := NewServiceMonitor(mockClient, mockAlarmService, "TestNamespace", time.Minute)

	// Add a stale state
	staleState := &ServiceState{
		Name:      "StaleService",
		Status:    "UP",
		Timestamp: time.Now().Add(-3 * time.Minute),
	}
	monitor.states["StaleService"] = staleState

	// Expect IsAlarmsEnabled check first
	mockAlarmService.On("IsAlarmsEnabled").Return(true)

	// Expect CreateAlarm call for stale service
	mockAlarmService.On("CreateAlarm", mock.Anything, mock.MatchedBy(func(alarm Alarm) bool {
		return alarm.Name == "StaleService-StaleState"
	})).Return(nil)

	monitor.checkAndUpdateStates(context.Background())

	mockAlarmService.AssertExpectations(t)
} 
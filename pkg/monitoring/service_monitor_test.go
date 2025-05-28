package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCloudWatchClient is a mock implementation of CloudWatch client
type MockCloudWatchClient struct {
	mock.Mock
}

func (m *MockCloudWatchClient) PutMetricData(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error) {
	args := m.Called(ctx, params)
	return &cloudwatch.PutMetricDataOutput{}, args.Error(1)
}

func (m *MockCloudWatchClient) PutMetricAlarm(ctx context.Context, params *cloudwatch.PutMetricAlarmInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricAlarmOutput, error) {
	args := m.Called(ctx, params)
	return &cloudwatch.PutMetricAlarmOutput{}, args.Error(1)
}

func TestServiceMonitor_UpdateServiceState(t *testing.T) {
	mockClient := new(MockCloudWatchClient)
	monitor := NewServiceMonitor(mockClient, "TestNamespace", time.Minute)

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

	// Expect PutMetricData calls
	mockClient.On("PutMetricData", mock.Anything, mock.MatchedBy(func(input *cloudwatch.PutMetricDataInput) bool {
		return *input.Namespace == "TestNamespace"
	})).Return(&cloudwatch.PutMetricDataOutput{}, nil)

	err := monitor.UpdateServiceState(state)
	assert.NoError(t, err)

	// Verify the state was stored
	storedState, exists := monitor.states[state.Name]
	assert.True(t, exists)
	assert.Equal(t, state.Status, storedState.Status)
	assert.Equal(t, state.Message, storedState.Message)

	mockClient.AssertExpectations(t)
}

func TestServiceMonitor_CreateServiceAlarm(t *testing.T) {
	mockClient := new(MockCloudWatchClient)
	monitor := NewServiceMonitor(mockClient, "TestNamespace", time.Minute)

	// Expect PutMetricAlarm call
	mockClient.On("PutMetricAlarm", mock.Anything, mock.MatchedBy(func(input *cloudwatch.PutMetricAlarmInput) bool {
		return *input.AlarmName == "TestAlarm" && *input.Namespace == "TestNamespace"
	})).Return(&cloudwatch.PutMetricAlarmOutput{}, nil)

	err := monitor.CreateServiceAlarm(context.Background(), "TestService", "TestAlarm", 0.5, types.ComparisonOperatorLessThanThreshold)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestServiceMonitor_GetStatusValue(t *testing.T) {
	monitor := NewServiceMonitor(nil, "TestNamespace", time.Minute)

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
	mockClient := new(MockCloudWatchClient)
	monitor := NewServiceMonitor(mockClient, "TestNamespace", time.Minute)

	// Add a stale state
	staleState := &ServiceState{
		Name:      "StaleService",
		Status:    "UP",
		Timestamp: time.Now().Add(-3 * time.Minute),
	}
	monitor.states["StaleService"] = staleState

	// Expect PutMetricAlarm call for stale service
	mockClient.On("PutMetricAlarm", mock.Anything, mock.MatchedBy(func(input *cloudwatch.PutMetricAlarmInput) bool {
		return *input.AlarmName == "StaleService-StaleState"
	})).Return(&cloudwatch.PutMetricAlarmOutput{}, nil)

	monitor.checkAndUpdateStates(context.Background())

	mockClient.AssertExpectations(t)
} 
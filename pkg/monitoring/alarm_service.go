package monitoring

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

// CloudWatchClient is an interface that wraps the required CloudWatch operations
type CloudWatchClient interface {
	PutMetricData(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error)
}

// AlarmService defines the interface for alarm management
type AlarmService interface {
	// CreateAlarm creates a new alarm
	CreateAlarm(ctx context.Context, alarm Alarm) error

	// UpdateAlarm updates an existing alarm
	UpdateAlarm(ctx context.Context, alarm Alarm) error

	// DeleteAlarm deletes an alarm
	DeleteAlarm(ctx context.Context, alarmName string) error

	// GetAlarmState gets the current state of an alarm
	GetAlarmState(ctx context.Context, alarmName string) (AlarmState, error)

	IsAlarmsEnabled() bool
}

// AlarmState represents the state of an alarm
type AlarmState string

const (
	AlarmStateOK       AlarmState = "OK"
	AlarmStateALARM   AlarmState = "ALARM"
	AlarmStateUnknown AlarmState = "UNKNOWN"
)

// ComparisonOperator represents the comparison operator for an alarm
type ComparisonOperator string

const (
	GreaterThanThreshold      ComparisonOperator = "GreaterThanThreshold"
	GreaterThanOrEqualToThreshold ComparisonOperator = "GreaterThanOrEqualToThreshold"
	LessThanThreshold        ComparisonOperator = "LessThanThreshold"
	LessThanOrEqualToThreshold ComparisonOperator = "LessThanOrEqualToThreshold"
)

// Alarm represents an alarm configuration
type Alarm struct {
	Name               string
	Description        string
	MetricName         string
	Namespace          string
	ComparisonOperator ComparisonOperator
	Threshold          float64
	Period            time.Duration
	EvaluationPeriods int
	Actions           []AlarmAction
	Labels            map[string]string
}

// AlarmAction represents an action to take when an alarm triggers
type AlarmAction struct {
	Type    string
	Target  string
	Payload map[string]interface{}
} 
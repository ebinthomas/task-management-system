package monitoring

import (
	"context"
	"time"
)

// AlarmService defines the interface for alarm providers
type AlarmService interface {
	// CreateAlarm creates a new alarm
	CreateAlarm(ctx context.Context, alarm Alarm) error

	// UpdateAlarm updates an existing alarm
	UpdateAlarm(ctx context.Context, alarm Alarm) error

	// DeleteAlarm deletes an alarm
	DeleteAlarm(ctx context.Context, alarmName string) error

	// GetAlarmState gets the current state of an alarm
	GetAlarmState(ctx context.Context, alarmName string) (AlarmState, error)
}

// AlarmState represents the current state of an alarm
type AlarmState string

const (
	AlarmStateOK       AlarmState = "OK"
	AlarmStateAlert    AlarmState = "ALERT"
	AlarmStateUnknown  AlarmState = "UNKNOWN"
)

// ComparisonOperator defines how to compare metric values
type ComparisonOperator string

const (
	GreaterThanThreshold          ComparisonOperator = "GreaterThanThreshold"
	GreaterThanOrEqualToThreshold ComparisonOperator = "GreaterThanOrEqualToThreshold"
	LessThanThreshold            ComparisonOperator = "LessThanThreshold"
	LessThanOrEqualToThreshold   ComparisonOperator = "LessThanOrEqualToThreshold"
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
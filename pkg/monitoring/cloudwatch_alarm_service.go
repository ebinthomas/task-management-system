package monitoring

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// CloudWatchAlarmService implements AlarmService using AWS CloudWatch
type CloudWatchAlarmService struct {
	client    *cloudwatch.Client
	namespace string
}

// NewCloudWatchAlarmService creates a new CloudWatch alarm service
func NewCloudWatchAlarmService(client *cloudwatch.Client, namespace string) *CloudWatchAlarmService {
	return &CloudWatchAlarmService{
		client:    client,
		namespace: namespace,
	}
}

// CreateAlarm implements AlarmService.CreateAlarm
func (c *CloudWatchAlarmService) CreateAlarm(ctx context.Context, alarm Alarm) error {
	// Validate alarm configuration
	if alarm.Name == "" {
		return fmt.Errorf("alarm name is required")
	}
	if alarm.MetricName == "" {
		return fmt.Errorf("metric name is required")
	}
	if alarm.Period.Seconds() < 60 {
		return fmt.Errorf("period must be at least 60 seconds")
	}
	if alarm.EvaluationPeriods < 1 {
		return fmt.Errorf("evaluation periods must be at least 1")
	}

	input := &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(alarm.Name),
		AlarmDescription:   aws.String(alarm.Description),
		MetricName:         aws.String(alarm.MetricName),
		Namespace:          aws.String(alarm.Namespace),
		Period:             aws.Int32(int32(alarm.Period.Seconds())),
		EvaluationPeriods:  aws.Int32(int32(alarm.EvaluationPeriods)),
		Threshold:          aws.Float64(alarm.Threshold),
		ComparisonOperator: c.convertComparisonOperator(alarm.ComparisonOperator),
		Statistic:          types.StatisticAverage,
	}

	// Add alarm actions if configured
	for _, action := range alarm.Actions {
		if action.Type == "sns" {
			input.AlarmActions = append(input.AlarmActions, action.Target)
		}
	}

	// Add dimensions from labels
	for key, value := range alarm.Labels {
		input.Dimensions = append(input.Dimensions, types.Dimension{
			Name:  aws.String(key),
			Value: aws.String(value),
		})
	}

	_, err := c.client.PutMetricAlarm(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create CloudWatch alarm: %w", err)
	}

	return nil
}

// UpdateAlarm implements AlarmService.UpdateAlarm
func (c *CloudWatchAlarmService) UpdateAlarm(ctx context.Context, alarm Alarm) error {
	// CloudWatch's PutMetricAlarm is idempotent, so we can reuse CreateAlarm
	return c.CreateAlarm(ctx, alarm)
}

// DeleteAlarm implements AlarmService.DeleteAlarm
func (c *CloudWatchAlarmService) DeleteAlarm(ctx context.Context, alarmName string) error {
	_, err := c.client.DeleteAlarms(ctx, &cloudwatch.DeleteAlarmsInput{
		AlarmNames: []string{alarmName},
	})
	if err != nil {
		return fmt.Errorf("failed to delete CloudWatch alarm: %w", err)
	}

	return nil
}

// GetAlarmState implements AlarmService.GetAlarmState
func (c *CloudWatchAlarmService) GetAlarmState(ctx context.Context, alarmName string) (AlarmState, error) {
	output, err := c.client.DescribeAlarms(ctx, &cloudwatch.DescribeAlarmsInput{
		AlarmNames: []string{alarmName},
	})
	if err != nil {
		return AlarmStateUnknown, fmt.Errorf("failed to get CloudWatch alarm state: %w", err)
	}

	if len(output.MetricAlarms) == 0 {
		return AlarmStateUnknown, fmt.Errorf("alarm not found: %s", alarmName)
	}

	switch output.MetricAlarms[0].StateValue {
	case "OK":
		return AlarmStateOK, nil
	case "ALARM":
		return AlarmStateALARM, nil
	default:
		return AlarmStateUnknown, nil
	}
}

// IsAlarmsEnabled implements AlarmService.IsAlarmsEnabled
func (c *CloudWatchAlarmService) IsAlarmsEnabled() bool {
	// Check if alarms are enabled via environment variable
	return os.Getenv("ENABLE_ALARMS") == "true"
}

// convertComparisonOperator converts our generic operator to CloudWatch's type
func (c *CloudWatchAlarmService) convertComparisonOperator(op ComparisonOperator) types.ComparisonOperator {
	switch op {
	case GreaterThanThreshold:
		return types.ComparisonOperatorGreaterThanThreshold
	case GreaterThanOrEqualToThreshold:
		return types.ComparisonOperatorGreaterThanOrEqualToThreshold
	case LessThanThreshold:
		return types.ComparisonOperatorLessThanThreshold
	case LessThanOrEqualToThreshold:
		return types.ComparisonOperatorLessThanOrEqualToThreshold
	default:
		return types.ComparisonOperatorGreaterThanThreshold
	}
} 
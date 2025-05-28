package monitoring

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// ServiceState represents different states of service components
type ServiceState struct {
	Name       string
	Status     string
	Message    string
	Timestamp  time.Time
	Metrics    map[string]float64
}

// ServiceMonitor handles monitoring of service states and alarm management
type ServiceMonitor struct {
	client        *cloudwatch.Client
	alarmService  AlarmService
	namespace     string
	checkInterval time.Duration
	states        map[string]*ServiceState
	alarmsEnabled bool
	stopCh        chan struct{}
}

// NewServiceMonitor creates a new service monitor instance
func NewServiceMonitor(client *cloudwatch.Client, alarmService AlarmService, namespace string, interval time.Duration) *ServiceMonitor {
	return &ServiceMonitor{
		client:        client,
		alarmService:  alarmService,
		namespace:     namespace,
		checkInterval: interval,
		states:        make(map[string]*ServiceState),
		alarmsEnabled: os.Getenv("ENABLE_ALARMS") == "true",
		stopCh:        make(chan struct{}),
	}
}

// Start begins monitoring service states
func (sm *ServiceMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(sm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopCh:
			return
		case <-ticker.C:
			sm.checkAndUpdateStates(ctx)
		}
	}
}

// Stop gracefully stops the service monitor
func (sm *ServiceMonitor) Stop() {
	close(sm.stopCh)
}

// UpdateServiceState updates the state of a service component
func (sm *ServiceMonitor) UpdateServiceState(state ServiceState) error {
	if state.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if state.Timestamp.IsZero() {
		state.Timestamp = time.Now()
	}

	sm.states[state.Name] = &state

	// Skip metric publishing if metrics are disabled
	if os.Getenv("ENABLE_METRICS") != "true" {
		return nil
	}

	// Create context with timeout for metric publishing
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publish state metrics to CloudWatch
	_, err := sm.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(sm.namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(state.Name + "Status"),
				Value:      aws.Float64(sm.getStatusValue(state.Status)),
				Timestamp:  aws.Time(state.Timestamp),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("ServiceName"),
						Value: aws.String(state.Name),
					},
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to publish metrics for service %s: %w", state.Name, err)
	}

	// Publish custom metrics if provided
	for metricName, value := range state.Metrics {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout publishing metrics for service %s", state.Name)
		default:
			_, err := sm.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
				Namespace: aws.String(sm.namespace),
				MetricData: []types.MetricDatum{
					{
						MetricName: aws.String(metricName),
						Value:      aws.Float64(value),
						Timestamp:  aws.Time(state.Timestamp),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("ServiceName"),
								Value: aws.String(state.Name),
							},
						},
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to publish metric %s for service %s: %w", metricName, state.Name, err)
			}
		}
	}

	return nil
}

// checkAndUpdateStates periodically checks service states and updates alarms
func (sm *ServiceMonitor) checkAndUpdateStates(ctx context.Context) {
	if !sm.alarmsEnabled {
		return // Skip alarm checks if alarms are disabled
	}

	for name, state := range sm.states {
		if time.Since(state.Timestamp) > sm.checkInterval*2 {
			log.Printf("Warning: Service %s state is stale", name)
			
			// Create alarm for stale service state
			alarm := Alarm{
				Name:               name + "-StaleState",
				Description:        "Service state has not been updated recently",
				MetricName:        name + "Status",
				Namespace:         sm.namespace,
				ComparisonOperator: LessThanThreshold,
				Threshold:         0,
				Period:           time.Minute,
				EvaluationPeriods: 2,
				Labels: map[string]string{
					"ServiceName": name,
				},
			}

			err := sm.alarmService.CreateAlarm(ctx, alarm)
			if err != nil {
				log.Printf("Failed to create/update alarm for service %s: %v", name, err)
			}
		}
	}
}

// getStatusValue converts status string to numeric value for metrics
func (sm *ServiceMonitor) getStatusValue(status string) float64 {
	switch status {
	case "UP":
		return 1.0
	case "DOWN":
		return 0.0
	case "DEGRADED":
		return 0.5
	default:
		return -1.0
	}
}

// CreateServiceAlarm creates an alarm for a service
func (sm *ServiceMonitor) CreateServiceAlarm(ctx context.Context, serviceName, alarmName string, threshold float64, operator ComparisonOperator) error {
	if !sm.alarmsEnabled {
		log.Printf("Alarms are disabled, skipping alarm creation for %s", serviceName)
		return nil
	}

	alarm := Alarm{
		Name:               alarmName,
		Description:        "Alarm for service " + serviceName,
		MetricName:        serviceName + "Status",
		Namespace:         sm.namespace,
		ComparisonOperator: operator,
		Threshold:         threshold,
		Period:           time.Minute,
		EvaluationPeriods: 2,
		Labels: map[string]string{
			"ServiceName": serviceName,
		},
	}

	return sm.alarmService.CreateAlarm(ctx, alarm)
}

// IsAlarmsEnabled returns whether alarms are enabled
func (sm *ServiceMonitor) IsAlarmsEnabled() bool {
	return sm.alarmsEnabled
} 
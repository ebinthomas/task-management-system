package metrics

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

var (
	metricsEnabled bool
	cwClient       *cloudwatch.Client
	once           sync.Once
	namespace      = "TaskAPI"
)

// Initialize sets up the metrics client based on environment configuration
func Initialize() error {
	var initErr error
	once.Do(func() {
		// Check if metrics are enabled via environment variable
		metricsEnabled = os.Getenv("ENABLE_METRICS") == "true"
		if !metricsEnabled {
			log.Println("CloudWatch metrics collection is disabled")
			return
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(os.Getenv("AWS_REGION")),
		)
		if err != nil {
			initErr = fmt.Errorf("failed to initialize AWS config: %v", err)
			return
		}

		// Initialize CloudWatch client
		cwClient = cloudwatch.NewFromConfig(cfg)
		
		// Test the CloudWatch connection
		_, err = cwClient.ListMetrics(context.Background(), &cloudwatch.ListMetricsInput{
			Namespace: aws.String(namespace),
		})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to CloudWatch: %v", err)
			return
		}

		log.Println("CloudWatch metrics collection initialized successfully")
	})
	return initErr
}

// IsEnabled returns whether metrics collection is enabled
func IsEnabled() bool {
	return metricsEnabled && cwClient != nil
}

// RecordRequestDuration records the duration of an HTTP request
func RecordRequestDuration(method, path string, duration float64) {
	if !IsEnabled() {
		return
	}

	_, err := cwClient.PutMetricData(context.Background(), &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("RequestDuration"),
				Unit:       types.StandardUnitSeconds,
				Value:      aws.Float64(duration),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("Method"),
						Value: aws.String(method),
					},
					{
						Name:  aws.String("Path"),
						Value: aws.String(path),
					},
				},
				Timestamp: aws.Time(time.Now()),
			},
		},
	})

	if err != nil {
		log.Printf("Error publishing duration metric to CloudWatch: %v", err)
	}
}

// RecordAPICall records API call counts with status codes
func RecordAPICall(method, path string, statusCode int) {
	if !IsEnabled() {
		return
	}

	_, err := cwClient.PutMetricData(context.Background(), &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("APICallCount"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(1.0),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("Method"),
						Value: aws.String(method),
					},
					{
						Name:  aws.String("Path"),
						Value: aws.String(path),
					},
					{
						Name:  aws.String("StatusCode"),
						Value: aws.String(fmt.Sprintf("%d", statusCode)),
					},
				},
				Timestamp: aws.Time(time.Now()),
			},
		},
	})

	if err != nil {
		log.Printf("Error publishing API call metric to CloudWatch: %v", err)
	}
}

// RecordCacheOperation records cache hits and misses
func RecordCacheOperation(operation string, success bool) {
	if !IsEnabled() {
		return
	}

	_, err := cwClient.PutMetricData(context.Background(), &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("CacheOperations"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(1.0),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("Operation"),
						Value: aws.String(operation),
					},
					{
						Name:  aws.String("Result"),
						Value: aws.String(map[bool]string{true: "Success", false: "Failure"}[success]),
					},
				},
				Timestamp: aws.Time(time.Now()),
			},
		},
	})

	if err != nil {
		log.Printf("Error publishing cache metric to CloudWatch: %v", err)
	}
} 
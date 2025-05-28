package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"sample/task-management-system/pkg/api"
	"sample/task-management-system/pkg/auth"
	"sample/task-management-system/pkg/middleware"
	"sample/task-management-system/pkg/repository/postgres"
	"sample/task-management-system/pkg/service"
	"sample/task-management-system/pkg/cache"
	"sample/task-management-system/pkg/health"
	"sample/task-management-system/pkg/metrics"
	"sample/task-management-system/pkg/monitoring"
)

func main() {
	// Enable verbose logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	
	// Initialize metrics if enabled
	if err := metrics.Initialize(); err != nil {
		log.Printf("Warning: Failed to initialize metrics: %v", err)
	}
	
	// Load configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "taskdb")
	serverPort := getEnv("SERVER_PORT", "8080")
	
	// Auth configuration
	authSecret := []byte(getEnv("AUTH_SECRET", ""))
	authIssuer := getEnv("AUTH_ISSUER", "")
	
	if len(authSecret) == 0 || authIssuer == "" {
		log.Fatal("AUTH_SECRET and AUTH_ISSUER must be set")
	}

	// Initialize AWS CloudWatch client and alarm service if monitoring is enabled
	var serviceMonitor *monitoring.ServiceMonitor
	if os.Getenv("ENABLE_METRICS") == "true" {
		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(os.Getenv("AWS_REGION")),
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize AWS config: %v", err)
		} else {
			cwClient := cloudwatch.NewFromConfig(cfg)
			
			// Initialize alarm service based on configuration
			var alarmService monitoring.AlarmService
			alarmProvider := os.Getenv("ALARM_PROVIDER")
			switch alarmProvider {
			case "cloudwatch":
				alarmService = monitoring.NewCloudWatchAlarmService(cwClient, "TaskAPI")
			default:
				log.Printf("Warning: Unknown alarm provider %s, defaulting to CloudWatch", alarmProvider)
				alarmService = monitoring.NewCloudWatchAlarmService(cwClient, "TaskAPI")
			}

			// Initialize service monitor
			serviceMonitor = monitoring.NewServiceMonitor(cwClient, alarmService, "TaskAPI", 1*time.Minute)
			go serviceMonitor.Start(context.Background())
			
			// Create default alarms
			if err := setupDefaultAlarms(context.Background(), serviceMonitor); err != nil {
				log.Printf("Warning: Failed to setup default alarms: %v", err)
			}
		}
	}

	log.Printf("Connecting to database: host=%s port=%s user=%s dbname=%s", dbHost, dbPort, dbUser, dbName)

	// Create database URL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Initialize dependencies
	taskRepo := postgres.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := api.NewTaskHandler(taskService)

	// Set up the router
	router := mux.NewRouter()

	// Configure auth middleware
	authConfig := auth.AuthConfig{
		JWTSecret:    authSecret,
		AllowedRoles: auth.DefaultRoles,
		PublicPaths:  []string{"/health"},
	}

	// Add global middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.NewSafetyLimiter().Limit)
	router.Use(auth.AuthMiddleware(authConfig))
	
	// Initialize Redis cache
	log.Printf("Connecting to Redis at %s", os.Getenv("REDIS_ADDR"))
	redisCache, err := cache.NewRedisCache(
		os.Getenv("REDIS_ADDR"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err)
	}
	log.Println("Successfully connected to Redis")

	// Create middleware instances
	cacheMiddleware := middleware.NewCacheMiddleware(redisCache, 5*time.Minute)

	// API v1 routes
	v1Router := router.PathPrefix("/api/v1").Subrouter()
	
	// Tasks routes for v1
	tasksRouter := v1Router.PathPrefix("/tasks").Subrouter()
	tasksRouter.Use(auth.ResourceOwnershipMiddleware("task"))
	
	// Configure router to handle trailing slashes
	tasksRouter.StrictSlash(true)
	
	taskHandler.RegisterRoutes(tasksRouter)

	// Apply cache middleware
	handler := cacheMiddleware.CacheHandler(router)

	// Initialize health check handler with service monitor
	healthHandler := health.NewHandler(
		"1.0", // API version
		db,      // database connection
		redisCache, // Redis client
		serviceMonitor, // Service monitor
	)

	// Add global health check route
	router.Handle("/health", healthHandler).Methods(http.MethodGet)

	// Start the server
	log.Printf("Server starting on port %s", serverPort)
	if err := http.ListenAndServe(":"+serverPort, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// setupDefaultAlarms creates the default set of alarms
func setupDefaultAlarms(ctx context.Context, monitor *monitoring.ServiceMonitor) error {
	alarms := []struct {
		service   string
		name      string
		threshold float64
	}{
		{
			service:   "database",
			name:      "DatabaseDown",
			threshold: 0.5,
		},
		{
			service:   "cache",
			name:      "CacheDown",
			threshold: 0.5,
		},
		{
			service:   "system",
			name:      "SystemDegraded",
			threshold: 0.5,
		},
	}

	for _, alarm := range alarms {
		err := monitor.CreateServiceAlarm(
			ctx,
			alarm.service,
			alarm.name,
			alarm.threshold,
			monitoring.LessThanThreshold,
		)
		if err != nil {
			return fmt.Errorf("failed to create alarm %s: %v", alarm.name, err)
		}
	}

	return nil
} 
package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"sample/task-management-system/pkg/monitoring"
)

// Status represents the status of a service component
type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time             `json:"timestamp"`
	Version   string                `json:"version"`
	Services  map[string]Component  `json:"services"`
	System    SystemInfo            `json:"system"`
}

// Component represents a service component's health
type Component struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

// SystemInfo represents system-level information
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutines"`
	NumCPU       int    `json:"num_cpu"`
	HeapInUse    uint64 `json:"heap_in_use"`
}

// Handler handles health check requests
type Handler struct {
	version  string
	db       *sql.DB
	cache    interface {
		Ping(ctx context.Context) error
	}
	monitor  interface {
		UpdateServiceState(state monitoring.ServiceState) error
	}
}

// NewHandler creates a new health check handler
func NewHandler(version string, db *sql.DB, cache interface{ Ping(ctx context.Context) error }, monitor interface{ UpdateServiceState(state monitoring.ServiceState) error }) *Handler {
	return &Handler{
		version:  version,
		db:       db,
		cache:    cache,
		monitor:  monitor,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	response := h.checkHealth(ctx)

	w.Header().Set("Content-Type", "application/json")
	if response.Status == StatusDown {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

// checkHealth performs the health check
func (h *Handler) checkHealth(ctx context.Context) HealthResponse {
	services := make(map[string]Component)
	overallStatus := StatusUp

	// Check database if configured
	dbComponent := h.checkDatabase(ctx)
	services["database"] = dbComponent
	if h.db != nil && dbComponent.Status == StatusDown {
		overallStatus = StatusDown
	}

	// Update database state in service monitor
	if h.monitor != nil {
		h.monitor.UpdateServiceState(monitoring.ServiceState{
			Name:      "database",
			Status:    string(dbComponent.Status),
			Message:   dbComponent.Message,
			Timestamp: time.Now(),
			Metrics:   map[string]float64{},
		})
	}

	// Check Redis if available
	if h.cache != nil {
		cacheComponent := h.checkRedis(ctx)
		services["cache"] = cacheComponent
		if cacheComponent.Status == StatusDown {
			overallStatus = StatusDown
		}

		// Update cache state in service monitor
		if h.monitor != nil {
			h.monitor.UpdateServiceState(monitoring.ServiceState{
				Name:      "cache",
				Status:    string(cacheComponent.Status),
				Message:   cacheComponent.Message,
				Timestamp: time.Now(),
				Metrics:   map[string]float64{},
			})
		}
	}

	// Get system info
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	sysInfo := SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:      runtime.NumCPU(),
		HeapInUse:    mem.HeapInuse,
	}

	// Update system metrics in service monitor
	if h.monitor != nil {
		h.monitor.UpdateServiceState(monitoring.ServiceState{
			Name:      "system",
			Status:    string(overallStatus),
			Message:   "System metrics updated",
			Timestamp: time.Now(),
			Metrics: map[string]float64{
				"num_goroutines": float64(sysInfo.NumGoroutine),
				"heap_in_use":    float64(sysInfo.HeapInUse),
			},
		})
	}

	return HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Version:   h.version,
		Services:  services,
		System:    sysInfo,
	}
}

// checkDatabase verifies database connectivity
func (h *Handler) checkDatabase(ctx context.Context) Component {
	if h.db == nil {
		return Component{
			Status:  StatusDown,
			Message: "Database connection not configured",
		}
	}

	err := h.db.PingContext(ctx)
	if err != nil {
		return Component{
			Status:  StatusDown,
			Message: "Failed to connect to database: " + err.Error(),
		}
	}

	return Component{
		Status:  StatusUp,
		Message: "Database connection successful",
	}
}

// checkRedis verifies Redis connectivity
func (h *Handler) checkRedis(ctx context.Context) Component {
	if h.cache == nil {
		return Component{
			Status:  StatusDown,
			Message: "Redis connection not configured",
		}
	}

	err := h.cache.Ping(ctx)
	if err != nil {
		return Component{
			Status:  StatusDown,
			Message: "Failed to connect to Redis: " + err.Error(),
		}
	}

	return Component{
		Status:  StatusUp,
		Message: "Redis connection successful",
	}
} 
package middleware

import (
	"log"
	"net/http"
	"time"

	"sample/task-management-system/pkg/metrics"
)

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs request details and records metrics
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log incoming request
		log.Printf("Incoming request: %s %s", r.Method, r.RequestURI)

		// Create a response wrapper to capture the status code
		rw := newResponseWriter(w)

		// Process the request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Log completion
		log.Printf("Completed request: %s %s (status: %d, duration: %.2fs)",
			r.Method, r.RequestURI, rw.statusCode, duration)

		// Record metrics if enabled
		metrics.RecordRequestDuration(r.Method, r.URL.Path, duration)
		metrics.RecordAPICall(r.Method, r.URL.Path, rw.statusCode)
	})
} 
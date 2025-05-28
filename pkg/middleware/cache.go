package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"sample/task-management-system/pkg/cache"
)

// CacheMiddleware handles caching of HTTP responses
type CacheMiddleware struct {
	cache    *cache.RedisCache
	duration time.Duration
}

func NewCacheMiddleware(cache *cache.RedisCache, expiration time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		cache:    cache,
		duration: expiration,
	}
}

// buildCacheKey generates a consistent and efficient cache key
func (m *CacheMiddleware) buildCacheKey(r *http.Request) string {
	// Extract path parts
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	
	// Get API version
	version := "v1"
	if len(parts) > 1 {
		version = parts[1]
	}
	
	// Sort and filter query parameters
	params := r.URL.Query()
	var queryParts []string
	
	// Get sorted keys for consistent ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		if isCacheableParam(k) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	
	// Build normalized query string
	for _, k := range keys {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, params.Get(k)))
	}
	
	// Build final cache key
	keyParts := []string{
		version,
		"tasks", // Always use "tasks" as the resource type
	}

	// Add user ID if present
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		keyParts = append(keyParts, userID)
	}
	
	// Add resource ID if present (for single resource requests)
	if len(parts) > 3 {
		keyParts = append(keyParts, parts[3])
	}
	
	if len(queryParts) > 0 {
		keyParts = append(keyParts, strings.Join(queryParts, "&"))
	}
	
	key := strings.Join(keyParts, ":")
	log.Printf("Cache key generated: %s for path: %s", key, r.URL.Path)
	return key
}

// buildCachePatterns generates patterns to match related cache keys
func (m *CacheMiddleware) buildCachePatterns(r *http.Request) []string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	version := "v1"
	if len(parts) > 1 {
		version = parts[1]
	}
	
	// Always include the base pattern that matches all task-related keys
	patterns := []string{
		fmt.Sprintf("%s:tasks:*", version),
	}

	// Add user-specific pattern if user ID is present
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		patterns = append(patterns, fmt.Sprintf("%s:tasks:%s:*", version, userID))
	}

	// For single resource operations, add specific resource pattern
	if len(parts) > 3 {
		resourceID := parts[3]
		patterns = append(patterns, 
			fmt.Sprintf("%s:tasks:*:%s", version, resourceID),
			fmt.Sprintf("%s:tasks:*:%s:*", version, resourceID),
		)
	}

	log.Printf("Cache patterns to invalidate: %v for path: %s", patterns, r.URL.Path)
	return patterns
}

// invalidateRelatedCaches removes all cached entries related to the modified resource
func (m *CacheMiddleware) invalidateRelatedCaches(r *http.Request) error {
	patterns := m.buildCachePatterns(r)
	for _, pattern := range patterns {
		log.Printf("Attempting to invalidate cache pattern: %s", pattern)
		keys, err := m.cache.Keys(r.Context(), pattern)
		if err != nil {
			log.Printf("Failed to get keys for pattern %s: %v", pattern, err)
			continue
		}
		
		for _, key := range keys {
			if err := m.cache.Delete(r.Context(), key); err != nil {
				log.Printf("Failed to delete cache key %s: %v", key, err)
				continue
			}
			log.Printf("Successfully invalidated cache key: %s", key)
		}
	}
	return nil
}

// isCacheableParam determines if a query parameter should be included in the cache key
func isCacheableParam(param string) bool {
	cacheableParams := map[string]bool{
		"status": true,
		"limit":  true,
		"page":   true,
		"sort":   true,
		"order":  true,
	}
	return cacheableParams[param]
}

func (m *CacheMiddleware) CacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle write operations (POST, PUT, DELETE)
		if r.Method != http.MethodGet {
			// Invalidate related caches before processing the request
			log.Printf("Write operation detected (%s %s), invalidating caches", r.Method, r.URL.Path)
			if err := m.invalidateRelatedCaches(r); err != nil {
				log.Printf("Cache invalidation failed: %v", err)
			}
			next.ServeHTTP(w, r)
			return
		}

		// Handle read operations (GET)
		cacheKey := m.buildCacheKey(r)

		// Try to get from cache
		var cachedResponse []byte
		err := m.cache.Get(r.Context(), cacheKey, &cachedResponse)
		if err == nil {
			log.Printf("Cache HIT for key: %s", cacheKey)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.Write(cachedResponse)
			return
		}
		log.Printf("Cache MISS for key: %s", cacheKey)

		// Create a response recorder
		buf := &bytes.Buffer{}
		recorder := &responseRecorder{
			ResponseWriter: w,
			buf:           buf,
		}

		// Call the next handler
		next.ServeHTTP(recorder, r)

		// Only cache successful responses
		if recorder.status == http.StatusOK || recorder.status == http.StatusCreated {
			if err := m.cache.Set(r.Context(), cacheKey, buf.Bytes(), m.duration); err != nil {
				log.Printf("Failed to set cache for key %s: %v", cacheKey, err)
			} else {
				log.Printf("Successfully cached response for key: %s", cacheKey)
			}
		}
	})
}

type responseRecorder struct {
	http.ResponseWriter
	buf    *bytes.Buffer
	status int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.buf.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
} 
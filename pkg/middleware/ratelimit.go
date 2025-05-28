package middleware

import (
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	client      *redis.Client
	maxRequests int
	window      time.Duration
}

func NewRateLimiter(redisURL string, maxRequests int, window time.Duration) (*RateLimiter, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	return &RateLimiter{
		client:      client,
		maxRequests: maxRequests,
		window:      window,
	}, nil
}

func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as key
		key := "ratelimit:" + r.RemoteAddr

		// Use Redis INCR to count requests
		val, err := rl.client.Incr(r.Context(), key).Result()
		if err != nil {
			http.Error(w, "Rate limiting error", http.StatusInternalServerError)
			return
		}

		// Set expiry for the first request in window
		if val == 1 {
			rl.client.Expire(r.Context(), key, rl.window)
		}

		// Check if request count exceeds limit
		if val > int64(rl.maxRequests) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Local rate limiter as fallback
type LocalRateLimiter struct {
	limiter *rate.Limiter
}

func NewLocalRateLimiter(r rate.Limit, b int) *LocalRateLimiter {
	return &LocalRateLimiter{
		limiter: rate.NewLimiter(r, b),
	}
}

func (l *LocalRateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SafetyLimiter provides basic protection against extreme cases
type SafetyLimiter struct {
	limiter *rate.Limiter
}

// NewSafetyLimiter creates a new safety limiter with very permissive limits
func NewSafetyLimiter() *SafetyLimiter {
	// Allow 1000 requests per second with burst of 100
	return &SafetyLimiter{
		limiter: rate.NewLimiter(rate.Limit(1000), 100),
	}
}

// Limit provides basic rate limiting as a safety net
func (l *SafetyLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.limiter.Allow() {
			http.Error(w, "Service Protection", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
} 
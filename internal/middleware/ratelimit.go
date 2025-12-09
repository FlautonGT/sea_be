package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"seaply/internal/database"
	"seaply/internal/utils"
)

type RateLimiter struct {
	redis *database.RedisClient
}

type RateLimitConfig struct {
	Requests int           // Number of requests allowed
	Window   time.Duration // Time window
}

var (
	// Default rate limits
	DefaultRateLimit = RateLimitConfig{Requests: 100, Window: time.Minute}
	AuthRateLimit    = RateLimitConfig{Requests: 10, Window: time.Minute}
	OrderRateLimit   = RateLimitConfig{Requests: 30, Window: time.Minute}
	AdminRateLimit   = RateLimitConfig{Requests: 200, Window: time.Minute}
)

func NewRateLimiter(redis *database.RedisClient) *RateLimiter {
	return &RateLimiter{redis: redis}
}

// RateLimit creates a rate limiting middleware
func (rl *RateLimiter) RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get identifier (IP or user ID)
			identifier := getClientIP(r)
			if userID := GetUserIDFromContext(r.Context()); userID != "" {
				identifier = userID
			}

			// Create rate limit key
			key := rl.redis.RateLimitKey(identifier, r.URL.Path)

			// Check and increment
			allowed, remaining, resetAt, err := rl.checkAndIncrement(r.Context(), key, cfg)
			if err != nil {
				// On error, allow the request but log it
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Requests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))

			if !allowed {
				utils.WriteErrorJSON(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
					"Terlalu banyak permintaan. Silakan coba lagi nanti.",
					fmt.Sprintf("Limit: %d requests per %s", cfg.Requests, cfg.Window))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) checkAndIncrement(ctx context.Context, key string, cfg RateLimitConfig) (bool, int, time.Time, error) {
	// Get current count
	count, err := rl.redis.Client.Incr(ctx, key).Result()
	if err != nil {
		return true, cfg.Requests, time.Now().Add(cfg.Window), err
	}

	// Set expiry on first request
	if count == 1 {
		rl.redis.Client.Expire(ctx, key, cfg.Window)
	}

	// Get TTL for reset time
	ttl, err := rl.redis.Client.TTL(ctx, key).Result()
	if err != nil {
		ttl = cfg.Window
	}
	resetAt := time.Now().Add(ttl)

	remaining := cfg.Requests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return count <= int64(cfg.Requests), remaining, resetAt, nil
}

// IPRateLimit rate limits by IP only
func (rl *RateLimiter) IPRateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identifier := getClientIP(r)
			key := rl.redis.RateLimitKey(identifier, r.URL.Path)

			allowed, remaining, resetAt, err := rl.checkAndIncrement(r.Context(), key, cfg)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Requests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))

			if !allowed {
				utils.WriteErrorJSON(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
					"Terlalu banyak permintaan. Silakan coba lagi nanti.",
					fmt.Sprintf("Limit: %d requests per %s", cfg.Requests, cfg.Window))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

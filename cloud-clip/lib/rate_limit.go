package lib

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	burst    int
	window   time.Duration
}

func NewRateLimiter(limit int, burst int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		burst:    burst,
		window:   time.Second,
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[clientID]

	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	rl.requests[clientID] = validRequests

	if len(validRequests) >= rl.burst {
		return false
	}

	rl.requests[clientID] = append(validRequests, now)
	return true
}

func (s *ClipboardServer) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	if s.config.Server.RateLimit <= 0 {
		return next
	}

	limiter := NewRateLimiter(s.config.Server.RateLimit, s.config.Server.RateLimitBurst)

	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := get_remote_ip(r)

		if !limiter.Allow(clientIP) {
			s.logger.Printf("速率限制触发: IP: %s, 路径: %s", clientIP, r.URL.Path)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

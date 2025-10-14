package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type visitor struct {
	tokens     int
	lastAccess time.Time
}

func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerWindow,
		window:   window,
	}

	go rl.cleanupVisitors()

	return rl
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastAccess) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:     rl.rate,
			lastAccess: time.Now(),
		}
		rl.visitors[ip] = v
		return v
	}

	elapsed := time.Since(v.lastAccess)
	if elapsed >= rl.window {
		v.tokens = rl.rate
		v.lastAccess = time.Now()
	}

	return v
}

func (rl *RateLimiter) Allow(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}

		if !rl.Allow(ip) {
			http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

func (rl *RateLimiter) RequestEventMiddleware(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		ip := e.Request.RemoteAddr
		if xff := e.Request.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}

		if !rl.Allow(ip) {
			return e.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Too many requests. Please try again later.",
			})
		}

		return next(e)
	}
}

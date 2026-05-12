package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"nex_play_auth/github.com/pkg/response"

	"golang.org/x/time/rate"
)

// ipEntry tracks a per IP limiter and when it was last used
type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter holds a token bucket limiter for every unique client IP
// Stale entries are cleaned up in a background goroutine
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*ipEntry
	rps     rate.Limit // requests per second
	burst   int        // maximum burst size
}

// Create NewRateLimiter and starts the cleanup goroutine
func NewRateLimiter(rps, burst int) *RateLimiter {

	rl := &RateLimiter{

		clients: make(map[string]*ipEntry),
		rps:     rate.Limit(rps),
		burst:   burst,
	}

	go rl.cleanup()

	return rl
}

// Middleware returns an http.Handler middleware that enforces the rate limit
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip := clientIP(r)

		if !rl.allow(ip) {

			w.Header().Set("Retry-After", "1")

			response.Error(w, http.StatusTooManyRequests, "too many requests, please slow down")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ip string) bool {

	rl.mu.Lock()

	defer rl.mu.Unlock()

	entry, exists := rl.clients[ip]

	if !exists {

		entry = &ipEntry{limiter: rate.NewLimiter(rl.rps, rl.burst)}

		rl.clients[ip] = entry
	}

	entry.lastSeen = time.Now()

	return entry.limiter.Allow()
}

// cleanup removes entries that have not been seen for 3 minutes
// Runs every minute in the background
func (rl *RateLimiter) cleanup() {

	ticker := time.NewTicker(time.Minute)

	defer ticker.Stop()

	for range ticker.C {

		rl.mu.Lock()

		for ip, entry := range rl.clients {

			if time.Since(entry.lastSeen) > 3*time.Minute {

				delete(rl.clients, ip)
			}
		}

		rl.mu.Unlock()
	}
}

// clientIP extracts the real client IP
func clientIP(r *http.Request) string {

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {

		if idx := indexOf(xff, ','); idx != -1 {

			return xff[:idx]
		}

		return xff
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {

		return r.RemoteAddr
	}

	return ip
}

func indexOf(s string, b byte) int {

	for i := 0; i < len(s); i++ {

		if s[i] == b {

			return i
		}
	}

	return -1
}

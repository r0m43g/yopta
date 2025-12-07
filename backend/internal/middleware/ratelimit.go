// backend/internal/middleware/ratelimit.go
package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter represents a comprehensive rate limiting system that tracks and controls
// request frequency based on client identifiers (like IP addresses) and specific API endpoints.
// It provides both global rate limiting (per client IP) and endpoint-specific rate limiting,
// allowing for more granular control over high-risk or resource-intensive operations.
// The implementation uses an in-memory approach with efficient cleanup to prevent memory leaks.
type RateLimiter struct {
	sync.RWMutex                               // Mutex for thread-safe concurrent access
	ipRequests      map[string][]time.Time     // Stores request timestamps by IP address
	pathRequests    map[string][]time.Time     // Stores request timestamps by API path
	pathConfigs     map[string]RateLimitConfig // Configuration settings for specific paths
	defaultConfig   RateLimitConfig            // Default configuration for paths without specific settings
	cleanupInterval time.Duration              // How often to remove expired entries
}

// RateLimitConfig defines the parameters for rate limiting rules.
// This allows different endpoints to have different thresholds and time windows.
type RateLimitConfig struct {
	MaxRequests int           // Maximum number of requests allowed in the time window
	Window      time.Duration // The time period for measuring request frequency
}

// NewRateLimiter creates and initializes a new rate limiter with the specified configuration.
// It also starts a background goroutine to periodically clean up expired entries.
//
// Parameters:
//   - defaultMaxRequests: The default maximum number of requests allowed per time window
//   - defaultWindow: The default time window for rate limiting
//   - cleanupInterval: How often to run the cleanup routine to remove expired entries
//
// Returns a configured RateLimiter instance ready for use.
func NewRateLimiter(
	defaultMaxRequests int,
	defaultWindow time.Duration,
	cleanupInterval time.Duration,
) *RateLimiter {
	limiter := &RateLimiter{
		ipRequests:      make(map[string][]time.Time),
		pathRequests:    make(map[string][]time.Time),
		pathConfigs:     make(map[string]RateLimitConfig),
		defaultConfig:   RateLimitConfig{MaxRequests: defaultMaxRequests, Window: defaultWindow},
		cleanupInterval: cleanupInterval,
	}

	// Start background cleanup of expired entries to prevent memory leaks
	go limiter.cleanup()

	return limiter
}

// SetPathConfig configures custom rate limits for a specific API path.
// This allows creating stricter limits for sensitive operations (like authentication)
// or more generous limits for less critical operations.
//
// Parameters:
//   - path: The API endpoint path to configure
//   - maxRequests: Maximum number of requests allowed within the time window
//   - window: The time window for measuring request frequency
func (rl *RateLimiter) SetPathConfig(path string, maxRequests int, window time.Duration) {
	rl.Lock()
	defer rl.Unlock()
	rl.pathConfigs[path] = RateLimitConfig{
		MaxRequests: maxRequests,
		Window:      window,
	}
}

// cleanup periodically removes old request records to prevent memory leaks.
// This function runs as a background goroutine and removes entries that
// are outside their respective time windows, freeing up memory.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.Lock()
		now := time.Now()

		// Clean IP records
		for ip, times := range rl.ipRequests {
			newTimes := []time.Time{}
			for _, t := range times {
				if now.Sub(t) < rl.defaultConfig.Window {
					newTimes = append(newTimes, t)
				}
			}
			if len(newTimes) > 0 {
				rl.ipRequests[ip] = newTimes
			} else {
				delete(rl.ipRequests, ip)
			}
		}

		// Clean path records using their specific time windows
		for path, times := range rl.pathRequests {
			config := rl.getConfigForPath(path)
			newTimes := []time.Time{}
			for _, t := range times {
				if now.Sub(t) < config.Window {
					newTimes = append(newTimes, t)
				}
			}
			if len(newTimes) > 0 {
				rl.pathRequests[path] = newTimes
			} else {
				delete(rl.pathRequests, path)
			}
		}

		rl.Unlock()
	}
}

// getConfigForPath returns the rate limit configuration for a specific path.
// If no specific configuration exists, it returns the default configuration.
func (rl *RateLimiter) getConfigForPath(path string) RateLimitConfig {
	if config, ok := rl.pathConfigs[path]; ok {
		return config
	}
	return rl.defaultConfig
}

// isIPAllowed checks if an IP address is allowed to make another request.
// This implements global rate limiting based on client IP address.
// Returns true if the request should be allowed, false if it exceeds the limit.
func (rl *RateLimiter) isIPAllowed(ip string) bool {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	times := rl.ipRequests[ip]

	// Filter out requests outside the current window
	newTimes := []time.Time{}
	for _, t := range times {
		if now.Sub(t) < rl.defaultConfig.Window {
			newTimes = append(newTimes, t)
		}
	}

	// Check if we're at the limit
	if len(newTimes) >= rl.defaultConfig.MaxRequests {
		rl.ipRequests[ip] = newTimes
		return false
	}

	// Record this request
	newTimes = append(newTimes, now)
	rl.ipRequests[ip] = newTimes
	return true
}

// isPathAllowed checks if a path is allowed to receive another request.
// This implements endpoint-specific rate limiting to protect sensitive
// or resource-intensive operations.
// Returns true if the request should be allowed, false if it exceeds the limit.
func (rl *RateLimiter) isPathAllowed(path string) bool {
	rl.Lock()
	defer rl.Unlock()

	config := rl.getConfigForPath(path)
	now := time.Now()
	times := rl.pathRequests[path]

	// Filter out requests outside the current window
	newTimes := []time.Time{}
	for _, t := range times {
		if now.Sub(t) < config.Window {
			newTimes = append(newTimes, t)
		}
	}

	// Check if we're at the limit
	if len(newTimes) >= config.MaxRequests {
		rl.pathRequests[path] = newTimes
		return false
	}

	// Record this request
	newTimes = append(newTimes, now)
	rl.pathRequests[path] = newTimes
	return true
}

// GetClientIP extracts the client IP address from the request.
// It accounts for common proxy headers to get the real client IP
// even when behind load balancers, proxies, or CDNs.
func GetClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (common in proxied environments)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		return xff[:i]
	}
	// Check for X-Real-IP header (used by some proxies)
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	// Extract from RemoteAddr as a fallback
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RateLimitMiddleware creates middleware that enforces rate limiting rules.
// It applies both IP-based and path-based limits and returns appropriate
// HTTP 429 Too Many Requests responses when limits are exceeded.
// The middleware also sets Retry-After headers to inform clients about
// when they should retry their requests.
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetClientIP(r)
			path := r.URL.Path

			// Check both IP-based and path-based rate limiting
			ipAllowed := limiter.isIPAllowed(ip)
			pathAllowed := limiter.isPathAllowed(path)

			if !ipAllowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write(
					[]byte(`{"error": "Per daug užklausų. Bandykite vėliau."}`),
				) // Rate limit exceeded. Please try again later.
				return
			}

			if !pathAllowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write(
					[]byte(
						`{"error": "Šis API endpointas šiuo metu ribojamas. Bandykite vėliau."}`, // This endpoint is currently rate limited. Please try again later.
					),
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


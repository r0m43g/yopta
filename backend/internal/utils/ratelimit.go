// backend/internal/utils/ratelimit.go
package utils

import (
	"log"
	"os"
	"strconv"
	"time"

	"yopta-template/internal/middleware"
)

// InitRateLimiter initializes and configures the application's rate limiting system.
// Rate limiting helps protect against abuse and DoS attacks by controlling the number
// of requests a client can make within a specified time window.
//
// The function reads configuration from environment variables and applies different
// rate limiting rules for various API endpoints, with stricter limits for authentication
// and sensitive operations.
//
// Returns: A configured RateLimiter instance ready to be used in HTTP middleware.
func InitRateLimiter() *middleware.RateLimiter {
	// Get default rate limit values from environment variables with fallbacks
	// These limits apply to general API endpoints
	defaultMaxRequests := getEnvAsInt("RATE_LIMIT_DEFAULT_MAX", 100)
	defaultWindow := getEnvAsDuration("RATE_LIMIT_DEFAULT_WINDOW", 1*time.Minute)
	cleanupInterval := getEnvAsDuration("RATE_LIMIT_CLEANUP_INTERVAL", 5*time.Minute)

	// Create a new rate limiter with the default configuration
	limiter := middleware.NewRateLimiter(defaultMaxRequests, defaultWindow, cleanupInterval)

	// Configure stricter limits for authentication endpoints
	// These endpoints are common targets for brute force attacks and require
	// more protection than regular API endpoints
	authMaxRequests := getEnvAsInt("RATE_LIMIT_AUTH_MAX", 5)
	authWindow := getEnvAsDuration("RATE_LIMIT_AUTH_WINDOW", 1*time.Minute)

	// Apply the stricter configuration to authentication endpoints
	limiter.SetPathConfig("/api/v1/login", authMaxRequests, authWindow)
	limiter.SetPathConfig("/api/v1/register", authMaxRequests, authWindow)
	limiter.SetPathConfig("/api/v1/refresh-token", authMaxRequests, authWindow)
	limiter.SetPathConfig("/api/v1/verify-email", authMaxRequests, authWindow)
	limiter.SetPathConfig("/api/v1/change-password", authMaxRequests, authWindow)

	log.Printf("Rate limiting configured: Default: %d requests/%s, Auth: %d requests/%s",
		defaultMaxRequests, defaultWindow, authMaxRequests, authWindow)

	// Configure limits for admin endpoints
	// Admin endpoints need special protection as they often provide privileged operations
	adminMaxRequests := getEnvAsInt("RATE_LIMIT_ADMIN_MAX", 20)
	adminWindow := getEnvAsDuration("RATE_LIMIT_ADMIN_WINDOW", 1*time.Minute)
	limiter.SetPathConfig("/api/v1/users", adminMaxRequests, adminWindow)
	limiter.SetPathConfig("/api/v1/add-user", adminMaxRequests, adminWindow)
	limiter.SetPathConfig("/api/v1/update-user-role", adminMaxRequests, adminWindow)
	limiter.SetPathConfig("/api/v1/update-user-status", adminMaxRequests, adminWindow)
	limiter.SetPathConfig("/api/v1/delete-user", adminMaxRequests, adminWindow)

	return limiter
}

// getEnvAsInt retrieves an environment variable and converts it to an integer.
// If the variable doesn't exist or can't be converted, it returns the fallback value.
//
// Parameters:
//   - key: The name of the environment variable to retrieve
//   - fallback: The default value to use if the environment variable is not set or invalid
//
// Returns: The environment variable as an integer, or the fallback value
func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsDuration retrieves an environment variable and converts it to a time.Duration.
// If the variable doesn't exist or can't be converted, it returns the fallback value.
//
// Parameters:
//   - key: The name of the environment variable to retrieve
//   - fallback: The default duration to use if the environment variable is not set or invalid
//
// Returns: The environment variable as a duration, or the fallback value
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

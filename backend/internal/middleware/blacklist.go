// backend/internal/middleware/blacklist.go
package middleware

import (
	"database/sql"
	"net/http"
	"strings"
)

// BlacklistMiddleware checks incoming request IPs against the blacklist
// and rejects requests from blacklisted IPs with a 502 Bad Gateway response.
// This creates a digital bouncer that keeps unwanted visitors out of our digital nightclub.
func BlacklistMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP address - the digital fingerprint
			clientIP := r.RemoteAddr
			// Check for X-Forwarded-For header for clients behind proxies
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				// Use the first IP in the chain (the original client)
				clientIP = strings.Split(forwardedFor, ",")[0]
				clientIP = strings.TrimSpace(clientIP)
			}

			// Check if IP is in the blacklist
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM blacklisted_ips WHERE ip_address = ?", clientIP).
				Scan(&count)
			if err != nil {
				// If there's a database error, let the request through
				// Better to potentially allow a bad actor than block legitimate users
				next.ServeHTTP(w, r)
				return
			}

			if count > 0 {
				// IP is blacklisted, return 502 Bad Gateway
				w.WriteHeader(http.StatusBadGateway)
				return
			}

			// IP is not blacklisted, proceed with the request
			next.ServeHTTP(w, r)
		})
	}
}

// RegistrationEnabledMiddleware checks if registration is enabled
// and blocks access to registration endpoints if it's disabled.
// This is like having a "No Vacancy" sign for your digital hotel.
func RegistrationEnabledMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip check for paths other than registration
			if !strings.HasSuffix(r.URL.Path, "/register") {
				next.ServeHTTP(w, r)
				return
			}

			// Check if registration is enabled
			var enabled string
			err := db.QueryRow("SELECT setting_value FROM system_settings WHERE setting_key = 'registration_enabled'").
				Scan(&enabled)
			if err != nil || enabled != "true" {
				// If there's a database error or registration is disabled, return error
				http.Error(w, "Registracija laikinai išjungta", http.StatusForbidden)
				return
			}

			// Registration is enabled, proceed with the request
			next.ServeHTTP(w, r)
		})
	}
}

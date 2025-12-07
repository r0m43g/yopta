// backend/internal/handlers/logs.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"yopta-template/internal/cache"
	"yopta-template/internal/models"
)

// GetLogs returns a list of logs with pagination and filtering capabilities
// This handler serves as the "time machine" for your application, letting administrators
// peer into the history of system activities and user interactions
func GetLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination parameters - because scrolling through thousands of logs
		// would be about as fun as reading the entire Oxford English Dictionary in one sitting
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page < 1 {
			page = 1 // Default to first page if parameter is invalid
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit < 1 || limit > 100 {
			limit = 20 // Default value - a reasonable amount to digest at once
		}

		// Collect filter parameters - our magnifying glass for finding specific logs
		filters := make(map[string]any)

		if userID := r.URL.Query().Get("user_id"); userID != "" {
			filters["user_id"] = userID
		}

		if method := r.URL.Query().Get("method"); method != "" {
			filters["method"] = method
		}

		if path := r.URL.Query().Get("path"); path != "" {
			filters["path"] = path
		}

		if statusCode := r.URL.Query().Get("status_code"); statusCode != "" {
			code, err := strconv.Atoi(statusCode)
			if err == nil {
				filters["status_code"] = code
			}
		}

		if clientIP := r.URL.Query().Get("client_ip"); clientIP != "" {
			filters["client_ip"] = clientIP
		}

		if fromDate := r.URL.Query().Get("from_date"); fromDate != "" {
			filters["from_date"] = fromDate
		}

		if toDate := r.URL.Query().Get("to_date"); toDate != "" {
			filters["to_date"] = toDate
		}

		// Retrieve logs from database with filters and pagination
		logs, total, err := models.GetLogs(db, page, limit, filters)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti žurnalų: "+err.Error(),
				http.StatusInternalServerError,
			) // Failed to retrieve logs
			return
		}

		// Format response with pagination metadata
		// Think of this as putting the logs into a neat binder with labeled tabs
		response := map[string]any{
			"data":      logs,
			"total":     total,
			"page":      page,
			"limit":     limit,
			"pages":     (total + limit - 1) / limit, // Calculate total pages (ceiling division)
			"timestamp": time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetLogStatistics returns aggregated statistics about system logs
// This handler gathers insights about request methods, status codes,
// most accessed paths, and user activity - it's like having a data analyst
// summarize your system's behavior in real-time
func GetLogStatistics(db *sql.DB, appCache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get time range parameters
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		if from == "" {
			// Default to last 7 days if not specified
			fromDate := time.Now().AddDate(0, 0, -7)
			from = fromDate.Format("2006-01-02T15:04:05Z")
		}

		if to == "" {
			to = time.Now().Format("2006-01-02T15:04:05Z")
		}

		// Create cache key based on parameters
		// Caching is our secret weapon for making statistics lightning fast
		cacheKey := fmt.Sprintf("log_stats_%s_%s", from, to)

		// Try to get data from cache first
		if cachedStats, found := appCache.Get(cacheKey); found {
			w.Header().Set("Content-Type", "application/json")
			w.Header().
				Set("X-Cache", "HIT")
				// Let the client know we're working smarter, not harder
			json.NewEncoder(w).Encode(cachedStats)
			return
		}

		// HTTP methods statistics
		// Which HTTP verbs are the popular kids in our API?
		var methodStats []map[string]any
		rows, err := db.Query(`
			SELECT method, COUNT(*) as count
			FROM activity_logs
			WHERE timestamp BETWEEN ? AND ?
			GROUP BY method
			ORDER BY count DESC
		`, from, to)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti metodo statistikos: "+err.Error(), // Failed to retrieve method statistics
				http.StatusInternalServerError,
			)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var method string
			var count int
			if err = rows.Scan(&method, &count); err != nil {
				continue
			}
			methodStats = append(methodStats, map[string]any{
				"method": method,
				"count":  count,
			})
		}

		// Status code statistics
		// A breakdown of success vs failures - our digital report card
		var statusStats []map[string]any
		rows, err = db.Query(`
			SELECT status_code, COUNT(*) as count
			FROM activity_logs
			WHERE timestamp BETWEEN ? AND ?
			GROUP BY status_code
			ORDER BY count DESC
		`, from, to)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti statuso statistikos: "+err.Error(), // Failed to retrieve status statistics
				http.StatusInternalServerError,
			)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var status int
			var count int
			if err = rows.Scan(&status, &count); err != nil {
				continue
			}
			statusStats = append(statusStats, map[string]any{
				"status": status,
				"count":  count,
			})
		}

		// Path statistics (top 10 most accessed)
		// Which routes are the tourist attractions in our application?
		var pathStats []map[string]any
		rows, err = db.Query(`
			SELECT path, COUNT(*) as count
			FROM activity_logs
			WHERE timestamp BETWEEN ? AND ?
			GROUP BY path
			ORDER BY count DESC
			LIMIT 10
		`, from, to)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti kelio statistikos: "+err.Error(), // Failed to retrieve path statistics
				http.StatusInternalServerError,
			)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var path string
			var count int
			if err = rows.Scan(&path, &count); err != nil {
				continue
			}
			pathStats = append(pathStats, map[string]any{
				"path":  path,
				"count": count,
			})
		}

		// User statistics (top 10 most active users)
		// Who are our power users? The VIPs of our digital world
		var userStats []map[string]any
		rows, err = db.Query(`
			SELECT user_id, COUNT(*) as count
			FROM activity_logs
			WHERE timestamp BETWEEN ? AND ? AND user_id IS NOT NULL
			GROUP BY user_id
			ORDER BY count DESC
			LIMIT 10
		`, from, to)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti vartotojo statistikos: "+err.Error(), // Failed to retrieve user statistics
				http.StatusInternalServerError,
			)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var userID string
			var count int
			if err = rows.Scan(&userID, &count); err != nil {
				continue
			}
			userStats = append(userStats, map[string]any{
				"user_id": userID,
				"count":   count,
			})
		}

		// General statistics - the big picture overview
		var totalRequests, totalErrors int
		var avgDuration float64

		err = db.QueryRow(`
			SELECT
				COUNT(*) as total,
				AVG(duration) as avg_duration,
				SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as errors
			FROM activity_logs
			WHERE timestamp BETWEEN ? AND ?
		`, from, to).Scan(&totalRequests, &avgDuration, &totalErrors)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti bendros statistikos: "+err.Error(), // Failed to retrieve general statistics
				http.StatusInternalServerError,
			)
			return
		}

		// Build complete response
		response := map[string]any{
			"from":            from,
			"to":              to,
			"total_requests":  totalRequests,
			"total_errors":    totalErrors,
			"error_rate":      float64(totalErrors) / float64(totalRequests) * 100,
			"avg_duration_ms": avgDuration,
			"method_stats":    methodStats,
			"status_stats":    statusStats,
			"path_stats":      pathStats,
			"user_stats":      userStats,
		}

		// Cache the response for 10 minutes
		// Future requests will be faster - efficiency is our superpower
		appCache.Set(cacheKey, response, 10*time.Minute)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "MISS") // Indicate cache miss
		json.NewEncoder(w).Encode(response)
	}
}

// ClearOldLogs removes log entries older than a specified date
// This handler is the digital janitor of your application, keeping the database
// clean and performant by removing dusty old logs that are no longer needed
func ClearOldLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get date before which logs should be deleted
		beforeDate := r.URL.Query().Get("before")
		if beforeDate == "" {
			// Default to logs older than 30 days if not specified
			beforeDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02T15:04:05Z")
		}

		// Delete old logs
		result, err := db.Exec("DELETE FROM activity_logs WHERE timestamp < ?", beforeDate)
		if err != nil {
			http.Error(
				w,
				"Nepavyko išvalyti žurnalų: "+err.Error(),
				http.StatusInternalServerError,
			) // Failed to clear logs
			return
		}

		// Get number of deleted rows
		rowsAffected, _ := result.RowsAffected()

		// Build response
		response := map[string]any{
			"status":       "success",
			"deleted_logs": rowsAffected,
			"before_date":  beforeDate,
			"timestamp":    time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}


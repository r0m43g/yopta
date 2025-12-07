// backend/internal/handlers/clientlogs.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// ClientLog represents the structure of a client-side log entry
// This is the contract between frontend and backend for logging data
type ClientLog struct {
	Level     string          `json:"level"`     // Severity level (debug, info, warn, error)
	Message   string          `json:"message"`   // Log message content
	Timestamp string          `json:"timestamp"` // When the log was created on client
	URL       string          `json:"url"`       // Full URL where the log was generated
	Path      string          `json:"path"`      // URL path component only
	Route     string          `json:"route"`     // Named route in the frontend router
	Data      json.RawMessage `json:"data"`      // Additional contextual data as JSON
}

// ClientLogRequest represents the batch structure for client logs
// Using batches reduces network overhead and server load
type ClientLogRequest struct {
	Logs []ClientLog `json:"logs"` // Array of logs in a single request
}

// SaveClientLogs saves logs sent from the client side
// This handler accepts an array of logs from the frontend
// and stores them in the client_logs database table
// It's our "digital historian" - documenting the user's journey!
func SaveClientLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from the request context
		userID, _ := getUserIDFromContext(r)

		// Decode the request
		var request ClientLogRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Check if there are logs to save
		if len(request.Logs) == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Nėra įrašų, kuriuos reikia išsaugoti"}`)) // No logs to save
			return
		}

		// Begin transaction for batch insertion
		// Transactions ensure all-or-nothing operation and improve performance
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}
		defer tx.Rollback()

		// Prepare statement for insertion
		stmt, err := tx.Prepare(`
			INSERT INTO client_logs (
				user_id, level, message, client_timestamp, 
				url, path, route, data, server_timestamp, 
				ip_address, user_agent
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}
		defer stmt.Close()

		// Get client IP address - digital fingerprinting at work
		clientIP := r.RemoteAddr
		forwardedFor := r.Header.Get("X-Forwarded-For")
		if forwardedFor != "" {
			clientIP = forwardedFor
		}

		// Get User-Agent
		userAgent := r.Header.Get("User-Agent")

		// Write each log to the database
		for _, log := range request.Logs {
			// Validate timestamp
			timestamp := log.Timestamp
			if timestamp == "" {
				timestamp = time.Now().Format(time.RFC3339)
			}

			// Serialize data if not empty
			var dataJSON []byte
			if log.Data != nil {
				dataJSON = log.Data
			} else {
				dataJSON = []byte("{}")
			}

			// Execute insert statement
			_, err = stmt.Exec(
				userID,
				log.Level,
				log.Message,
				timestamp,
				log.URL,
				log.Path,
				log.Route,
				dataJSON,
				time.Now().Format(time.RFC3339),
				clientIP,
				userAgent,
			)
			if err != nil {
				http.Error(
					w,
					"Klaida išsaugant žurnalą",
					http.StatusInternalServerError,
				) // Error saving logs
				return
			}
		}

		// Commit transaction - success!
		if err = tx.Commit(); err != nil {
			http.Error(
				w,
				"Klaida vykdant operaciją",
				http.StatusInternalServerError,
			) // Error committing transaction
			return
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": fmt.Sprintf(
				"Sėkmingai išsaugota %d žurnalo įrašų",
				len(request.Logs),
			), // Successfully saved X logs
			"count": len(request.Logs),
		})
	}
}

// GetClientLogs returns client logs with pagination and filtering
// This handler provides a window into user behavior and application performance
// Useful for troubleshooting, analytics, and understanding user journeys
func GetClientLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get pagination parameters
		page, limit, err := getPaginationParams(r)
		if err != nil {
			http.Error(
				w,
				"Neteisingi puslapiavimo parametrai",
				http.StatusBadRequest,
			) // Invalid pagination parameters
			return
		}

		// Build base SQL query
		query := `
			SELECT 
				id, user_id, level, message, client_timestamp, 
				url, path, route, data, server_timestamp, 
				ip_address, user_agent
			FROM client_logs 
			WHERE 1=1
		`
		countQuery := `SELECT COUNT(*) FROM client_logs WHERE 1=1`

		// Add filters
		var params []interface{}
		query, countQuery, params = applyClientLogFilters(r, query, countQuery, params)

		// Add sorting and pagination
		query += " ORDER BY server_timestamp DESC LIMIT ? OFFSET ?"
		params = append(params, limit, (page-1)*limit)

		// Get total record count
		var total int
		err = db.QueryRow(countQuery, params[:len(params)-2]...).Scan(&total)
		if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}

		// Get records
		rows, err := db.Query(query, params...)
		if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}
		defer rows.Close()

		// Process results
		logs := []map[string]interface{}{}
		for rows.Next() {
			var (
				id              int64
				userID          sql.NullInt64
				level           string
				message         string
				clientTimestamp string
				url             sql.NullString
				path            sql.NullString
				route           sql.NullString
				data            []byte
				serverTimestamp string
				ipAddress       string
				userAgent       sql.NullString
			)

			if err := rows.Scan(
				&id, &userID, &level, &message, &clientTimestamp,
				&url, &path, &route, &data, &serverTimestamp,
				&ipAddress, &userAgent,
			); err != nil {
				http.Error(
					w,
					"Duomenų bazės klaida",
					http.StatusInternalServerError,
				) // Database error
				return
			}

			// Unpack JSON data
			var dataMap map[string]interface{}
			if err := json.Unmarshal(data, &dataMap); err != nil {
				dataMap = map[string]interface{}{"raw": string(data)}
			}

			// Build log entry
			log := map[string]interface{}{
				"id":               id,
				"level":            level,
				"message":          message,
				"client_timestamp": clientTimestamp,
				"server_timestamp": serverTimestamp,
				"ip_address":       ipAddress,
				"data":             dataMap,
			}

			// Add optional fields
			if userID.Valid {
				log["user_id"] = userID.Int64
			}
			if url.Valid {
				log["url"] = url.String
			}
			if path.Valid {
				log["path"] = path.String
			}
			if route.Valid {
				log["route"] = route.String
			}
			if userAgent.Valid {
				log["user_agent"] = userAgent.String
			}

			logs = append(logs, log)
		}

		// Build response
		response := map[string]interface{}{
			"data":       logs,
			"total":      total,
			"page":       page,
			"limit":      limit,
			"page_count": (total + limit - 1) / limit,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// DeleteClientLogs removes old client logs
// This handler helps maintain database performance by removing outdated logs
// Like digital spring cleaning - keeping our log storage neat and tidy!
func DeleteClientLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get date before which to delete logs
		beforeDate := r.URL.Query().Get("before")
		if beforeDate == "" {
			// Default to deleting logs older than 30 days
			beforeDate = time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
		}

		// Delete old logs
		result, err := db.Exec("DELETE FROM client_logs WHERE server_timestamp < ?", beforeDate)
		if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}

		// Get number of deleted records
		rowsAffected, _ := result.RowsAffected()

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"deleted":     rowsAffected,
			"before_date": beforeDate,
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	}
}

// GetClientLogStatistics returns statistics about client logs
// This handler provides aggregate data about log levels, paths, and users
// It's like having a data analyst on standby - always ready with insights!
func GetClientLogStatistics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get time range
		fromDate := r.URL.Query().Get("from")
		toDate := r.URL.Query().Get("to")

		if fromDate == "" {
			// Default to 7 days ago
			fromDate = time.Now().AddDate(0, 0, -7).Format(time.RFC3339)
		}

		if toDate == "" {
			toDate = time.Now().Format(time.RFC3339)
		}

		// Get statistics by log level
		levelStats, err := getLevelStatistics(db, fromDate, toDate)
		if err != nil {
			http.Error(
				w,
				"Klaida gaunant lygio statistiką",
				http.StatusInternalServerError,
			) // Error fetching level statistics
			return
		}

		// Get statistics by path
		pathStats, err := getPathStatistics(db, fromDate, toDate)
		if err != nil {
			http.Error(
				w,
				"Klaida gaunant kelio statistiką",
				http.StatusInternalServerError,
			) // Error fetching path statistics
			return
		}

		// Get statistics by user
		userStats, err := getUserStatistics(db, fromDate, toDate)
		if err != nil {
			http.Error(
				w,
				"Klaida gaunant vartotojo statistiką",
				http.StatusInternalServerError,
			) // Error fetching user statistics
			return
		}

		// Get overall statistics
		var totalLogs int
		var errorCount int
		err = db.QueryRow(`
			SELECT 
				COUNT(*) as total,
				SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END) as errors
			FROM client_logs
			WHERE server_timestamp BETWEEN ? AND ?
		`, fromDate, toDate).Scan(&totalLogs, &errorCount)
		if err != nil {
			http.Error(
				w,
				"Klaida gaunant bendrą statistiką",
				http.StatusInternalServerError,
			) // Error fetching general statistics
			return
		}

		// Build response
		response := map[string]interface{}{
			"from":         fromDate,
			"to":           toDate,
			"total_logs":   totalLogs,
			"error_count":  errorCount,
			"error_rate":   float64(errorCount) / float64(totalLogs) * 100,
			"level_stats":  levelStats,
			"path_stats":   pathStats,
			"user_stats":   userStats,
			"generated_at": time.Now().Format(time.RFC3339),
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetClientLogByID returns a specific client log by its ID
// This handler provides detailed information about a single log entry
// Useful for drilling down into specific issues or events
func GetClientLogByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get log ID from URL
		logID := chi.URLParam(r, "id")
		if logID == "" {
			http.Error(w, "Reikalingas žurnalo ID", http.StatusBadRequest) // Log ID required
			return
		}

		// Get log from DB
		var (
			id              int64
			userID          sql.NullInt64
			level           string
			message         string
			clientTimestamp string
			url             sql.NullString
			path            sql.NullString
			route           sql.NullString
			data            []byte
			serverTimestamp string
			ipAddress       string
			userAgent       sql.NullString
		)

		err := db.QueryRow(`
			SELECT 
				id, user_id, level, message, client_timestamp, 
				url, path, route, data, server_timestamp, 
				ip_address, user_agent
			FROM client_logs 
			WHERE id = ?
		`, logID).Scan(
			&id, &userID, &level, &message, &clientTimestamp,
			&url, &path, &route, &data, &serverTimestamp,
			&ipAddress, &userAgent,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Žurnalas nerastas", http.StatusNotFound) // Log not found
			return
		} else if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}

		// Unpack JSON data
		var dataMap map[string]interface{}
		if err := json.Unmarshal(data, &dataMap); err != nil {
			dataMap = map[string]interface{}{"raw": string(data)}
		}

		// Build response
		log := map[string]interface{}{
			"id":               id,
			"level":            level,
			"message":          message,
			"client_timestamp": clientTimestamp,
			"server_timestamp": serverTimestamp,
			"ip_address":       ipAddress,
			"data":             dataMap,
		}

		// Add optional fields
		if userID.Valid {
			log["user_id"] = userID.Int64
		}
		if url.Valid {
			log["url"] = url.String
		}
		if path.Valid {
			log["path"] = path.String
		}
		if route.Valid {
			log["route"] = route.String
		}
		if userAgent.Valid {
			log["user_agent"] = userAgent.String
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(log)
	}
}

// Helper functions

// getPaginationParams extracts pagination parameters from the request
// Returns page number, limit per page, and error if any
func getPaginationParams(r *http.Request) (int, int, error) {
	// Get page number
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = parseInt(pageStr)
		if err != nil || page < 1 {
			return 0, 0, fmt.Errorf("invalid page parameter")
		}
	}

	// Get items per page
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		limit, err = parseInt(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return 0, 0, fmt.Errorf("invalid limit parameter")
		}
	}

	return page, limit, nil
}

// parseInt converts a string to an integer
// Simple wrapper around fmt.Sscanf for better error handling
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// applyClientLogFilters applies filters to SQL query
// This function dynamically builds SQL WHERE conditions based on request parameters
// Returns updated query, count query, and parameters
func applyClientLogFilters(
	r *http.Request,
	query, countQuery string,
	params []interface{},
) (string, string, []interface{}) {
	// Filter by log level
	if level := r.URL.Query().Get("level"); level != "" {
		query += " AND level = ?"
		countQuery += " AND level = ?"
		params = append(params, level)
	}

	// Filter by message content
	if message := r.URL.Query().Get("message"); message != "" {
		query += " AND message LIKE ?"
		countQuery += " AND message LIKE ?"
		params = append(params, "%"+message+"%")
	}

	// Filter by path
	if path := r.URL.Query().Get("path"); path != "" {
		query += " AND path LIKE ?"
		countQuery += " AND path LIKE ?"
		params = append(params, "%"+path+"%")
	}

	// Filter by route
	if route := r.URL.Query().Get("route"); route != "" {
		query += " AND route = ?"
		countQuery += " AND route = ?"
		params = append(params, route)
	}

	// Filter by user ID
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		query += " AND user_id = ?"
		countQuery += " AND user_id = ?"
		params = append(params, userID)
	}

	// Filter by IP address
	if ipAddress := r.URL.Query().Get("ip_address"); ipAddress != "" {
		query += " AND ip_address LIKE ?"
		countQuery += " AND ip_address LIKE ?"
		params = append(params, "%"+ipAddress+"%")
	}

	// Filter by time range (client timestamp)
	if fromDate := r.URL.Query().Get("from"); fromDate != "" {
		query += " AND client_timestamp >= ?"
		countQuery += " AND client_timestamp >= ?"
		params = append(params, fromDate)
	}

	if toDate := r.URL.Query().Get("to"); toDate != "" {
		query += " AND client_timestamp <= ?"
		countQuery += " AND client_timestamp <= ?"
		params = append(params, toDate)
	}

	return query, countQuery, params
}

// getLevelStatistics returns statistics by log level
// Shows distribution of logs across different severity levels
func getLevelStatistics(db *sql.DB, fromDate, toDate string) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
		SELECT 
			level, 
			COUNT(*) as count
		FROM client_logs
		WHERE server_timestamp BETWEEN ? AND ?
		GROUP BY level
		ORDER BY count DESC
	`, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []map[string]interface{}{}
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		stats = append(stats, map[string]interface{}{
			"level": level,
			"count": count,
		})
	}

	return stats, nil
}

// getPathStatistics returns statistics by URL path
// Shows which parts of the application are most frequently used
func getPathStatistics(db *sql.DB, fromDate, toDate string) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
		SELECT 
			path, 
			COUNT(*) as count
		FROM client_logs
		WHERE server_timestamp BETWEEN ? AND ? 
		AND path IS NOT NULL
		GROUP BY path
		ORDER BY count DESC
		LIMIT 10
	`, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []map[string]interface{}{}
	for rows.Next() {
		var path sql.NullString
		var count int
		if err := rows.Scan(&path, &count); err != nil {
			return nil, err
		}
		if path.Valid {
			stats = append(stats, map[string]interface{}{
				"path":  path.String,
				"count": count,
			})
		}
	}

	return stats, nil
}

// getUserStatistics returns statistics by user
// Shows which users are most active in the system
func getUserStatistics(db *sql.DB, fromDate, toDate string) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
		SELECT
			cl.user_id,
			u.username,
			COUNT(*) as count
		FROM client_logs cl
		LEFT JOIN users u ON cl.user_id = u.id
		WHERE cl.server_timestamp BETWEEN ? AND ?
		AND cl.user_id IS NOT NULL
		GROUP BY cl.user_id
		ORDER BY count DESC
		LIMIT 10
	`, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []map[string]interface{}{}
	for rows.Next() {
		var userID int64
		var username sql.NullString
		var count int
		if err := rows.Scan(&userID, &username, &count); err != nil {
			return nil, err
		}

		userStat := map[string]interface{}{
			"user_id": userID,
			"count":   count,
		}

		if username.Valid {
			userStat["username"] = username.String
		} else {
			userStat["username"] = "Nežinomas vartotojas" // Unknown User
		}

		stats = append(stats, userStat)
	}

	return stats, nil
}


// backend/internal/models/log.go
package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// Global connection variables that allow the logging system to access
// the database throughout the application lifecycle
var (
	dbConnection *sql.DB    // Database connection used by the logging system
	dbMutex      sync.Mutex // Mutex to ensure thread-safe access to the database connection
)

// LogEntry represents the structure of a log record in the system.
// This comprehensive structure captures detailed information about HTTP requests and responses,
// providing a full audit trail of system activity, user actions, and potential security incidents.
// Each field serves a specific purpose in the logging and monitoring ecosystem:
type LogEntry struct {
	ID           int64  `json:"id,omitempty"`            // Auto-generated database ID, omitted when zero
	Timestamp    string `json:"timestamp"`               // When the request occurred (ISO 8601/RFC 3339 format)
	Method       string `json:"method"`                  // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path         string `json:"path"`                    // Request URL path
	Query        string `json:"query,omitempty"`         // URL query parameters, omitted when empty
	UserID       any    `json:"user_id,omitempty"`       // ID of the authenticated user, or nil for guests
	UserRole     any    `json:"user_role,omitempty"`     // Role of the authenticated user, or nil for guests
	ClientIP     string `json:"client_ip"`               // Client's IP address for geographic and security analysis
	UserAgent    string `json:"user_agent,omitempty"`    // Browser/client identification string
	RequestID    string `json:"request_id,omitempty"`    // Unique identifier to trace a request through the system
	RequestBody  string `json:"request_body,omitempty"`  // Content of the request (sanitized for sensitive data)
	ResponseBody string `json:"response_body,omitempty"` // Content returned to the client
	StatusCode   int    `json:"status_code"`             // HTTP response status code
	Duration     int64  `json:"duration_ms"`             // Request processing time in milliseconds
	Error        string `json:"error,omitempty"`         // Error message if the request failed
}

// SetDBConnection establishes a database connection for the logging system.
// This function should be called during application initialization to provide
// the logging system with access to the database.
//
// The mutex ensures that concurrent calls to this function don't create race conditions.
// This is essential in high-concurrency environments where multiple goroutines
// might attempt to set or access the connection simultaneously.
//
// Parameters:
//   - db: An initialized sql.DB connection to be used for logging operations
func SetDBConnection(db *sql.DB) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	dbConnection = db
}

// GetDBConnection retrieves the database connection for logging operations.
// This function should be used by logging components to obtain the shared
// database connection when they need to perform database operations.
//
// The mutex ensures thread-safe access to the connection, preventing
// potential race conditions in concurrent environments.
//
// Returns:
//   - The shared database connection or nil if not yet established
func GetDBConnection() *sql.DB {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	return dbConnection
}

// SaveLogEntry persists a log entry to the database.
// This function handles the conversion of complex data types and null values
// to ensure proper storage in the database. It's designed to be resilient against
// different data formats that might be encountered when logging heterogeneous requests.
//
// Parameters:
//   - db: The database connection to use
//   - entry: The LogEntry to save
//
// Returns:
//   - error: Any error that occurred during the database operation
func SaveLogEntry(db *sql.DB, entry LogEntry) error {
	// Convert userID to string representation if it's a number
	var userIDStr, userRoleStr sql.NullString

	if entry.UserID != nil {
		// Check if userID is a number
		switch v := entry.UserID.(type) {
		case int:
			userIDStr = sql.NullString{String: fmt.Sprintf("%d", v), Valid: true}
		case float64:
			userIDStr = sql.NullString{String: fmt.Sprintf("%d", int(v)), Valid: true}
		default:
			// If not a number, serialize as JSON
			userIDJSON, _ := json.Marshal(entry.UserID)
			userIDStr = sql.NullString{String: string(userIDJSON), Valid: true}
		}
	}

	if entry.UserRole != nil {
		userRoleJSON, _ := json.Marshal(entry.UserRole)
		userRoleStr = sql.NullString{String: string(userRoleJSON), Valid: true}
	}

	// Prepare and execute SQL query
	stmt, err := db.Prepare(`
		INSERT INTO activity_logs (
			timestamp, method, path, query, user_id, user_role,
			client_ip, user_agent, request_id, request_body,
			response_body, status_code, duration, error
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf(
			"Klaida ruošiant SQL užklausą žurnalo įrašui: %v",
			err,
		) // Error preparing SQL query for log
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		entry.Timestamp,
		entry.Method,
		entry.Path,
		entry.Query,
		userIDStr,
		userRoleStr,
		entry.ClientIP,
		entry.UserAgent,
		entry.RequestID,
		entry.RequestBody,
		entry.ResponseBody,
		entry.StatusCode,
		entry.Duration,
		entry.Error,
	)
	if err != nil {
		log.Printf(
			"Klaida vykdant SQL užklausą žurnalo įrašui: %v",
			err,
		) // Error executing SQL query for log
		return err
	}

	return nil
}

// GetLogs retrieves logs from the database with pagination and filtering.
// This function supports flexible filtering options for advanced log analysis
// and implements pagination to efficiently handle large volumes of log data.
//
// Parameters:
//   - db: The database connection
//   - page: The page number (1-based) for pagination
//   - limit: Maximum number of records to return per page
//   - filters: Map of column names to filter values
//
// Returns:
//   - []LogEntry: Array of log entries matching the criteria
//   - int: Total number of matching records (before pagination)
//   - error: Any error that occurred during the database operation
func GetLogs(db *sql.DB, page, limit int, filters map[string]any) ([]LogEntry, int, error) {
	// Build the base SQL query
	baseQuery := `
		SELECT
			id, timestamp, method, path, query, user_id, user_role,
			client_ip, user_agent, request_id, request_body,
			response_body, status_code, duration, error
		FROM activity_logs
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM activity_logs WHERE 1=1"

	// Parameters for prepared statement
	var params []any

	// Add filter conditions
	for key, value := range filters {
		switch key {
		case "user_id":
			baseQuery += " AND user_id = ?"
			countQuery += " AND user_id = ?"
			params = append(params, value)
		case "method":
			baseQuery += " AND method = ?"
			countQuery += " AND method = ?"
			params = append(params, value)
		case "path":
			baseQuery += " AND path LIKE ?"
			countQuery += " AND path LIKE ?"
			params = append(params, fmt.Sprintf("%%%v%%", value))
		case "status_code":
			baseQuery += " AND status_code = ?"
			countQuery += " AND status_code = ?"
			params = append(params, value)
		case "client_ip":
			baseQuery += " AND client_ip = ?"
			countQuery += " AND client_ip = ?"
			params = append(params, value)
		case "from_date":
			baseQuery += " AND timestamp >= ?"
			countQuery += " AND timestamp >= ?"
			params = append(params, value)
		case "to_date":
			baseQuery += " AND timestamp <= ?"
			countQuery += " AND timestamp <= ?"
			params = append(params, value)
		}
	}

	// Count total matching records
	var total int
	err := db.QueryRow(countQuery, params...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination and sorting
	baseQuery += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	params = append(params, limit, (page-1)*limit)

	// Execute the query
	rows, err := db.Query(baseQuery, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Process results
	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		var userIDStr, userRoleStr sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Method,
			&log.Path,
			&log.Query,
			&userIDStr,
			&userRoleStr,
			&log.ClientIP,
			&log.UserAgent,
			&log.RequestID,
			&log.RequestBody,
			&log.ResponseBody,
			&log.StatusCode,
			&log.Duration,
			&log.Error,
		)
		if err != nil {
			return nil, 0, err
		}

		// Convert string user_id to number if possible
		if userIDStr.Valid {
			// Try to interpret as integer
			var userID int
			if _, err := fmt.Sscanf(userIDStr.String, "%d", &userID); err == nil {
				log.UserID = userID
			} else {
				// If that fails, try as JSON
				var userIDValue any
				if err := json.Unmarshal([]byte(userIDStr.String), &userIDValue); err == nil {
					log.UserID = userIDValue
				} else {
					// If that fails too, use the string as is
					log.UserID = userIDStr.String
				}
			}
		}

		// Convert string user_role if possible
		if userRoleStr.Valid {
			var userRole any
			if err := json.Unmarshal([]byte(userRoleStr.String), &userRole); err == nil {
				log.UserRole = userRole
			} else {
				log.UserRole = userRoleStr.String
			}
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}


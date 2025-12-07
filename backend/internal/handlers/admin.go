package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"yopta-template/internal/cache"
	"yopta-template/internal/models"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

// Users retrieves all users from the database for the admin panel.
// This handler queries the database for user information excluding sensitive data like passwords.
// It returns a JSON array of all users with their attributes (id, username, email, role, theme, etc.).
// This endpoint is intended for administrative use only and should be protected by appropriate
// authorization middleware.
// Cache implementation optimizes performance for frequently accessed data.
func Users(db *sql.DB, appCache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try to get data from cache first - working smarter, not harder!
		if cachedUsers, found := appCache.Get("all_users"); found {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT") // Add header indicating cache hit
			json.NewEncoder(w).Encode(cachedUsers)
			return
		}

		// If not in cache, fetch from database
		rows, err := db.Query(
			"SELECT users.id, username, email, role, theme, avatar, verified, user_status, created_at FROM users LEFT JOIN profiles ON users.id = user_id",
		)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti vartotojų sąrašo",
				http.StatusInternalServerError,
			) // Failed to retrieve users
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var user models.User
			err = rows.Scan(
				&user.ID,
				&user.Username,
				&user.Email,
				&user.Role,
				&user.Theme,
				&user.Avatar,
				&user.Verified,
				&user.UserStatus,
				&user.CreatedAt,
			)
			if err != nil {
				http.Error(
					w,
					"Nepavyko gauti vartotojų sąrašo",
					http.StatusInternalServerError,
				) // Failed to retrieve users
				return
			}
			users = append(users, user)
		}

		// Save result in cache (for 5 minutes)
		// This significantly reduces database load for frequently accessed data
		appCache.Set("all_users", users, 5*time.Minute)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "MISS") // Add header indicating cache miss
		json.NewEncoder(w).Encode(users)
	}
}

// AddUser creates a new user in the database via the admin panel.
// This handler processes a JSON request containing user details, hashes the password,
// and adds the user to the database. It returns the created user (without password)
// in the response body with a 201 Created status.
// This endpoint is intended for administrative use only and should be protected
// by appropriate authorization middleware.
func AddUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(
				w,
				"Nepavyko pridėti vartotojo: neteisingas užklausos formatas",
				http.StatusInternalServerError,
			) // Failed to add user: invalid request format
			return
		}

		// Hash the password - never store passwords in plain text!
		// Security is not optional, it's a requirement
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(user.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida apdorojant slaptažodį",
				http.StatusInternalServerError,
			) // Error processing password
			return
		}
		res, err := db.Exec(
			"INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)",
			user.Username,
			user.Email,
			hashedPassword,
			user.Role,
		)
		if err != nil {
			http.Error(
				w,
				"Nepavyko pridėti vartotojo",
				http.StatusInternalServerError,
			) // Failed to add user
			return
		}

		// Retrieve the inserted user id to create a profile
		userID, err := res.LastInsertId()
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti vartotojo ID",
				http.StatusInternalServerError,
			) // Failed to retrieve user ID
			return
		}

		// Create user profile with default values
		_, err = db.Exec(
			"INSERT INTO profiles (user_id, theme, avatar) VALUES (?, ?, ?)",
			userID,
			user.Theme,
			"none",
		)
		if err != nil {
			http.Error(
				w,
				"Nepavyko sukurti vartotojo profilio",
				http.StatusInternalServerError,
			) // Failed to add user profile
			return
		}

		// Clear password before sending the response
		// We should never, ever send passwords back to the client
		user.Password = ""
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

// DeleteUser removes a user from the database by their ID.
// This handler also removes the user's avatar file if exists.
// It performs a cascade delete that removes all associated data due to
// foreign key constraints set up in the database schema.
func DeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(
				w,
				"Nenurodytas vartotojo ID",
				http.StatusBadRequest,
			) // User ID not specified
			return
		}

		// Get the path to the user's avatar before deletion
		var avatarPath string
		err := db.QueryRow("SELECT avatar FROM profiles WHERE user_id = ?", id).Scan(&avatarPath)

		// Delete the user (and profile cascadingly due to ON DELETE CASCADE)
		_, err = db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			http.Error(
				w,
				"Nepavyko ištrinti vartotojo",
				http.StatusInternalServerError,
			) // Failed to delete user
			return
		}

		// If avatar path exists and is not 'none', delete the avatar file
		// This prevents orphaned files and keeps storage clean
		if err == nil && avatarPath != "" && avatarPath != "none" &&
			!strings.HasPrefix(avatarPath, "http") {
			// Delete avatar file
			if deleteErr := DeleteUserImage(avatarPath); deleteErr != nil {
				// Log error but don't abort user deletion
				log.Printf("Klaida trinant vartotojo %s avatarą: %v", id, deleteErr)
			}
		}

		w.Write([]byte("Vartotojas sėkmingai ištrintas")) // User deleted successfully
	}
}

// UpdateUserRole updates a user's role in the system.
// This handler accepts a JSON with user ID and new role,
// validates the role, and updates the data in the database.
// Access to this endpoint should be restricted to administrators only.
func UpdateUserRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type roleRequest struct {
			UserID int    `json:"user_id"`
			Role   string `json:"role"`
		}

		var req roleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Validate role (allowable values from DB schema)
		// Role validation is critical for security - we can't allow arbitrary role values!
		if req.Role != "admin" && req.Role != "user" && req.Role != "viewer" {
			http.Error(
				w,
				"Nurodyta neteisinga rolė",
				http.StatusBadRequest,
			) // Invalid role specified
			return
		}

		// Check if user exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", req.UserID).
			Scan(&exists)
		if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}

		if !exists {
			http.Error(w, "Vartotojas nerastas", http.StatusNotFound) // User not found
			return
		}

		// Update user role
		_, err = db.Exec("UPDATE users SET role = ? WHERE id = ?", req.Role, req.UserID)
		if err != nil {
			http.Error(
				w,
				"Nepavyko atnaujinti vartotojo rolės",
				http.StatusInternalServerError,
			) // Failed to update user role
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Vartotojo rolė sėkmingai atnaujinta", // User role updated successfully
		})
	}
}

// UpdateUserStatus updates a user's status in the system.
// This handler accepts a JSON with user ID and new status,
// validates the status, and updates the data in the database.
// Access to this endpoint should be restricted to administrators only.
func UpdateUserStatus(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type statusRequest struct {
			UserID int    `json:"user_id"`
			Status string `json:"status"`
		}

		var req statusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Validate status (allowable values from DB schema)
		// Similar to role validation, we must restrict to predefined values
		if req.Status != "active" && req.Status != "inactive" && req.Status != "suspended" {
			http.Error(
				w,
				"Nurodytas neteisingas statusas",
				http.StatusBadRequest,
			) // Invalid status specified
			return
		}

		// Check if user exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", req.UserID).
			Scan(&exists)
		if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}

		if !exists {
			http.Error(w, "Vartotojas nerastas", http.StatusNotFound) // User not found
			return
		}

		// Update user status
		_, err = db.Exec("UPDATE users SET user_status = ? WHERE id = ?", req.Status, req.UserID)
		if err != nil {
			http.Error(
				w,
				"Nepavyko atnaujinti vartotojo statuso",
				http.StatusInternalServerError,
			) // Failed to update user status
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Vartotojo statusas sėkmingai atnaujintas", // User status updated successfully
		})
	}
}

// ClearCache completely flushes the application's memory cache.
// This is useful for administrators who need to force fresh data retrieval
// after significant data changes or when troubleshooting cache-related issues.
// Like emptying the cookie jar - sometimes you just need a clean slate!
func ClearCache(appCache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCache.Flush()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Talpykla sėkmingai išvalyta")) // Cache cleared successfully
	}
}

// UpdateUserVerification updates a user's email verification status in the system.
// This handler accepts a JSON with user ID and new verification status,
// validates the input, and updates the data in the database.
// Access to this endpoint should be restricted to administrators only.
//
// Parameters in request:
// - user_id: The ID of the user whose verification status will be changed
// - verified: Boolean value indicating the new verification status
//
// Returns a success message if the update was successful or an appropriate error message.
func UpdateUserVerification(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type verificationRequest struct {
			UserID   int  `json:"user_id"`
			Verified bool `json:"verified"`
		}

		var req verificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Check if user exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", req.UserID).
			Scan(&exists)
		if err != nil {
			http.Error(w, "Duomenų bazės klaida", http.StatusInternalServerError) // Database error
			return
		}

		if !exists {
			http.Error(w, "Vartotojas nerastas", http.StatusNotFound) // User not found
			return
		}

		// Update user verification status
		_, err = db.Exec("UPDATE users SET verified = ? WHERE id = ?", req.Verified, req.UserID)
		if err != nil {
			http.Error(
				w,
				"Nepavyko atnaujinti vartotojo patvirtinimo būsenos",
				http.StatusInternalServerError,
			) // Failed to update user verification status
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Vartotojo patvirtinimo būsena sėkmingai atnaujinta", // User verification status successfully updated
		})
	}
}

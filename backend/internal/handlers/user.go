// backend/internal/handlers/user.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// UserProfile represents a user's profile information that's safe to return to clients.
// This acts as our "digital passport" for users - containing their identity and preferences
// while carefully excluding sensitive information like passwords.
type UserProfile struct {
	ID       int    `json:"id"`       // Unique identifier
	Username string `json:"username"` // Display name
	Email    string `json:"email"`    // Email address
	Role     string `json:"role"`     // User role (admin, user, etc.)
	Theme    string `json:"theme"`    // UI theme preference
	Avatar   string `json:"avatar"`   // Profile image path
}

// AuthPing returns HTTP 200 OK status if the user is authenticated.
// This handler is like a "digital doorbell" - a simple way to check
// if the user's authentication session is still active and valid.
// It can be used by frontend applications to verify session status
// without retrieving any actual data.
func AuthPing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// GetProfile retrieves the user profile data from the database.
// This handler fetches user details including username, email, role, theme and avatar
// based on the user ID extracted from the request context (provided by JWT middleware).
// Think of it as opening a user's "digital business card" - showing their public information.
func GetProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromContext(r)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti vartotojo ID",
				http.StatusUnauthorized,
			) // Failed to retrieve user ID
			return
		}

		var profile UserProfile
		err = db.QueryRow("SELECT users.id, username, email, role, theme, avatar FROM users LEFT JOIN profiles ON user_id = users.id WHERE users.id = ?", userID).
			Scan(&profile.ID, &profile.Username, &profile.Email, &profile.Role, &profile.Theme, &profile.Avatar)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Vartotojas nerastas", http.StatusNotFound) // User not found
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}
}

// UpdateProfileTheme updates the user's theme preference in the database.
// This handler allows users to personalize their experience by changing
// the visual theme of the application - like choosing the wallpaper
// for their digital workspace.
func UpdateProfileTheme(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromContext(r)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti vartotojo ID",
				http.StatusUnauthorized,
			) // Failed to retrieve user ID
			return
		}

		var profile UserProfile
		if err = json.NewDecoder(r.Body).Decode(&profile); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Update the theme in the profiles table
		_, err = db.Exec(
			"UPDATE profiles SET theme = ? WHERE user_id = ?",
			profile.Theme,
			userID,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida atnaujinant profilį",
				http.StatusInternalServerError,
			) // Error updating profile
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Profilis sėkmingai atnaujintas")) // Profile successfully updated
	}
}

// PasswordChangeRequest represents the structure of a password change request
// containing both the old password for verification and the new password to set.
// This structure ensures we have all the information needed for securely changing passwords.
type PasswordChangeRequest struct {
	OldPassword string `json:"old_password"` // Current password for verification
	NewPassword string `json:"new_password"` // New password to set
}

// ChangePassword allows users to change their password after verifying their existing password.
// This handler is like a "digital locksmith" that follows a strict protocol:
// 1. Verifies the user's identity from the request context
// 2. Validates the request format
// 3. Retrieves the current hashed password from the database
// 4. Compares the old password with the stored hash to authenticate
// 5. Validates the new password (minimum length check)
// 6. Hashes the new password securely with bcrypt
// 7. Updates the password in the database
//
// Returns a success message if all steps complete successfully.
func ChangePassword(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromContext(r)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti vartotojo ID",
				http.StatusUnauthorized,
			) // Failed to retrieve user ID
			return
		}

		var req PasswordChangeRequest
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Get current hashed password from database
		var storedHashedPassword string
		err = db.QueryRow("SELECT password FROM users WHERE id = ?", userID).
			Scan(&storedHashedPassword)
		if err != nil {
			http.Error(w, "Vartotojas nerastas", http.StatusNotFound) // User not found
			return
		}

		// Verify old password - the first line of defense against unauthorized changes
		if err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(req.OldPassword)); err != nil {
			http.Error(
				w,
				"Neteisingas dabartinis slaptažodis",
				http.StatusUnauthorized,
			) // Incorrect old password
			return
		}

		// Validate new password - quality control for security
		if len(req.NewPassword) < 8 {
			http.Error(
				w,
				"Naujas slaptažodis turi būti bent 8 simbolių ilgio",
				http.StatusBadRequest,
			) // New password must be at least 8 characters long
			return
		}

		// Hash new password - never store passwords in plain text!
		newHashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(req.NewPassword),
			bcrypt.DefaultCost,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida apdorojant naują slaptažodį",
				http.StatusInternalServerError,
			) // Error processing new password
			return
		}

		// Update password in database
		_, err = db.Exec(
			"UPDATE users SET password = ? WHERE id = ?",
			string(newHashedPassword),
			userID,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida atnaujinant slaptažodį",
				http.StatusInternalServerError,
			) // Error updating password
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Slaptažodis sėkmingai atnaujintas")) // Password successfully updated
	}
}

// getUserIDFromContext extracts the user ID from the request context.
// The user ID is set in the context by the JWT authentication middleware.
// This helper ensures we're always accessing the right user's data.
// Returns the user ID as an integer if found, or an error if not present or invalid.
func getUserIDFromContext(r *http.Request) (int, error) {
	// User ID is set by JWT middleware
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		return 0, http.ErrNoCookie
	}
	userID, ok := userIDValue.(int)
	if !ok {
		return 0, http.ErrNoCookie
	}
	return userID, nil
}

// Test is a simple handler for testing purposes.
// Sometimes we all need a simple "Hello World" equivalent
// to verify our system is working as expected!
func Test() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test"))
	}
}


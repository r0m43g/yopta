// backend/internal/handlers/profile.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// UpdateProfileAvatar handles the request to update a user's avatar image.
// This handler is like a digital photo studio: it takes the raw image data in base64 format,
// processes it to ensure quality and security, saves it to disk, and updates the user's profile.
// The image becomes the user's digital face in our application - their visual identity!
func UpdateProfileAvatar(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the user ID from the JWT token stored in the request context
		// This ensures we're updating the correct user's profile
		userID, err := getUserIDFromContext(r)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti vartotojo ID",
				http.StatusUnauthorized,
			) // Failed to retrieve user ID
			return
		}

		// Parse the request body to get the avatar data
		var profile struct {
			Avatar string `json:"avatar"`
		}
		if err = json.NewDecoder(r.Body).Decode(&profile); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Validate the avatar data - empty avatars need not apply!
		if profile.Avatar == "" {
			http.Error(w, "Avataro duomenys tušti", http.StatusBadRequest) // Avatar data is empty
			return
		}

		// Ensure the data is actually base64 image data and not a file path
		// Security first - we don't want path traversal attacks!
		if !strings.HasPrefix(profile.Avatar, "data:") {
			http.Error(
				w,
				"Neteisingas avataro duomenų formatas",
				http.StatusBadRequest,
			) // Invalid avatar data format
			return
		}

		// Process and upload the image to the server
		// This transforms the raw base64 string into a proper image file
		avatarPath, err := UploadBase64Image(profile.Avatar, userID, "avatars")
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Klaida įkeliant avatarą: %v", err), // Error uploading avatar
				http.StatusInternalServerError,
			)
			return
		}

		// Get the current avatar to delete it if it exists
		// We're keeping our server tidy - no orphaned avatar files allowed!
		var currentAvatar string
		if err := db.QueryRow("SELECT avatar FROM profiles WHERE user_id = ?", userID).Scan(&currentAvatar); err == nil {
			// If current avatar exists and is not the default "none"
			if currentAvatar != "none" && currentAvatar != "" &&
				!strings.Contains(currentAvatar, "yopta") {
				// Skip deletion for default avatars or external URLs
				if !strings.HasPrefix(currentAvatar, "http") {
					DeleteUserImage(currentAvatar)
				}
			}
		}

		// Update the profile record in the database with the new avatar path
		_, err = db.Exec(
			"UPDATE profiles SET avatar = ? WHERE user_id = ?",
			avatarPath,
			userID,
		)
		if err != nil {
			fmt.Println(err)
			http.Error(
				w,
				"Klaida atnaujinant profilį",
				http.StatusInternalServerError,
			) // Error updating profile
			return
		}

		// Return success response with the new avatar path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Avataras sėkmingai atnaujintas", // Avatar updated successfully
			"avatar":  avatarPath,
		})
	}
}

// DeleteProfileAvatar removes a user's avatar and sets it back to the default value.
// This handler is the digital equivalent of removing a portrait from a wall -
// it cleans up the image file from the server and resets the user's visual representation
// back to the default state.
func DeleteProfileAvatar(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the request context
		userID, err := getUserIDFromContext(r)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti vartotojo ID",
				http.StatusUnauthorized,
			) // Failed to retrieve user ID
			return
		}

		// Retrieve the current avatar path
		var currentAvatar string
		if err := db.QueryRow("SELECT avatar FROM profiles WHERE user_id = ?", userID).Scan(&currentAvatar); err == nil {
			// If avatar exists and is not the default "none"
			if currentAvatar != "none" && currentAvatar != "" &&
				!strings.Contains(currentAvatar, "yopta") {
				// Skip deletion for default avatars or external URLs
				if !strings.HasPrefix(currentAvatar, "http") {
					DeleteUserImage(currentAvatar)
				}
			}
		}

		// Reset avatar to default value in the database
		_, err = db.Exec(
			"UPDATE profiles SET avatar = 'none' WHERE user_id = ?",
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

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Avataras sėkmingai pašalintas", // Avatar removed successfully
		})
	}
}


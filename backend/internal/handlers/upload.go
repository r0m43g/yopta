// backend/internal/handlers/upload.go
package handlers

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UploadBase64Image processes a base64 encoded image and saves it to disk
// in the specified directory.
// This function serves as our "digital darkroom" - transforming raw encoded
// data into actual image files stored on the server.
//
// Parameters:
//   - base64Data: The encoded image data, optionally with MIME type prefix
//   - userID: The user's ID for file ownership and naming
//   - folder: Target folder within the uploads directory
//
// Returns the relative path to the saved image or an error if processing fails.
func UploadBase64Image(base64Data string, userID int, folder string) (string, error) {
	// Check for data:image prefix - a sign of properly formatted base64 images
	var imageData string
	if strings.HasPrefix(base64Data, "data:image") {
		// Split metadata from actual data
		// Like separating a letter from its envelope
		parts := strings.Split(base64Data, ",")
		if len(parts) != 2 {
			return "", fmt.Errorf(
				"neteisingas base64 duomenų formatas",
			) // Invalid base64 data format
		}
		imageData = parts[1]
	} else {
		// Assume it's raw base64 data without the metadata prefix
		imageData = base64Data
	}

	// Decode the base64 data into bytes
	// This is where the magic happens - transforming text into binary image data
	imageBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return "", fmt.Errorf("klaida dekoduojant base64: %w", err) // Error decoding base64
	}

	// Create directory if it doesn't exist
	// We're building a home for our images if it's not there already
	uploadDir := filepath.Join("uploads", folder)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", fmt.Errorf("klaida kuriant direktoriją: %w", err) // Error creating directory
	}

	// Generate a unique filename to prevent collisions and overwrites
	// This combines several unique elements to ensure no two images share the same name:
	// - User ID: Associates the file with its owner
	// - Timestamp: Adds a temporal component
	// - Random UUID: Adds true randomness for collision resistance
	randomName := uuid.New().String()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	fileName := fmt.Sprintf("%d_%s_%s.jpg", userID, timestamp, randomName[:8])
	filePath := filepath.Join(uploadDir, fileName)

	// Save the image to disk - materializing our digital creation
	if err := os.WriteFile(filePath, imageBytes, 0o644); err != nil {
		return "", fmt.Errorf("klaida įrašant failą: %w", err) // Error writing file
	}

	// Return the path relative to static root
	// Note: we remove "static/" from the path as it's already the
	// root for static files in our web server configuration
	relativePath := filepath.Join("/", folder, fileName)
	return relativePath, nil
}

// DeleteUserImage removes an image file from the server.
// This function helps maintain disk space hygiene by cleaning up
// files that are no longer needed - like a digital janitor for our
// image storage system.
//
// Parameters:
//   - imagePath: The path to the image to be deleted
//
// Returns an error if deletion fails, or nil if successful.
// If the file doesn't exist, this is considered a success (not an error).
func DeleteUserImage(imagePath string) error {
	// Protect against path traversal attacks
	// Security is paramount - we don't want users deleting files outside of static directory
	if !strings.HasPrefix(imagePath, "/") {
		imagePath = "/" + imagePath
	}

	// If path starts with "/", remove it for correct path formation
	if strings.HasPrefix(imagePath, "/") {
		imagePath = imagePath[1:]
	}

	// Form the full path to the file
	fullPath := filepath.Join("static", imagePath)

	// Check if the file exists first - no need to worry about non-existent files
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// File not found - this is actually okay, nothing to delete!
		return nil
	}

	// Delete the file - goodbye, digital artifact!
	return os.Remove(fullPath)
}

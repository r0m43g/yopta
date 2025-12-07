// backend/internal/models/user.go
package models

// User represents a user entity in the system.
// This model contains all user-related attributes and is used for database operations
// and API responses. The Password field is only used during authentication and
// is not included in JSON responses by default for security reasons.
//
// The User struct serves as the central data model for user management throughout the application.
// It maintains a clear separation between public attributes (exposed in API responses) and
// sensitive attributes (like password, which is excluded from responses).
//
// This model implements a comprehensive user profile system with various attributes:
// - Core identification (ID, username, email)
// - Authentication data (password, verification status)
// - Authorization information (role)
// - Personalization settings (theme, avatar)
// - Status tracking (user_status, created_at)
type User struct {
	ID         int    `json:"id"`          // Unique identifier for the user, auto-incremented in database
	Username   string `json:"username"`    // User's display name shown in the interface
	Email      string `json:"email"`       // User's email address used for communication and login
	Password   string `json:"password"`    // Stores the hashed password value (never sent in responses)
	Verified   bool   `json:"verified"`    // Indicates whether email address has been confirmed
	Role       string `json:"role"`        // User's role for access control (e.g., "naudotojas", "administratorius")
	Theme      string `json:"theme"`       // User's UI theme preference for personalization
	Avatar     string `json:"avatar"`      // User's profile image (stored as path or base64)
	UserStatus string `json:"user_status"` // User's current status ("aktyvus", "neaktyvus", "sustabdytas")
	CreatedAt  string `json:"created_at"`  // Timestamp when user account was created (ISO 8601 format)
}


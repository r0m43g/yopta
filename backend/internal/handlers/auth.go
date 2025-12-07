// backend/internal/handlers/auth.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"yopta-template/internal/utils"

	regexp "github.com/dlclark/regexp2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	VerifyCode string `json:"verify_code"`
}

// passwordRegex defines the password strength requirements:
// - At least 8 characters long (security 101: longer is stronger!)
// - Contains at least one letter (uppercase or lowercase)
// - Contains at least one digit (because '1' is not the same as 'one')
// - Contains at least one special character (!@#$%^&*) - adds unpredictability
// Together, these requirements create a decent balance between security and usability
var passwordRegex = regexp.MustCompile(`^(?=.*[A-Za-z])(?=.*\d)(?=.*[!@#$%^&*]).{8,}$`, 0)

// Register handles the registration of a new user.
// This handler:
// 1. Validates the request payload
// 2. Checks password strength using the regex pattern
// 3. Hashes the password securely
// 4. Creates a new user in the database
// 5. Generates a verification code for email confirmation
// 6. Sends a verification email to the user
// The handler requires a database connection and JWT configuration parameters.
func Register(db *sql.DB, jwtSecret, jwtExpiry string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Validate password strength - our first line of defense against weak passwords
		if match, _ := passwordRegex.MatchString(creds.Password); !match {
			http.Error(
				w,
				"Silpnas slaptažodis: mažiausiai 8 simboliai, turi būti skaičius ir specialus simbolis",
				http.StatusBadRequest,
			) // Weak password: minimum 8 characters, must include a digit and special character
			return
		}

		// Password hashing - one of the most crucial security measures
		// We use bcrypt for its adaptive nature and built-in salt management
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(creds.Password),
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
		creds.Role = "user" // Default role for newly registered users

		// Insert user into database - welcome to our system!
		res, err := db.Exec(
			"INSERT INTO users (username, email, password, role, verified) VALUES (?, ?, ?, ?, 0)",
			creds.Username,
			creds.Email,
			string(hashedPassword),
			creds.Role,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida registruojant vartotoją",
				http.StatusInternalServerError,
			) // Error registering user
			return
		}
		userID, err := res.LastInsertId()
		if err != nil {
			http.Error(
				w,
				"Klaida gaunant naują vartotojo ID",
				http.StatusInternalServerError,
			) // Error retrieving new user ID
			return
		}

		// Create user profile - their digital home in our system
		_, err = db.Exec(
			"INSERT INTO profiles (user_id, theme, avatar) VALUES (?, 'dim', 'none')",
			userID,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida kuriant vartotojo profilį",
				http.StatusInternalServerError,
			) // Error creating user profile
			return
		}

		// Generate verification code - like a digital doorbell, must be rung to enter
		verifyCode := uuid.NewString()[:8]
		expiresAt := time.Now().Add(15 * time.Minute)
		_, err = db.Exec(
			"INSERT INTO email_verifications (user_id, code, expires_at) VALUES (?, ?, ?)",
			userID, verifyCode, expiresAt,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida kuriant patvirtinimo kodą",
				http.StatusInternalServerError,
			) // Error creating verification code
			return
		}

		// Prepare email content using template
		data := map[string]any{
			"Name": creds.Username,
			"Code": verifyCode,
		}
		message, err := utils.RenderTemplate("internal/templates/mail.html", data)
		if err != nil {
			fmt.Println(err)
			http.Error(
				w,
				"Klaida apdorojant šabloną",
				http.StatusInternalServerError,
			) // Error rendering template
			return
		}

		// Send verification email
		err = utils.SendEmail(creds.Email, "Registracija Yopta.top", message)
		if err != nil {
			fmt.Println(err)
			http.Error(
				w,
				"Klaida siunčiant el. laišką",
				http.StatusInternalServerError,
			) // Error sending email
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(
			[]byte(
				"Vartotojas sukurtas. Patikrinkite savo el. paštą, kad gautumėte patvirtinimo kodą",
			),
		) // User created. Check your email for verification code
	}
}

// VerifyEmail handles email verification using a verification code.
// This handler:
// 1. Receives a verification code from the request
// 2. Validates the code against the database
// 3. Checks if the code has expired
// 4. Updates the user's verified status in the database
// 5. Removes the used verification code from the database
// Returns a success message if verification is successful.
func VerifyEmail(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Find the code in the email_verifications table
		var userID int
		var expiresAt time.Time
		err := db.QueryRow(
			"SELECT user_id, expires_at FROM email_verifications WHERE code = ?",
			creds.VerifyCode,
		).Scan(&userID, &expiresAt)
		if err != nil {
			fmt.Println(err)
			http.Error(
				w,
				"Neteisingas patvirtinimo kodas",
				http.StatusUnauthorized,
			) // Invalid verification code
			return
		}

		// Check if code has expired - time waits for no verification code!
		if time.Now().After(expiresAt) {
			fmt.Println(err)
			http.Error(
				w,
				"Patvirtinimo kodas baigėsi",
				http.StatusUnauthorized,
			) // Verification code expired
			return
		}

		// Update verified status in users table
		_, err = db.Exec("UPDATE users SET verified = 1 WHERE id = ?", userID)
		if err != nil {
			http.Error(
				w,
				"Klaida patvirtinant el. paštą",
				http.StatusInternalServerError,
			) // Error confirming email
			return
		}

		// Delete the used code (one-time use)
		_, _ = db.Exec("DELETE FROM email_verifications WHERE code = ?", creds.VerifyCode)

		w.WriteHeader(http.StatusOK)
		w.Write(
			[]byte("El. paštas patvirtintas. Dabar galite prisijungti."),
		) // Email verified. You can now log in.
	}
}

// Login handles user authentication.
// This handler:
// 1. Validates the provided credentials
// 2. Retrieves the user from the database
// 3. Verifies that the user's email is confirmed
// 4. Compares the password with the stored hash
// 5. Generates a JWT access token
// 6. Creates and stores a refresh token
// Returns both tokens as JSON if authentication is successful.
func Login(db *sql.DB, jwtSecret, jwtExpiry string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(
				w,
				"Neteisingas užklausos formatas",
				http.StatusBadRequest,
			) // Invalid request format
			return
		}

		// Retrieve user data from database
		var id int
		var email string
		var storedHashedPassword string
		var username string
		var role string
		var verified bool
		var user_status string
		err := db.QueryRow("SELECT id, username, email, password, role, verified, user_status FROM users WHERE email = ?", creds.Email).
			Scan(&id, &username, &email, &storedHashedPassword, &role, &verified, &user_status)
		if err != nil {
			http.Error(
				w,
				"Neteisingi prisijungimo duomenys",
				http.StatusUnauthorized,
			) // Invalid credentials
			return
		}

		// Check if email is verified - security is a process, not just a password
		if !verified {
			http.Error(w, "El. paštas nepatvirtintas", http.StatusForbidden) // Email not verified
			return
		}

		// Check if user is active - inactive users need not apply
		if user_status != "active" {
			http.Error(
				w,
				"Vartotojas sustabdytas arba užblokuotas. Susisiekite su administratoriumi.",
				http.StatusForbidden,
			) // User is suspended or blocked. Please contact administration.
			return
		}

		// Compare password with stored hash - the moment of truth!
		if err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(creds.Password)); err != nil {
			http.Error(
				w,
				"Neteisingi prisijungimo duomenys",
				http.StatusUnauthorized,
			) // Invalid credentials
			return
		}

		// Generate tokens - the keys to our digital kingdom
		accessToken, err := generateJWT(id, role, jwtSecret, jwtExpiry)
		if err != nil {
			http.Error(
				w,
				"Klaida generuojant prieigos raktą",
				http.StatusInternalServerError,
			) // Error generating token
			return
		}

		refreshToken, err := storeRefreshToken(db, id)
		if err != nil {
			http.Error(
				w,
				"Klaida su atnaujinimo raktu",
				http.StatusInternalServerError,
			) // Error with refresh token
			return
		}

		// Send tokens to client
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	}
}

// RefreshToken issues a new access token using a valid refresh token.
// This handler:
// 1. Retrieves the refresh token from the X-Refresh-Token header
// 2. Validates the token against the database
// 3. Checks if the token has expired
// 4. Generates a new JWT access token
// 5. Optionally creates a new refresh token
// 6. Invalidates the old refresh token
// Returns both new tokens as JSON if the refresh is successful.
func RefreshToken(db *sql.DB, jwtSecret, jwtExpiry string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get refresh token from X-Refresh-Token header
		refresh := r.Header.Get("X-Refresh-Token")
		if refresh == "" {
			http.Error(
				w,
				"Trūksta X-Refresh-Token antraštės",
				http.StatusUnauthorized,
			) // Missing X-Refresh-Token header
			return
		}

		// Check if token exists in database
		var userID int
		var role string
		var expiresAt time.Time
		err := db.QueryRow(
			"SELECT user_id, role, expires_at FROM refresh_tokens LEFT JOIN users ON users.id = user_id WHERE token = ?",
			refresh,
		).Scan(&userID, &role, &expiresAt)
		if err != nil {
			http.Error(
				w,
				"Neteisingas atnaujinimo raktas",
				http.StatusUnauthorized,
			) // Invalid refresh token
			return
		}

		// Check expiration - even refresh tokens have an expiration date
		if time.Now().After(expiresAt) {
			http.Error(
				w,
				"Atnaujinimo raktas baigėsi",
				http.StatusUnauthorized,
			) // Refresh token expired
			return
		}

		// Delete old refresh token to prevent reuse - security is a clean house
		_, _ = db.Exec("DELETE FROM refresh_tokens WHERE user_id = ?", userID)
		// Generate new access token - fresh keys for a new session
		newAccess, err := generateJWT(userID, role, jwtSecret, jwtExpiry)
		if err != nil {
			http.Error(
				w,
				"Klaida generuojant prieigos raktą",
				http.StatusInternalServerError,
			) // Error generating access token
			return
		}

		// Generate new refresh token (otherwise old one would be valid until expiration)
		newRefresh, err := storeRefreshToken(db, userID)
		if err != nil {
			http.Error(
				w,
				"Klaida generuojant naują atnaujinimo raktą",
				http.StatusInternalServerError,
			) // Error generating new refresh token
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token":  newAccess,
			"refresh_token": newRefresh,
		})
	}
}

// generateJWT creates a JSON Web Token with specified expiration time and user details.
// This function:
// 1. Parses the expiration duration from the provided string
// 2. Calculates the expiration timestamp
// 3. Creates a JWT with claims (user_id, role, expiration)
// 4. Signs the token with the secret key using HMAC-SHA256
// Returns the signed token string or an error.
func generateJWT(userID int, role string, jwtSecret, jwtExpiry string) (string, error) {
	duration, err := time.ParseDuration(jwtExpiry)
	if err != nil {
		return "", err
	}
	expirationTime := time.Now().Add(duration)

	// Create JWT claims - the payload of our token
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     expirationTime.Unix(),
	}

	// Create a new token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// storeRefreshToken creates and stores a refresh token for a user in the database.
// This function:
// 1. Generates a UUID as the refresh token
// 2. Sets an expiration time (12 hours from creation)
// 3. Stores the token in the database with the user ID and expiration
// Returns the token string if successful, or an error if the database operation fails.
func storeRefreshToken(db *sql.DB, userID int) (string, error) {
	// Generate UUID as refresh token - universally unique, collision-resistant
	token := uuid.NewString()
	expires := time.Now().Add(12 * time.Hour) // 12 hours

	_, err := db.Exec(
		"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expires,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

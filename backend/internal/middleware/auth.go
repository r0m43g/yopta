// backend/internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware authenticates and authorizes requests by validating JWT tokens.
// This middleware functions as a security gateway for protected routes, ensuring
// that only users with valid credentials can access sensitive functionality.
//
// The authentication flow works as follows:
// 1. Extracts the JWT token from the Authorization header (Bearer format)
// 2. Validates the token signature using the provided secret key
// 3. Verifies the token's validity and expiration time
// 4. Extracts user information (ID and role) from the token claims
// 5. Stores this information in the request context for downstream handlers
//
// If validation fails at any step, the middleware returns an appropriate error response
// and prevents the request from reaching protected handlers.
func JWTMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header - the digital key to our system
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(
					w,
					"Trūksta autorizacijos rakto",
					http.StatusUnauthorized,
				) // Missing authorization token
				return
			}

			// Validate token format (should be "Bearer [token]")
			// This ensures we're receiving tokens in the expected format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(
					w,
					"Neteisingas rakto formatas",
					http.StatusUnauthorized,
				) // Invalid token format
				return
			}

			tokenString := parts[1]

			// Parse and validate the token using our secret key
			// The secret key is like our digital signature that verifies authentic tokens
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Verify signing method is correct to prevent algorithm switching attacks
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrTokenInvalidId
				}
				return []byte(jwtSecret), nil
			})

			// If token is invalid or expired, deny access
			if err != nil || !token.Valid {
				http.Error(
					w,
					"Negaliojantis arba pasibaigęs raktas",
					http.StatusUnauthorized,
				) // Invalid or expired token
				return
			}

			// Extract user_id from token claims
			// This information identifies the user making the request
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// Parse user ID from the claims
				userIDFloat, ok := claims["user_id"].(float64)
				if !ok {
					http.Error(
						w,
						"Neteisingi rakto duomenys",
						http.StatusUnauthorized,
					) // Invalid token data
					return
				}

				// Extract role from token claims
				// Role determines the user's permissions in the system
				role, ok := claims["role"].(string)
				if !ok {
					http.Error(
						w,
						"Neteisingi rakto duomenys",
						http.StatusUnauthorized,
					) // Invalid token data
					return
				}

				userID := int(userIDFloat)

				// Create custom context types to ensure type safety
				type ctxRole string
				type ctxUserID int

				// Store user_id in request context for later use by handlers
				// This makes user identification available throughout the request lifecycle
				ctx := context.WithValue(r.Context(), "user_id", userID)

				// Store role in request context
				// This enables role-based access control in downstream handlers
				ctx = context.WithValue(ctx, "role", role)

				// Update the request with the enhanced context
				r = r.WithContext(ctx)
			} else {
				http.Error(w, "Neteisingi rakto duomenys", http.StatusUnauthorized) // Invalid token data
				return
			}

			// If authentication successful, proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}


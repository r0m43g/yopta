// backend/internal/middleware/role.go
package middleware

import (
	"net/http"
)

// RoleMiddleware creates middleware for checking a user's role permissions.
// This middleware is designed to be executed after JWTMiddleware, ensuring
// that the user's role is already established in the request context.
//
// The middleware implements a role-based access control (RBAC) system
// that restricts access to certain endpoints based on user roles.
// By defining allowed roles for specific routes, we can enforce
// proper authorization throughout the application.
//
// Parameters:
//   - allowedRoles: A variable-length slice of strings containing the roles
//     that are permitted to access the protected endpoint. This allows for
//     flexible configuration of permissions for different API routes.
//
// Returns an http.Handler middleware function that checks if the user has
// one of the allowed roles. If not, it returns a 403 Forbidden error.
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the role from the context, which was set by JWTMiddleware
			role, ok := r.Context().Value("role").(string)
			if !ok {
				http.Error(
					w,
					"Neautorizuota: trūksta rolės informacijos",
					http.StatusUnauthorized,
				) // Unauthorized: role information missing
				return
			}

			// Check if the user's role is in the list of allowed roles
			allowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(
					w,
					"Draudžiama: nepakanka teisių",
					http.StatusForbidden,
				) // Forbidden: insufficient permissions
				return
			}

			// If the role is appropriate, pass control to the next handler
			next.ServeHTTP(w, r)
		})
	}
}


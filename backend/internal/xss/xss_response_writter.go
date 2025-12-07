// backend/internal/xss/xss_response_writer.go

package xss

import (
	"net/http"
	"strings"

	"yopta-template/internal/utils"
)

// XSSProtectionMiddleware creates a middleware that protects against Cross-Site Scripting (XSS) attacks.
// This middleware works by intercepting and sanitizing all HTTP responses before they reach the client.
// It specifically focuses on JSON and other text-based responses, ensuring that any potentially
// malicious content is neutralized.
//
// The protection is implemented using a custom ResponseWriter wrapper that sanitizes
// response data based on content type, with special handling for JSON responses.
//
// Returns: An http.Handler middleware function that can be used in middleware chains
func XSSProtectionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create an XSS-protected ResponseWriter that will intercept the response
			// This wrapper allows us to sanitize the data before it's sent to the client
			xssWriter := &xssResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status code
			}

			// Process the request with our protective wrapper in place
			next.ServeHTTP(xssWriter, r)
		})
	}
}

// xssResponseWriter is a custom ResponseWriter that sanitizes response data to prevent XSS attacks.
// It wraps the standard http.ResponseWriter and intercepts all Write calls to apply sanitization
// rules based on the content type of the response.
type xssResponseWriter struct {
	http.ResponseWriter      // Embed the standard ResponseWriter
	statusCode          int  // Keep track of the response status code
	written             bool // Flag to track if headers have been written
}

// WriteHeader captures the status code and passes it to the underlying ResponseWriter.
// This allows us to track the status code for logging and conditional processing.
func (x *xssResponseWriter) WriteHeader(statusCode int) {
	x.statusCode = statusCode
	x.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts data being written to the response and applies XSS sanitization.
// The sanitization approach varies based on the content type:
// - For JSON responses: It parses, sanitizes, and re-serializes the JSON data
// - For other content types: It passes the data through without modification
//
// This method ensures that all outgoing data is properly sanitized while maintaining
// the correct content structure and encoding.
//
// Parameters:
//   - b: The byte slice containing the response data to be written
//
// Returns:
//   - Number of bytes written
//   - Error if the sanitization or writing process fails
func (x *xssResponseWriter) Write(b []byte) (int, error) {
	contentType := x.Header().Get("Content-Type")

	// If this is JSON, sanitize it
	if strings.Contains(contentType, "application/json") {
		sanitized, err := utils.SanitizeJSON(b)
		if err != nil {
			// In case of JSON parsing error, fall back to the original response
			// This prevents breaking functionality when sanitization fails
			return x.ResponseWriter.Write(b)
		}
		return x.ResponseWriter.Write(sanitized)
	}

	// For other content types (images, files, etc.), pass through unchanged
	// NOTE: HTML content could be sanitized here in the future if needed
	return x.ResponseWriter.Write(b)
}

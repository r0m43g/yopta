// backend/internal/utils/sanitizer.go

package utils

import (
	"encoding/json"
	"html"
	"strings"
)

// SanitizeString cleanses a string from potentially dangerous XSS vectors.
// This function applies multiple layers of protection:
// 1. Basic HTML escaping of special characters (<, >, &, ", ')
// 2. Blocking of common attack vectors like javascript: and data: URI schemes
//
// The function is designed to be used on any user-supplied strings before
// they're included in HTML responses or stored in the database.
//
// Parameters:
//   - input: The string to be sanitized
//
// Returns: A sanitized version of the input string with dangerous elements neutralized
func SanitizeString(input string) string {
	// Apply standard HTML escaping for basic protection
	// This converts <, >, &, ", ' to their HTML entity equivalents
	escaped := html.EscapeString(input)

	// Additional protection against common attack vectors
	// Blocks javascript: and data: URI schemes which can be used for XSS
	escaped = strings.ReplaceAll(escaped, "javascript:", "blocked:")
	escaped = strings.ReplaceAll(escaped, "data:", "blocked:")

	return escaped
}

// SanitizeJSON sanitizes JSON data to prevent XSS attacks.
// This function works by:
// 1. Unmarshaling the JSON into a Go data structure
// 2. Recursively sanitizing all string values
// 3. Re-marshaling the data back to JSON
//
// This approach ensures that even deeply nested JSON objects are properly sanitized.
//
// Parameters:
//   - jsonData: Raw JSON data bytes to be sanitized
//
// Returns:
//   - Sanitized JSON data bytes
//   - Error if JSON parsing fails
func SanitizeJSON(jsonData []byte) ([]byte, error) {
	var data interface{}

	// Parse JSON into a generic interface{} structure
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	// Sanitize all values in the data structure
	sanitizedData := sanitizeValue(data)

	// Convert back to JSON
	return json.Marshal(sanitizedData)
}

// sanitizeValue recursively sanitizes each value in a data structure.
// It handles different types appropriately:
// - Strings are sanitized using SanitizeString
// - Arrays/slices have each element sanitized
// - Maps/objects have each value sanitized
// - Other primitive types (numbers, booleans, null) are left unchanged
//
// This approach ensures complete sanitization of complex nested structures.
//
// Parameters:
//   - value: The value to sanitize (can be any JSON-compatible type)
//
// Returns: The sanitized value with the same structure but cleaned strings
func sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// For strings, apply our string sanitization function
		return SanitizeString(v)
	case []interface{}:
		// For arrays/slices, recursively sanitize each element
		sanitizedSlice := make([]interface{}, len(v))
		for i, item := range v {
			sanitizedSlice[i] = sanitizeValue(item)
		}
		return sanitizedSlice
	case map[string]interface{}:
		// For maps/objects, recursively sanitize each value
		sanitizedMap := make(map[string]interface{})
		for key, val := range v {
			sanitizedMap[key] = sanitizeValue(val)
		}
		return sanitizedMap
	default:
		// For primitive types (numbers, bool, null), leave as is
		return v
	}
}


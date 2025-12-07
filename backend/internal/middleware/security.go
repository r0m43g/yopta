// backend/internal/middleware/security.go

package middleware

import (
	"net/http"
)

// SecurityHeaders adds essential HTTP security headers to protect against common web vulnerabilities.
// This middleware implements defense-in-depth security measures based on OWASP best practices.
//
// The middleware adds several critical security headers:
//   - Content-Security-Policy: Controls which resources can be loaded, preventing XSS attacks
//   - X-XSS-Protection: Enables the browser's built-in XSS auditors for older browsers
//   - X-Content-Type-Options: Prevents MIME-sniffing attacks where browsers may interpret files as
//     a different content-type than declared
//   - X-Frame-Options: Prevents clickjacking attacks by controlling if the page can be framed
//   - Referrer-Policy: Controls how much referrer information is included with requests
//   - Strict-Transport-Security: Forces HTTPS usage, enhancing transport security
//
// These headers work together to create multiple layers of protection against common
// web vulnerabilities, significantly improving the application's security posture.
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content-Security-Policy (CSP) - Restricts the sources from which resources can be loaded
			// This helps prevent cross-site scripting (XSS) and related attacks by strictly controlling
			// which domains can serve executable scripts, styles, fonts, etc.
			w.Header().
				Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:;")

			// X-XSS-Protection - Activates the browser's built-in XSS filter
			// While modern browsers rely more on CSP, this header provides backward compatibility
			// with older browsers. The "mode=block" directive stops page rendering if an attack is detected.
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// X-Content-Type-Options - Prevents MIME type sniffing
			// Browsers sometimes try to "sniff" the MIME type of a resource, potentially
			// executing a malicious file with an incorrect content type. This header
			// forces browsers to use the declared content type.
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// X-Frame-Options - Protection against clickjacking
			// Prevents the page from being embedded in frames/iframes on other domains,
			// which helps protect against UI redress attacks where malicious sites
			// overlay invisible elements over legitimate UI elements.
			w.Header().Set("X-Frame-Options", "DENY")

			// Referrer-Policy - Controls referrer information in HTTP requests
			// This limits the information sent in the Referer header when navigating
			// to other origins, protecting user privacy and preventing sensitive
			// information leakage through URLs.
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Strict-Transport-Security - Forces HTTPS usage
			// HSTS tells browsers to only use HTTPS for future visits, even if HTTP
			// links are clicked or typed. This is enabled only in production environments
			// to prevent issues during development.
			if isProd := false; isProd { // Replace with environment check in production
				w.Header().
					Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			next.ServeHTTP(w, r)
		})
	}
}


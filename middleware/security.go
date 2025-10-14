package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Content Security Policy
		e.Response.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com https://myhelcim.com; style-src 'self' 'unsafe-inline' https://unpkg.com https://maxcdn.bootstrapcdn.com; img-src 'self' data: https: blob:; font-src 'self' https://maxcdn.bootstrapcdn.com; connect-src 'self' https://myhelcim.com; frame-ancestors 'none';")

		// X-Frame-Options
		e.Response.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options
		e.Response.Header().Set("X-Content-Type-Options", "nosniff")

		// Referrer Policy
		e.Response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature Policy)
		e.Response.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		return next(e)
	}
}

// HTTPSEnforcement redirects HTTP requests to HTTPS in production
func HTTPSEnforcement(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Only enforce HTTPS in production
		if os.Getenv("GO_ENV") == "production" {
			// Check if request is already HTTPS
			if e.Request.TLS == nil {
				// Check for X-Forwarded-Proto header (for proxies/load balancers)
				if proto := e.Request.Header.Get("X-Forwarded-Proto"); proto == "http" || proto == "" {
					// Redirect to HTTPS
					httpsURL := "https://" + e.Request.Host + e.Request.URL.Path
					if e.Request.URL.RawQuery != "" {
						httpsURL += "?" + e.Request.URL.RawQuery
					}
					e.Response.Header().Set("Location", httpsURL)
					return e.Redirect(http.StatusMovedPermanently, httpsURL)
				}
			}

			// Add HSTS header for HTTPS requests
			e.Response.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		return next(e)
	}
}

// ValidateInput sanitizes and validates common input patterns
func ValidateInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Basic length check (adjust as needed)
	if len(input) > 10000 {
		input = input[:10000]
	}

	return input
}

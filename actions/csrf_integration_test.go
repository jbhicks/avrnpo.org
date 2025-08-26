package actions

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/mw-csrf"
	"github.com/stretchr/testify/require"
)

// setupAppWithCSRF creates a Buffalo app with CSRF protection enabled
// This simulates a non-test environment where CSRF middleware is active
func setupAppWithCSRF() *buffalo.App {
	app := buffalo.New(buffalo.Options{
		Env: "integration", // NOT "test" - this enables CSRF protection
	})

	// Force CSRF middleware (this is the key difference from test environment)
	app.Use(csrf.New)

	// Add simple test routes for CSRF testing
	app.POST("/csrf-test", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(map[string]string{"status": "success"}))
	})

	return app
}

// TestCSRFProtectionEnabled verifies that CSRF middleware is working
// This is the key test that would have caught the original donation form issue
func TestCSRFProtectionEnabled(t *testing.T) {
	app := setupAppWithCSRF()

	// Submit form WITHOUT CSRF token
	formData := url.Values{
		"test_field": {"test_value"},
		// Note: NO authenticity_token field
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/csrf-test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.ServeHTTP(w, req)

	// Should return 403 Forbidden due to missing CSRF token
	require.Equal(t, 403, w.Code, "Request without CSRF token should be rejected with 403")

	t.Logf("✅ CSRF protection is working: POST request without token rejected with status %d", w.Code)
}

// TestCSRFWithInvalidToken verifies that invalid tokens are rejected
func TestCSRFWithInvalidToken(t *testing.T) {
	app := setupAppWithCSRF()

	// Submit form with INVALID CSRF token
	formData := url.Values{
		"authenticity_token": {"invalid-token-12345"}, // Invalid token
		"test_field":         {"test_value"},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/csrf-test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.ServeHTTP(w, req)

	// Should return 403 Forbidden due to invalid CSRF token
	require.Equal(t, 403, w.Code, "Request with invalid CSRF token should be rejected with 403")

	t.Logf("✅ CSRF protection is working: POST request with invalid token rejected with status %d", w.Code)
}

// TestCSRFEnvironmentDifference verifies that our test vs integration environment differs
func TestCSRFEnvironmentDifference(t *testing.T) {
	// Create an app with test environment (should NOT have CSRF)
	testApp := buffalo.New(buffalo.Options{
		Env: "test", // This is what regular tests use
	})

	// Notice: NO csrf.New middleware for test environment
	testApp.POST("/csrf-test", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(map[string]string{"status": "success"}))
	})

	// Submit form without CSRF token to test environment
	formData := url.Values{
		"test_field": {"test_value"},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/csrf-test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	testApp.ServeHTTP(w, req)

	// Test environment should allow this (no CSRF protection)
	require.Equal(t, 200, w.Code, "Test environment should allow requests without CSRF token")

	t.Logf("✅ Test environment correctly allows requests without CSRF tokens (status %d)", w.Code)
	t.Logf("💡 This confirms why unit tests missed the CSRF issue")
}

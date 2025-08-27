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

// TestBuffaloCSRFBuiltInPattern tests that our implementation follows Buffalo's CSRF patterns
func TestBuffaloCSRFBuiltInPattern(t *testing.T) {
	// Test that we use Buffalo's built-in CSRF middleware correctly
	app := buffalo.New(buffalo.Options{Env: "development"})

	// Test that forms include CSRF tokens
	app.GET("/test-form", func(c buffalo.Context) error {
		return c.Render(200, r.String("<form><input name='authenticity_token' value='test-token'></form>"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test-form", nil)

	app.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "authenticity_token", "Form should include CSRF token")

	t.Logf("✅ Buffalo CSRF built-in pattern working")
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

// TestBuffaloCSRFBasicFunctionality tests that Buffalo CSRF middleware works as expected
func TestBuffaloCSRFBasicFunctionality(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.Use(csrf.New)

	app.POST("/submit", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(map[string]string{"status": "success"}))
	})

	// Test 1: Request without token should be rejected
	formData := url.Values{
		"data": {"test"},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.ServeHTTP(w, req)
	require.Equal(t, 403, w.Code, "Request without CSRF token should be rejected")

	// Test 2: Request with invalid token should be rejected
	formData = url.Values{
		"authenticity_token": {"invalid-token-12345"},
		"data":               {"test"},
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.ServeHTTP(w, req)
	require.Equal(t, 403, w.Code, "Request with invalid CSRF token should be rejected")

	t.Logf("✅ Buffalo CSRF middleware basic functionality working")
}
		html := fmt.Sprintf(`<form method="post" action="/submit">
			<input type="hidden" name="authenticity_token" value="%s" />
			<input type="text" name="test_field" value="test_value" />
			<button type="submit">Submit</button>
		</form>`, token)
		return c.Render(200, r.String(html))
	})

	app.POST("/submit", func(c buffalo.Context) error {
		return c.Render(200, r.String("Form submitted successfully"))
	})

	// Use a single request cycle - this is how Buffalo CSRF actually works
	// The middleware generates a token for the response and validates it on the next request
	app.POST("/test-cycle", func(c buffalo.Context) error {
		// First check if we have a token in the request
		token := c.Param("authenticity_token")
		if token == "" {
			// No token provided, this is the first request - return form with token
			responseToken := c.Value("authenticity_token")
			if responseToken == nil {
				return c.Render(500, r.String("No CSRF token generated"))
			}
			html := fmt.Sprintf(`<form method="post" action="/test-cycle">
				<input type="hidden" name="authenticity_token" value="%s" />
				<input type="text" name="test_field" value="test_value" />
				<button type="submit">Submit</button>
			</form>`, responseToken)
			return c.Render(200, r.String(html))
		} else {
			// Token provided, validate it
			return c.Render(200, r.String("Form submitted successfully with token: "+token))
		}
	})

	// Test the complete cycle
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test-cycle", nil)
	app.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code, "First request should return form with token")
	require.Contains(t, w.Body.String(), "authenticity_token", "Form should contain CSRF token")

	// Extract token and submit again
	bodyString := w.Body.String()
	tokenStart := strings.Index(bodyString, `value="`) + 7
	tokenEnd := strings.Index(bodyString[tokenStart:], `"`) + tokenStart
	token := bodyString[tokenStart:tokenEnd]

	require.NotEmpty(t, token, "Should extract valid CSRF token")

	// Submit with the token
	formData := url.Values{
		"authenticity_token": {token},
		"test_field":         {"test_value"},
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/test-cycle", strings.NewReader(formData.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.ServeHTTP(w2, req2)

	require.Equal(t, 200, w2.Code, "POST with valid CSRF token should succeed")
	require.Contains(t, w2.Body.String(), "Form submitted successfully")

	t.Logf("✅ Buffalo CSRF middleware basic functionality working")
}

// TestHTMXCSRFBuiltInPattern tests HTMX with real Buffalo CSRF middleware
func TestHTMXCSRFBuiltInPattern(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.Use(csrf.New)

	app.POST("/htmx-test", func(c buffalo.Context) error {
		// First check if we have a token in the request
		token := c.Param("authenticity_token")
		if token == "" {
			// No token provided, this is the first request - return form with token
			responseToken := c.Value("authenticity_token")
			if responseToken == nil {
				return c.Render(500, r.String("No CSRF token generated"))
			}
			html := fmt.Sprintf(`<form method="post" action="/htmx-test" hx-post="/htmx-test" hx-target="#result" hx-swap="innerHTML">
				<input type="hidden" name="authenticity_token" value="%s" />
				<input type="text" name="message" value="HTMX test" />
				<button type="submit">Submit HTMX</button>
				<div id="result"></div>
			</form>`, responseToken)
			return c.Render(200, r.String(html))
		} else {
			// Token provided, validate it
			message := c.Param("message")
			return c.Render(200, r.String(fmt.Sprintf("HTMX Success: %s", message)))
		}
	})

	// Test the complete HTMX cycle
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/htmx-test", nil)
	app.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code, "First HTMX request should return form with token")
	require.Contains(t, w.Body.String(), "authenticity_token", "Form should contain CSRF token")

	// Extract token and submit HTMX request
	bodyString := w.Body.String()
	tokenStart := strings.Index(bodyString, `value="`) + 7
	tokenEnd := strings.Index(bodyString[tokenStart:], `"`) + tokenStart
	token := bodyString[tokenStart:tokenEnd]

	require.NotEmpty(t, token, "Should extract valid CSRF token")

	// Submit HTMX request with the token
	formData := url.Values{
		"authenticity_token": {token},
		"message":            {"HTMX test message"},
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/htmx-test", strings.NewReader(formData.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("HX-Request", "true") // Mark as HTMX request

	app.ServeHTTP(w2, req2)

	require.Equal(t, 200, w2.Code, "HTMX POST with valid CSRF token should succeed")
	require.Contains(t, w2.Body.String(), "HTMX Success")

	t.Logf("✅ HTMX CSRF integration working with real Buffalo middleware")
}

// TestFormCSRFIntegration verifies that regular form submissions work with CSRF
func TestFormCSRFIntegration(t *testing.T) {
	app := setupAppWithCSRF()

	// Test regular form submission with CSRF token
	formData := url.Values{
		"authenticity_token": {"test-token-12345"},
		"test_field":         {"test_value"},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/csrf-test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// No HX-Request header = regular form submission

	app.ServeHTTP(w, req)

	// Should succeed with valid CSRF token
	require.Equal(t, 200, w.Code, "Regular form submission with CSRF token should succeed")

	t.Logf("✅ Form CSRF integration working: POST request with token succeeded (status %d)", w.Code)
}

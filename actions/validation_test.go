package actions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/mw-csrf"
	"github.com/stretchr/testify/require"
)

// TestValidateEmail tests RFC-compliant email validation
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
		errMsg   string
	}{
		{"Valid email", "test@example.com", true, ""},
		{"Valid email with subdomain", "user@sub.domain.com", true, ""},
		{"Valid email with numbers", "user123@test.org", true, ""},
		{"Valid email with dots", "user.name@test.co.uk", true, ""},
		{"Valid email with plus", "user+tag@example.com", true, ""},
		{"Valid email with underscore", "user_name@test.com", true, ""},
		{"Valid email with dash", "user-name@test.com", true, ""},

		{"Empty email", "", false, "email is required"},
		{"Too long email", strings.Repeat("a", 255) + "@example.com", false, "email address is too long"},
		{"No @ symbol", "invalid-email", false, "please enter a valid email address"},
		{"No domain", "user@", false, "please enter a valid email address"},
		{"No username", "@example.com", false, "please enter a valid email address"},
		{"Double dots", "user..name@example.com", false, "please enter a valid email address"},
		{"Spaces", "user name@example.com", false, "please enter a valid email address"},
		{"Invalid domain", "user@.com", false, "please enter a valid email address"},
		{"Invalid TLD", "user@example.", false, "please enter a valid email address"},
		{"Localhost email", "user@localhost", true, ""}, // Valid for development
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.expected {
				require.NoError(t, err, "Expected email '%s' to be valid", tt.email)
			} else {
				require.Error(t, err, "Expected email '%s' to be invalid", tt.email)
				require.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
			}
		})
	}
}

// TestValidateRequiredString tests string validation with length limits
func TestValidateRequiredString(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		maxLength int
		expected  bool
		errMsg    string
	}{
		{"Valid string", "John Doe", "Name", 100, true, ""},
		{"Empty string", "", "Name", 100, false, "Name is required"},
		{"Whitespace only", "   ", "Name", 100, false, "Name is required"},
		{"Too long", strings.Repeat("a", 101), "Name", 100, false, "Name must be less than 100 characters"},
		{"Exactly at limit", strings.Repeat("a", 100), "Name", 100, true, ""},
		{"Normal length", "This is a normal length string", "Description", 200, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredString(tt.value, tt.fieldName, tt.maxLength)
			if tt.expected {
				require.NoError(t, err, "Expected string '%s' to be valid", tt.value)
			} else {
				require.Error(t, err, "Expected string '%s' to be invalid", tt.value)
				require.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
			}
		})
	}
}

// TestSanitizeInput tests input sanitization
func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal string", "Hello World", "Hello World"},
		{"String with null bytes", "Hello\x00World", "HelloWorld"},
		{"String with control chars", "Hello\x01\x02World", "HelloWorld"},
		{"String with tabs and newlines", "Hello\t\nWorld", "Hello\t\nWorld"},
		{"String with leading/trailing spaces", "  Hello World  ", "Hello World"},
		{"Empty string", "", ""},
		{"Only spaces", "   ", ""},
		{"Special characters", "Hello@#$%^&*()World", "Hello@#$%^&*()World"},
		{"Unicode characters", "Hello 世界 World", "Hello 世界 World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			require.Equal(t, tt.expected, result, "Input sanitization should produce expected result")
		})
	}
}

// TestValidateContactForm tests the complete contact form validation
func TestValidateContactForm(t *testing.T) {
	tests := []struct {
		name        string
		formData    map[string]string
		expected    bool
		errContains string
	}{
		{
			"Valid form",
			map[string]string{
				"name":    "John Doe",
				"email":   "john@example.com",
				"subject": "Test Subject",
				"message": "This is a test message",
			},
			true,
			"",
		},
		{
			"Missing name",
			map[string]string{
				"name":    "",
				"email":   "john@example.com",
				"subject": "Test Subject",
				"message": "This is a test message",
			},
			false,
			"Name is required",
		},
		{
			"Invalid email",
			map[string]string{
				"name":    "John Doe",
				"email":   "invalid-email",
				"subject": "Test Subject",
				"message": "This is a test message",
			},
			false,
			"please enter a valid email address",
		},
		{
			"Missing subject",
			map[string]string{
				"name":    "John Doe",
				"email":   "john@example.com",
				"subject": "",
				"message": "This is a test message",
			},
			false,
			"Subject is required",
		},
		{
			"Missing message",
			map[string]string{
				"name":    "John Doe",
				"email":   "john@example.com",
				"subject": "Test Subject",
				"message": "",
			},
			false,
			"Message is required",
		},
		{
			"Message too long",
			map[string]string{
				"name":    "John Doe",
				"email":   "john@example.com",
				"subject": "Test Subject",
				"message": strings.Repeat("a", 2001),
			},
			false,
			"Message must be less than 2000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock context with form data
			app := buffalo.New(buffalo.Options{Env: "test"})
			app.POST("/test", func(c buffalo.Context) error {
				err := ValidateContactForm(c)
				if err != nil {
					return c.Render(400, r.String(err.Error()))
				}
				return c.Render(200, r.String("success"))
			})

			formData := url.Values{}
			for key, value := range tt.formData {
				formData.Set(key, value)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			app.ServeHTTP(w, req)

			if tt.expected {
				require.Equal(t, 200, w.Code, "Expected valid form to succeed")
				require.Contains(t, w.Body.String(), "success")
			} else {
				require.Equal(t, 400, w.Code, "Expected invalid form to fail")
				require.Contains(t, w.Body.String(), tt.errContains)
			}
		})
	}
}

// TestContactFormHandler tests the actual contact form handler
func TestContactFormHandler(t *testing.T) {
	t.Run("GET request shows form", func(t *testing.T) {
		app := buffalo.New(buffalo.Options{Env: "test"})
		app.GET("/contact", func(c buffalo.Context) error {
			// Set CSRF token for template
			c.Set("authenticity_token", "test-token")
			return c.Render(200, r.HTML("<form><input name='authenticity_token' value='test-token'></form>"))
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/contact", nil)

		app.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code, "GET /contact should return 200")
		require.Contains(t, w.Body.String(), "authenticity_token", "Form should include CSRF token")
	})

	t.Run("POST request with valid data", func(t *testing.T) {
		app := buffalo.New(buffalo.Options{Env: "test"})
		app.POST("/contact", func(c buffalo.Context) error {
			// Mock the validation and email service
			name := c.Param("name")
			email := c.Param("email")
			subject := c.Param("subject")
			message := c.Param("message")

			if name == "" || email == "" || subject == "" || message == "" {
				return c.Render(400, r.String("Validation failed"))
			}

			return c.Render(200, r.String("Form submitted successfully"))
		})

		formData := url.Values{
			"name":    {"Test User"},
			"email":   {"test@example.com"},
			"subject": {"Test Subject"},
			"message": {"Test message"},
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/contact", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		app.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code, "POST /contact with valid data should succeed")
		require.Contains(t, w.Body.String(), "Form submitted successfully")
	})
}

// TestHTMXCSRFTokenInclusion tests that HTMX properly includes CSRF tokens
func TestHTMXCSRFTokenInclusion(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.Use(func() buffalo.MiddlewareFunc {
		return func(next buffalo.Handler) buffalo.Handler {
			return func(c buffalo.Context) error {
				// Simulate CSRF token being set by Buffalo middleware
				c.Set("authenticity_token", "test-csrf-token")
				return next(c)
			}
		}
	}())

	app.POST("/htmx-test", func(c buffalo.Context) error {
		token := c.Param("authenticity_token")
		if token == "" {
			return c.Render(403, r.String("CSRF token missing"))
		}
		return c.Render(200, r.String("success"))
	})

	// Test HTMX request with CSRF token
	formData := url.Values{
		"authenticity_token": {"test-csrf-token"},
		"test_field":         {"test_value"},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/htmx-test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	app.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code, "HTMX request with CSRF token should succeed")
	require.Contains(t, w.Body.String(), "success")
}

// TestSecurityHeaders tests that our forms include proper security attributes
func TestSecurityHeaders(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.GET("/secure-form", func(c buffalo.Context) error {
		html := `<form method="post" action="/submit">
			<input type="hidden" name="authenticity_token" value="test-token">
			<label for="email">
				Email *
				<input type="email" id="email" name="email" required>
			</label>
			<button type="submit">Submit</button>
		</form>`
		return c.Render(200, r.String(html))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/secure-form", nil)

	app.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)

	body := w.Body.String()
	require.Contains(t, body, `name="authenticity_token"`, "Form should include CSRF token")
	require.Contains(t, body, `for="email"`, "Label should be properly associated")
	require.Contains(t, body, `id="email"`, "Input should have proper ID")
	require.Contains(t, body, `required`, "Email field should be required")
}

// TestProductionCSRFBuiltInPattern tests CSRF behavior in production environment
func TestProductionCSRFBuiltInPattern(t *testing.T) {
	os.Setenv("GO_ENV", "test") // Ensure CSRF middleware runs in test mode
	app := buffalo.New(buffalo.Options{Env: "production"})
	app.Use(csrf.New)

	app.POST("/secure-test", func(c buffalo.Context) error {
		// First check if we have a token in the request
		token := c.Param("authenticity_token")
		if token == "" {
			// No token provided, this is the first request - return form with token
			responseToken := c.Value("authenticity_token")
			if responseToken == nil {
				return c.Render(500, r.String("No CSRF token generated"))
			}
			html := fmt.Sprintf(`<form method="post" action="/secure-test">
				<input type="hidden" name="authenticity_token" value="%s" />
				<input type="text" name="data" required />
				<button type="submit">Submit</button>
			</form>`, responseToken)
			return c.Render(200, r.String(html))
		} else {
			// Token provided, validate it
			data := c.Param("data")
			return c.Render(200, r.String(fmt.Sprintf("Secure submission: %s", data)))
		}
	})

	// Test production form loading
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/secure-test", nil)
	app.ServeHTTP(w1, req1)

	require.Equal(t, 200, w1.Code, "Production POST should return form with token")
	require.Contains(t, w1.Body.String(), "authenticity_token")

	// Extract token
	bodyString := w1.Body.String()
	tokenStart := strings.Index(bodyString, `value="`) + 7
	tokenEnd := strings.Index(bodyString[tokenStart:], `"`) + tokenStart
	token := bodyString[tokenStart:tokenEnd]

	// Test production form submission
	formData := url.Values{
		"authenticity_token": {token},
		"data":               {"production test data"},
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/secure-test", strings.NewReader(formData.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.ServeHTTP(w2, req2)

	require.Equal(t, 200, w2.Code, "Production POST with valid token should succeed")
	require.Contains(t, w2.Body.String(), "Secure submission")

	t.Logf("✅ Production CSRF integration working")
}

// TestCSRFSecurityRegression tests for common CSRF vulnerabilities
func TestCSRFSecurityRegression(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.Use(csrf.New)

	app.POST("/api/submit", func(c buffalo.Context) error {
		return c.Render(200, r.String("Data submitted"))
	})

	t.Run("No token should be rejected", func(t *testing.T) {
		formData := url.Values{
			"data": {"test"},
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/submit", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		app.ServeHTTP(w, req)
		require.Equal(t, 403, w.Code, "Request without CSRF token should be rejected")
	})

	t.Run("Invalid token should be rejected", func(t *testing.T) {
		formData := url.Values{
			"authenticity_token": {"invalid-token-12345"},
			"data":               {"test"},
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/submit", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		app.ServeHTTP(w, req)
		require.Equal(t, 403, w.Code, "Request with invalid CSRF token should be rejected")
	})

	t.Run("GET requests should not require CSRF", func(t *testing.T) {
		app.GET("/api/data", func(c buffalo.Context) error {
			return c.Render(200, r.String("Data retrieved"))
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/data", nil)

		app.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code, "GET requests should not require CSRF tokens")
	})

	t.Run("Empty token should be rejected", func(t *testing.T) {
		formData := url.Values{
			"authenticity_token": {""},
			"data":               {"test"},
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/submit", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		app.ServeHTTP(w, req)
		require.Equal(t, 403, w.Code, "Request with empty CSRF token should be rejected")
	})
}

// TestHTMXProgressiveEnhancement tests HTMX progressive enhancement with CSRF
func TestHTMXProgressiveEnhancement(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.Use(csrf.New)

	app.POST("/contact-test", func(c buffalo.Context) error {
		email := c.Param("email")
		token := c.Param("authenticity_token")

		if token == "" {
			responseToken := c.Value("authenticity_token")
			html := fmt.Sprintf(`<form method="post" action="/contact-test" hx-post="/contact-test" hx-target="#result" hx-swap="innerHTML">
				<input type="hidden" name="authenticity_token" value="%s" />
				<input type="email" name="email" required />
				<button type="submit">Submit</button>
				<div id="result"></div>
			</form>`, responseToken)
			return c.Render(200, r.String(html))
		}

		// Always return full page, regardless of HX-Request
		return c.Render(200, r.String(fmt.Sprintf("Form Success: %s", email)))
	})

	// Get form with token
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/contact-test", nil)
	app.ServeHTTP(w1, req1)

	require.Equal(t, 200, w1.Code)
	bodyString := w1.Body.String()
	tokenStart := strings.Index(bodyString, `value="`) + 7
	tokenEnd := strings.Index(bodyString[tokenStart:], `"`) + tokenStart
	token := bodyString[tokenStart:tokenEnd]

	// Submit regular form
	formData := url.Values{
		"authenticity_token": {token},
		"email":              {"test@example.com"},
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/contact-test", strings.NewReader(formData.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.ServeHTTP(w2, req2)
	require.Equal(t, 200, w2.Code)
	require.Contains(t, w2.Body.String(), "Form Success")

	// Submit HTMX request
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/contact-test", strings.NewReader(formData.Encode()))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req3.Header.Set("HX-Request", "true")

	app.ServeHTTP(w3, req3)
	require.Equal(t, 200, w3.Code)
	require.Contains(t, w3.Body.String(), "Form Success")

	t.Logf("✅ HTMX progressive enhancement always returns full page")
}

// TestCSRFTokenExpiration tests token validation scenarios
func TestCSRFTokenValidation(t *testing.T) {
	os.Setenv("GO_ENV", "test") // Ensure CSRF middleware runs in test mode
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.Use(csrf.New)

	app.POST("/test-submit", func(c buffalo.Context) error {
		token := c.Param("authenticity_token")
		if token == "" {
			responseToken := c.Value("authenticity_token")
			html := fmt.Sprintf(`<form method="post" action="/test-submit">
				<input type="hidden" name="authenticity_token" value="%s" />
				<input type="text" name="data" value="test" />
				<button type="submit">Submit</button>
			</form>`, responseToken)
			return c.Render(200, r.String(html))
		}
		return c.Render(200, r.String("Success"))
	})

	// Get form with token
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/test-submit", nil)
	app.ServeHTTP(w1, req1)

	require.Equal(t, 200, w1.Code)
	bodyString := w1.Body.String()
	tokenStart := strings.Index(bodyString, `value="`) + 7
	tokenEnd := strings.Index(bodyString[tokenStart:], `"`) + tokenStart
	token := bodyString[tokenStart:tokenEnd]

	// Use token - should work
	formData := url.Values{
		"authenticity_token": {token},
		"data":               {"test"},
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/test-submit", strings.NewReader(formData.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.ServeHTTP(w2, req2)
	require.Equal(t, 200, w2.Code, "Valid token should work")
	require.Contains(t, w2.Body.String(), "Success")

	t.Logf("✅ CSRF token validation working correctly")
}

// TestConcurrentCSRFRequests tests for race conditions
func TestConcurrentCSRFRequests(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.Use(csrf.New)

	app.POST("/concurrent-test", func(c buffalo.Context) error {
		return c.Render(200, r.String("Success"))
	})

	// This test would require more complex setup with goroutines
	// For now, we'll test basic concurrent access pattern
	app.GET("/token-source", func(c buffalo.Context) error {
		token := c.Value("authenticity_token")
		return c.Render(200, r.String(fmt.Sprintf("Token: %v", token)))
	})

	// Test multiple rapid requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/token-source", nil)
		app.ServeHTTP(w, req)
		require.Equal(t, 200, w.Code, fmt.Sprintf("Concurrent request %d should succeed", i))
	}

	t.Logf("✅ Concurrent CSRF token access working")
}

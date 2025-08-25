package actions

import (
	"net/http"
	"strings"

	"github.com/stretchr/testify/require"
)

func (as *ActionSuite) Test_CSRF_Protection_ContactForm() {
	// Test that contact form requires CSRF token
	res := as.HTML("/contact").Post(map[string]interface{}{
		"name":    "Test User",
		"email":   "test@example.com",
		"subject": "Test Subject",
		"message": "Test message",
	})

	// Should fail without CSRF token
	require.Equal(as.T(), http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_Protection_UserRegistration() {
	// Test that user registration requires CSRF token
	res := as.HTML("/users").Post(map[string]interface{}{
		"first_name":            "Test",
		"last_name":             "User",
		"email":                 "test@example.com",
		"password":              "password123",
		"password_confirmation": "password123",
		"accept_terms":          "true",
	})

	// Should fail without CSRF token
	require.Equal(as.T(), http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_Protection_Auth() {
	// Test that authentication requires CSRF token
	res := as.HTML("/auth").Post(map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	})

	// Should fail without CSRF token
	require.Equal(as.T(), http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_Token_Generation() {
	// Test that CSRF token is generated and included in templates
	res := as.HTML("/auth/new").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	body := res.Body.String()

	// Check that CSRF meta tags are present
	require.Contains(as.T(), body, `<meta name="csrf-param" content="authenticity_token"`)
	require.Contains(as.T(), body, `<meta name="csrf-token" content="`)
}

func (as *ActionSuite) Test_CSRF_FormFor_Token_Generation() {
	// Test that formFor() helper includes CSRF tokens
	res := as.HTML("/users/new").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	body := res.Body.String()

	// formFor should automatically include CSRF token as hidden input
	require.Contains(as.T(), body, `name="authenticity_token"`)
	require.Contains(as.T(), body, `type="hidden"`)
}

func (as *ActionSuite) Test_CSRF_API_Endpoints_Excluded() {
	// Test that API endpoints (like webhooks) are excluded from CSRF protection
	// This test assumes the webhook endpoint doesn't require authentication
	res := as.HTML("/api/donations/webhook").Post(map[string]interface{}{
		"test": "data",
	})

	// Should not return 403 Forbidden for CSRF (might return other errors)
	require.NotEqual(as.T(), http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_Valid_Token_Accepted() {
	// Get a page with CSRF token first
	getRes := as.HTML("/auth/new").Get()
	require.Equal(as.T(), http.StatusOK, getRes.Code)

	// Extract CSRF token from the response
	body := getRes.Body.String()
	tokenStart := strings.Index(body, `<meta name="csrf-token" content="`)
	if tokenStart == -1 {
		as.T().Skip("CSRF token not found in response")
		return
	}

	tokenStart += len(`<meta name="csrf-token" content="`)
	tokenEnd := strings.Index(body[tokenStart:], `"`)
	if tokenEnd == -1 {
		as.T().Skip("CSRF token end not found in response")
		return
	}

	token := body[tokenStart : tokenStart+tokenEnd]
	require.NotEmpty(as.T(), token)

	// Now make a request with the valid token
	// Note: This test may still fail due to user validation, but should not fail with CSRF error
	res := as.HTML("/auth").Post(map[string]interface{}{
		"email":              "invalid@example.com",
		"password":           "wrongpassword",
		"authenticity_token": token,
	})

	// Should not return 403 Forbidden (CSRF error)
	// May return 401 Unauthorized (auth error) which is expected
	require.NotEqual(as.T(), http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_GET_Requests_Not_Protected() {
	// Test that GET requests are not subject to CSRF protection
	res := as.HTML("/").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	res = as.HTML("/blog").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	res = as.HTML("/team").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)
}

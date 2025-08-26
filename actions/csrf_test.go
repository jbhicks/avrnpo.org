package actions

import (
	"net/http"

	"github.com/stretchr/testify/require"
)

// Note: CSRF protection is disabled in test environment (ENV == "test")
// These tests verify the test environment behavior is correct

func (as *ActionSuite) Test_CSRF_Disabled_In_Test_Environment_ContactForm() {
	// Test that contact form works without CSRF token in test environment
	res := as.HTML("/contact").Post(map[string]interface{}{
		"name":    "Test User",
		"email":   "test@example.com",
		"subject": "Test Subject",
		"message": "Test message",
	})

	// Should succeed without CSRF token in test environment (redirects after success)
	as.Equal(http.StatusSeeOther, res.Code)
}

func (as *ActionSuite) Test_CSRF_Disabled_In_Test_Environment_UserRegistration() {
	// Test that user registration works without CSRF token in test environment
	res := as.HTML("/users").Post(map[string]interface{}{
		"first_name":            "Test",
		"last_name":             "User",
		"email":                 "test@example.com",
		"password":              "password123",
		"password_confirmation": "password123",
		"accept_terms":          "on",
	})

	// Should succeed without CSRF token in test environment
	as.Equal(http.StatusFound, res.Code) // Redirect after successful registration
}

func (as *ActionSuite) Test_CSRF_Token_Generation_Disabled_In_Test() {
	// Test that CSRF token generation is disabled in test environment
	res := as.HTML("/auth/new").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	body := res.Body.String()

	// Check that CSRF meta tags are NOT present in test environment
	require.NotContains(as.T(), body, `<meta name="csrf-param" content="authenticity_token"`)
	require.NotContains(as.T(), body, `<meta name="csrf-token" content="`)
}

func (as *ActionSuite) Test_CSRF_API_Endpoints_Work_In_Test() {
	// Test that API endpoints work correctly in test environment
	res := as.HTML("/api/donations/webhook").Post(map[string]interface{}{
		"test": "data",
	})

	// Should not return 403 Forbidden (webhook expects specific data format)
	require.NotEqual(as.T(), http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_GET_Requests_Work() {
	// Test that GET requests work correctly (not affected by CSRF anyway)
	res := as.HTML("/").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	res = as.HTML("/blog").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	res = as.HTML("/team").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)
}

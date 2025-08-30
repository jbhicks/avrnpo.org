package actions

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/stretchr/testify/require"
)

// Note: CSRF protection is enabled in all environments with Buffalo's standard middleware
// These tests verify that CSRF protection works correctly

func (as *ActionSuite) Test_CSRF_Enabled_In_Test_Environment_ContactForm() {
	// Test that contact form requires CSRF token in test environment
	res := as.HTML("/contact").Post(map[string]interface{}{
		"name":    "Test User",
		"email":   "test@example.com",
		"subject": "Test Subject",
		"message": "Test message",
	})

	// Should fail without CSRF token (403 Forbidden)
	as.Equal(http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_Enabled_In_Test_Environment_UserRegistration() {
	// Test that user registration requires CSRF token in test environment
	res := as.HTML("/users").Post(map[string]interface{}{
		"first_name":            "Test",
		"last_name":             "User",
		"email":                 "test@example.com",
		"password":              "password123",
		"password_confirmation": "password123",
		"accept_terms":          "on",
	})

	// Should fail without CSRF token (403 Forbidden)
	as.Equal(http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_Token_Generation_Enabled_In_Test() {
	// Test that CSRF token generation is enabled in test environment
	res := as.HTML("/auth/new").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	body := res.Body.String()

	// Check that CSRF token is present as hidden input (Buffalo's standard approach)
	require.Contains(as.T(), body, `<input name="authenticity_token" type="hidden"`)
	require.Contains(as.T(), body, `value="`)
}

func (as *ActionSuite) Test_CSRF_API_Endpoints_WORK_In_Test() {
	// Test that API endpoints work correctly in test environment
	payload := map[string]interface{}{"test": "data"}
	jsonBody, _ := json.Marshal(payload)
	// Attach signature using helper and ensure verifier token present
	os.Setenv("HELCIM_WEBHOOK_VERIFIER_TOKEN", "test_verifier_token")

	req := as.HTML("/api/donations/webhook")
	req.Headers["Content-Type"] = "application/json"
	// Provide the signed header
	req.Headers["X-Helcim-Signature"] = AttachHelcimSignature(jsonBody)
	// Use Post for raw body by passing the json as string
	res := req.Post(string(jsonBody))

	// Should not return 403 Forbidden (webhook expects specific data format)
	require.NotEqual(as.T(), http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_GET_Requests_WORK() {
	// Test that GET requests work correctly (not affected by CSRF anyway)
	res := as.HTML("/").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	res = as.HTML("/blog").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)

	res = as.HTML("/team").Get()
	require.Equal(as.T(), http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_CSRF_Donation_API_MissingToken() {
	// Test that donation API requires CSRF token
	res := as.JSON("/api/donations/initialize").Post(map[string]interface{}{
		"amount": "100",
	})
	as.Equal(http.StatusForbidden, res.Code)
}

func (as *ActionSuite) Test_CSRF_Donation_API_ValidToken() {
	// Test that donation API works with valid CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	req := as.JSON("/api/donations/initialize")
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	req.Headers["X-CSRF-Token"] = token // For JSON API, use header

	res := req.Post(map[string]interface{}{
		"amount": "100",
	})
	as.NotEqual(http.StatusForbidden, res.Code) // Should not be 403
}

func init() {}

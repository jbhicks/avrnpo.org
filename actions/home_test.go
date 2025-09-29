package actions

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (as *ActionSuite) Test_HomeHandler() {
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	// Check for main page structure (the AVR content)
	as.Contains(res.Body.String(), "American Veterans Rebuilding")
	as.Contains(res.Body.String(), "THE AVR MISSION") // Actual home content
	as.Contains(res.Body.String(), "Rebuilding the American Veteran's Self")
}

func (as *ActionSuite) Test_HomeHandler_Enhanced_Content() {
	// Test enhanced content loading with progressive enhancement
	req := as.HTML("/")
	res := req.Get()

	as.Equal(http.StatusOK, res.Code)
	// Check for actual AVR content that should be in the partial
	as.Contains(res.Body.String(), "THE AVR MISSION")
	as.Contains(res.Body.String(), "Technical Training")
	as.Contains(res.Body.String(), "Occupational Licensing")
	as.Contains(res.Body.String(), "Home Ownership Options")
}

func (as *ActionSuite) Test_HomeHandler_LoggedIn() {
	// Create a user through the signup endpoint (which works)
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("mark-%d@example.com", timestamp)

	// First, get CSRF token from signup page
	cookie, token := fetchCSRF(as.T(), App(), "/users/new")

	signupData := map[string]interface{}{
		"Email":                email,
		"Password":             "password",
		"PasswordConfirmation": "password",
		"FirstName":            "Mark",
		"LastName":             "Smith",
		"accept_terms":         "on", // Add required terms acceptance
		"authenticity_token":   token,
	}

	// Create user via web interface to ensure it's properly committed
	signupReq := as.HTML("/users")
	if cookie != "" {
		signupReq.Headers["Cookie"] = cookie
	}
	signupRes := signupReq.Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Extract session cookie from signup response
	sessionCookie := ""
	for _, c := range signupRes.Result().Cookies() {
		if strings.Contains(c.Name, "session") || c.Name == "_avrnpo.org_session" {
			sessionCookie = c.String()
			break
		}
	}

	// Get new CSRF token for login using the session cookie
	loginCookie, loginToken := fetchCSRF(as.T(), App(), "/auth/new")

	// Combine cookies if we have both
	combinedCookie := sessionCookie
	if loginCookie != "" && sessionCookie != "" && !strings.Contains(sessionCookie, loginCookie) {
		combinedCookie = sessionCookie + "; " + loginCookie
	} else if loginCookie != "" {
		combinedCookie = loginCookie
	}

	loginData := map[string]interface{}{
		"Email":              email,
		"Password":           "password",
		"authenticity_token": loginToken,
	}

	// POST to login endpoint to get proper session
	loginReq := as.HTML("/auth")
	if combinedCookie != "" {
		loginReq.Headers["Cookie"] = combinedCookie
	}
	loginRes := loginReq.Post(loginData)
	as.Equal(http.StatusFound, loginRes.Code)

	// Extract final session cookie from login response
	finalSessionCookie := sessionCookie
	for _, c := range loginRes.Result().Cookies() {
		if strings.Contains(c.Name, "session") || c.Name == "_avrnpo.org_session" {
			finalSessionCookie = c.String()
			break
		}
	}

	// Test that logged in users see the main shell with authenticated nav
	homeReq := as.HTML("/")
	if finalSessionCookie != "" {
		homeReq.Headers["Cookie"] = finalSessionCookie
	}
	res := homeReq.Get()
	as.Equal(http.StatusOK, res.Code)

	// Check that we get the basic content
	body := res.Body.String()
	as.Contains(body, "THE AVR MISSION") // Main content should be there

	// Test enhanced content for logged in user
	enhancedReq := as.HTML("/")
	if finalSessionCookie != "" {
		enhancedReq.Headers["Cookie"] = finalSessionCookie
	}
	enhancedRes := enhancedReq.Get()
	as.Equal(http.StatusOK, enhancedRes.Code)
	// Verify the basic template content is there
	as.Contains(enhancedRes.Body.String(), "THE AVR MISSION")
	as.Contains(enhancedRes.Body.String(), "Technical Training")

	// Test that the dashboard is accessible (this proves authentication works)
	dashboardReq := as.HTML("/dashboard")
	if finalSessionCookie != "" {
		dashboardReq.Headers["Cookie"] = finalSessionCookie
	}
	res = dashboardReq.Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Dashboard")
	as.Contains(res.Body.String(), "Welcome back") // Dashboard-specific content
}

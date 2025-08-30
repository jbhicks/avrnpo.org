package actions

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"avrnpo.org/models"
)

func (as *ActionSuite) Test_Users_New() {
	res := as.HTML("/users/new").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_Users_Create() {
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("mark-%d@example.com", timestamp)
	u := map[string]interface{}{
		"Email":                email,
		"Password":             "password",
		"PasswordConfirmation": "password",
		"FirstName":            "Mark",
		"LastName":             "Smith",
		"accept_terms":         "on", // Add required terms acceptance
	}

	// Fetch CSRF and session for signup
	cookie, token := fetchCSRF(as.T(), as.App, "/users/new")
	form := u
	form["authenticity_token"] = token
	req := as.HTML("/users")
	req.Headers["Cookie"] = cookie
	res := req.Post(form)
	as.Equal(http.StatusFound, res.Code)

	// Verify the redirect location
	location := res.Header().Get("Location")
	as.Equal("/", location, "Should redirect to home page after successful user creation")

	// Test that we can authenticate with the created user immediately
	// This implicitly tests that the user was created and can be found
	authData := &models.User{
		Email:    email,
		Password: "password",
	}

	authRes := as.HTML("/auth").Post(authData)
	as.Equal(http.StatusFound, authRes.Code, "Should be able to authenticate with newly created user")
}

func (as *ActionSuite) Test_ProfileSettings_LoggedIn() {
	timestamp := time.Now().UnixNano()

	// Create a user through the signup endpoint (which works)
	signupData := map[string]interface{}{
		"Email":                fmt.Sprintf("profile-test-%d@example.com", timestamp),
		"Password":             "password",
		"PasswordConfirmation": "password",
		"FirstName":            "Profile",
		"LastName":             "Test",
		"accept_terms":         "on", // Add required terms acceptance
	}

	// Create user via web interface to ensure it's properly committed
	// Fetch CSRF and session for signup
	signupCookie, signupToken := fetchCSRF(as.T(), as.App, "/users/new")
	signupData["authenticity_token"] = signupToken
	signupReq := as.HTML("/users")
	signupReq.Headers["Cookie"] = signupCookie
	signupRes := signupReq.Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Now login with the same user using MockLogin to obtain session + CSRF
	cookie, _ := MockLogin(as.T(), as.App, fmt.Sprintf("profile-test-%d@example.com", timestamp), "password")

	// Access profile settings while logged in
	req := as.HTML("/profile")
	req.Headers["Cookie"] = cookie
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Profile Settings")
}

func (as *ActionSuite) Test_ProfileSettings_RequiresAuth() {
	res := as.HTML("/profile").Get()
	as.Equal(http.StatusFound, res.Code) // Should redirect to signin
}

func (as *ActionSuite) Test_ProfileUpdate_LoggedIn() {
	timestamp := time.Now().UnixNano()

	// Create a user through the signup endpoint (which works)
	signupData := map[string]interface{}{
		"Email":                fmt.Sprintf("profile-update-%d@example.com", timestamp),
		"Password":             "password",
		"PasswordConfirmation": "password",
		"FirstName":            "Update",
		"LastName":             "Test",
		"accept_terms":         "on", // Add required terms acceptance
	}

	// Create user via web interface to ensure it's properly committed
	// Fetch CSRF and session for signup
	signupCookie, signupToken := fetchCSRF(as.T(), as.App, "/users/new")
	signupData["authenticity_token"] = signupToken
	signupReq := as.HTML("/users")
	signupReq.Headers["Cookie"] = signupCookie
	signupRes := signupReq.Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Log in via MockLogin to obtain session+token
	cookie, token := MockLogin(as.T(), as.App, fmt.Sprintf("profile-update-%d@example.com", timestamp), "password")

	// Update profile data (include authenticity_token and session cookie)
	form := map[string]interface{}{
		"FirstName":          "UpdatedFirst",
		"LastName":           "UpdatedLast",
		"authenticity_token": token,
	}
	req := as.HTML("/profile")
	req.Headers["Cookie"] = cookie
	res := req.Post(form)
	as.Equal(http.StatusFound, res.Code) // Should redirect after successful update

	// Verify the profile was updated by checking the profile page
	profileReq := as.HTML("/profile")
	profileReq.Headers["Cookie"] = cookie
	profileRes := profileReq.Get()
	as.Equal(http.StatusOK, profileRes.Code)
	as.Contains(profileRes.Body.String(), "UpdatedFirst")
	as.Contains(profileRes.Body.String(), "UpdatedLast")
}

func (as *ActionSuite) Test_AccountSettings_LoggedIn() {
	timestamp := time.Now().UnixNano()
	userEmail := fmt.Sprintf("account-test-%d@example.com", timestamp)

	// Create a user through the signup endpoint (which works)
	signupData := map[string]interface{}{
		"Email":                userEmail,
		"Password":             "password",
		"PasswordConfirmation": "password",
		"FirstName":            "Account",
		"LastName":             "Test",
		"accept_terms":         "on", // Add required terms acceptance
	}

	// Create user via web interface to ensure it's properly committed
	// Fetch CSRF and session for signup
	signupCookie, signupToken := fetchCSRF(as.T(), as.App, "/users/new")
	signupData["authenticity_token"] = signupToken
	signupReq := as.HTML("/users")
	signupReq.Headers["Cookie"] = signupCookie
	signupRes := signupReq.Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Login via MockLogin to obtain session cookie
	cookie, _ := MockLogin(as.T(), as.App, userEmail, "password")

	// Assert we can see the account settings page with user data
	req := as.HTML("/account")
	req.Headers["Cookie"] = cookie
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Account Settings")
	as.Contains(res.Body.String(), userEmail)
}

func (as *ActionSuite) Test_AccountSettings_RequiresAuth() {
	res := as.HTML("/account").Get()
	as.Equal(http.StatusFound, res.Code) // Should redirect to signin
}

func (as *ActionSuite) Test_AccountSettings_HTMX_Partial() {
	timestamp := time.Now().UnixNano()

	// Create and login user
	signupData := map[string]interface{}{
		"Email":                fmt.Sprintf("htmx-test-%d@example.com", timestamp),
		"Password":             "password",
		"PasswordConfirmation": "password",
		"FirstName":            "HTMX",
		"LastName":             "Test",
		"accept_terms":         "on", // Add required terms acceptance
	}

	// Fetch CSRF and session for signup
	signupCookie, signupToken := fetchCSRF(as.T(), as.App, "/users/new")
	signupData["authenticity_token"] = signupToken
	signupReq := as.HTML("/users")
	signupReq.Headers["Cookie"] = signupCookie
	signupRes := signupReq.Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Use MockLogin for session-backed requests
	_, _ = MockLogin(as.T(), as.App, fmt.Sprintf("htmx-test-%d@example.com", timestamp), "password")

	// Test HTMX request (now returns full page with progressive enhancement)
	req := as.HTML("/account")
	req.Headers["HX-Request"] = "true"
	htmxRes := req.Get()

	as.Equal(http.StatusOK, htmxRes.Code)
	as.Contains(htmxRes.Body.String(), "Account Settings")
	// HTMX response now returns full page (single-template architecture)
	as.Contains(htmxRes.Body.String(), "American Veterans Rebuilding")
	as.Contains(htmxRes.Body.String(), "<nav")

	// Test regular request (should return full page)
	regularRes := as.HTML("/account").Get()

	as.Equal(http.StatusOK, regularRes.Code)
	as.Contains(regularRes.Body.String(), "Account Settings")
	// Regular response SHOULD contain navigation (it's a full page)
	as.Contains(regularRes.Body.String(), "American Veterans Rebuilding")
	as.Contains(regularRes.Body.String(), "<nav")
}

// Debug test to see what validation errors are happening
func (as *ActionSuite) Test_Debug_User_Creation() {
	timestamp := time.Now().Unix()

	// First test direct database creation
	u := &models.User{
		Email:                fmt.Sprintf("direct-test-%d@example.com", timestamp),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Direct",
		LastName:             "Test",
	}

	// Test direct database creation
	tx := as.DB
	verrs, err := u.Create(tx)

	as.T().Logf("Direct Create() errors: %v, validation errors: %v", err, verrs.String())
	as.NoError(err)
	if verrs.HasAny() {
		as.T().Logf("Validation errors from direct Create(): %v", verrs.String())
	}
	as.False(verrs.HasAny(), "Expected no validation errors, got: %v", verrs.String())

	// Now test web interface with different email
	signupData := &models.User{
		Email:                fmt.Sprintf("debug-test-%d@example.com", timestamp+1),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Debug",
		LastName:             "Test",
	}

	// Create user via web interface
	res := as.HTML("/users").Post(signupData)

	// Print the response code and body to debug
	as.T().Logf("Web Response Code: %d", res.Code)
	if res.Code != http.StatusFound {
		as.T().Logf("Expected 302 but got %d", res.Code)
		bodyStr := res.Body.String()
		if strings.Contains(bodyStr, "text-red-600") {
			as.T().Log("Found error styling in response - there are validation errors")
		} else {
			as.T().Log("No error styling found - validation might be passing but redirect failing")
		}
	}
}

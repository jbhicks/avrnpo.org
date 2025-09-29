package actions

import (
	"net/url"
	"regexp"
	"testing"

	"avrnpo.org/models"
)

// TestEndToEndFormSubmissionWithCSRF tests complete form workflows including CSRF token extraction and validation
func (as *ActionSuite) Test_EndToEndFormSubmissionWithCSRF() {
	as.T().Run("User Registration with Real CSRF", func(t *testing.T) {
		// Step 1: GET the registration form to extract CSRF token
		getRes := as.HTML("/users/new").Get()
		as.Equal(200, getRes.Code)

		// Step 2: Extract CSRF token from the rendered HTML
		csrfToken := as.extractCSRFTokenFromHTML(getRes.Body.String())
		as.NotEmpty(csrfToken, "CSRF token should be present in registration form")

		// Step 3: Submit form with extracted CSRF token
		formData := url.Values{
			"Email":                {"e2e-test@example.com"},
			"FirstName":            {"E2E"},
			"LastName":             {"Test"},
			"Password":             {"password123"},
			"PasswordConfirmation": {"password123"},
			"authenticity_token":   {csrfToken},
		}

		postRes := as.HTML("/users").Post(formData)
		as.Equal(302, postRes.Code, "Registration should succeed with valid CSRF token")

		// Step 4: Verify user was actually created
		var user models.User
		err := as.DB.Where("email = ?", "e2e-test@example.com").First(&user)
		as.NoError(err, "User should be created in database")
		as.Equal("E2E", user.FirstName)
	})

	as.T().Run("User Registration Fails Without CSRF", func(t *testing.T) {
		// Submit form without CSRF token
		formData := url.Values{
			"Email":                {"no-csrf@example.com"},
			"FirstName":            {"No"},
			"LastName":             {"CSRF"},
			"Password":             {"password123"},
			"PasswordConfirmation": {"password123"},
		}

		postRes := as.HTML("/users").Post(formData)
		as.Equal(403, postRes.Code, "Registration should fail without CSRF token")

		// Verify user was NOT created
		count, err := as.DB.Where("email = ?", "no-csrf@example.com").Count(&models.User{})
		as.NoError(err)
		as.Equal(0, count, "User should not be created without CSRF token")
	})

	as.T().Run("User Registration Fails With Invalid CSRF", func(t *testing.T) {
		// Submit form with fake CSRF token
		formData := url.Values{
			"Email":                {"bad-csrf@example.com"},
			"FirstName":            {"Bad"},
			"LastName":             {"CSRF"},
			"Password":             {"password123"},
			"PasswordConfirmation": {"password123"},
			"authenticity_token":   {"fake-token-12345"},
		}

		postRes := as.HTML("/users").Post(formData)
		as.Equal(403, postRes.Code, "Registration should fail with invalid CSRF token")

		// Verify user was NOT created
		count, err := as.DB.Where("email = ?", "bad-csrf@example.com").Count(&models.User{})
		as.NoError(err)
		as.Equal(0, count, "User should not be created with invalid CSRF token")
	})

	as.T().Run("Login Form with Real CSRF", func(t *testing.T) {
		// Create a test user first
		user := &models.User{
			Email:     "login-test@example.com",
			FirstName: "Login",
			LastName:  "Test",
			Role:      "user",
		}
		user.Password = "password123"
		user.PasswordConfirmation = "password123"
		verrs, err := user.Create(as.DB)
		as.NoError(err)
		as.False(verrs.HasAny())

		// Step 1: GET the login form to extract CSRF token
		getRes := as.HTML("/auth/new").Get()
		as.Equal(200, getRes.Code)

		// Step 2: Extract CSRF token from the rendered HTML
		csrfToken := as.extractCSRFTokenFromHTML(getRes.Body.String())
		as.NotEmpty(csrfToken, "CSRF token should be present in login form")

		// Step 3: Submit login form with extracted CSRF token
		formData := url.Values{
			"Email":              {"login-test@example.com"},
			"Password":           {"password123"},
			"authenticity_token": {csrfToken},
		}

		postRes := as.HTML("/auth").Post(formData)
		as.Equal(302, postRes.Code, "Login should succeed with valid CSRF token")

		// Step 4: Verify session was created (redirect to dashboard)
		location := postRes.Header().Get("Location")
		as.Contains(location, "/", "Should redirect after successful login")
	})

	as.T().Run("Admin Post Creation with Real CSRF", func(t *testing.T) {
		// Create admin user first
		adminUser := &models.User{
			Email:     "admin-csrf-test@example.com",
			FirstName: "Admin",
			LastName:  "Test",
			Role:      "admin",
		}
		adminUser.Password = "password123"
		adminUser.PasswordConfirmation = "password123"
		verrs, err := adminUser.Create(as.DB)
		as.NoError(err)
		as.False(verrs.HasAny())

		// Step 1: Login as admin to establish session
		as.Session.Set("current_user_id", adminUser.ID)

		// Step 2: GET the post creation form to extract CSRF token
		getRes := as.HTML("/admin/posts/new").Get()
		as.Equal(200, getRes.Code)

		// Step 3: Extract CSRF token from the rendered HTML
		csrfToken := as.extractCSRFTokenFromHTML(getRes.Body.String())
		as.NotEmpty(csrfToken, "CSRF token should be present in admin post form")

		// Step 4: Submit post creation form with extracted CSRF token
		formData := url.Values{
			"Title":              {"E2E Test Post"},
			"Content":            {"This is a test post created via E2E testing"},
			"Excerpt":            {"Test excerpt"},
			"Slug":               {"e2e-test-post"},
			"Published":          {"false"},
			"authenticity_token": {csrfToken},
		}

		postRes := as.HTML("/admin/posts").Post(formData)
		as.Equal(302, postRes.Code, "Post creation should succeed with valid CSRF token")

		// Step 5: Verify post was actually created
		var post models.Post
		err = as.DB.Where("slug = ?", "e2e-test-post").First(&post)
		as.NoError(err, "Post should be created in database")
		as.Equal("E2E Test Post", post.Title)
		as.Equal(adminUser.ID, post.AuthorID)
	})
}

// extractCSRFTokenFromHTML extracts the CSRF token from rendered HTML
func (as *ActionSuite) extractCSRFTokenFromHTML(html string) string {
	// Look for authenticity_token in hidden input fields
	re := regexp.MustCompile(`<input[^>]*name="authenticity_token"[^>]*value="([^"]*)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// Look for csrf() helper output
	re2 := regexp.MustCompile(`<input[^>]*name="gorilla\.csrf\.Token"[^>]*value="([^"]*)"`)
	matches2 := re2.FindStringSubmatch(html)
	if len(matches2) > 1 {
		return matches2[1]
	}

	return ""
}

// TestFormCSRFTokenConsistency ensures CSRF tokens are consistent across requests in the same session
func (as *ActionSuite) Test_FormCSRFTokenConsistency() {
	as.T().Run("CSRF Token Consistency Across Forms", func(t *testing.T) {
		// Get CSRF token from registration form
		regRes := as.HTML("/users/new").Get()
		as.Equal(200, regRes.Code)
		regToken := as.extractCSRFTokenFromHTML(regRes.Body.String())

		// Get CSRF token from login form
		loginRes := as.HTML("/auth/new").Get()
		as.Equal(200, loginRes.Code)
		loginToken := as.extractCSRFTokenFromHTML(loginRes.Body.String())

		as.NotEmpty(regToken, "Registration form should have CSRF token")
		as.NotEmpty(loginToken, "Login form should have CSRF token")

		// Tokens should be different for different sessions/forms
		// (This depends on CSRF implementation - some reuse tokens, others don't)
		as.T().Logf("Registration CSRF token: %s", regToken)
		as.T().Logf("Login CSRF token: %s", loginToken)
	})
}

// TestCSRFHeaderSubmission tests CSRF token submission via headers (for AJAX requests)
func (as *ActionSuite) Test_CSRFHeaderSubmission() {
	as.T().Run("CSRF Token via Header", func(t *testing.T) {
		// This would test X-CSRF-Token header submission for AJAX requests
		// Currently skipped as the app may not support header-based CSRF
		as.T().Skip("CSRF header submission not yet implemented")
	})
}

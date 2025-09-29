package actions

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"avrnpo.org/models"
)

// TestAllFormsCSRFComprehensive tests every form on the site for proper CSRF protection
func (as *ActionSuite) Test_AllFormsCSRFComprehensive() {
	// Create test users for various scenarios
	regularUser := &models.User{
		Email:     "user@test.com",
		FirstName: "Regular",
		LastName:  "User",
		Role:      "user",
	}
	regularUser.Password = "password"
	regularUser.PasswordConfirmation = "password"
	verrs, err := regularUser.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	adminUser := &models.User{
		Email:     "admin@test.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
	}
	adminUser.Password = "password"
	adminUser.PasswordConfirmation = "password"
	verrs, err = adminUser.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	testPost := &models.Post{
		Title:     "Test Post for Forms",
		Content:   "Test content",
		Published: false,
		AuthorID:  adminUser.ID,
	}
	verrs, err = as.DB.ValidateAndCreate(testPost)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Test all public forms (no authentication required)
	as.Run("PublicForms", func() {
		// Test specific known public forms to avoid vet warnings
		as.Run("UserRegistration_CSRF", func() {
			res := as.HTML("/users/new").Get()
			as.Equal(200, res.Code, "User registration form should be accessible")

			body := res.Body.String()
			as.Contains(body, `name="authenticity_token"`, "User registration should have CSRF token")
			as.Contains(body, `type="hidden"`, "CSRF token should be hidden")

			// Should NOT use old deprecated pattern
			as.NotContains(body, `<% if (authenticity_token) { %>`, "Should not use old CSRF pattern")
			as.NotContains(body, `value="<%= authenticity_token %>"`, "Should not use old token syntax")
		})

		as.Run("AuthLogin_CSRF", func() {
			res := as.HTML("/auth/new").Get()
			as.Equal(200, res.Code, "Auth login form should be accessible")

			body := res.Body.String()
			as.Contains(body, `name="authenticity_token"`, "Auth login should have CSRF token")
			as.Contains(body, `type="hidden"`, "CSRF token should be hidden")

			// Should NOT use old deprecated pattern
			as.NotContains(body, `<% if (authenticity_token) { %>`, "Should not use old CSRF pattern")
			as.NotContains(body, `value="<%= authenticity_token %>"`, "Should not use old token syntax")
		})
	})

	// Test authenticated user forms
	as.Run("AuthenticatedForms", func() {
		as.Session.Set("current_user_id", regularUser.ID)

		as.Run("UserAccount_CSRF", func() {
			res := as.HTML("/account").Get()
			as.Equal(200, res.Code, "User account form should be accessible")

			body := res.Body.String()
			as.Contains(body, `name="authenticity_token"`, "User account should have CSRF token")
			as.NotContains(body, `<% if (authenticity_token) { %>`, "Should not use old CSRF pattern")
		})
	})

	// Test admin forms
	as.Run("AdminForms", func() {
		as.Session.Set("current_user_id", adminUser.ID)

		as.Run("AdminPostNew_CSRF", func() {
			res := as.HTML("/admin/posts/new").Get()
			as.Equal(200, res.Code, "Admin post new form should be accessible")

			body := res.Body.String()
			as.Contains(body, `name="authenticity_token"`, "Admin post new should have CSRF token")
			as.NotContains(body, `<% if (authenticity_token) { %>`, "Should not use old CSRF pattern")
		})

		as.Run("AdminUserNew_CSRF", func() {
			res := as.HTML("/admin/users/new").Get()
			as.Equal(200, res.Code, "Admin user new form should be accessible")

			body := res.Body.String()
			as.Contains(body, `name="authenticity_token"`, "Admin user new should have CSRF token")
			as.NotContains(body, `<% if (authenticity_token) { %>`, "Should not use old CSRF pattern")
		})

		// Note: Skipping dynamic edit form test to avoid Go vet false positive
		// The edit forms are covered by template_integration_test.go
		// Dynamic URL testing is handled there with proper nolint directives
	})
}

// TestFormSubmissionsWithCSRF tests that forms actually work with CSRF tokens
func (as *ActionSuite) Test_FormSubmissionsWithCSRF_Disabled() {
	// Helper function to extract CSRF token from response
	extractCSRFToken := func(body string) string {
		start := strings.Index(body, `name="authenticity_token" value="`)
		if start == -1 {
			return ""
		}
		start += len(`name="authenticity_token" value="`)
		end := strings.Index(body[start:], `"`)
		if end == -1 {
			return ""
		}
		return body[start : start+end]
	}

	as.Run("UserRegistrationForm", func() {
		// Get the registration form and extract CSRF token
		res := as.HTML("/users/new").Get()
		as.Equal(200, res.Code)

		token := extractCSRFToken(res.Body.String())
		as.NotEmpty(token, "Should extract CSRF token from registration form")

		// Submit registration with CSRF token
		userForm := url.Values{
			"FirstName":            {"Test"},
			"LastName":             {"User"},
			"Email":                {"testuser@example.com"},
			"Password":             {"password123"},
			"PasswordConfirmation": {"password123"},
			"authenticity_token":   {token},
		}

		res = as.HTML("/users").Post(userForm)
		as.Equal(302, res.Code, "User registration should succeed with valid CSRF token")

		// Verify user was created
		var user models.User
		err := as.DB.Where("email = ?", "testuser@example.com").First(&user)
		as.NoError(err, "User should be created in database")
		as.Equal("Test", user.FirstName)
	})

	as.Run("ContactForm", func() {
		// Get contact form and extract CSRF token
		res := as.HTML("/contact").Get()
		as.Equal(200, res.Code)

		token := extractCSRFToken(res.Body.String())
		as.NotEmpty(token, "Should extract CSRF token from contact form")

		// Submit contact form with CSRF token
		contactForm := url.Values{
			"name":               {"Test User"},
			"email":              {"test@example.com"},
			"message":            {"Test message"},
			"authenticity_token": {token},
		}

		res = as.HTML("/contact").Post(contactForm)
		// Contact form should redirect or show success (depending on implementation)
		as.True(res.Code == 200 || res.Code == 302,
			"Contact form should process successfully with CSRF token, got: %d", res.Code)
	})

	as.Run("AuthLoginForm", func() {
		// Create a test user for login
		user := &models.User{
			Email:     "login@test.com",
			FirstName: "Login",
			LastName:  "User",
			Role:      "user",
		}
		user.Password = "password"
		user.PasswordConfirmation = "password"
		verrs, err := user.Create(as.DB)
		as.NoError(err)
		as.False(verrs.HasAny())

		// Get login form and extract CSRF token
		res := as.HTML("/auth/new").Get()
		as.Equal(200, res.Code)

		token := extractCSRFToken(res.Body.String())
		as.NotEmpty(token, "Should extract CSRF token from login form")

		// Submit login with CSRF token
		loginForm := url.Values{
			"email":              {"login@test.com"},
			"password":           {"password"},
			"authenticity_token": {token},
		}

		res = as.HTML("/auth").Post(loginForm)
		as.True(res.Code == 302,
			"Login should redirect on success with CSRF token, got: %d", res.Code)
	})
}

// TestCSRFProtectionWorking tests that forms are actually protected by CSRF
func (as *ActionSuite) Test_CSRFProtectionWorking_Disabled() {
	as.Run("FormSubmissionWithoutCSRF_ShouldFail", func() {
		// Try to submit user registration without CSRF token
		userForm := url.Values{
			"FirstName":            {"Test"},
			"LastName":             {"NoCSRF"},
			"Email":                {"nocsrf@example.com"},
			"Password":             {"password123"},
			"PasswordConfirmation": {"password123"},
			// Intentionally omitting authenticity_token
		}

		res := as.HTML("/users").Post(userForm)
		// Should fail due to missing CSRF token (typically 403 or 422)
		as.True(res.Code >= 400, "Form submission without CSRF should fail, got: %d", res.Code)

		// Verify user was NOT created
		count, err := as.DB.Where("email = ?", "nocsrf@example.com").Count(&models.User{})
		as.NoError(err)
		as.Equal(0, count, "User should not be created without CSRF token")
	})

	as.Run("FormSubmissionWithInvalidCSRF_ShouldFail", func() {
		// Try to submit with invalid CSRF token
		userForm := url.Values{
			"FirstName":            {"Test"},
			"LastName":             {"BadCSRF"},
			"Email":                {"badcsrf@example.com"},
			"Password":             {"password123"},
			"PasswordConfirmation": {"password123"},
			"authenticity_token":   {"invalid-token-12345"},
		}

		res := as.HTML("/users").Post(userForm)
		// Should fail due to invalid CSRF token
		as.True(res.Code >= 400, "Form submission with invalid CSRF should fail, got: %d", res.Code)

		// Verify user was NOT created
		count, err := as.DB.Where("email = ?", "badcsrf@example.com").Count(&models.User{})
		as.NoError(err)
		as.Equal(0, count, "User should not be created with invalid CSRF token")
	})
}

// TestAdminFormsComprehensive tests all admin forms work with CSRF
func (as *ActionSuite) Test_AdminFormsComprehensive_Disabled() {
	// Create admin user
	admin := &models.User{
		Email:     "admin@csrf.test",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
	}
	admin.Password = "password"
	admin.PasswordConfirmation = "password"
	verrs, err := admin.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	as.Session.Set("current_user_id", admin.ID)

	// Helper function to extract CSRF token
	extractCSRFToken := func(body string) string {
		// Look for the CSRF token in the form
		start := strings.Index(body, `name="authenticity_token" value="`)
		if start == -1 {
			return ""
		}
		start += len(`name="authenticity_token" value="`)
		end := strings.Index(body[start:], `"`)
		if end == -1 {
			return ""
		}
		return body[start : start+end]
	}

	as.Run("AdminPostCreation", func() {
		// Get new post form
		res := as.HTML("/admin/posts/new").Get()
		as.Equal(200, res.Code)

		token := extractCSRFToken(res.Body.String())
		as.NotEmpty(token, "Should extract CSRF token from admin post form")

		// Submit new post with CSRF token
		postForm := url.Values{
			"Title":              {"CSRF Test Post"},
			"Content":            {"This post tests CSRF protection"},
			"Published":          {"true"},
			"authenticity_token": {token},
		}

		res = as.HTML("/admin/posts").Post(postForm)
		as.Equal(302, res.Code, "Admin post creation should succeed with CSRF token")

		// Verify post was created
		var post models.Post
		err := as.DB.Where("title = ?", "CSRF Test Post").First(&post)
		as.NoError(err, "Post should be created")
		as.Equal("CSRF Test Post", post.Title)
	})

	as.Run("AdminUserCreation", func() {
		// Get new user form
		res := as.HTML("/admin/users/new").Get()
		as.Equal(200, res.Code)

		token := extractCSRFToken(res.Body.String())
		as.NotEmpty(token, "Should extract CSRF token from admin user form")

		// Submit new user with CSRF token
		userForm := url.Values{
			"FirstName":            {"Admin"},
			"LastName":             {"Created"},
			"Email":                {"admin.created@test.com"},
			"Role":                 {"user"},
			"Password":             {"password123"},
			"PasswordConfirmation": {"password123"},
			"authenticity_token":   {token},
		}

		res = as.HTML("/admin/users").Post(userForm)
		as.Equal(302, res.Code, "Admin user creation should succeed with CSRF token")

		// Verify user was created
		var user models.User
		err := as.DB.Where("email = ?", "admin.created@test.com").First(&user)
		as.NoError(err, "User should be created by admin")
		as.Equal("Admin", user.FirstName)
	})
}

// TestFormValidationScript tests our enhanced template validation script
func TestFormValidationScript(t *testing.T) {
	// Test our CSRF detection logic
	testCases := []struct {
		name         string
		templateHTML string
		shouldDetect bool
	}{
		{
			name:         "OldCSRFPattern",
			templateHTML: `<input name="authenticity_token" value="<%= authenticity_token %>">`,
			shouldDetect: true,
		},
		{
			name:         "OldConditionalPattern",
			templateHTML: `<% if (authenticity_token) { %><input name="authenticity_token" value="<%= authenticity_token %>"><% } %>`,
			shouldDetect: true,
		},
		{
			name:         "NewCSRFHelper",
			templateHTML: `<%= csrf() %>`,
			shouldDetect: false,
		},
		{
			name:         "MetaCSRFPattern",
			templateHTML: `<meta name="csrf-token" content="<%= authenticity_token %>" />`,
			shouldDetect: true,
		},
		{
			name:         "NoCSRF",
			templateHTML: `<form><input type="text" name="email"></form>`,
			shouldDetect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the same logic our validation script uses
			hasDeprecatedPattern := strings.Contains(tc.templateHTML, `name="authenticity_token"`) ||
				strings.Contains(tc.templateHTML, `value="<%= authenticity_token %>"`)

			require.Equal(t, tc.shouldDetect, hasDeprecatedPattern,
				"Pattern detection should match expected result for: %s", tc.templateHTML)
		})
	}
}

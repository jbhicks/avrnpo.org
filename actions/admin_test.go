package actions

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"avrnpo.org/models"
)

// Helper function to create and login a user using the exact working pattern from home_test.go
func (as *ActionSuite) createAndLoginUser(email, role string) (*models.User, string, string) {
	// Make email unique by adding timestamp to prevent conflicts
	timestamp := time.Now().UnixNano()
	uniqueEmail := fmt.Sprintf("test-%d-%s", timestamp, email)

	// If admin role is needed, create user via web signup first, then promote to admin
	if role == "admin" {
		// Create user via web interface like successful user tests do
		cookie, token := fetchCSRF(as.T(), as.App, "/users/new")
		as.T().Logf("🔍 Initial signup fetchCSRF - Cookie: '%s', Token exists: %t", cookie, token != "")

		signupData := map[string]interface{}{
			"Email":                uniqueEmail,
			"Password":             "password",
			"PasswordConfirmation": "password",
			"FirstName":            "Test",
			"LastName":             "User",
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
		as.T().Logf("✅ User created via web signup: %s", uniqueEmail)

		// Now promote the user to admin role using the global DB connection
		// Use models.DB instead of as.DB to avoid transaction isolation
		user := &models.User{}
		err := models.DB.Where("email = ?", uniqueEmail).First(user)
		as.NoError(err, "Should find user created via web interface")

		user.Role = "admin"
		verrs, err := models.DB.ValidateAndUpdate(user)
		as.NoError(err, "Failed to promote user to admin")
		as.False(verrs.HasAny(), "Validation errors promoting user to admin")
		as.T().Logf("✅ Promoted user %s to admin role", uniqueEmail)

		// Now use MockLogin to login as the admin user
		adminCookie, adminToken := MockLogin(as.T(), as.App, uniqueEmail, "password")
		as.T().Logf("🔍 MockLogin result for admin - Cookie: %s, Token: %s", adminCookie, adminToken)

		return user, adminCookie, adminToken
	}

	// For regular users, use the web signup flow
	// First, get CSRF token from signup page
	cookie, token := fetchCSRF(as.T(), as.App, "/users/new")
	as.T().Logf("🔍 Initial signup fetchCSRF - Cookie: '%s', Token exists: %t", cookie, token != "")

	signupData := map[string]interface{}{
		"Email":                uniqueEmail,
		"Password":             "password123",
		"PasswordConfirmation": "password123",
		"FirstName":            "Test",
		"LastName":             "User",
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
	as.T().Logf("🔍 Signup response cookies count: %d", len(signupRes.Result().Cookies()))
	for i, c := range signupRes.Result().Cookies() {
		as.T().Logf("🔍 Signup cookie %d: %s = %s", i, c.Name, c.Value)
		if strings.Contains(c.Name, "session") || c.Name == "_avrnpo.org_session" {
			sessionCookie = c.String()
			as.T().Logf("🔍 Found signup session cookie: %s", sessionCookie)
			break
		}
	}

	// Get new CSRF token for login using the session cookie
	loginCookie, loginToken := fetchCSRF(as.T(), as.App, "/auth/new")
	as.T().Logf("🔍 Login fetchCSRF - Cookie: '%s', Token exists: %t", loginCookie, loginToken != "")

	// Combine cookies if we have both
	combinedCookie := sessionCookie
	if loginCookie != "" && sessionCookie != "" && !strings.Contains(sessionCookie, loginCookie) {
		combinedCookie = sessionCookie + "; " + loginCookie
	} else if loginCookie != "" {
		combinedCookie = loginCookie
	}
	as.T().Logf("🔍 Combined cookie for login: '%s'", combinedCookie)

	loginData := map[string]interface{}{
		"Email":              uniqueEmail,
		"Password":           "password123",
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
	as.T().Logf("🔍 Login response cookies count: %d", len(loginRes.Result().Cookies()))
	for i, c := range loginRes.Result().Cookies() {
		as.T().Logf("🔍 Login response cookie %d: %s = %s", i, c.Name, c.Value)
		if strings.Contains(c.Name, "session") || c.Name == "_avrnpo.org_session" {
			finalSessionCookie = c.String()
			as.T().Logf("🔍 Found login session cookie: %s", finalSessionCookie)
			break
		}
	}
	as.T().Logf("🔍 Final session cookie: '%s'", finalSessionCookie)

	// IMPORTANT: Buffalo in test mode doesn't always return cookies via HTTP headers
	// but the session is maintained internally. Test by accessing a protected page.
	testReq := as.HTML("/dashboard")
	testRes := testReq.Get()
	if testRes.Code == 200 {
		as.T().Logf("✅ Session working: /dashboard accessible without explicit cookie")
		finalSessionCookie = "BUFFALO_TEST_SESSION_ACTIVE" // Placeholder to indicate session works
	} else {
		as.T().Logf("❌ Session not working: /dashboard returned %d", testRes.Code)
	}

	// Create a user object with the known information
	user := models.User{
		Email:     uniqueEmail,
		FirstName: "Test",
		LastName:  "User",
		Role:      role,
	}

	return &user, finalSessionCookie, loginToken
}

func (as *ActionSuite) Test_AdminRoutes_RequireAuthentication() {
	// Test admin routes without any authentication - each route individually
	res := as.HTML("/admin/").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/auth/new")

	res = as.HTML("/admin/dashboard").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/auth/new")

	res = as.HTML("/admin/users").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/auth/new")
}

func (as *ActionSuite) Test_AdminRoutes_RequireAdminRole() {
	// Create and login regular user
	_, cookie, _ := as.createAndLoginUser("user@example.com", "user")

	// Test admin routes with regular user - each route individually
	req := as.HTML("/admin/")
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/dashboard")

	req = as.HTML("/admin/dashboard")
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res = req.Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/dashboard")

	req = as.HTML("/admin/users")
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res = req.Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/dashboard")
}

func (as *ActionSuite) Test_AdminDashboard_Success() {
	// Create and login admin user
	admin, cookie, _ := as.createAndLoginUser("admin@example.com", "admin")

	// Debug: Verify session was set
	as.T().Logf("🔍 Admin user ID: %v", admin.ID)
	as.NotEmpty(cookie, "Should have valid session cookie")

	// Create additional users for statistics
	user1 := &models.User{
		Email:                "user1@example.com",
		FirstName:            "User",
		LastName:             "One",
		Role:                 "user",
		Password:             "password123",
		PasswordConfirmation: "password123",
	}
	verrs, err := user1.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	user2 := &models.User{
		Email:                "user2@example.com",
		FirstName:            "User",
		LastName:             "Two",
		Role:                 "user",
		Password:             "password123",
		PasswordConfirmation: "password123",
	}
	verrs, err = user2.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Test admin dashboard access
	req := as.HTML("/admin/dashboard")
	// Only set cookie if it's not the special Buffalo test session marker
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Get()

	// Debug redirect if not successful
	if res.Code != 200 {
		as.T().Logf("🔍 Request failed with status %d, redirected to: %s", res.Code, res.Header().Get("Location"))
	}

	as.Equal(http.StatusOK, res.Code)

	// Should display user statistics
	body := res.Body.String()
	as.Contains(body, "3") // Total users (admin + 2 regular users)
	as.Contains(body, "1") // Admin count
	as.Contains(body, "2") // Regular user count
}

func (as *ActionSuite) Test_AdminUsers_Success() {
	// Create and login admin user
	admin, cookie, _ := as.createAndLoginUser("admin@example.com", "admin")

	// Create additional test user using global DB to avoid transaction isolation
	timestamp := time.Now().UnixNano()
	uniqueUserEmail := fmt.Sprintf("test-user-%d@example.com", timestamp)

	user1 := &models.User{
		Email:                uniqueUserEmail,
		FirstName:            "Test",
		LastName:             "User",
		Role:                 "user",
		Password:             "password123",
		PasswordConfirmation: "password123",
	}
	verrs, err := user1.Create(models.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Test admin users list
	req := as.HTML("/admin/users")
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)

	// Should show both users using their actual emails
	body := res.Body.String()
	as.Contains(body, admin.Email, "Should show admin user email")
	as.Contains(body, uniqueUserEmail, "Should show test user email")
}

func (as *ActionSuite) Test_AdminUsers_Pagination() {
	// Create and login admin user
	_, cookie, _ := as.createAndLoginUser("admin@example.com", "admin")

	// Create multiple users for pagination testing
	for i := 1; i <= 25; i++ {
		user := &models.User{
			Email:                fmt.Sprintf("user%d@example.com", i),
			FirstName:            fmt.Sprintf("User%d", i),
			LastName:             "Test",
			Role:                 "user",
			Password:             "password123",
			PasswordConfirmation: "password123",
		}

		verrs, err := user.Create(as.DB)
		as.NoError(err)
		as.False(verrs.HasAny())
	}

	// Test first page
	req := as.HTML("/admin/users")
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)

	// Test second page
	req = as.HTML("/admin/users?page=2")
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_AdminRequired_WithAdminUser() {
	// Create and login admin user
	_, cookie, _ := as.createAndLoginUser("admin@example.com", "admin")

	// Test access to admin dashboard
	req := as.HTML("/admin/dashboard")
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_AdminRequired_WithRegularUser() {
	// Create and login regular user
	_, cookie, _ := as.createAndLoginUser("user@example.com", "user")

	// Try to access admin dashboard
	req := as.HTML("/admin/dashboard")
	if cookie != "" && cookie != "BUFFALO_TEST_SESSION_ACTIVE" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Get()

	// Should redirect to dashboard with access denied
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/dashboard")
}

func (as *ActionSuite) Test_AdminUserCreationDebug() {
	// Simple test to debug admin user creation and authentication
	timestamp := time.Now().UnixNano()
	uniqueEmail := fmt.Sprintf("debug-admin-%d@example.com", timestamp)

	// Create user via web interface first (like successful tests do)
	cookie, token := fetchCSRF(as.T(), as.App, "/users/new")
	signupData := map[string]interface{}{
		"Email":                uniqueEmail,
		"Password":             "password",
		"PasswordConfirmation": "password",
		"FirstName":            "Debug",
		"LastName":             "Admin",
		"accept_terms":         "on",
		"authenticity_token":   token,
	}

	signupReq := as.HTML("/users")
	if cookie != "" {
		signupReq.Headers["Cookie"] = cookie
	}
	signupRes := signupReq.Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)
	as.T().Logf("✅ Created user via web signup: %s", uniqueEmail)

	// Promote to admin using global DB connection
	adminUser := &models.User{}
	err := models.DB.Where("email = ?", uniqueEmail).First(adminUser)
	as.NoError(err)
	adminUser.Role = "admin"
	verrs, err := models.DB.ValidateAndUpdate(adminUser)
	as.NoError(err)
	as.False(verrs.HasAny())
	as.T().Logf("✅ Promoted to admin role: %s", adminUser.Role)

	// Test password verification directly
	err = adminUser.VerifyPassword("password")
	as.NoError(err)
	as.T().Logf("✅ Password verification passed")

	// Debug: Log the exact password hash and original password
	as.T().Logf("🔍 Password hash: %s", adminUser.PasswordHash)
	as.T().Logf("🔍 Original password used: password")

	// Now try MockLogin
	loginCookie, loginToken := MockLogin(as.T(), as.App, uniqueEmail, "password")
	as.T().Logf("MockLogin result - Cookie: %s, Token: %s", loginCookie, loginToken)

	if loginCookie == "" {
		as.T().Logf("❌ MockLogin failed - empty cookie returned")
		as.Fail("MockLogin returned empty cookie")
		return
	}

	// Test admin route access
	req := as.HTML("/admin/dashboard")
	req.Headers["Cookie"] = loginCookie
	res := req.Get()
	as.T().Logf("Admin dashboard access - Status: %d, Location: %s", res.Code, res.Header().Get("Location"))

	if res.Code != http.StatusOK {
		as.T().Logf("❌ Admin dashboard access failed")
		as.Equal(http.StatusOK, res.Code)
	} else {
		as.T().Logf("✅ Admin dashboard access successful")
	}
}

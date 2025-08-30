package actions

import (
	"fmt"
	"net/http"
	"time"
)

func (as *ActionSuite) Test_Auth_Signin() {
	res := as.HTML("/auth/").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/auth/new", res.Header().Get("Location"))
}

func (as *ActionSuite) Test_Auth_New() {
	res := as.HTML("/auth/new").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Sign In")
}

func (as *ActionSuite) Test_Auth_Create() {
	timestamp := time.Now().UnixNano()

	// Fetch CSRF token from signup page
	cookie, token := fetchCSRF(as.T(), as.App, "/users/new")

	// Create a user through the signup endpoint (which works)
	signupData := map[string]interface{}{
		"Email":                fmt.Sprintf("mark-%d@example.com", timestamp),
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

	// Extract email for auth tests
	userEmail := fmt.Sprintf("mark-%d@example.com", timestamp)

	tcases := []struct {
		Email       string
		Password    string
		Status      int
		RedirectURL string

		Identifier string
	}{
		{userEmail, "password", http.StatusFound, "/", "Valid"},
		{"noexist@example.com", "password", http.StatusUnauthorized, "", "Email Invalid"},
		{userEmail, "invalidPassword", http.StatusUnauthorized, "", "Password Invalid"},
	}

	for _, tcase := range tcases {
		as.Run(tcase.Identifier, func() {
			// Fetch token for login
			_, loginToken := fetchCSRF(as.T(), as.App, "/auth/new")

			res := as.HTML("/auth").Post(map[string]interface{}{
				"Email":             tcase.Email,
				"Password":          tcase.Password,
				"authenticity_token": loginToken,
			})

			as.Equal(tcase.Status, res.Code)
			as.Equal(tcase.RedirectURL, res.Location())
		})
	}
}

func (as *ActionSuite) Test_Auth_Redirect() {
	timestamp := time.Now().UnixNano()

	// Fetch CSRF token for signup
	cookie, token := fetchCSRF(as.T(), as.App, "/users/new")

	// Create a user through the signup endpoint (which works)
	signupData := map[string]interface{}{
		"Email":                fmt.Sprintf("redirect-%d@example.com", timestamp),
		"Password":             "password",
		"PasswordConfirmation": "password",
		"FirstName":            "Redirect",
		"LastName":             "Test",
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

	// Create auth data for login attempts
	userEmail := fmt.Sprintf("redirect-%d@example.com", timestamp)
	authData := map[string]interface{}{
		"Email":    userEmail,
		"Password": "password",
	}

	tcases := []struct {
		redirectURL    interface{}
		resultLocation string

		identifier string
	}{
		{"/some/url", "/some/url", "RedirectURL defined"},
		{nil, "/", "RedirectURL nil"},
		{"", "/", "RedirectURL empty"},
	}

	for _, tcase := range tcases {
		as.Run(tcase.identifier, func() {
			as.Session.Set("redirectURL", tcase.redirectURL)

			// Fetch token for login
			_, loginToken := fetchCSRF(as.T(), as.App, "/auth/new")

			req := as.HTML("/auth")
			res := req.Post(map[string]interface{}{
				"Email":             authData["Email"],
				"Password":          authData["Password"],
				"authenticity_token": loginToken,
			})

			as.Equal(http.StatusFound, res.Code)
			as.Equal(res.Location(), tcase.resultLocation)
		})
	}

	for _, tcase := range tcases {
		as.Run(tcase.identifier, func() {
			as.Session.Set("redirectURL", tcase.redirectURL)

			// Fetch token for signup (but this is login test, so use login token)
			_, _ = fetchCSRF(as.T(), as.App, "/auth/new")

			req := as.HTML("/auth")
			res := req.Post(signupData)

			as.Equal(http.StatusFound, res.Code)
			as.Equal(res.Location(), tcase.resultLocation)
		})
	}
}

func (as *ActionSuite) Test_Auth_Create_Password_Preservation() {
	timestamp := time.Now().UnixNano()

	// Fetch CSRF token for signup
	cookie, token := fetchCSRF(as.T(), as.App, "/users/new")

	// This test specifically verifies that the plaintext password is preserved
	// during the authentication process and not overwritten by the database query.
	// This would catch the bug where tx.First(u) overwrites the Password field.

	// Create a user through the signup endpoint (which works)
	signupData := map[string]interface{}{
		"Email":                fmt.Sprintf("test.password.preservation-%d@example.com", timestamp),
		"Password":             "secretpassword123",
		"PasswordConfirmation": "secretpassword123",
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

	// Now attempt to login with the correct password
	// This should succeed if the password is properly preserved during auth
	userEmail := fmt.Sprintf("test.password.preservation-%d@example.com", timestamp)

	// Fetch token for login
	loginCookie, loginToken := fetchCSRF(as.T(), as.App, "/auth/new")

	loginReq := as.HTML("/auth")
	if loginCookie != "" {
		loginReq.Headers["Cookie"] = loginCookie
	}
	res := loginReq.Post(map[string]interface{}{
		"Email":             userEmail,
		"Password":          "secretpassword123", // Same password used during creation
		"authenticity_token": loginToken,
	})

	// Should redirect to home page on successful authentication
	as.Equal(http.StatusFound, res.Code, "Authentication should succeed with correct password")
	as.Equal("/", res.Location(), "Should redirect to home page after successful login")

	// Verify session was set
	sessionUserID := as.Session.Get("current_user_id")
	as.NotNil(sessionUserID, "Session should contain current_user_id after successful login")
}

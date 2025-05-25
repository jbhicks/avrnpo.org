package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// HomeHandler serves the public landing page
func HomeHandler(c buffalo.Context) error {
	// Check if we have a current_user_id in the session (like the test sets)
	userID := c.Session().Get("current_user_id")

	// Set a simple boolean flag for the template
	if userID != nil {
		c.Set("user_logged_in", true)
		// Also try to get the user object from the middleware
		if user := c.Value("current_user"); user != nil {
			c.Set("current_user", user)
		}
	} else {
		c.Set("user_logged_in", false)
		c.Set("current_user", nil)
	}

	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// DashboardHandler serves the protected dashboard for authenticated users
func DashboardHandler(c buffalo.Context) error {
	// This will be protected by the Authorize middleware
	return c.Render(http.StatusOK, r.HTML("home/dashboard.plush.html"))
}

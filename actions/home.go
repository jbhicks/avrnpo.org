package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// HomeHandler is a default handler to serve up a home page.
func HomeHandler(c buffalo.Context) error {
	// Check if user is authenticated
	if c.Value("current_user") == nil {
		// Redirect to login page if not authenticated
		return c.Redirect(http.StatusFound, "/auth/new")
	}

	// If authenticated, render the home page
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

package actions

import (
	"fmt"
	"avrnpo.org/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
)

// HomeHandler serves the public landing page
func HomeHandler(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	userID := c.Session().Get("current_user_id")
	if userID != nil {
		c.Set("user_logged_in", true)
		// current_user is already set by SetCurrentUser middleware
		// Don't override it here
	} else {
		c.Set("user_logged_in", false)
		// current_user should already be nil from middleware
	}

	// Fetch recent published blog posts for homepage
	posts := []models.Post{}
	err := tx.Where("published = ?", true).
		Order("created_at desc").
		Limit(3).
		Eager("User").
		All(&posts)
	if err != nil {
		c.Logger().Errorf("Error fetching posts for homepage: %v", err)
		// Don't fail the homepage if posts can't be loaded
		posts = []models.Post{}
	}
	c.Set("recentPosts", posts)

	// Render the home page (using application layout for consistency)
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// DashboardHandler serves the protected dashboard for authenticated users
func DashboardHandler(c buffalo.Context) error {
	// Get current_user
	currentUser, ok := c.Value("current_user").(*models.User)
	if !ok || currentUser == nil {
		// This should ideally not happen if AuthMiddleware is working
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// You can pass additional data to the template if needed
	c.Set("user", currentUser) // This is the same as current_user, but explicit for template

	// Since we're using single-template architecture, just render the dashboard template
	return c.Render(http.StatusOK, r.HTML("home/dashboard.plush.html"))
}

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
		if user := c.Value("current_user"); user != nil {
			c.Set("current_user", user)
		}
	} else {
		c.Set("user_logged_in", false)
		c.Set("current_user", nil)
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

	htmxRequest := IsHTMX(c.Request())
	c.LogField("is_htmx_request_in_handler_for_home", htmxRequest)

	if htmxRequest {
		// For HTMX requests, render only the content part
		return c.Render(http.StatusOK, rHTMX.HTML("home/_index_content"))
	}

	// For direct page loads, render the main index with home content
	c.Set("currentPath", "/")
	c.Set("initialContent", "home/_index_content")
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

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.HTML("home/dashboard.plush.html"))
	}

	// Direct access - render full page with navigation
	return c.Render(http.StatusOK, r.HTML("home/dashboard_full.plush.html"))
}

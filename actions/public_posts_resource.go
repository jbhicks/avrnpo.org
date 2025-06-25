package actions

import (
	"fmt"
	"net/http"
	"avrnpo.org/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
)

// PublicPostsResource handles public blog post display (non-admin)
type PublicPostsResource struct {
	buffalo.BaseResource
}

// List displays all published posts for public blog (GET /blog)
func (ppr PublicPostsResource) List(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	posts := []models.Post{}
	// Get published posts ordered by created_at desc
	if err := tx.Where("published = ?", true).Order("created_at desc").All(&posts); err != nil {
		return err
	}

	c.Set("posts", posts)
	
	// Set base URL for social sharing
	req := c.Request()
	scheme := "http"
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	c.Set("baseURL", scheme+"://"+req.Host)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.HTML("blog/_index_simple.plush.html"))
	}

	// Direct access - render full blog index page
	return c.Render(http.StatusOK, r.HTML("blog/index_full.plush.html"))
}

// Show displays a single published post by slug (GET /blog/{slug})
func (ppr PublicPostsResource) Show(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	slug := c.Param("slug")
	post := &models.Post{}

	// Find published post by slug with user relationship
	if err := tx.Eager("User").Where("slug = ? AND published = ?", slug, true).First(post); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	c.Set("post", post)
	
	// Set base URL for social sharing
	req := c.Request()
	scheme := "http"
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	c.Set("baseURL", scheme+"://"+req.Host)

	// Always return full page - hx-boost with hx-select will extract the content
	return c.Render(http.StatusOK, r.HTML("blog/show_full.plush.html"))
}

// Create, Update, Destroy methods not implemented for public resource
// These will return 404 as they inherit from BaseResource

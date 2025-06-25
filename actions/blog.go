package actions

import (
	"fmt"
	"avrnpo.org/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
)

// BlogIndex displays all published posts
func BlogIndex(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	posts := []models.Post{}
	// Get published posts ordered by created_at desc (simplified - no user loading for now)
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
	return c.Render(200, r.HTML("blog/index.plush.html"))
}

// BlogShow displays a single post by slug
func BlogShow(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	slug := c.Param("slug")
	post := &models.Post{}

	// Find published post by slug with user relationship
	if err := tx.Eager("User").Where("slug = ? AND published = ?", slug, true).First(post); err != nil {
		return c.Error(404, err)
	}
	c.Set("post", post)
	
	// Set base URL for social sharing
	req := c.Request()
	scheme := "http"
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	c.Set("baseURL", scheme+"://"+req.Host)

	// Always return full page - hx-boost will handle navigation
	return c.Render(200, r.HTML("blog/show.plush.html"))
}
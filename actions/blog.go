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

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(200, r.HTML("blog/_index_simple.plush.html"))
	}

	// Direct access - render universal layout with blog content
	c.Set("blogContent", true)
	return c.Render(200, r.HTML("home/index.plush.html"))
}

// BlogShow displays a single post by slug
func BlogShow(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	slug := c.Param("slug")
	post := &models.Post{}

	// Find published post by slug
	if err := tx.Where("slug = ? AND published = ?", slug, true).First(post); err != nil {
		return c.Error(404, err)
	}
	c.Set("post", post)

	// Always return full page - hx-boost with hx-select will extract the content
	return c.Render(200, r.HTML("blog/show_full.plush.html"))
}
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

	// Get published posts ordered by created_at desc
	if err := tx.Where("published_at IS NOT NULL").Order("created_at desc").All(&posts); err != nil {
		return err
	}

	// Load authors for each post
	for i := range posts {
		if err := tx.Load(&posts[i], "User"); err != nil {
			return err
		}
	}

	c.Set("posts", posts)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(200, r.HTML("blog/index.plush.html"))
	}

	// Direct access - render full page with navigation
	return c.Render(200, r.HTML("blog/index_full.plush.html"))
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
	if err := tx.Where("slug = ? AND published_at IS NOT NULL", slug).First(post); err != nil {
		return c.Error(404, err)
	}

	// Load author
	if err := tx.Load(post, "User"); err != nil {
		return err
	}

	c.Set("post", post)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(200, r.HTML("blog/show.plush.html"))
	}

	// Direct access - render full page with navigation
	return c.Render(200, r.HTML("blog/show_full.plush.html"))
}

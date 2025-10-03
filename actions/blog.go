package actions

import (
	"avrnpo.org/models"
	"fmt"
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
)

// BlogIndex displays all published posts
func BlogIndex(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	posts := []models.Post{}
	// Get published posts ordered by published_at desc with user relationships
	if err := tx.Eager("User").Where("published = ?", true).Order("published_at desc").All(&posts); err != nil {
		return errors.WithStack(err)
	}

	c.Set("posts", posts)

	// Set base URL for social sharing
	req := c.Request()
	scheme := "http"
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	c.Set("baseURL", scheme+"://"+req.Host)

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

	return c.Render(200, r.HTML("blog/show.plush.html"))
}

// BlogLoadMore loads additional posts for infinite scroll
func BlogLoadMore(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	pageStr := c.Param("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Define posts per page (e.g., 10)
	postsPerPage := 10

	posts := []models.Post{}
	// Get published posts with pagination
	if err := tx.Where("published = ?", true).Order("created_at desc").Paginate(page, postsPerPage).All(&posts); err != nil {
		return errors.WithStack(err)
	}

	// Check if there are more posts
	hasMore := len(posts) == postsPerPage

	// Return JSON response
	return c.Render(200, r.JSON(map[string]interface{}{
		"posts":    posts,
		"hasMore":  hasMore,
		"nextPage": page + 1,
	}))
}

package actions

import (
	"avrnpo.org/models"
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
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

	// Render blog index page
	return c.Render(http.StatusOK, r.HTML("blog/index.plush.html"))
}

// Show displays a single post by slug (GET /blog/{slug})
func (ppr PublicPostsResource) Show(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	slug := c.Param("public_post_id")
	post := &models.Post{}

	// Find post by slug with user relationship
	if err := tx.Eager("User").Where("slug = ?", slug).First(post); err != nil {
		return c.Error(404, err)
	}

	// Check if current user is authenticated and can preview
	currentUser := c.Value("current_user")
	if currentUser != nil {
		user := currentUser.(*models.User)
		// Allow preview if user is admin or the post author
		if user.Role != "admin" && user.ID != post.AuthorID {
			// Not authorized to preview, check if published
			if !post.Published {
				return c.Error(404, fmt.Errorf("post not found"))
			}
		}
	} else {
		// For non-authenticated users, only show published posts
		if !post.Published {
			return c.Error(404, fmt.Errorf("post not found"))
		}
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
	return c.Render(http.StatusOK, r.HTML("blog/show.plush.html"))
}

// Create, Update, Destroy methods not implemented for public resource
// These will return 404 as they inherit from BaseResource

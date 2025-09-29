package actions

import (
	"avrnpo.org/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
)

// PostsResource handles CRUD operations for blog posts in admin area
type PostsResource struct {
	buffalo.BaseResource
}

// List displays all posts for admin management (GET /admin/posts)
func (pr PostsResource) List(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	posts := []models.Post{}
	// Get all posts (published and unpublished) with pagination
	query := tx.Eager("User").Order("created_at desc")

	// Handle search functionality
	if search := c.Param("search"); search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
		c.Set("search", search)
	}

	// Handle filtering by status
	if status := c.Param("status"); status != "" {
		if status == "published" {
			query = query.Where("published = ?", true)
		} else if status == "draft" {
			query = query.Where("published = ?", false)
		}
		c.Set("status", status)
	}

	if err := query.All(&posts); err != nil {
		return errors.WithStack(err)
	}

	c.Set("posts", posts)

	// Always return the complete page - Single Template Architecture
	return c.Render(http.StatusOK, r.HTML("admin/posts/index.plush.html"))
}

// Show displays a single post for admin (GET /admin/posts/{post_id})
func (pr PostsResource) Show(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	postID := c.Param("post_id")
	post := &models.Post{}

	if err := tx.Eager("User").Find(post, postID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("post", post)

	// Always return the complete page - Single Template Architecture
	return c.Render(http.StatusOK, r.HTML("admin/posts/show.plush.html"))
}

// New displays the form for creating a new post (GET /admin/posts/new)
func (pr PostsResource) New(c buffalo.Context) error {
	post := &models.Post{}
	c.Set("post", post)
	c.Set("csrf", c.Value("authenticity_token"))

	// Always return the complete page - Single Template Architecture
	return c.Render(http.StatusOK, r.HTML("admin/posts/new.plush.html"))
}

// Create handles creation of new blog posts (POST /admin/posts)
func (pr PostsResource) Create(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	post := &models.Post{}
	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}

	// Set the author to current user
	if currentUser := c.Value("current_user"); currentUser != nil {
		if user, ok := currentUser.(*models.User); ok {
			post.AuthorID = user.ID
		}
	}

	// Generate slug from title if not provided
	if post.Slug == "" {
		post.GenerateSlug()
	}

	// Validate and create post
	if verrs, err := tx.ValidateAndCreate(post); err != nil {
		return errors.WithStack(err)
	} else if verrs.HasAny() {
		c.Set("post", post)
		c.Set("errors", verrs)
		c.Set("csrf", c.Value("authenticity_token"))

		// Always return complete page for validation errors
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/posts/new.plush.html"))
	}

	c.Flash().Add("success", "Post created successfully!")

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/posts/%d", post.ID))
}

// Edit displays the form for editing a post (GET /admin/posts/{post_id}/edit)
func (pr PostsResource) Edit(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	postID := c.Param("post_id")
	post := &models.Post{}

	if err := tx.Find(post, postID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("post", post)
	c.Set("csrf", c.Value("authenticity_token"))

	// Always return the complete page - Single Template Architecture
	return c.Render(http.StatusOK, r.HTML("admin/posts/edit.plush.html"))
}

// Update handles updating blog posts (PUT /admin/posts/{post_id})
func (pr PostsResource) Update(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	postID := c.Param("post_id")
	post := &models.Post{}

	if err := tx.Find(post, postID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}

	// Generate slug from title if changed
	if post.Slug == "" {
		post.GenerateSlug()
	}

	if verrs, err := tx.ValidateAndUpdate(post); err != nil {
		return errors.WithStack(err)
	} else if verrs.HasAny() {
		c.Set("post", post)
		c.Set("errors", verrs)
		c.Set("csrf", c.Value("authenticity_token"))

		// Always return complete page for validation errors
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/posts/edit.plush.html"))
	}

	c.Flash().Add("success", "Post updated successfully!")

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/posts/%d", post.ID))
}

// Destroy deletes a blog post (DELETE /admin/posts/{post_id})
func (pr PostsResource) Destroy(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	postID := c.Param("post_id")
	post := &models.Post{}

	if err := tx.Find(post, postID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(post); err != nil {
		return errors.WithStack(err)
	}

	c.Flash().Add("success", "Post deleted successfully!")

	return c.Redirect(http.StatusSeeOther, "/admin/posts")
}

// Bulk handles bulk operations on posts (POST /admin/posts/bulk)
func (pr PostsResource) Bulk(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	action := c.Param("action")
	postIDsParam := c.Request().FormValue("post_ids")
	var postIDs []string
	if postIDsParam != "" {
		postIDs = []string{postIDsParam}
	}
	// Handle multiple post_ids if sent as array
	if values := c.Request().Form["post_ids"]; len(values) > 0 {
		postIDs = values
	}

	if action == "" || len(postIDs) == 0 {
		c.Flash().Add("danger", "No action or posts selected")
		return c.Redirect(http.StatusSeeOther, "/admin/posts")
	}

	// Convert string IDs to integers
	var ids []int
	for _, idStr := range postIDs {
		if id, err := strconv.Atoi(idStr); err == nil {
			ids = append(ids, id)
		}
	}

	switch action {
	case "publish":
		now := time.Now()
		err := tx.RawQuery("UPDATE posts SET published = true, published_at = ? WHERE id IN (?)", now, ids).Exec()
		if err != nil {
			return errors.WithStack(err)
		}
		c.Flash().Add("success", fmt.Sprintf("Published %d posts", len(ids)))

	case "unpublish":
		err := tx.RawQuery("UPDATE posts SET published = false, published_at = NULL WHERE id IN (?)", ids).Exec()
		if err != nil {
			return errors.WithStack(err)
		}
		c.Flash().Add("success", fmt.Sprintf("Unpublished %d posts", len(ids)))

	case "delete":
		err := tx.RawQuery("DELETE FROM posts WHERE id IN (?)", ids).Exec()
		if err != nil {
			return errors.WithStack(err)
		}
		c.Flash().Add("success", fmt.Sprintf("Deleted %d posts", len(ids)))

	default:
		c.Flash().Add("danger", "Unknown bulk action")
	}

	return c.Redirect(http.StatusSeeOther, "/admin/posts")
}

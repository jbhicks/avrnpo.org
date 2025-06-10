package actions

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"

	"avrnpo.org/models"
	"avrnpo.org/pkg/logging"
)

// AdminRequired middleware ensures only admins can access admin routes
func AdminRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if !ok || user == nil {
			if c.Value("test_mode") != nil {
				if !ok {
					logging.Debug("AdminRequired: current_user not found in context or wrong type", logging.Fields{})
				} else {
					logging.Debug("AdminRequired: current_user is nil", logging.Fields{})
				}
			}
			c.Flash().Add("danger", "Access denied. Administrator privileges required.")
			return c.Redirect(http.StatusFound, "/dashboard")
		}

		if user.Role != "admin" {
			// Log unauthorized admin access attempt
			logging.SecurityEvent(c, "unauthorized_admin_access", "failure", "insufficient_privileges", logging.Fields{
				"user_id": user.ID.String(),
				"email":   user.Email,
				"role":    user.Role,
			})

			if c.Value("test_mode") != nil {
				logging.Debug("AdminRequired: User is not admin", logging.Fields{
					"role": user.Role,
				})
			}
			c.Flash().Add("danger", "Access denied. Administrator privileges required.")
			return c.Redirect(http.StatusFound, "/dashboard")
		}

		// Log successful admin access
		logging.UserAction(c, user.ID.String(), "admin_access", "User accessed admin area", logging.Fields{
			"email": user.Email,
		})

		if c.Value("test_mode") != nil {
			logging.Debug("AdminRequired: Admin access granted", logging.Fields{
				"user_id": user.ID.String(),
				"email":   user.Email,
			})
		}
		return next(c)
	}
}

// AdminDashboard shows the admin dashboard
func AdminDashboard(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	// Get user statistics
	userCount, err := tx.Count("users")
	if err != nil {
		return errors.WithStack(err)
	}

	adminCount, err := tx.Where("role = ?", "admin").Count("users")
	if err != nil {
		return errors.WithStack(err)
	}

	// Get post statistics
	totalPosts, err := tx.Count("posts")
	if err != nil {
		return errors.WithStack(err)
	}

	publishedPosts, err := tx.Where("published = ?", true).Count("posts")
	if err != nil {
		return errors.WithStack(err)
	}

	draftPosts := totalPosts - publishedPosts

	// Get recent posts (this month)
	recentPosts, err := tx.Where("created_at >= date_trunc('month', now())").Count("posts")
	if err != nil {
		return errors.WithStack(err)
	}

	// Get recent posts for display
	posts := []models.Post{}
	if err := tx.Order("created_at desc").Limit(5).All(&posts); err != nil {
		return errors.WithStack(err)
	}

	// Load authors for each post
	for i := range posts {
		if err := tx.Load(&posts[i], "Author"); err != nil {
			return errors.WithStack(err)
		}
	}

	c.Set("userCount", userCount)
	c.Set("adminCount", adminCount)
	c.Set("regularUserCount", userCount-adminCount)
	c.Set("totalPosts", totalPosts)
	c.Set("publishedPosts", publishedPosts)
	c.Set("draftPosts", draftPosts)
	c.Set("recentPosts", recentPosts)
	c.Set("posts", posts)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.HTML("admin/index.plush.html"))
	}

	// Direct access - render full page with navigation
	return c.Render(http.StatusOK, r.HTML("admin/index.plush.html"))
}

// AdminUsers lists all users for admin management
func AdminUsers(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	users := []models.User{}
	q := tx.PaginateFromParams(c.Params())

	if err := q.Order("created_at desc").All(&users); err != nil {
		return errors.WithStack(err)
	}

	c.Set("users", users)
	c.Set("pagination", q.Paginator)

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("admin/users.plush.html"))
	}
	return c.Render(http.StatusOK, r.HTML("admin/users.plush.html"))
}

// AdminUserShow shows a specific user for admin editing
func AdminUserShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("user", user)

	// Provide role options - Buffalo SelectTag expects slice of maps with "value" and "label" keys
	roleOptions := []map[string]interface{}{
		{"value": "user", "label": "User"},
		{"value": "admin", "label": "Administrator"},
	}
	c.Set("roleOptions", roleOptions)

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("admin/user_edit.plush.html"))
	}
	return c.Render(http.StatusOK, r.HTML("admin/user_edit.plush.html"))
}

// AdminUserUpdate updates a user as admin
func AdminUserUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Create a copy for updates
	updatedUser := &models.User{}
	*updatedUser = *user

	// Bind form data
	if err := c.Bind(updatedUser); err != nil {
		return errors.WithStack(err)
	}

	// Preserve sensitive fields that shouldn't be changed via this form
	updatedUser.ID = user.ID
	updatedUser.PasswordHash = user.PasswordHash
	updatedUser.CreatedAt = user.CreatedAt
	updatedUser.Password = ""
	updatedUser.PasswordConfirmation = ""

	verrs, err := tx.ValidateAndUpdate(updatedUser)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", updatedUser)
		c.Set("errors", verrs)

		// Provide role options for re-render - Buffalo SelectTag expects slice of maps
		roleOptions := []map[string]interface{}{
			{"value": "user", "label": "User"},
			{"value": "admin", "label": "Administrator"},
		}
		c.Set("roleOptions", roleOptions)

		if c.Request().Header.Get("HX-Request") == "true" {
			return c.Render(http.StatusOK, rHTMX.HTML("admin/user_edit.plush.html"))
		}
		return c.Render(http.StatusOK, r.HTML("admin/user_edit.plush.html"))
	}

	// Log admin user update
	adminUser := c.Value("current_user").(*models.User)
	logging.UserAction(c, adminUser.ID.String(), "admin_update_user", fmt.Sprintf("Admin updated user %s", updatedUser.Email), logging.Fields{
		"admin_email":    adminUser.Email,
		"target_user_id": updatedUser.ID.String(),
		"target_email":   updatedUser.Email,
		"updated_role":   updatedUser.Role,
		"previous_role":  user.Role,
	})

	c.Flash().Add("success", "User updated successfully!")
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/admin/users")
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, "/admin/users")
}

// AdminUserDelete deletes a user (admin only)
func AdminUserDelete(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Prevent deletion of the current admin user
	currentUser := c.Value("current_user").(*models.User)
	if user.ID == currentUser.ID {
		c.Flash().Add("danger", "You cannot delete your own account.")
		return c.Redirect(http.StatusFound, "/admin/users")
	}

	if err := tx.Destroy(user); err != nil {
		return errors.WithStack(err)
	}

	// Log admin user deletion
	adminUser := c.Value("current_user").(*models.User)
	logging.UserAction(c, adminUser.ID.String(), "admin_delete_user", fmt.Sprintf("Admin deleted user %s", user.Email), logging.Fields{
		"admin_email":     adminUser.Email,
		"deleted_user_id": user.ID.String(),
		"deleted_email":   user.Email,
		"deleted_role":    user.Role,
	})

	c.Flash().Add("success", "User deleted successfully!")
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/admin/users")
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, "/admin/users")
}

// ============================================================================
// ADMIN BLOG POST HANDLERS
// ============================================================================

// AdminPostsIndex shows all posts for admin management
func AdminPostsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	posts := []models.Post{}
	if err := tx.Order("created_at desc").All(&posts); err != nil {
		return errors.WithStack(err)
	}

	// Load authors for each post
	for i := range posts {
		if err := tx.Load(&posts[i], "Author"); err != nil {
			return errors.WithStack(err)
		}
	}

	c.Set("posts", posts)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.HTML("admin/posts/index.plush.html"))
	}

	return c.Render(http.StatusOK, r.HTML("admin/posts/index.plush.html"))
}

// AdminPostsNew shows the new post creation form
func AdminPostsNew(c buffalo.Context) error {
	post := &models.Post{}
	c.Set("post", post)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.HTML("admin/posts/new.plush.html"))
	}

	return c.Render(http.StatusOK, r.HTML("admin/posts/new.plush.html"))
}

// AdminPostsCreate handles creation of new blog posts
func AdminPostsCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	post := &models.Post{}
	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}

	// Set the author to the current user
	currentUser := c.Value("current_user").(*models.User)
	post.AuthorID = currentUser.ID

	// Generate slug if not provided
	if post.Slug == "" {
		post.GenerateSlug()
	}

	// Handle published status based on form action
	action := c.Param("action")
	if action == "publish" {
		post.Published = true
	} else {
		post.Published = false
	}

	// Validate and save
	verrs, err := tx.ValidateAndCreate(post)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("post", post)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/posts/new.plush.html"))
	}

	// Log post creation
	logging.UserAction(c, currentUser.ID.String(), "post_created", fmt.Sprintf("Created blog post: %s", post.Title), logging.Fields{
		"post_id":   fmt.Sprintf("%d", post.ID),
		"post_slug": post.Slug,
		"published": post.Published,
	})

	c.Flash().Add("success", fmt.Sprintf("Post \"%s\" created successfully!", post.Title))

	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/admin/posts")
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, "/admin/posts")
}

// AdminPostsEdit shows the edit form for a blog post
func AdminPostsEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	post := &models.Post{}
	if err := tx.Find(post, c.Param("post_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Load the author
	if err := tx.Load(post, "Author"); err != nil {
		return errors.WithStack(err)
	}

	c.Set("post", post)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.HTML("admin/posts/edit.plush.html"))
	}

	return c.Render(http.StatusOK, r.HTML("admin/posts/edit.plush.html"))
}

// AdminPostsUpdate handles updating blog posts
func AdminPostsUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	post := &models.Post{}
	if err := tx.Find(post, c.Param("post_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}

	// Generate slug if changed
	if post.Slug == "" {
		post.GenerateSlug()
	}

	// Validate and save
	verrs, err := tx.ValidateAndUpdate(post)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("post", post)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/posts/edit.plush.html"))
	}

	// Log post update
	currentUser := c.Value("current_user").(*models.User)
	logging.UserAction(c, currentUser.ID.String(), "post_updated", fmt.Sprintf("Updated blog post: %s", post.Title), logging.Fields{
		"post_id":   fmt.Sprintf("%d", post.ID),
		"post_slug": post.Slug,
		"published": post.Published,
	})

	c.Flash().Add("success", fmt.Sprintf("Post \"%s\" updated successfully!", post.Title))

	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/admin/posts")
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, "/admin/posts")
}

// AdminPostsDestroy deletes a blog post
func AdminPostsDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	post := &models.Post{}
	if err := tx.Find(post, c.Param("post_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(post); err != nil {
		return errors.WithStack(err)
	}

	// Log post deletion
	currentUser := c.Value("current_user").(*models.User)
	logging.UserAction(c, currentUser.ID.String(), "post_deleted", fmt.Sprintf("Deleted blog post: %s", post.Title), logging.Fields{
		"post_id":   fmt.Sprintf("%d", post.ID),
		"post_slug": post.Slug,
	})

	c.Flash().Add("success", fmt.Sprintf("Post \"%s\" deleted successfully!", post.Title))

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.String(""))
	}
	return c.Redirect(http.StatusFound, "/admin/posts")
}

// AdminPostsShow displays a single post for admin
func AdminPostsShow(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	post := &models.Post{}
	if err := tx.Find(post, c.Param("post_id")); err != nil {
		c.Flash().Add("error", "Post not found")
		return c.Redirect(302, "/admin/posts")
	}

	// Load the user who created the post
	if err := tx.Load(post, "User"); err != nil {
		return err
	}

	c.Set("post", post)
	return c.Render(200, r.HTML("admin/posts/show.plush.html"))
}

// AdminPostsDelete deletes a post
func AdminPostsDelete(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	post := &models.Post{}
	if err := tx.Find(post, c.Param("post_id")); err != nil {
		c.Flash().Add("error", "Post not found")
		return c.Redirect(302, "/admin/posts")
	}

	if err := tx.Destroy(post); err != nil {
		return err
	}

	// Log post deletion
	currentUser := c.Value("current_user").(*models.User)
	logging.UserAction(c, currentUser.ID.String(), "post_deleted", fmt.Sprintf("Deleted blog post: %s", post.Title), logging.Fields{
		"post_id":   fmt.Sprintf("%d", post.ID),
		"post_slug": post.Slug,
	})

	c.Flash().Add("success", fmt.Sprintf("Post \"%s\" deleted successfully!", post.Title))

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.String(""))
	}
	return c.Redirect(http.StatusFound, "/admin/posts")
}

// AdminPostsBulk handles bulk operations on posts
func AdminPostsBulk(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	action := c.Param("bulk_action")
	postIDsStr := c.Param("post_ids")

	if postIDsStr == "" {
		c.Flash().Add("error", "No posts selected")
		return c.Redirect(302, "/admin/posts")
	}

	// Parse post IDs
	postIDStrings := strings.Split(postIDsStr, ",")
	postIDs := make([]int, 0, len(postIDStrings))

	for _, idStr := range postIDStrings {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			postIDs = append(postIDs, id)
		}
	}

	if len(postIDs) == 0 {
		c.Flash().Add("error", "No valid posts selected")
		return c.Redirect(302, "/admin/posts")
	}

	currentUser := c.Value("current_user").(*models.User)

	switch action {
	case "publish":
		now := time.Now()
		err := tx.RawQuery("UPDATE posts SET published_at = ? WHERE id IN (?)", now, postIDs).Exec()
		if err != nil {
			return err
		}
		logging.UserAction(c, currentUser.ID.String(), "posts_bulk_published", "Bulk published posts", logging.Fields{
			"post_count": len(postIDs),
		})
		c.Flash().Add("success", fmt.Sprintf("Published %d post(s)", len(postIDs)))

	case "unpublish":
		err := tx.RawQuery("UPDATE posts SET published_at = NULL WHERE id IN (?)", postIDs).Exec()
		if err != nil {
			return err
		}
		logging.UserAction(c, currentUser.ID.String(), "posts_bulk_unpublished", "Bulk unpublished posts", logging.Fields{
			"post_count": len(postIDs),
		})
		c.Flash().Add("success", fmt.Sprintf("Unpublished %d post(s)", len(postIDs)))

	case "delete":
		err := tx.RawQuery("DELETE FROM posts WHERE id IN (?)", postIDs).Exec()
		if err != nil {
			return err
		}
		logging.UserAction(c, currentUser.ID.String(), "posts_bulk_deleted", "Bulk deleted posts", logging.Fields{
			"post_count": len(postIDs),
		})
		c.Flash().Add("success", fmt.Sprintf("Deleted %d post(s)", len(postIDs)))

	default:
		c.Flash().Add("error", "Invalid bulk action")
	}

	return c.Redirect(302, "/admin/posts")
}

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
			return c.Redirect(http.StatusFound, "/auth/new")
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
		if err := tx.Load(&posts[i], "User"); err != nil {
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
		if err := tx.Load(&posts[i], "User"); err != nil {
			return errors.WithStack(err)
		}
	}

	c.Set("posts", posts)

	return c.Render(http.StatusOK, r.HTML("admin/posts/index.plush.html"))
}

// AdminPostsNew shows the new post creation form
func AdminPostsNew(c buffalo.Context) error {
	post := &models.Post{}
	c.Set("post", post)

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

	// Handle published status - use form data if provided, otherwise check action
	action := c.Param("action")
	if action == "publish" {
		post.Published = true
	} else if action == "draft" {
		post.Published = false
	}
	// If no specific action, keep the Published value from the form

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
	if err := tx.Load(post, "User"); err != nil {
		return errors.WithStack(err)
	}

	c.Set("post", post)

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

	return c.Redirect(http.StatusFound, "/admin/posts")
}

// AdminPostsBulk handles bulk operations on posts
func AdminPostsBulk(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	action := c.Param("action")
	postIDsParam := c.Request().Form["post_ids"]

	if len(postIDsParam) == 0 {
		c.Flash().Add("error", "Please select at least one post")
		return c.Render(200, r.HTML("admin/posts/index.plush.html"))
	}

	// Parse post IDs
	postIDInts := make([]int, 0, len(postIDsParam))
	for _, idStr := range postIDsParam {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			postIDInts = append(postIDInts, id)
		}
	}

	if len(postIDInts) == 0 {
		c.Flash().Add("error", "No valid posts selected")
		return c.Render(200, r.HTML("admin/posts/index.plush.html"))
	}

	currentUser := c.Value("current_user").(*models.User)

	switch action {
	case "publish":
		now := time.Now()
		err := tx.RawQuery("UPDATE posts SET published_at = ? WHERE id IN (?)", now, postIDInts).Exec()
		if err != nil {
			return err
		}
		logging.UserAction(c, currentUser.ID.String(), "posts_bulk_published", "Bulk published posts", logging.Fields{
			"post_count": len(postIDInts),
		})
		c.Flash().Add("success", fmt.Sprintf("Published %d post(s)", len(postIDInts)))

	case "unpublish":
		err := tx.RawQuery("UPDATE posts SET published_at = NULL WHERE id IN (?)", postIDInts).Exec()
		if err != nil {
			return err
		}
		logging.UserAction(c, currentUser.ID.String(), "posts_bulk_unpublished", "Bulk unpublished posts", logging.Fields{
			"post_count": len(postIDInts),
		})
		c.Flash().Add("success", fmt.Sprintf("Unpublished %d post(s)", len(postIDInts)))

	case "delete":
		// For delete action, require confirmation
		if c.Param("confirm_delete") != "true" {
			c.Flash().Add("warning", fmt.Sprintf("Are you sure you want to delete %d post(s)? This action cannot be undone.", len(postIDInts)))
			// Return the current page with confirmation message
			return c.Render(200, r.HTML("admin/posts/index.plush.html"))
		}

		err := tx.RawQuery("DELETE FROM posts WHERE id IN (?)", postIDInts).Exec()
		if err != nil {
			return err
		}
		logging.UserAction(c, currentUser.ID.String(), "posts_bulk_deleted", "Bulk deleted posts", logging.Fields{
			"post_count": len(postIDInts),
		})
		c.Flash().Add("success", fmt.Sprintf("Deleted %d post(s)", len(postIDInts)))

	default:
		c.Flash().Add("error", "Invalid bulk action")
	}

	return c.Render(200, r.HTML("admin/posts/index.plush.html"))
}

// AdminDonationsIndex shows the donations management page
func AdminDonationsIndex(c buffalo.Context) error {
	// Get current user
	currentUser, ok := c.Value("current_user").(*models.User)
	if !ok || currentUser == nil {
		return c.Redirect(http.StatusFound, "/")
	}

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Parse query parameters
	page := 1
	if p := c.Param("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	status := c.Param("status")
	search := c.Param("search")

	// Build query
	query := tx.Q()

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Where("donor_name ILIKE ? OR donor_email ILIKE ?",
			"%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	totalCount, err := query.Count(&models.Donation{})
	if err != nil {
		c.Logger().Errorf("Error counting donations: %v", err)
		c.Flash().Add("error", "Error loading donations")
		return c.Redirect(http.StatusFound, "/admin")
	}

	// Calculate pagination
	perPage := 20
	totalPages := (totalCount + perPage - 1) / perPage

	// Get donations for current page
	donations := []models.Donation{}
	err = query.Order("created_at desc").
		Paginate(page, perPage).
		All(&donations)
	if err != nil {
		c.Logger().Errorf("Error fetching donations: %v", err)
		c.Flash().Add("error", "Error loading donations")
		return c.Redirect(http.StatusFound, "/admin")
	}

	// Calculate summary statistics
	stats, err := getDonationStats(tx)
	if err != nil {
		c.Logger().Errorf("Error getting donation stats: %v", err)
		// Continue without stats
		stats = DonationStats{}
	}

	// Set template data
	c.Set("donations", donations)
	c.Set("stats", stats)
	c.Set("currentPage", page)
	c.Set("totalPages", totalPages)
	c.Set("totalCount", totalCount)
	c.Set("currentStatus", status)
	c.Set("currentSearch", search)
	c.Set("user", currentUser)

	// TODO: Admin donations templates not implemented yet
	// Redirect to main admin dashboard for now
	c.Flash().Add("info", "Donation management interface coming soon!")
	return c.Redirect(http.StatusSeeOther, "/admin")
}

// DonationStats holds donation statistics
type DonationStats struct {
	TotalDonations  int     `json:"total_donations"`
	CompletedCount  int     `json:"completed_count"`
	PendingCount    int     `json:"pending_count"`
	FailedCount     int     `json:"failed_count"`
	TotalAmount     float64 `json:"total_amount"`
	CompletedAmount float64 `json:"completed_amount"`
	AverageAmount   float64 `json:"average_amount"`
	MonthlyTotal    float64 `json:"monthly_total"`
	RecurringCount  int     `json:"recurring_count"`
}

// getDonationStats calculates donation statistics
func getDonationStats(tx *pop.Connection) (DonationStats, error) {
	stats := DonationStats{}

	// Total donations count
	totalCount, err := tx.Count(&models.Donation{})
	if err != nil {
		return stats, err
	}
	stats.TotalDonations = totalCount

	// Count by status
	completed, _ := tx.Where("status = ?", "completed").Count(&models.Donation{})
	pending, _ := tx.Where("status = ?", "pending").Count(&models.Donation{})
	failed, _ := tx.Where("status = ?", "failed").Count(&models.Donation{})

	stats.CompletedCount = completed
	stats.PendingCount = pending
	stats.FailedCount = failed

	// Amount calculations
	var totalAmountResult struct {
		TotalAmount     float64 `db:"total_amount"`
		CompletedAmount float64 `db:"completed_amount"`
		AverageAmount   float64 `db:"average_amount"`
	}

	err = tx.RawQuery(`
		SELECT 
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END), 0) as completed_amount,
			COALESCE(AVG(CASE WHEN status = 'completed' THEN amount ELSE NULL END), 0) as average_amount
		FROM donations
	`).First(&totalAmountResult)

	if err == nil {
		stats.TotalAmount = totalAmountResult.TotalAmount
		stats.CompletedAmount = totalAmountResult.CompletedAmount
		stats.AverageAmount = totalAmountResult.AverageAmount
	}

	// Monthly total (current month)
	var monthlyResult struct {
		MonthlyTotal float64 `db:"monthly_total"`
	}

	err = tx.RawQuery(`
		SELECT COALESCE(SUM(amount), 0) as monthly_total
		FROM donations 
		WHERE status = 'completed' 
		AND created_at >= date_trunc('month', now())
	`).First(&monthlyResult)

	if err == nil {
		stats.MonthlyTotal = monthlyResult.MonthlyTotal
	}

	// Recurring donations count
	recurringCount, _ := tx.Where("donation_type = ?", "recurring").Count(&models.Donation{})
	stats.RecurringCount = recurringCount

	return stats, nil
}

// AdminDonationShow shows a single donation details
func AdminDonationShow(c buffalo.Context) error {
	// Get current user
	currentUser, ok := c.Value("current_user").(*models.User)
	if !ok || currentUser == nil {
		return c.Redirect(http.StatusFound, "/")
	}

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Get donation ID from URL
	donationID := c.Param("donation_id")
	if donationID == "" {
		c.Flash().Add("error", "Donation ID is required")
		return c.Redirect(http.StatusFound, "/admin/donations")
	}

	// Find the donation
	donation := &models.Donation{}
	err := tx.Find(donation, donationID)
	if err != nil {
		c.Logger().Errorf("Error finding donation %s: %v", donationID, err)
		c.Flash().Add("error", "Donation not found")
		return c.Redirect(http.StatusFound, "/admin/donations")
	}

	// Set template data
	c.Set("donation", donation)
	c.Set("user", currentUser)

	// TODO: Admin donation show template not implemented yet
	// Redirect to main admin dashboard for now
	c.Flash().Add("info", "Donation details view coming soon!")
	return c.Redirect(http.StatusSeeOther, "/admin")
}

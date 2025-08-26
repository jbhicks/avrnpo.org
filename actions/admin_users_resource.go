package actions

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"

	"avrnpo.org/models"
	"avrnpo.org/pkg/logging"
)

// AdminUsersResource handles CRUD operations for user management in admin area
type AdminUsersResource struct {
	buffalo.BaseResource
}

// List displays all users for admin management (GET /admin/users)
func (aur AdminUsersResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	users := []models.User{}
	q := tx.PaginateFromParams(c.Params())

	if err := q.Order("created_at desc").All(&users); err != nil {
		return errors.WithStack(err)
	}

	c.Set("users", users)
	c.Set("pagination", q.Paginator)

	// Always return the complete page - Single Template Architecture
	// This ensures direct access (bookmarks, reloads) works correctly
	return c.Render(http.StatusOK, r.HTML("admin/users/index.plush.html"))
}

// Show displays a specific user for admin editing (GET /admin/users/{user_id})
func (aur AdminUsersResource) Show(c buffalo.Context) error {
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

	// Always return the complete page - Single Template Architecture
	return c.Render(http.StatusOK, r.HTML("admin/users/show.plush.html"))
}

// New displays the form for creating a new user (GET /admin/users/new)
func (aur AdminUsersResource) New(c buffalo.Context) error {
	user := &models.User{}
	c.Set("user", user)

	// Provide role options
	roleOptions := []map[string]interface{}{
		{"value": "user", "label": "User"},
		{"value": "admin", "label": "Administrator"},
	}
	c.Set("roleOptions", roleOptions)

	// Always return the complete page - Single Template Architecture
	return c.Render(http.StatusOK, r.HTML("admin/users/new.plush.html"))
}

// Create handles creation of new users by admin (POST /admin/users)
func (aur AdminUsersResource) Create(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := c.Bind(user); err != nil {
		return errors.WithStack(err)
	}

	// Validate and create the user
	verrs, err := user.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", user)
		c.Set("errors", verrs)

		// Provide role options for re-render
		roleOptions := []map[string]interface{}{
			{"value": "user", "label": "User"},
			{"value": "admin", "label": "Administrator"},
		}
		c.Set("roleOptions", roleOptions)

		// Always return complete page for validation errors
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/users/new.plush.html"))
	}

	// Log admin user creation
	adminUser := c.Value("current_user").(*models.User)
	logging.UserAction(c, adminUser.ID.String(), "admin_create_user", fmt.Sprintf("Admin created user %s", user.Email), logging.Fields{
		"admin_email":     adminUser.Email,
		"created_user_id": user.ID.String(),
		"created_email":   user.Email,
		"created_role":    user.Role,
	})

	c.Flash().Add("success", fmt.Sprintf("User \"%s\" created successfully!", user.Email))

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/users/%s", user.ID))
}

// Edit displays the form for editing a user (GET /admin/users/{user_id}/edit)
func (aur AdminUsersResource) Edit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("user", user)

	// Provide role options
	roleOptions := []map[string]interface{}{
		{"value": "user", "label": "User"},
		{"value": "admin", "label": "Administrator"},
	}
	c.Set("roleOptions", roleOptions)

	// Always return the complete page - Single Template Architecture
	return c.Render(http.StatusOK, r.HTML("admin/users/edit.plush.html"))
}

// Update handles updating users by admin (PUT /admin/users/{user_id})
func (aur AdminUsersResource) Update(c buffalo.Context) error {
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

		// Provide role options for re-render
		roleOptions := []map[string]interface{}{
			{"value": "user", "label": "User"},
			{"value": "admin", "label": "Administrator"},
		}
		c.Set("roleOptions", roleOptions)

		// Always return complete page for validation errors
		return c.Render(http.StatusUnprocessableEntity, r.HTML("admin/users/edit.plush.html"))
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

	c.Flash().Add("success", fmt.Sprintf("User \"%s\" updated successfully!", updatedUser.Email))

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/users/%s", updatedUser.ID))
}

// Destroy deletes a user (DELETE /admin/users/{user_id})
func (aur AdminUsersResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Prevent deletion of the current admin user
	currentUser := c.Value("current_user").(*models.User)
	if user.ID == currentUser.ID {
		c.Flash().Add("danger", "You cannot delete your own account.")
		return c.Redirect(http.StatusSeeOther, "/admin/users")
	}

	// Check for confirmation
	if c.Param("confirm_delete") != "true" {
		c.Flash().Add("warning", fmt.Sprintf("Are you sure you want to delete user \"%s\"? This action cannot be undone.", user.Email))
		return c.Redirect(http.StatusSeeOther, "/admin/users")
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

	c.Flash().Add("success", fmt.Sprintf("User \"%s\" deleted successfully!", user.Email))

	return c.Redirect(http.StatusSeeOther, "/admin/users")
}

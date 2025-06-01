package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"

	"my_go_saas_template/models"
)

// AdminRequired middleware ensures only admins can access admin routes
func AdminRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if !ok || user == nil {
			if c.Value("test_mode") != nil {
				if !ok {
					c.Logger().Debugf("AdminRequired: current_user not found in context or wrong type")
				} else {
					c.Logger().Debugf("AdminRequired: current_user is nil")
				}
			}
			c.Flash().Add("danger", "Access denied. Administrator privileges required.")
			return c.Redirect(http.StatusFound, "/dashboard")
		}

		if user.Role != "admin" {
			if c.Value("test_mode") != nil {
				c.Logger().Debugf("AdminRequired: User is not admin. Role=%s", user.Role)
			}
			c.Flash().Add("danger", "Access denied. Administrator privileges required.")
			return c.Redirect(http.StatusFound, "/dashboard")
		}

		if c.Value("test_mode") != nil {
			c.Logger().Debugf("AdminRequired: Admin access granted for user: ID=%s, Email=%s", user.ID, user.Email)
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

	c.Set("userCount", userCount)
	c.Set("adminCount", adminCount)
	c.Set("regularUserCount", userCount-adminCount)

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("admin/dashboard.plush.html"))
	}
	return c.Render(http.StatusOK, r.HTML("admin/dashboard.plush.html"))
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

	// Provide role options
	roleOptions := map[string]string{
		"user":  "User",
		"admin": "Administrator",
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

		// Provide role options for re-render
		roleOptions := map[string]string{
			"user":  "User",
			"admin": "Administrator",
		}
		c.Set("roleOptions", roleOptions)

		if c.Request().Header.Get("HX-Request") == "true" {
			return c.Render(http.StatusOK, rHTMX.HTML("admin/user_edit.plush.html"))
		}
		return c.Render(http.StatusOK, r.HTML("admin/user_edit.plush.html"))
	}

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

	c.Flash().Add("success", "User deleted successfully!")
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/admin/users")
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, "/admin/users")
}

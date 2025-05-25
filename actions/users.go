package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"my_go_saas_template/models"
)

// UsersNew renders the users form
func UsersNew(c buffalo.Context) error {
	u := models.User{}
	c.Set("user", u)
	return c.Render(http.StatusOK, r.HTML("users/new.plush.html"))
}

// UsersCreate registers a new user with the application.
func UsersCreate(c buffalo.Context) error {
	u := &models.User{}
	if err := c.Bind(u); err != nil {
		return errors.WithStack(err)
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := u.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", u)
		c.Set("errors", verrs)
		return c.Render(http.StatusOK, r.HTML("users/new.plush.html"))
	}

	c.Session().Set("current_user_id", u.ID)
	c.Flash().Add("success", "Welcome to my-go-saas-template!")

	return c.Redirect(http.StatusFound, "/")
}

// ProfileSettings shows the user profile settings page
func ProfileSettings(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	c.Set("user", user)
	return c.Render(http.StatusOK, r.HTML("users/profile.plush.html"))
}

// ProfileUpdate updates the user's profile information
func ProfileUpdate(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)

	// Create a copy to avoid modifying the session user
	updatedUser := &models.User{}
	*updatedUser = *user

	// Bind only the profile fields we want to update
	if err := c.Bind(updatedUser); err != nil {
		return errors.WithStack(err)
	}

	// Preserve the password hash and other sensitive fields
	updatedUser.ID = user.ID
	updatedUser.Email = user.Email // Don't allow email changes in profile
	updatedUser.PasswordHash = user.PasswordHash
	updatedUser.CreatedAt = user.CreatedAt
	updatedUser.Password = "" // Clear password fields for profile updates
	updatedUser.PasswordConfirmation = ""

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndUpdate(updatedUser)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", updatedUser)
		c.Set("errors", verrs)
		return c.Render(http.StatusOK, r.HTML("users/profile.plush.html"))
	}

	c.Flash().Add("success", "Profile updated successfully!")
	return c.Redirect(http.StatusFound, "/profile")
}

// AccountSettings shows the user account settings page
func AccountSettings(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	c.Set("user", user)
	return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
}

// AccountUpdate updates the user's account settings
func AccountUpdate(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)

	// Get the current password for verification
	currentPassword := c.Param("current_password")
	newPassword := c.Param("new_password")
	confirmPassword := c.Param("confirm_password")

	tx := c.Value("tx").(*pop.Connection)

	// If changing password, verify current password first
	if newPassword != "" {
		if currentPassword == "" {
			c.Flash().Add("danger", "Current password is required to change password")
			c.Set("user", user)
			return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
		}

		// Verify current password
		err := user.VerifyPassword(currentPassword)
		if err != nil {
			c.Flash().Add("danger", "Current password is incorrect")
			c.Set("user", user)
			return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
		}

		// Check password confirmation
		if newPassword != confirmPassword {
			c.Flash().Add("danger", "New passwords do not match")
			c.Set("user", user)
			return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
		}

		// Create a copy for updating
		updatedUser := &models.User{}
		*updatedUser = *user

		// Set the new password fields
		updatedUser.Password = newPassword
		updatedUser.PasswordConfirmation = confirmPassword

		// Hash the new password
		ph, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return errors.WithStack(err)
		}
		updatedUser.PasswordHash = string(ph)

		verrs, err := tx.ValidateAndUpdate(updatedUser)
		if err != nil {
			return errors.WithStack(err)
		}

		if verrs.HasAny() {
			c.Set("user", user)
			c.Set("errors", verrs)
			return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
		}

		c.Flash().Add("success", "Password updated successfully!")
	} else {
		c.Flash().Add("info", "No changes were made")
	}

	return c.Redirect(http.StatusFound, "/account")
}

// SetCurrentUser attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err != nil {
				// If user not found, clear the session and continue
				// This handles cases where user was deleted but session still exists
				c.Session().Delete("current_user_id")
			} else {
				c.Set("current_user", u)
			}
		}
		return next(c)
	}
}

// Authorize require a user be logged in before accessing a route
func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid == nil {
			c.Session().Set("redirectURL", c.Request().URL.String())

			err := c.Session().Save()
			if err != nil {
				return errors.WithStack(err)
			}

			c.Flash().Add("danger", "You must be authorized to see that page")
			return c.Redirect(http.StatusFound, "/auth/new")
		}
		return next(c)
	}
}

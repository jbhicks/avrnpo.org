package actions

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"avrnpo.org/models"
	"avrnpo.org/pkg/logging"
	"avrnpo.org/services"
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

	// Check if accept_terms checkbox was checked
	acceptTerms := c.Param("accept_terms")
	verrs := validate.NewErrors()

	if acceptTerms != "on" {
		verrs.Add("accept_terms", "You must accept the Terms of Service and Privacy Policy")
	}

	tx := c.Value("tx").(*pop.Connection)
	userVerrs, err := u.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Combine validation errors
	verrs.Append(userVerrs)

	if verrs.HasAny() {
		// Log failed registration attempt
		logging.SecurityEvent(c, "registration_failed", "failure", "validation_errors", logging.Fields{
			"email":             u.Email,
			"validation_errors": verrs.Error(),
		})

		c.Set("user", u)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("users/new.plush.html"))
	}

	// Log successful user registration
	logging.UserAction(c, u.Email, "register", "User registration successful", logging.Fields{
		"user_id":   u.ID.String(),
		"user_role": u.Role,
	})

	c.Session().Set("current_user_id", u.ID)
	c.Flash().Add("success", "Welcome to American Veterans Rebuilding!")

	return c.Redirect(http.StatusFound, "/")
}

// ProfileSettings shows the user profile settings page
func ProfileSettings(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	c.Set("user", user)

	// If user is admin, provide role options
	if user.Role == "admin" {
		roleOptions := map[string]string{
			"user":  "User",
			"admin": "Administrator",
		}
		c.Set("roleOptions", roleOptions)
	}

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

	// Only allow role changes for admins
	if user.Role != "admin" {
		updatedUser.Role = user.Role // Preserve original role for non-admins
	}

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

// SubscriptionsList shows all subscriptions for the current user
func SubscriptionsList(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	tx := c.Value("tx").(*pop.Connection)

	// Get all donations for this user with subscription IDs
	var donations []models.Donation
	err := tx.Where("user_id = ? AND subscription_id IS NOT NULL", user.ID).Order("created_at desc").All(&donations)
	if err != nil {
		c.Flash().Add("danger", "Unable to load your subscriptions")
		return c.Redirect(http.StatusFound, "/account")
	}

	// Group subscriptions by subscription_id (in case there are duplicates)
	subscriptionMap := make(map[string]*models.Donation)
	for i := range donations {
		if donations[i].SubscriptionID != nil {
			subscriptionMap[*donations[i].SubscriptionID] = &donations[i]
		}
	}

	// Convert to slice for template
	var subscriptions []*models.Donation
	for _, donation := range subscriptionMap {
		subscriptions = append(subscriptions, donation)
	}

	c.Set("subscriptions", subscriptions)
	return c.Render(http.StatusOK, r.HTML("users/subscriptions_list.plush.html"))
}

// SubscriptionDetails shows details for a specific subscription
func SubscriptionDetails(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	subscriptionID := c.Param("subscriptionId")
	tx := c.Value("tx").(*pop.Connection)

	// Find the donation record for this subscription
	donation := &models.Donation{}
	err := tx.Where("user_id = ? AND subscription_id = ?", user.ID, subscriptionID).First(donation)
	if err != nil {
		c.Flash().Add("danger", "Subscription not found")
		return c.Redirect(http.StatusFound, "/account/subscriptions")
	}

	// Get subscription details from Helcim
	helcimClient := services.NewHelcimClient()
	subscription, err := helcimClient.GetSubscription(subscriptionID)
	if err != nil {
		c.Flash().Add("warning", "Unable to load current subscription status from payment processor")
		// Still show the page with limited info
		subscription = nil
	}

	c.Set("donation", donation)
	c.Set("subscription", subscription)
	return c.Render(http.StatusOK, r.HTML("users/subscription_details.plush.html"))
}

// CancelSubscription cancels a user's subscription
func CancelSubscription(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	subscriptionID := c.Param("subscriptionId")
	tx := c.Value("tx").(*pop.Connection)

	// Verify this subscription belongs to the user
	donation := &models.Donation{}
	err := tx.Where("user_id = ? AND subscription_id = ?", user.ID, subscriptionID).First(donation)
	if err != nil {
		c.Flash().Add("danger", "Subscription not found")
		return c.Redirect(http.StatusFound, "/account/subscriptions")
	}

	// Check for confirmation
	if c.Param("confirm_cancel") != "true" {
		c.Flash().Add("warning", "Are you absolutely sure you want to cancel your recurring donation? This action cannot be undone.")
		return c.Redirect(http.StatusFound, fmt.Sprintf("/account/subscriptions/%s", subscriptionID))
	}

	// Cancel the subscription with Helcim
	helcimClient := services.NewHelcimClient()
	err = helcimClient.CancelSubscription(subscriptionID)
	if err != nil {
		// Log the error but don't expose internal details
		logging.Error("subscription_cancellation_failed", err, logging.Fields{
			"subscription_id": subscriptionID,
			"user_id":         user.ID.String(),
		})
		c.Flash().Add("danger", "Unable to cancel subscription. Please contact support.")
		return c.Redirect(http.StatusFound, fmt.Sprintf("/account/subscriptions/%s", subscriptionID))
	}

	// Update our local record
	donation.Status = "cancelled"
	err = tx.Update(donation)
	if err != nil {
		// Log the error but subscription is already cancelled with Helcim
		logging.Error("donation_status_update_failed", err, logging.Fields{
			"donation_id":     donation.ID.String(),
			"subscription_id": subscriptionID,
		})
	}

	// Log the successful cancellation
	logging.UserAction(c, user.Email, "cancel_subscription", "User cancelled recurring donation", logging.Fields{
		"subscription_id": subscriptionID,
		"donation_amount": donation.Amount,
	})

	c.Flash().Add("success", "Your subscription has been cancelled successfully")
	return c.Redirect(http.StatusFound, "/account/subscriptions")
}

// SetCurrentUser attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		sessionUID := c.Session().Get("current_user_id")
		if sessionUID != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, sessionUID)
			if err != nil {
				// If user not found, clear the session and continue
				c.Session().Delete("current_user_id")
				c.Set("current_user", nil)
			} else {
				c.Set("current_user", u)
			}
		} else {
			// Explicitly set current_user to nil when no session
			c.Set("current_user", nil)
		}
		return next(c)
	}
}

// Authorize require a user be logged in before accessing a route
func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		// Check if current_user was set by SetCurrentUser middleware
		user, ok := c.Value("current_user").(*models.User)

		if !ok || user == nil {
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

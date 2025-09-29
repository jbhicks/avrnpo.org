package grifts

import (
	"avrnpo.org/models"
	"fmt"
	"os"

	"github.com/gobuffalo/grift/grift"
	"golang.org/x/crypto/bcrypt"
)

var _ = grift.Namespace("admin", func() {

	grift.Desc("update_password", "Updates admin password from ADMIN_PASSWORD environment variable")
	grift.Add("update_password", func(c *grift.Context) error {
		db := models.DB

		// Get new password from environment variable
		newPassword := os.Getenv("ADMIN_PASSWORD")
		if newPassword == "" {
			return fmt.Errorf("ADMIN_PASSWORD environment variable is required")
		}

		// Find the admin user
		admin := &models.User{}
		if err := db.Where("email = ?", "admin@avrnpo.org").First(admin); err != nil {
			return fmt.Errorf("admin user not found: %w", err)
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Update the password
		admin.PasswordHash = string(hashedPassword)
		if err := db.Update(admin); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		fmt.Printf("✅ Successfully updated admin password for %s\n", admin.Email)
		fmt.Printf("   Password updated from environment variable\n")

		return nil
	})

})

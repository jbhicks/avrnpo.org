package grifts

import (
	"avrnpo.org/models"
	"fmt"
	"os"
	"strings"

	"github.com/gobuffalo/grift/grift"
	"golang.org/x/crypto/bcrypt"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		// Add DB seeding stuff here
		return nil
	})

	grift.Desc("promote_admin", "Promotes the first user (by email) to admin role")
	grift.Add("promote_admin", func(c *grift.Context) error {
		// Use the existing global DB connection
		db := models.DB

		// Find the first user
		user := &models.User{}
		if err := db.Order("created_at asc").First(user); err != nil {
			fmt.Println("No users found to promote")
			return err
		}

		// Update user role to admin
		user.Role = "admin"
		if err := db.Update(user); err != nil {
			return err
		}

		fmt.Printf("Successfully promoted user %s (%s) to admin\n", user.Email, user.FirstName+" "+user.LastName)
		return nil
	})

	grift.Desc("create_admin", "Creates an admin user from environment variables")
	grift.Add("create_admin", func(c *grift.Context) error {
		// Use the existing global DB connection
		db := models.DB

		// Get admin details from environment variables - all required
		email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
		password := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
		firstName := getEnvOrDefault("ADMIN_FIRST_NAME", "Admin")
		lastName := getEnvOrDefault("ADMIN_LAST_NAME", "User")

		// Validate required fields
		if email == "" || password == "" {
			return fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD environment variables are required")
		}

		// Check if admin user already exists
		existingUser := &models.User{}
		if err := db.Where("email = ?", email).First(existingUser); err == nil {
			// User exists, promote to admin if not already
			if existingUser.Role != "admin" {
				existingUser.Role = "admin"
				if err := db.Update(existingUser); err != nil {
					return fmt.Errorf("failed to promote existing user to admin: %w", err)
				}
				fmt.Printf("✅ Promoted existing user %s to admin\n", email)
			} else {
				fmt.Printf("ℹ️  Admin user %s already exists\n", email)
			}
			return nil
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Create new admin user
		admin := &models.User{
			FirstName:    firstName,
			LastName:     lastName,
			Email:        email,
			PasswordHash: string(hashedPassword),
			Role:         "admin",
		}

		// Validate and create the user
		if err := db.Create(admin); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		fmt.Printf("✅ Successfully created admin user:\n")
		fmt.Printf("   Email: %s\n", email)
		fmt.Printf("   Name: %s %s\n", firstName, lastName)
		fmt.Printf("   Role: admin\n")
		fmt.Printf("   Password: [Set from environment variable]\n")

		return nil
	})

	grift.Desc("create_admin_interactive", "Creates an admin user with interactive prompts")
	grift.Add("create_admin_interactive", func(c *grift.Context) error {
		// Use the existing global DB connection
		db := models.DB

		fmt.Println("🔧 Creating Admin User")
		fmt.Println("=====================")

		var email, password, firstName, lastName string

		fmt.Print("Email: ")
		fmt.Scanln(&email)

		fmt.Print("Password: ")
		fmt.Scanln(&password)

		fmt.Print("First Name: ")
		fmt.Scanln(&firstName)

		fmt.Print("Last Name: ")
		fmt.Scanln(&lastName)

		// Validate
		if email == "" || password == "" || firstName == "" || lastName == "" {
			return fmt.Errorf("all fields are required")
		}

		// Check if user already exists
		existingUser := &models.User{}
		if err := db.Where("email = ?", email).First(existingUser); err == nil {
			return fmt.Errorf("user with email %s already exists", email)
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Create admin user
		admin := &models.User{
			FirstName:    firstName,
			LastName:     lastName,
			Email:        email,
			PasswordHash: string(hashedPassword),
			Role:         "admin",
		}

		if err := db.Create(admin); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		fmt.Printf("✅ Admin user created successfully!\n")
		fmt.Printf("   Email: %s\n", email)
		fmt.Printf("   Name: %s %s\n", firstName, lastName)

		return nil
	})

	grift.Desc("check_posts", "Check all posts and published posts")
	grift.Add("check_posts", func(c *grift.Context) error {
		db := models.DB

		// Query all posts
		posts := []models.Post{}
		if err := db.All(&posts); err != nil {
			return fmt.Errorf("failed to query all posts: %w", err)
		}

		fmt.Printf("Found %d posts:\n", len(posts))
		for _, post := range posts {
			fmt.Printf("ID: %d, Title: %s, Slug: %s, Published: %t, Content Length: %d, Created: %s\n", post.ID, post.Title, post.Slug, post.Published, len(post.Content), post.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		// Query published posts
		publishedPosts := []models.Post{}
		if err := db.Where("published = ?", true).All(&publishedPosts); err != nil {
			return fmt.Errorf("failed to query published posts: %w", err)
		}

		fmt.Printf("\nFound %d published posts:\n", len(publishedPosts))
		for _, post := range publishedPosts {
			fmt.Printf("ID: %d, Title: %s, Published: %t\n", post.ID, post.Title, post.Published)
		}

		return nil
	})

})

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return defaultValue
}

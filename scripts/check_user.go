package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: "../pb_data",
	})

	if err := app.Bootstrap(); err != nil {
		log.Fatal(err)
	}

	email := os.Getenv("PB_ADMIN_EMAIL")
	if email == "" {
		email = "admin@avrnpo.org"
	}

	user, err := app.FindFirstRecordByFilter("users", "email = {:email}", map[string]any{
		"email": email,
	})

	if err != nil {
		fmt.Printf("❌ User NOT found: %v\n", err)
		return
	}

	fmt.Printf("✅ User EXISTS:\n")
	fmt.Printf("  Email: %s\n", user.GetString("email"))
	fmt.Printf("  Username: %s\n", user.GetString("username"))
	fmt.Printf("  Role: %s\n", user.GetString("role"))
	fmt.Printf("  ID: %s\n", user.Id)
}

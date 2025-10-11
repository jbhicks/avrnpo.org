//go:build ignore

package main

import (
	"log"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func main() {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: "./pb_data",
	})

	if err := app.Bootstrap(); err != nil {
		log.Fatal(err)
	}
	defer app.ResetBootstrapState()

	usersCollection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		log.Fatal("Failed to find users collection:", err)
	}

	user := core.NewRecord(usersCollection)
	user.Set("username", "testauthor")
	user.Set("email", "author@avrnpo.org")
	user.Set("role", "admin")
	user.SetPassword("TestPassword123")

	if err := app.Save(user); err != nil {
		log.Println("User might already exist:", err)
	} else {
		log.Println("Created user:", user.Id)
	}

	postsCollection, err := app.FindCollectionByNameOrId("posts")
	if err != nil {
		log.Fatal("Failed to find posts collection:", err)
	}

	post := core.NewRecord(postsCollection)
	post.Set("title", "Welcome to AVR NPO")
	post.Set("slug", "welcome-to-avr-npo")
	post.Set("content", "<p>Welcome to the Armed Services Vocational Aptitude Battery Research NPO blog! We're excited to share insights about our research and mission.</p><p>Stay tuned for more updates.</p>")
	post.Set("excerpt", "Welcome to the AVR NPO blog! Learn about our research and mission.")
	post.Set("published", true)
	post.Set("published_at", types.NowDateTime())
	post.Set("author", user.Id)
	post.Set("meta_title", "Welcome to AVR NPO - Latest Research and Updates")
	post.Set("meta_description", "Discover the mission and goals of the Armed Services Vocational Aptitude Battery Research NPO.")

	if err := app.Save(post); err != nil {
		log.Fatal("Failed to create post:", err)
	}

	log.Println("Created post:", post.Id)

	post2 := core.NewRecord(postsCollection)
	post2.Set("title", "Understanding the ASVAB")
	post2.Set("slug", "understanding-the-asvab")
	post2.Set("content", "<p>The Armed Services Vocational Aptitude Battery (ASVAB) is a crucial assessment for military recruitment.</p><p>Our research focuses on improving test validity and fairness.</p>")
	post2.Set("excerpt", "Learn about the ASVAB and our research approach to improving military recruitment assessments.")
	post2.Set("published", true)
	publishedAt := types.NowDateTime()
	publishedAt.Time().Add(-24 * time.Hour)
	post2.Set("published_at", publishedAt)
	post2.Set("author", user.Id)
	post2.Set("meta_title", "Understanding the ASVAB - Military Aptitude Testing")
	post2.Set("meta_description", "Research insights into the Armed Services Vocational Aptitude Battery and military recruitment.")

	if err := app.Save(post2); err != nil {
		log.Fatal("Failed to create post 2:", err)
	}

	log.Println("Created post 2:", post2.Id)
	log.Println("Test data seeded successfully!")
}

package main

import (
	"avrnpo.org/models"
	"fmt"
	"log"
)

func main() {
	// Query published posts
	posts := []models.Post{}
	err := models.DB.Where("published = ?", true).Order("created_at desc").All(&posts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d published posts:\n\n", len(posts))
	for _, post := range posts {
		publishedAt := "NULL"
		if post.PublishedAt != nil {
			publishedAt = post.PublishedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("ID: %d | Title: %s | Published: %t | PublishedAt: %s\n",
			post.ID, post.Title, post.Published, publishedAt)
	}
}

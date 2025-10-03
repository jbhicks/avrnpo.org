package grifts

import (
	"avrnpo.org/models"
	"fmt"
	"github.com/gobuffalo/grift/grift"
)

var _ = grift.Namespace("posts", func() {

	grift.Desc("list", "List all posts in the database")
	grift.Add("list", func(c *grift.Context) error {
		tx := models.DB

		posts := []models.Post{}
		err := tx.Order("created_at desc").All(&posts)
		if err != nil {
			return err
		}

		if len(posts) == 0 {
			fmt.Println("No posts found in database")
			return nil
		}

		fmt.Printf("Found %d posts:\n\n", len(posts))
		for _, post := range posts {
			publishedAt := "NULL"
			if post.PublishedAt != nil {
				publishedAt = post.PublishedAt.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("ID: %d | Title: %s | Published: %t | PublishedAt: %s | Created: %s\n",
				post.ID,
				post.Title,
				post.Published,
				publishedAt,
				post.CreatedAt.Format("2006-01-02 15:04:05"),
			)
		}
		return nil
	})
})

package grifts

import (
	"avrnpo.org/models"
	"fmt"
	"time"
	"github.com/gobuffalo/grift/grift"
)

var _ = grift.Namespace("posts", func() {

	grift.Desc("publish:all", "Publish all unpublished posts")
	grift.Add("publish:all", func(c *grift.Context) error {
		tx := models.DB

		now := time.Now()
		
		// Update all unpublished posts to be published
		sql := `
			UPDATE posts 
			SET published = true, published_at = $1
			WHERE published = false
		`

		err := tx.RawQuery(sql, now).Exec()
		if err != nil {
			return err
		}

		// Count how many posts were published
		var count int
		err = tx.RawQuery("SELECT COUNT(*) FROM posts WHERE published = true").First(&count)
		if err != nil {
			return err
		}

		fmt.Printf("Published all posts. Total published posts: %d\n", count)
		return nil
	})

	grift.Desc("publish", "Publish a specific post by ID (usage: buffalo task posts:publish -- --id=10)")
	grift.Add("publish", func(c *grift.Context) error {
		tx := models.DB

		postID := c.Value("id")
		if postID == nil {
			return fmt.Errorf("Please provide post ID with --id flag")
		}

		now := time.Now()
		
		sql := `
			UPDATE posts 
			SET published = true, published_at = $1
			WHERE id = $2
		`

		err := tx.RawQuery(sql, now, postID).Exec()
		if err != nil {
			return err
		}

		fmt.Printf("Published post ID %v\n", postID)
		return nil
	})
})

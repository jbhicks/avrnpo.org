package grifts

import (
	"avrnpo.org/models"
	"fmt"
	"github.com/gobuffalo/grift/grift"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("fix:published", "Fix published posts that are missing published_at timestamps")
	grift.Add("fix:published", func(c *grift.Context) error {
		tx := models.DB

		// Update all published posts that have NULL published_at
		// Use created_at as the published_at timestamp
		sql := `
			UPDATE posts 
			SET published_at = created_at 
			WHERE published = true AND published_at IS NULL
		`

		err := tx.RawQuery(sql).Exec()
		if err != nil {
			return err
		}

		// Count how many posts were fixed
		var count int
		err = tx.RawQuery("SELECT COUNT(*) FROM posts WHERE published = true AND published_at IS NOT NULL").First(&count)
		if err != nil {
			return err
		}

		fmt.Printf("Fixed published_at timestamps for published posts. Total published posts: %d\n", count)
		return nil
	})
})

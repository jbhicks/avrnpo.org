package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		postsCollection, err := app.FindCollectionByNameOrId("posts")
		if err != nil {
			return err
		}

		postsCollection.Fields.Add(
			&core.FileField{
				Name:      "content_images",
				MaxSelect: 20,
				MaxSize:   5242880,
				MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/avif"},
			},
		)

		return app.Save(postsCollection)
	}, func(app core.App) error {
		postsCollection, err := app.FindCollectionByNameOrId("posts")
		if err != nil {
			return err
		}

		contentImagesField := postsCollection.Fields.GetByName("content_images")
		if contentImagesField != nil {
			postsCollection.Fields.RemoveById(contentImagesField.GetId())
			return app.Save(postsCollection)
		}

		return nil
	})
}

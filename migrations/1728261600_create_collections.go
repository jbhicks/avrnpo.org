package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		usersCollection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		usersCollection.Fields.Add(
			&core.SelectField{
				Name:     "role",
				Required: true,
				Values:   []string{"user", "admin"},
			},
		)

		if err := app.Save(usersCollection); err != nil {
			return err
		}

		postsCollection := core.NewBaseCollection("posts")
		postsCollection.ListRule = types.Pointer(`published = true || @request.auth.role ?= 'admin'`)
		postsCollection.ViewRule = types.Pointer(`published = true || @request.auth.role ?= 'admin'`)
		postsCollection.CreateRule = types.Pointer(`@request.auth.role ?= 'admin'`)
		postsCollection.UpdateRule = types.Pointer(`@request.auth.role ?= 'admin'`)
		postsCollection.DeleteRule = types.Pointer(`@request.auth.role ?= 'admin'`)

		postsCollection.Fields.Add(
			&core.TextField{
				Name:     "title",
				Required: true,
				Max:      255,
			},
			&core.TextField{
				Name:     "slug",
				Required: true,
				Max:      255,
				Pattern:  `^[a-z0-9]+(?:-[a-z0-9]+)*$`,
			},
			&core.EditorField{
				Name:     "content",
				Required: true,
			},
			&core.TextField{
				Name: "excerpt",
				Max:  500,
			},
			&core.BoolField{
				Name: "published",
			},
			&core.DateField{
				Name: "published_at",
			},
			&core.RelationField{
				Name:          "author",
				CollectionId:  usersCollection.Id,
				CascadeDelete: false,
				MaxSelect:     1,
			},
			&core.FileField{
				Name:      "image",
				MaxSelect: 1,
				MaxSize:   5242880,
				MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
			},
			&core.TextField{
				Name: "image_alt",
				Max:  255,
			},
			&core.TextField{
				Name: "meta_title",
				Max:  255,
			},
			&core.TextField{
				Name: "meta_description",
				Max:  500,
			},
			&core.TextField{
				Name: "meta_keywords",
				Max:  500,
			},
			&core.TextField{
				Name: "og_title",
				Max:  255,
			},
			&core.TextField{
				Name: "og_description",
				Max:  500,
			},
			&core.FileField{
				Name:      "og_image",
				MaxSelect: 1,
				MaxSize:   5242880,
				MimeTypes: []string{"image/jpeg", "image/png", "image/webp"},
			},
		)

		postsCollection.AddIndex("idx_slug_unique", true, "slug", "")

		if err := app.Save(postsCollection); err != nil {
			return err
		}

		donationsCollection := core.NewBaseCollection("donations")
		donationsCollection.ListRule = types.Pointer(`@request.auth.role ?= 'admin' || donor_email = @request.auth.email`)
		donationsCollection.ViewRule = types.Pointer(`@request.auth.role ?= 'admin' || donor_email = @request.auth.email`)
		donationsCollection.CreateRule = types.Pointer(``)
		donationsCollection.UpdateRule = types.Pointer(`@request.auth.role ?= 'admin'`)
		donationsCollection.DeleteRule = types.Pointer(`@request.auth.role ?= 'admin'`)

		donationsCollection.Fields.Add(
			&core.TextField{
				Name: "helcim_transaction_id",
			},
			&core.TextField{
				Name: "checkout_token",
			},
			&core.TextField{
				Name: "secret_token",
			},
			&core.NumberField{
				Name:     "amount",
				Required: true,
			},
			&core.TextField{
				Name: "currency",
				Max:  3,
			},
			&core.TextField{
				Name:     "donor_name",
				Required: true,
			},
			&core.EmailField{
				Name:     "donor_email",
				Required: true,
			},
			&core.TextField{
				Name: "donor_phone",
			},
			&core.TextField{
				Name: "address_line1",
			},
			&core.TextField{
				Name: "address_line2",
			},
			&core.TextField{
				Name: "city",
			},
			&core.TextField{
				Name: "province",
			},
			&core.TextField{
				Name: "postal_code",
			},
			&core.TextField{
				Name: "country",
			},
			&core.SelectField{
				Name:     "donation_type",
				Required: true,
				Values:   []string{"one-time", "recurring", "monthly"},
			},
			&core.SelectField{
				Name:     "status",
				Required: true,
				Values:   []string{"pending", "completed", "failed", "cancelled"},
			},
			&core.TextField{
				Name: "comments",
				Max:  2000,
			},
			&core.TextField{
				Name: "subscription_id",
			},
			&core.TextField{
				Name: "customer_id",
			},
			&core.TextField{
				Name: "payment_plan_id",
			},
			&core.TextField{
				Name: "transaction_id",
			},
			&core.SelectField{
				Name:   "subscription_status",
				Values: []string{"active", "paused", "cancelled", "expired"},
			},
			&core.DateField{
				Name: "activation_date",
			},
			&core.DateField{
				Name: "next_billing_date",
			},
			&core.TextField{
				Name: "payment_method",
			},
			&core.JSONField{
				Name: "addon_ids",
			},
			&core.JSONField{
				Name: "addon_amounts",
			},
			&core.NumberField{
				Name: "payment_retry_count",
			},
			&core.DateField{
				Name: "last_payment_attempt",
			},
			&core.TextField{
				Name: "payment_failure_reason",
			},
			&core.DateField{
				Name: "last_status_sync",
			},
			&core.TextField{
				Name: "sync_error",
			},
			&core.TextField{
				Name: "error_message",
				Max:  2000,
			},
		)

		if err := app.Save(donationsCollection); err != nil {
			return err
		}

		contactCollection := core.NewBaseCollection("contact_submissions")
		contactCollection.ListRule = types.Pointer(`@request.auth.role ?= 'admin'`)
		contactCollection.ViewRule = types.Pointer(`@request.auth.role ?= 'admin'`)
		contactCollection.CreateRule = types.Pointer(``)
		contactCollection.UpdateRule = types.Pointer(`@request.auth.role ?= 'admin'`)
		contactCollection.DeleteRule = types.Pointer(`@request.auth.role ?= 'admin'`)

		contactCollection.Fields.Add(
			&core.TextField{
				Name:     "name",
				Required: true,
			},
			&core.EmailField{
				Name:     "email",
				Required: true,
			},
			&core.TextField{
				Name: "phone",
			},
			&core.TextField{
				Name:     "message",
				Required: true,
				Max:      5000,
			},
			&core.SelectField{
				Name:   "status",
				Values: []string{"new", "in_progress", "resolved", "spam"},
			},
		)

		if err := app.Save(contactCollection); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		collections := []string{"posts", "donations", "contact_submissions"}
		for _, name := range collections {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}
			if err := app.Delete(collection); err != nil {
				return err
			}
		}

		usersCollection, err := app.FindCollectionByNameOrId("users")
		if err == nil {
			roleField := usersCollection.Fields.GetByName("role")
			if roleField != nil {
				usersCollection.Fields.RemoveById(roleField.GetId())
				if err := app.Save(usersCollection); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

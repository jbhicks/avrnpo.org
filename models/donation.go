package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Donation represents a donation transaction
type Donation struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	UserID              *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	HelcimTransactionID *string    `json:"helcim_transaction_id,omitempty" db:"helcim_transaction_id"`
	CheckoutToken       string     `json:"checkout_token" db:"checkout_token"`
	SecretToken         string     `json:"secret_token" db:"secret_token"`
	Amount              float64    `json:"amount" db:"amount"`
	Currency            string     `json:"currency" db:"currency"`
	DonorName           string     `json:"donor_name" db:"donor_name"`
	DonorEmail          string     `json:"donor_email" db:"donor_email"`
	DonorPhone          *string    `json:"donor_phone,omitempty" db:"donor_phone"`
	AddressLine1        *string    `json:"address_line1,omitempty" db:"address_line1"`
	AddressLine2        *string    `json:"address_line2,omitempty" db:"address_line2"`
	City                *string    `json:"city,omitempty" db:"city"`
	State               *string    `json:"state,omitempty" db:"state"`
	Zip                 *string    `json:"zip,omitempty" db:"zip"`
	DonationType        string     `json:"donation_type" db:"donation_type"`
	Status              string     `json:"status" db:"status"`
	Comments            *string    `json:"comments,omitempty" db:"comments"`
	// Recurring payment fields
	SubscriptionID *string   `json:"subscription_id,omitempty" db:"subscription_id"`
	CustomerID     *string   `json:"customer_id,omitempty" db:"customer_id"`
	PaymentPlanID  *string   `json:"payment_plan_id,omitempty" db:"payment_plan_id"`
	TransactionID  *string   `json:"transaction_id,omitempty" db:"transaction_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (d Donation) String() string {
	jd, _ := json.Marshal(d)
	return string(jd)
}

// Donations is not required by pop and may be deleted
type Donations []Donation

// String is not required by pop and may be deleted
func (d Donations) String() string {
	jd, _ := json.Marshal(d)
	return string(jd)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (d *Donation) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: d.CheckoutToken, Name: "CheckoutToken"},
		&validators.StringIsPresent{Field: d.SecretToken, Name: "SecretToken"},
		&validators.StringIsPresent{Field: d.DonorName, Name: "DonorName"},
		&validators.StringIsPresent{Field: d.DonorEmail, Name: "DonorEmail"},
		&validators.EmailIsPresent{Field: d.DonorEmail, Name: "DonorEmail"},
		&validators.StringIsPresent{Field: d.Currency, Name: "Currency"},
		&validators.StringIsPresent{Field: d.DonationType, Name: "DonationType"},
		&validators.StringIsPresent{Field: d.Status, Name: "Status"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (d *Donation) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (d *Donation) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

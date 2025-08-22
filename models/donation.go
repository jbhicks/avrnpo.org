package models

import (
	"encoding/json"
	"strconv"
	"strings"
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
	SubscriptionID *string `json:"subscription_id,omitempty" db:"subscription_id"`
	CustomerID     *string `json:"customer_id,omitempty" db:"customer_id"`
	PaymentPlanID  *string `json:"payment_plan_id,omitempty" db:"payment_plan_id"`
	TransactionID  *string `json:"transaction_id,omitempty" db:"transaction_id"`

	// Enhanced subscription tracking
	SubscriptionStatus *string    `json:"subscription_status,omitempty" db:"subscription_status"`
	ActivationDate     *time.Time `json:"activation_date,omitempty" db:"activation_date"`
	NextBillingDate    *time.Time `json:"next_billing_date,omitempty" db:"next_billing_date"`
	PaymentMethod      *string    `json:"payment_method,omitempty" db:"payment_method"`

	// Add-on support (JSON-encoded arrays)
	AddonIDs     *string `json:"addon_ids,omitempty" db:"addon_ids"`
	AddonAmounts *string `json:"addon_amounts,omitempty" db:"addon_amounts"`

	// Failed payment tracking
	PaymentRetryCount    int        `json:"payment_retry_count" db:"payment_retry_count"`
	LastPaymentAttempt   *time.Time `json:"last_payment_attempt,omitempty" db:"last_payment_attempt"`
	PaymentFailureReason *string    `json:"payment_failure_reason,omitempty" db:"payment_failure_reason"`

	// Status sync tracking
	LastStatusSync *time.Time `json:"last_status_sync,omitempty" db:"last_status_sync"`
	SyncError      *string    `json:"sync_error,omitempty" db:"sync_error"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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
		// Note: CheckoutToken and SecretToken are populated after Helcim API call, so not required here
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

// AddAddon adds an add-on to the donation
func (d *Donation) AddAddon(addonID string, amount float64) {
	var ids, amounts []string

	if d.AddonIDs != nil && *d.AddonIDs != "" {
		ids = strings.Split(*d.AddonIDs, ",")
	}
	if d.AddonAmounts != nil && *d.AddonAmounts != "" {
		amounts = strings.Split(*d.AddonAmounts, ",")
	}

	ids = append(ids, addonID)
	amounts = append(amounts, strconv.FormatFloat(amount, 'f', 2, 64))

	idsStr := strings.Join(ids, ",")
	amountsStr := strings.Join(amounts, ",")
	d.AddonIDs = &idsStr
	d.AddonAmounts = &amountsStr
}

// GetAddons returns a map of addon IDs to amounts
func (d *Donation) GetAddons() map[string]float64 {
	addons := make(map[string]float64)

	if d.AddonIDs == nil || d.AddonAmounts == nil {
		return addons
	}

	ids := strings.Split(*d.AddonIDs, ",")
	amounts := strings.Split(*d.AddonAmounts, ",")

	for i, id := range ids {
		if i < len(amounts) && id != "" {
			if amount, err := strconv.ParseFloat(amounts[i], 64); err == nil {
				addons[id] = amount
			}
		}
	}

	return addons
}

// RemoveAddon removes an add-on from the donation
func (d *Donation) RemoveAddon(addonID string) {
	if d.AddonIDs == nil || d.AddonAmounts == nil {
		return
	}

	ids := strings.Split(*d.AddonIDs, ",")
	amounts := strings.Split(*d.AddonAmounts, ",")

	var newIds, newAmounts []string
	for i, id := range ids {
		if id != addonID && id != "" {
			newIds = append(newIds, id)
			if i < len(amounts) {
				newAmounts = append(newAmounts, amounts[i])
			}
		}
	}

	if len(newIds) == 0 {
		d.AddonIDs = nil
		d.AddonAmounts = nil
	} else {
		idsStr := strings.Join(newIds, ",")
		amountsStr := strings.Join(newAmounts, ",")
		d.AddonIDs = &idsStr
		d.AddonAmounts = &amountsStr
	}
}

// GetTotalAmount returns the base amount plus all add-on amounts
func (d *Donation) GetTotalAmount() float64 {
	total := d.Amount
	for _, amount := range d.GetAddons() {
		total += amount
	}
	return total
}

// IsRecurring returns true if this is a recurring donation
func (d *Donation) IsRecurring() bool {
	return d.SubscriptionID != nil && *d.SubscriptionID != ""
}

// CanRetryPayment returns true if payment can be retried
func (d *Donation) CanRetryPayment() bool {
	return d.PaymentRetryCount < 3 && d.IsRecurring()
}

// RecordPaymentFailure records a payment failure
func (d *Donation) RecordPaymentFailure(reason string) {
	d.PaymentRetryCount++
	now := time.Now()
	d.LastPaymentAttempt = &now
	d.PaymentFailureReason = &reason
}

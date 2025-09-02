package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDonation_AddAddon(t *testing.T) {
	donation := &Donation{}

	// Test adding first add-on
	donation.AddAddon("addon1", 5.50)
	assert.Equal(t, "addon1", *donation.AddonIDs)
	assert.Equal(t, "5.50", *donation.AddonAmounts)

	// Test adding second add-on
	donation.AddAddon("addon2", 10.25)
	assert.Equal(t, "addon1,addon2", *donation.AddonIDs)
	assert.Equal(t, "5.50,10.25", *donation.AddonAmounts)
}

func TestDonation_GetAddons(t *testing.T) {
	donation := &Donation{}

	// Test empty addons
	addons := donation.GetAddons()
	assert.Empty(t, addons)

	// Test with addons
	donation.AddAddon("addon1", 5.50)
	donation.AddAddon("addon2", 10.25)

	addons = donation.GetAddons()
	assert.Len(t, addons, 2)
	assert.Equal(t, 5.50, addons["addon1"])
	assert.Equal(t, 10.25, addons["addon2"])
}

func TestDonation_RemoveAddon(t *testing.T) {
	donation := &Donation{}

	// Setup test data
	donation.AddAddon("addon1", 5.50)
	donation.AddAddon("addon2", 10.25)
	donation.AddAddon("addon3", 2.75)

	// Remove middle addon
	donation.RemoveAddon("addon2")
	assert.Equal(t, "addon1,addon3", *donation.AddonIDs)
	assert.Equal(t, "5.50,2.75", *donation.AddonAmounts)

	// Remove all remaining addons
	donation.RemoveAddon("addon1")
	donation.RemoveAddon("addon3")
	assert.Nil(t, donation.AddonIDs)
	assert.Nil(t, donation.AddonAmounts)
}

func TestDonation_GetTotalAmount(t *testing.T) {
	donation := &Donation{Amount: 25.00}

	// Test base amount only
	assert.Equal(t, 25.00, donation.GetTotalAmount())

	// Test with add-ons
	donation.AddAddon("addon1", 5.50)
	donation.AddAddon("addon2", 2.50)
	assert.Equal(t, 33.00, donation.GetTotalAmount())
}

func TestDonation_IsRecurring(t *testing.T) {
	donation := &Donation{}

	// Test non-recurring
	assert.False(t, donation.IsRecurring())

	// Test recurring
	subscriptionID := "sub_123"
	donation.SubscriptionID = &subscriptionID
	assert.True(t, donation.IsRecurring())

	// Test empty subscription ID
	emptyID := ""
	donation.SubscriptionID = &emptyID
	assert.False(t, donation.IsRecurring())
}

func TestDonation_CanRetryPayment(t *testing.T) {
	donation := &Donation{}

	// Test non-recurring donation
	assert.False(t, donation.CanRetryPayment())

	// Test recurring donation with no retries
	subscriptionID := "sub_123"
	donation.SubscriptionID = &subscriptionID
	assert.True(t, donation.CanRetryPayment())

	// Test with max retries reached
	donation.PaymentRetryCount = 3
	assert.False(t, donation.CanRetryPayment())

	// Test with retries under limit
	donation.PaymentRetryCount = 2
	assert.True(t, donation.CanRetryPayment())
}

func TestDonation_RecordPaymentFailure(t *testing.T) {
	donation := &Donation{}

	// Record first failure
	donation.RecordPaymentFailure("Card declined")
	assert.Equal(t, 1, donation.PaymentRetryCount)
	assert.Equal(t, "Card declined", *donation.PaymentFailureReason)
	assert.NotNil(t, donation.LastPaymentAttempt)
	assert.WithinDuration(t, time.Now(), *donation.LastPaymentAttempt, time.Second)

	// Record second failure
	donation.RecordPaymentFailure("Insufficient funds")
	assert.Equal(t, 2, donation.PaymentRetryCount)
	assert.Equal(t, "Insufficient funds", *donation.PaymentFailureReason)
}

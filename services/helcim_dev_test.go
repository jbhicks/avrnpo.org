package services

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHelcimClient_Development_NoAPIKey(t *testing.T) {
	// Save original environment
	originalGoEnv := os.Getenv("GO_ENV")
	originalAPIKey := os.Getenv("HELCIM_PRIVATE_API_KEY")
	originalLiveTesting := os.Getenv("HELCIM_LIVE_TESTING")

	// Restore environment after test
	defer func() {
		if originalGoEnv != "" {
			os.Setenv("GO_ENV", originalGoEnv)
		} else {
			os.Unsetenv("GO_ENV")
		}
		if originalAPIKey != "" {
			os.Setenv("HELCIM_PRIVATE_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("HELCIM_PRIVATE_API_KEY")
		}
		if originalLiveTesting != "" {
			os.Setenv("HELCIM_LIVE_TESTING", originalLiveTesting)
		} else {
			os.Unsetenv("HELCIM_LIVE_TESTING")
		}
	}()

	// Ensure environment is development and API key is unset
	os.Setenv("GO_ENV", "development")
	os.Unsetenv("HELCIM_PRIVATE_API_KEY")
	os.Unsetenv("HELCIM_LIVE_TESTING")

	// Call NewHelcimClient; it should not panic and should return a non-nil client
	client := NewHelcimClient()
	if client == nil {
		t.Fatalf("expected non-nil HelcimClient in development without API key")
	}
}

func TestMockHelcimClient_AddOnMethods(t *testing.T) {
	client := &mockHelcimClient{}

	// Test CreateAddOn
	req := AddOnRequest{
		Name:        "Test Add-on",
		Description: "Test description",
		Amount:      5.00,
		Type:        "one_time",
		Quantity:    true,
	}

	addon, err := client.CreateAddOn(req)
	require.NoError(t, err)
	assert.Equal(t, req.Name, addon.Name)
	assert.Equal(t, req.Description, addon.Description)
	assert.Equal(t, req.Amount, addon.Amount)
	assert.Equal(t, req.Type, addon.Type)
	assert.Equal(t, req.Quantity, addon.Quantity)
	assert.Equal(t, "active", addon.Status)

	// Test GetAddOn
	addon, err = client.GetAddOn("123")
	require.NoError(t, err)
	assert.Equal(t, 123, addon.ID)
	assert.Equal(t, "Dev Add-on", addon.Name)

	// Test UpdateAddOn
	updates := map[string]interface{}{
		"name":   "Updated Add-on",
		"amount": 7.50,
	}
	addon, err = client.UpdateAddOn("123", updates)
	require.NoError(t, err)
	assert.Equal(t, "Updated Dev Add-on", addon.Name)
	assert.Equal(t, 7.50, addon.Amount)

	// Test DeleteAddOn
	err = client.DeleteAddOn("123")
	assert.NoError(t, err)

	// Test ListAddOns
	addons, err := client.ListAddOns()
	require.NoError(t, err)
	assert.Len(t, addons, 1)
	assert.Equal(t, "Dev Add-on", addons[0].Name)
}

func TestMockHelcimClient_SubscriptionAddOnMethods(t *testing.T) {
	client := &mockHelcimClient{}

	// Test LinkAddOnToSubscription
	req := SubscriptionAddOnRequest{
		AddOnID:  123,
		Quantity: 2,
	}

	subAddon, err := client.LinkAddOnToSubscription("sub_123", req)
	require.NoError(t, err)
	assert.Equal(t, 123456, subAddon.SubscriptionID)
	assert.Equal(t, req.AddOnID, subAddon.AddOnID)
	assert.Equal(t, req.Quantity, subAddon.Quantity)
	assert.Equal(t, "active", subAddon.Status)

	// Test UpdateSubscriptionAddOn
	updates := map[string]interface{}{
		"quantity": 3,
	}
	subAddon, err = client.UpdateSubscriptionAddOn("sub_123", "456", updates)
	require.NoError(t, err)
	assert.Equal(t, 123456, subAddon.SubscriptionID)
	assert.Equal(t, 456, subAddon.AddOnID)
	assert.Equal(t, 2, subAddon.Quantity) // Mock returns fixed values

	// Test DeleteSubscriptionAddOn
	err = client.DeleteSubscriptionAddOn("sub_123", "456")
	assert.NoError(t, err)
}

func TestMockHelcimClient_PaymentMethodMethods(t *testing.T) {
	client := &mockHelcimClient{}

	// Test SetCustomerCardDefault
	err := client.SetCustomerCardDefault("cust_123", "card_456")
	assert.NoError(t, err)

	// Test SetCustomerBankAccountDefault
	err = client.SetCustomerBankAccountDefault("cust_123", "bank_789")
	assert.NoError(t, err)
}

func TestMockHelcimClient_ProcedureMethods(t *testing.T) {
	client := &mockHelcimClient{}

	// Test ProcessSubscriptionPayment
	response, err := client.ProcessSubscriptionPayment("sub_123")
	require.NoError(t, err)
	assert.Contains(t, response.TransactionID, "dev_retry_txn_")
	assert.Equal(t, "APPROVED", response.Status)
	assert.Equal(t, 10.00, response.Amount)
	assert.NotEmpty(t, response.ProcessedAt)
}

func TestMockHelcimClient_StatusSync(t *testing.T) {
	client := &mockHelcimClient{}

	// Test SyncSubscriptionStatus
	sync, err := client.SyncSubscriptionStatus("sub_123")
	require.NoError(t, err)
	assert.Equal(t, "sub_123", sync.SubscriptionID)
	assert.Equal(t, "active", sync.Status)
	assert.Equal(t, "card", sync.PaymentMethod)
	assert.WithinDuration(t, time.Now(), sync.LastSyncAt, time.Second)
	assert.WithinDuration(t, time.Now().AddDate(0, 1, 0), sync.NextBillingDate, time.Hour)
}

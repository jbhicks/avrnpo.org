package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHelcimClient(t *testing.T) {
	// Test with valid API key
	originalToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
	defer func() {
		os.Setenv("HELCIM_PRIVATE_API_KEY", originalToken)
	}()

	os.Setenv("HELCIM_PRIVATE_API_KEY", "test-api-key-12345")
	client := NewHelcimClient()
	assert.NotNil(t, client)

	// Type assert to concrete type to access fields
	if concreteClient, ok := client.(*HelcimClient); ok {
		assert.Equal(t, "test-api-key-12345", concreteClient.APIToken)
		assert.Equal(t, "https://api.helcim.com/v2", concreteClient.BaseURL)
		assert.NotNil(t, concreteClient.Client)
	} else {
		t.Error("Expected HelcimClient type")
	}
}

func TestNewHelcimClient_MissingAPIKey(t *testing.T) {
	originalToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
	originalEnv := os.Getenv("GO_ENV")
	defer func() {
		os.Setenv("HELCIM_PRIVATE_API_KEY", originalToken)
		os.Setenv("GO_ENV", originalEnv)
	}()

	os.Unsetenv("HELCIM_PRIVATE_API_KEY")
	os.Setenv("GO_ENV", "production")

	assert.Panics(t, func() {
		NewHelcimClient()
	})
}

func TestNewHelcimClient_DevelopmentFallback(t *testing.T) {
	// Test with valid API key
	originalToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
	originalEnv := os.Getenv("GO_ENV")
	originalLiveTesting := os.Getenv("HELCIM_LIVE_TESTING")
	defer func() {
		os.Setenv("HELCIM_PRIVATE_API_KEY", originalToken)
		os.Setenv("GO_ENV", originalEnv)
		os.Setenv("HELCIM_LIVE_TESTING", originalLiveTesting)
	}()

	os.Unsetenv("HELCIM_PRIVATE_API_KEY")
	os.Setenv("GO_ENV", "development")
	os.Unsetenv("HELCIM_LIVE_TESTING") // Ensure live testing is disabled

	client := NewHelcimClient()
	assert.NotNil(t, client)
	// Should return mock client in development without API key
	assert.IsType(t, &mockHelcimClient{}, client)
}

func TestProcessPayment_IdempotencyKeyGeneration(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify idempotency key header is present and properly formatted
		idempotencyKey := r.Header.Get("Idempotency-Key")
		assert.NotEmpty(t, idempotencyKey, "Idempotency-Key header should be present")

		// Verify it's a valid UUID v4 format (36 characters)
		assert.Len(t, idempotencyKey, 36, "Idempotency key should be 36 characters (UUID v4)")

		// Verify it contains hyphens in UUID format
		assert.Contains(t, idempotencyKey, "-", "Idempotency key should contain hyphens (UUID format)")

		// Verify api-token header is present
		apiToken := r.Header.Get("api-token")
		assert.Equal(t, "test-api-key", apiToken)

		// Verify Content-Type
		contentType := r.Header.Get("Content-Type")
		assert.Equal(t, "application/json", contentType)

		// Parse request body to verify idempotency key is NOT in the JSON body
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		// Verify idempotencyKey is not in the request body
		_, hasIdempotencyKey := reqBody["idempotencyKey"]
		assert.False(t, hasIdempotencyKey, "idempotencyKey should not be in request body")

		// Return successful response
		response := PaymentAPIResponse{
			TransactionID: 123456,
			Status:        "APPROVED",
			Amount:        100.0,
			Currency:      "USD",
			CustomerCode:  "test-customer",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := &HelcimClient{
		APIToken: "test-api-key",
		BaseURL:  server.URL,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}

	// Test ProcessPayment
	req := PaymentAPIRequest{
		PaymentType:  "purchase",
		Amount:       100.0,
		Currency:     "USD",
		CustomerCode: "test-customer",
		CardData: CardData{
			CardToken: "test-card-token",
		},
		IPAddress:     "127.0.0.1",
		Description:   "Test payment",
		CustomerEmail: "test@example.com",
		CustomerName:  "Test User",
		BillingAddress: &BillingAddress{
			Name:       "Test User",
			Street1:    "123 Test St",
			City:       "Test City",
			Province:   "Test State",
			Country:    "USA",
			PostalCode: "12345",
		},
	}

	response, err := client.ProcessPayment(req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 123456, response.TransactionID)
	assert.Equal(t, "APPROVED", response.Status)
}

func TestCreatePaymentPlan_IdempotencyKeyGeneration(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify idempotency key header is present and properly formatted
		idempotencyKey := r.Header.Get("Idempotency-Key")
		assert.NotEmpty(t, idempotencyKey, "Idempotency-Key header should be present")
		assert.Len(t, idempotencyKey, 36, "Idempotency key should be 36 characters (UUID v4)")
		assert.Contains(t, idempotencyKey, "-", "Idempotency key should contain hyphens (UUID format)")

		// Verify api-token header is present
		apiToken := r.Header.Get("api-token")
		assert.Equal(t, "test-api-key", apiToken)

		// Parse request body to verify idempotency key is NOT in the JSON body
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		// Verify idempotencyKey is not in the request body
		_, hasIdempotencyKey := reqBody["idempotencyKey"]
		assert.False(t, hasIdempotencyKey, "idempotencyKey should not be in request body")

		// Return successful response
		response := map[string]interface{}{
			"status": "ok",
			"data": []PaymentPlan{
				{
					ID:              12345,
					Name:            "Test Plan",
					Description:     "Test payment plan",
					Type:            "subscription",
					Currency:        "USD",
					RecurringAmount: 50.0,
					BillingPeriod:   "monthly",
					Status:          "active",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := &HelcimClient{
		APIToken: "test-api-key",
		BaseURL:  server.URL,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}

	// Test CreatePaymentPlan
	plan, err := client.CreatePaymentPlan(50.0, "Test Plan")
	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, 12345, plan.ID)
	assert.Equal(t, "Test Plan", plan.Name)
	assert.Equal(t, 50.0, plan.RecurringAmount)
}

func TestCreateSubscription_IdempotencyKeyGeneration(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify idempotency key header is present and properly formatted
		idempotencyKey := r.Header.Get("Idempotency-Key")
		assert.NotEmpty(t, idempotencyKey, "Idempotency-Key header should be present")
		assert.Len(t, idempotencyKey, 36, "Idempotency key should be 36 characters (UUID v4)")
		assert.Contains(t, idempotencyKey, "-", "Idempotency key should contain hyphens (UUID format)")

		// Verify api-token header is present
		apiToken := r.Header.Get("api-token")
		assert.Equal(t, "test-api-key", apiToken)

		// Parse request body to verify idempotency key is NOT in the JSON body
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		// Verify idempotencyKey is not in the request body
		_, hasIdempotencyKey := reqBody["idempotencyKey"]
		assert.False(t, hasIdempotencyKey, "idempotencyKey should not be in request body")

		// Return successful response
		response := map[string]interface{}{
			"status": "ok",
			"data": []SubscriptionResponse{
				{
					ID:              67890,
					CustomerID:      "test-customer",
					PaymentPlanID:   12345,
					Amount:          50.0,
					Status:          "active",
					ActivationDate:  "2024-01-01",
					NextBillingDate: time.Now().AddDate(0, 1, 0),
					PaymentMethod:   "card",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := &HelcimClient{
		APIToken: "test-api-key",
		BaseURL:  server.URL,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}

	// Test CreateSubscription
	req := SubscriptionRequest{
		CustomerID:    "test-customer",
		PaymentPlanID: 12345,
		Amount:        50.0,
		PaymentMethod: "card",
	}

	response, err := client.CreateSubscription(req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 67890, response.ID)
	assert.Equal(t, "test-customer", response.CustomerID)
	assert.Equal(t, 12345, response.PaymentPlanID)
	assert.Equal(t, 50.0, response.Amount)
	assert.Equal(t, "active", response.Status)
}

func TestIdempotencyKey_UUIDFormat(t *testing.T) {
	// Test that generated UUIDs are valid UUID v4 format
	uuid1, err := uuid.NewV4()
	require.NoError(t, err)

	uuid2, err := uuid.NewV4()
	require.NoError(t, err)

	// Verify UUIDs are different (extremely unlikely to be the same)
	assert.NotEqual(t, uuid1.String(), uuid2.String())

	// Verify format: 8-4-4-4-12 characters
	uuidStr := uuid1.String()
	assert.Len(t, uuidStr, 36)

	// Count hyphens (should be 4)
	hyphenCount := strings.Count(uuidStr, "-")
	assert.Equal(t, 4, hyphenCount)

	// Verify structure: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	parts := strings.Split(uuidStr, "-")
	assert.Len(t, parts, 5)
	assert.Len(t, parts[0], 8)  // First part: 8 chars
	assert.Len(t, parts[1], 4)  // Second part: 4 chars
	assert.Len(t, parts[2], 4)  // Third part: 4 chars
	assert.Len(t, parts[3], 4)  // Fourth part: 4 chars
	assert.Len(t, parts[4], 12) // Fifth part: 12 chars

	// Verify all characters are valid hex
	for _, char := range uuidStr {
		if char != '-' {
			assert.Contains(t, "0123456789abcdefABCDEF", string(char), "UUID should only contain valid hex characters")
		}
	}
}

func TestIdempotencyKey_Uniqueness(t *testing.T) {
	// Generate multiple UUIDs and ensure they're all unique
	uuids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		uuid, err := uuid.NewV4()
		require.NoError(t, err)
		uuidStr := uuid.String()

		// Verify this UUID hasn't been seen before
		assert.False(t, uuids[uuidStr], "UUID should be unique")
		uuids[uuidStr] = true
	}

	// Verify we generated the expected number of unique UUIDs
	assert.Len(t, uuids, 1000)
}

func TestPaymentAPIRequest_NoIdempotencyKeyField(t *testing.T) {
	// Verify that PaymentAPIRequest struct does not have an IdempotencyKey field
	// This ensures we're not accidentally including it in the JSON body

	req := PaymentAPIRequest{
		PaymentType:  "purchase",
		Amount:       100.0,
		Currency:     "USD",
		CustomerCode: "test-customer",
		CardData: CardData{
			CardToken: "test-token",
		},
		IPAddress:     "127.0.0.1",
		Description:   "Test payment validation",
		CustomerEmail: "test@example.com",
		CustomerName:  "Test User",
		BillingAddress: &BillingAddress{
			Name:       "Test User",
			Street1:    "123 Test St",
			City:       "Test City",
			Province:   "Test State",
			Country:    "USA",
			PostalCode: "12345",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	// Parse back to map to check fields
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)

	// Verify idempotencyKey is not present in the JSON
	_, hasIdempotencyKey := parsed["idempotencyKey"]
	assert.False(t, hasIdempotencyKey, "PaymentAPIRequest should not have idempotencyKey field")

	// Verify expected fields are present
	assert.Equal(t, "purchase", parsed["paymentType"])
	assert.Equal(t, 100.0, parsed["amount"])
	assert.Equal(t, "USD", parsed["currency"])
	assert.Equal(t, "test-customer", parsed["customerCode"])

	// Check cardToken inside cardData object
	if cardData, ok := parsed["cardData"].(map[string]interface{}); ok {
		assert.Equal(t, "test-token", cardData["cardToken"])
	} else {
		t.Error("cardData should be a map")
	}
}

func TestMockHelcimClient_ProcessPayment(t *testing.T) {
	client := &mockHelcimClient{}

	req := PaymentAPIRequest{
		PaymentType:  "purchase",
		Amount:       100.0,
		Currency:     "USD",
		CustomerCode: "test-customer",
		CardData: CardData{
			CardToken: "test-token",
		},
		IPAddress:     "127.0.0.1",
		Description:   "Mock payment",
		CustomerEmail: "mock@example.com",
		CustomerName:  "Mock User",
		BillingAddress: &BillingAddress{
			Name:       "Mock User",
			Street1:    "123 Mock St",
			City:       "Mock City",
			Province:   "Mock State",
			Country:    "USA",
			PostalCode: "67890",
		},
	}

	response, err := client.ProcessPayment(req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "APPROVED", response.Status)
	assert.Equal(t, 100.0, response.Amount)
	assert.Equal(t, "USD", response.Currency)
	assert.Equal(t, "test-customer", response.CustomerCode)
	assert.Greater(t, response.TransactionID, 0)
}

func TestMockHelcimClient_CreatePaymentPlan(t *testing.T) {
	client := &mockHelcimClient{}

	plan, err := client.CreatePaymentPlan(50.0, "Test Plan")
	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "Test Plan", plan.Name)
	assert.Equal(t, 50.0, plan.RecurringAmount)
	assert.Equal(t, "subscription", plan.Type)
	assert.Equal(t, "active", plan.Status)
}

func TestMockHelcimClient_CreateSubscription(t *testing.T) {
	client := &mockHelcimClient{}

	req := SubscriptionRequest{
		CustomerID:    "test-customer",
		PaymentPlanID: 12345,
		Amount:        50.0,
		PaymentMethod: "card",
	}

	response, err := client.CreateSubscription(req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test-customer", response.CustomerID)
	assert.Equal(t, 12345, response.PaymentPlanID)
	assert.Equal(t, 50.0, response.Amount)
	assert.Equal(t, "active", response.Status)
	assert.Equal(t, "card", response.PaymentMethod)
}

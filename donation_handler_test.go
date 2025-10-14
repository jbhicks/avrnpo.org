package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"avrnpo.org/services"
)

type mockHelcimServer struct {
	server             *httptest.Server
	lastPaymentRequest *services.PaymentAPIRequest
	lastPlanRequest    map[string]interface{}
	lastSubRequest     map[string]interface{}
}

func newMockHelcimServer() *mockHelcimServer {
	mock := &mockHelcimServer{}

	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/payment/purchase"):
			var req services.PaymentAPIRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mock.lastPaymentRequest = &req

			resp := services.PaymentAPIResponse{
				TransactionID: 123456,
				Status:        "APPROVED",
				Amount:        req.Amount,
				Currency:      req.Currency,
				CustomerCode:  req.CustomerCode,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case strings.Contains(r.URL.Path, "/payment-plans"):
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mock.lastPlanRequest = req

			resp := map[string]interface{}{
				"status": "ok",
				"data": []services.PaymentPlan{
					{
						ID:              12345,
						Name:            "Test Plan",
						Type:            "subscription",
						Currency:        "USD",
						RecurringAmount: 50.0,
						BillingPeriod:   "monthly",
						Status:          "active",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case strings.Contains(r.URL.Path, "/subscriptions"):
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mock.lastSubRequest = req

			resp := map[string]interface{}{
				"status": "ok",
				"data": []services.SubscriptionResponse{
					{
						ID:              67890,
						CustomerID:      "test-customer",
						PaymentPlanID:   12345,
						Amount:          50.0,
						Status:          "active",
						ActivationDate:  time.Now().Format("2006-01-02"),
						NextBillingDate: time.Now().AddDate(0, 1, 0),
						PaymentMethod:   "card",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))

	return mock
}

func (m *mockHelcimServer) Close() {
	m.server.Close()
}

func TestDonationHandler_OneTimeDonation_FieldMapping(t *testing.T) {
	mockHelcim := newMockHelcimServer()
	defer mockHelcim.Close()

	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	helcimClient = &services.HelcimClient{
		APIToken: "test-api-key",
		BaseURL:  mockHelcim.server.URL,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}

	emailService = services.NewEmailService()

	donationData := map[string]interface{}{
		"donation_type": "one-time",
		"amount":        100.00,
		"donor_name":    "John Doe",
		"donor_email":   "john@example.com",
		"address_line1": "123 Main St",
		"city":          "Springfield",
		"province":      "IL",
		"country":       "USA",
		"postal_code":   "62701",
		"status":        "pending",
	}

	collection, err := testApp.FindCollectionByNameOrId("donations")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	for k, v := range donationData {
		record.Set(k, v)
	}
	require.NoError(t, testApp.Save(record))

	reqBody := map[string]interface{}{
		"donationId":    record.Id,
		"customerCode":  "test-customer-code",
		"amount":        100.00,
		"transactionId": "test-transaction-id",
	}

	jsonData, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/donations/process", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleProcessPayment(requestEvent)
	require.NoError(t, err)

	assert.Nil(t, mockHelcim.lastPaymentRequest, "Payment should NOT be sent to Helcim (already processed by HelcimPay.js)")

	updatedRecord, err := testApp.FindRecordById("donations", record.Id)
	require.NoError(t, err)

	assert.Equal(t, "test-transaction-id", updatedRecord.GetString("helcim_transaction_id"))
	assert.Equal(t, "test-customer-code", updatedRecord.GetString("customer_id"))
	assert.Equal(t, "completed", updatedRecord.GetString("status"))
}

func TestDonationHandler_MonthlyDonation_FieldMapping(t *testing.T) {
	mockHelcim := newMockHelcimServer()
	defer mockHelcim.Close()

	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	helcimClient = &services.HelcimClient{
		APIToken: "test-api-key",
		BaseURL:  mockHelcim.server.URL,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}

	emailService = services.NewEmailService()

	donationData := map[string]interface{}{
		"donation_type": "monthly",
		"amount":        50.00,
		"donor_name":    "Jane Smith",
		"donor_email":   "jane@example.com",
		"address_line1": "456 Oak Ave",
		"city":          "Portland",
		"province":      "OR",
		"country":       "USA",
		"postal_code":   "97201",
		"status":        "pending",
	}

	collection, err := testApp.FindCollectionByNameOrId("donations")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	for k, v := range donationData {
		record.Set(k, v)
	}
	require.NoError(t, testApp.Save(record))

	reqBody := map[string]interface{}{
		"donationId":    record.Id,
		"customerCode":  "test-customer-monthly",
		"cardToken":     "test-card-token-monthly",
		"amount":        50.00,
		"transactionId": "test-transaction-monthly",
	}

	jsonData, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/donations/process", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleProcessPayment(requestEvent)
	require.NoError(t, err)

	assert.NotNil(t, mockHelcim.lastPlanRequest, "Payment plan request should have been sent")
	planArray, ok := mockHelcim.lastPlanRequest["paymentPlans"].([]interface{})
	require.True(t, ok, "paymentPlans should be an array")
	require.Greater(t, len(planArray), 0, "paymentPlans array should not be empty")

	plan := planArray[0].(map[string]interface{})
	assert.Equal(t, fmt.Sprintf("Monthly $%.2f Donation", 50.00), plan["name"])
	assert.Equal(t, "subscription", plan["type"])
	assert.Equal(t, "USD", plan["currency"])
	assert.Equal(t, 50.00, plan["recurringAmount"])
	assert.Equal(t, "monthly", plan["billingPeriod"])
	assert.Equal(t, float64(1), plan["billingPeriodIncrements"])
	assert.Equal(t, "Sign-up", plan["dateBilling"])
	assert.Equal(t, "forever", plan["termType"])
	assert.Equal(t, "card", plan["paymentMethod"])
	assert.Equal(t, "active", plan["status"])

	assert.NotNil(t, mockHelcim.lastSubRequest, "Subscription request should have been sent")
	subArray, ok := mockHelcim.lastSubRequest["subscriptions"].([]interface{})
	require.True(t, ok, "subscriptions should be an array")
	require.Greater(t, len(subArray), 0, "subscriptions array should not be empty")

	subscription := subArray[0].(map[string]interface{})
	assert.Equal(t, "test-customer-monthly", subscription["customerCode"])
	assert.Equal(t, float64(12345), subscription["paymentPlanId"])
	assert.Equal(t, 50.00, subscription["recurringAmount"])
	assert.Equal(t, "card", subscription["paymentMethod"])
	assert.NotEmpty(t, subscription["dateActivated"])
}

func TestDonationHandler_BillingAddressFields_AllPresent(t *testing.T) {
	t.Skip("Skipping: One-time donations are now processed client-side via HelcimPay.js, not server-side")
}

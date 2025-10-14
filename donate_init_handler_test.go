package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"avrnpo.org/services"
)

type mockHelcimClientInit struct {
	lastInitRequest *services.InitializeRequest
}

func (m *mockHelcimClientInit) Initialize(req services.InitializeRequest) (*services.InitializeResponse, error) {
	m.lastInitRequest = &req
	return &services.InitializeResponse{
		CheckoutToken: "test-checkout-token",
		SecretToken:   "test-secret-token",
	}, nil
}

func (m *mockHelcimClientInit) ProcessPayment(req services.PaymentAPIRequest) (*services.PaymentAPIResponse, error) {
	return nil, nil
}

func (m *mockHelcimClientInit) CreatePaymentPlan(amount float64, planName string) (*services.PaymentPlan, error) {
	return nil, nil
}

func (m *mockHelcimClientInit) CreateSubscription(req services.SubscriptionRequest) (*services.SubscriptionResponse, error) {
	return nil, nil
}

func (m *mockHelcimClientInit) GetSubscription(subscriptionID string) (*services.SubscriptionResponse, error) {
	return nil, nil
}

func (m *mockHelcimClientInit) CancelSubscription(subscriptionID string) error {
	return nil
}

func (m *mockHelcimClientInit) UpdateSubscription(subscriptionID string, updates map[string]interface{}) (*services.SubscriptionResponse, error) {
	return nil, nil
}

func (m *mockHelcimClientInit) ListSubscriptionsByCustomer(customerID string) ([]services.SubscriptionResponse, error) {
	return nil, nil
}

func TestDonateInitHandler_ValidOneTimeDonation(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	mockHelcim := &mockHelcimClientInit{}
	helcimClient = mockHelcim

	formData := url.Values{
		"donation_type": {"one-time"},
		"amount":        {"100.00"},
		"name":          {"Jane Doe"},
		"email":         {"jane@example.com"},
		"address_line1": {"456 Elm St"},
		"city":          {"Boston"},
		"province":      {"MA"},
		"postal_code":   {"02101"},
		"country":       {"USA"},
	}

	req := httptest.NewRequest(http.MethodPost, "/donate", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleDonatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 200, res.Code)

	var response map[string]interface{}
	err = json.Unmarshal(res.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["donationId"])
	assert.Equal(t, "test-checkout-token", response["checkoutToken"])

	collection, err := testApp.FindCollectionByNameOrId("donations")
	require.NoError(t, err)

	records, err := testApp.FindAllRecords(collection)
	require.NoError(t, err)
	require.Len(t, records, 1)

	record := records[0]
	assert.Equal(t, 100.00, record.GetFloat("amount"))
	assert.Equal(t, "USD", record.GetString("currency"))
	assert.Equal(t, "one-time", record.GetString("donation_type"))
	assert.Equal(t, "pending", record.GetString("status"))
	assert.Equal(t, "Jane Doe", record.GetString("donor_name"))
	assert.Equal(t, "jane@example.com", record.GetString("donor_email"))
	assert.Equal(t, "456 Elm St", record.GetString("address_line1"))
	assert.Equal(t, "Boston", record.GetString("city"))
	assert.Equal(t, "MA", record.GetString("province"))
	assert.Equal(t, "02101", record.GetString("postal_code"))
	assert.Equal(t, "USA", record.GetString("country"))
	assert.Equal(t, "test-checkout-token", record.GetString("checkout_token"))
	assert.Equal(t, "test-secret-token", record.GetString("secret_token"))

	require.NotNil(t, mockHelcim.lastInitRequest)
	assert.Equal(t, "purchase", mockHelcim.lastInitRequest.PaymentType)
	assert.Equal(t, 100.00, mockHelcim.lastInitRequest.Amount)
	assert.Equal(t, "USD", mockHelcim.lastInitRequest.Currency)
	require.NotNil(t, mockHelcim.lastInitRequest.CustomerRequest)
	assert.Equal(t, "Jane Doe", mockHelcim.lastInitRequest.CustomerRequest.ContactName)
	assert.Equal(t, "jane@example.com", mockHelcim.lastInitRequest.CustomerRequest.Email)
	assert.Equal(t, "Jane Doe", mockHelcim.lastInitRequest.CustomerRequest.BillingAddress.Name)
	assert.Equal(t, "456 Elm St", mockHelcim.lastInitRequest.CustomerRequest.BillingAddress.Street1)
	assert.Equal(t, "Boston", mockHelcim.lastInitRequest.CustomerRequest.BillingAddress.City)
	assert.Equal(t, "MA", mockHelcim.lastInitRequest.CustomerRequest.BillingAddress.Province)
	assert.Equal(t, "USA", mockHelcim.lastInitRequest.CustomerRequest.BillingAddress.Country)
	assert.Equal(t, "02101", mockHelcim.lastInitRequest.CustomerRequest.BillingAddress.PostalCode)
}

func TestDonateInitHandler_InvalidDonationType(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"donation_type": {"invalid"},
		"amount":        {"100.00"},
		"name":          {"Jane Doe"},
		"email":         {"jane@example.com"},
	}

	req := httptest.NewRequest(http.MethodPost, "/donate", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleDonatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

func TestDonateInitHandler_InvalidAmount(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"donation_type": {"one-time"},
		"amount":        {"not-a-number"},
		"name":          {"Jane Doe"},
		"email":         {"jane@example.com"},
	}

	req := httptest.NewRequest(http.MethodPost, "/donate", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleDonatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

func TestDonateInitHandler_MissingName(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"donation_type": {"one-time"},
		"amount":        {"100.00"},
		"name":          {""},
		"email":         {"jane@example.com"},
	}

	req := httptest.NewRequest(http.MethodPost, "/donate", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleDonatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

func TestDonateInitHandler_InvalidEmail(t *testing.T) {
	testApp, err := tests.NewTestApp()
	require.NoError(t, err)
	defer testApp.Cleanup()

	formData := url.Values{
		"donation_type": {"one-time"},
		"amount":        {"100.00"},
		"name":          {"Jane Doe"},
		"email":         {"invalid-email"},
	}

	req := httptest.NewRequest(http.MethodPost, "/donate", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	requestEvent := &core.RequestEvent{
		App: testApp,
		Event: router.Event{
			Request:  req,
			Response: res,
		},
	}

	err = handleDonatePost(requestEvent)
	require.NoError(t, err)

	assert.Equal(t, 400, res.Code)
}

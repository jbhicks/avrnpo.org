package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HelcimAPI defines the methods used by the application
type HelcimAPI interface {
	ProcessPayment(req PaymentAPIRequest) (*PaymentAPIResponse, error)
	CreatePaymentPlan(amount float64, planName string) (*PaymentPlan, error)
	CreateSubscription(req SubscriptionRequest) (*SubscriptionResponse, error)
	GetSubscription(subscriptionID string) (*SubscriptionResponse, error)
	CancelSubscription(subscriptionID string) error
	UpdateSubscription(subscriptionID string, updates map[string]interface{}) (*SubscriptionResponse, error)
	ListSubscriptionsByCustomer(customerID string) ([]SubscriptionResponse, error)
}

// HelcimClient is the real implementation of HelcimAPI
type HelcimClient struct {
	APIToken string
	BaseURL  string
	Client   *http.Client
}

// Payment API structures
type PaymentAPIRequest struct {
	Amount       float64  `json:"amount"`
	Currency     string   `json:"currency"`
	CustomerCode string   `json:"customerCode"`
	CardData     CardData `json:"cardData"`
}

type CardData struct {
	CardToken string `json:"cardToken"`
}

type PaymentAPIResponse struct {
	TransactionID string  `json:"transactionId"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	CustomerCode  string  `json:"customerCode"`
}

// Recurring API structures
type PaymentPlan struct {
	ID                      int     `json:"id"`
	Name                    string  `json:"name"`
	Description             string  `json:"description"`
	Type                    string  `json:"type"`
	Currency                string  `json:"currency"`
	RecurringAmount         float64 `json:"recurringAmount"`
	BillingPeriod           string  `json:"billingPeriod"`
	BillingPeriodIncrements int     `json:"billingPeriodIncrements"`
	DateBilling             string  `json:"dateBilling"`
	TermType                string  `json:"termType"`
	PaymentMethod           string  `json:"paymentMethod"`
	Status                  string  `json:"status"`
}

type CustomerRequest struct {
	ContactName    string         `json:"contactName"`
	Email          string         `json:"email"`
	BillingAddress BillingAddress `json:"billingAddress"`
}

type BillingAddress struct {
	Name       string `json:"name"`
	Street1    string `json:"street1"`
	City       string `json:"city"`
	Province   string `json:"province"`
	Country    string `json:"country"`
	PostalCode string `json:"postalCode"`
}

type SubscriptionRequest struct {
	CustomerID    string  `json:"customerId"`
	PaymentPlanID int     `json:"paymentPlanId"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"paymentMethod"` // "card" for credit card
}

type SubscriptionResponse struct {
	ID              int       `json:"id"`
	CustomerID      string    `json:"customerId"`
	PaymentPlanID   int       `json:"paymentPlanId"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	ActivationDate  string    `json:"activationDate"`
	NextBillingDate time.Time `json:"nextBillingDate"`
	PaymentMethod   string    `json:"paymentMethod"`
}

// NewHelcimClient creates a new Helcim API client
func NewHelcimClient() HelcimAPI {
	apiKey := os.Getenv("HELCIM_PRIVATE_API_KEY")
	goEnv := os.Getenv("GO_ENV")

	// In development, prefer a safe fallback rather than panicking
	if apiKey == "" {
		if goEnv == "development" {
			fmt.Printf("[Helcim] Development mode: HELCIM_PRIVATE_API_KEY not set — returning mockHelcimClient\n")
			return &mockHelcimClient{}
		}

		// Non-development environments must provide an API key
		panic("[Helcim] FATAL: HELCIM_PRIVATE_API_KEY is not set or empty! Test runner did not load .env or environment variable.")
	}

	// Print partial key for safety
	if len(apiKey) > 10 {
		fmt.Printf("[Helcim] API key loaded: %s\n", apiKey[:6]+"..."+apiKey[len(apiKey)-4:])
	} else {
		fmt.Printf("[Helcim] API key loaded (short): %s\n", apiKey)
	}

	return &HelcimClient{
		APIToken: apiKey,
		BaseURL:  "https://api.helcim.com/v2",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessPayment processes a one-time payment using the Payment API
func (h *HelcimClient) ProcessPayment(req PaymentAPIRequest) (*PaymentAPIResponse, error) {
	url := fmt.Sprintf("%s/payment/purchase", h.BaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-token", h.APIToken)

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result PaymentAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CreatePaymentPlan creates a new payment plan for recurring donations
func (h *HelcimClient) CreatePaymentPlan(amount float64, planName string) (*PaymentPlan, error) {
	url := fmt.Sprintf("%s/payment-plans", h.BaseURL)

	// Create payment plan request according to Helcim API docs
	request := map[string]interface{}{
		"paymentPlans": []map[string]interface{}{
			{
				"name":                    planName,
				"description":             fmt.Sprintf("Monthly donation plan for $%.2f", amount),
				"type":                    "subscription", // Bill on sign-up
				"currency":                "USD",
				"recurringAmount":         amount,
				"billingPeriod":           "monthly",
				"billingPeriodIncrements": 1,
				"dateBilling":             "Sign-up",
				"termType":                "forever", // Indefinite billing
				"paymentMethod":           "card",
				"taxType":                 "no_tax",
				"status":                  "active",
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-token", h.APIToken)

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - Helcim returns array of created payment plans
	var responseData []PaymentPlan
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(responseData) == 0 {
		return nil, fmt.Errorf("no payment plan returned in response")
	}

	return &responseData[0], nil
}

// CreateSubscription creates a new subscription using the Recurring API
func (h *HelcimClient) CreateSubscription(req SubscriptionRequest) (*SubscriptionResponse, error) {
	url := fmt.Sprintf("%s/subscriptions", h.BaseURL)

	// Create subscription request according to Helcim API docs
	request := map[string]interface{}{
		"subscriptions": []map[string]interface{}{
			{
				"customerId":     req.CustomerID,
				"paymentPlanId":  req.PaymentPlanID,
				"amount":         req.Amount,
				"paymentMethod":  req.PaymentMethod,
				"activationDate": time.Now().Format("2006-01-02"), // Activate immediately
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-token", h.APIToken)

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - Helcim returns array of created subscriptions
	var responseData []SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(responseData) == 0 {
		return nil, fmt.Errorf("no subscription returned in response")
	}

	return &responseData[0], nil
}

// GetSubscription retrieves a subscription by ID
func (h *HelcimClient) GetSubscription(subscriptionID string) (*SubscriptionResponse, error) {
	url := fmt.Sprintf("%s/subscriptions/%s", h.BaseURL, subscriptionID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-token", h.APIToken)

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CancelSubscription cancels a subscription by ID
func (h *HelcimClient) CancelSubscription(subscriptionID string) error {
	url := fmt.Sprintf("%s/subscriptions/%s", h.BaseURL, subscriptionID)

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-token", h.APIToken)

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateSubscription updates a subscription's details
func (h *HelcimClient) UpdateSubscription(subscriptionID string, updates map[string]interface{}) (*SubscriptionResponse, error) {
	url := fmt.Sprintf("%s/subscriptions", h.BaseURL)

	// Add the subscription ID to the updates
	subscriptionUpdate := make(map[string]interface{})
	for k, v := range updates {
		subscriptionUpdate[k] = v
	}
	subscriptionUpdate["id"] = subscriptionID

	// Wrap in subscriptions array as required by Helcim API
	request := map[string]interface{}{
		"subscriptions": []map[string]interface{}{subscriptionUpdate},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-token", h.APIToken)

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - Helcim returns array of updated subscriptions
	var responseData []SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(responseData) == 0 {
		return nil, fmt.Errorf("no subscription returned in response")
	}

	return &responseData[0], nil
}

// ListSubscriptionsByCustomer retrieves all subscriptions for a customer
func (h *HelcimClient) ListSubscriptionsByCustomer(customerID string) ([]SubscriptionResponse, error) {
	url := fmt.Sprintf("%s/subscriptions?customerId=%s", h.BaseURL, customerID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-token", h.APIToken)

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result []SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// mockHelcimClient implements HelcimAPI for development/testing
type mockHelcimClient struct{}

func (m *mockHelcimClient) ProcessPayment(req PaymentAPIRequest) (*PaymentAPIResponse, error) {
	// Simulate an approved transaction
	return &PaymentAPIResponse{
		TransactionID: fmt.Sprintf("dev_txn_%d", time.Now().UnixNano()),
		Status:        "APPROVED",
		Amount:        req.Amount,
		Currency:      req.Currency,
		CustomerCode:  req.CustomerCode,
	}, nil
}

func (m *mockHelcimClient) CreatePaymentPlan(amount float64, planName string) (*PaymentPlan, error) {
	return &PaymentPlan{
		ID:              int(time.Now().Unix() % 1000000),
		Name:            planName,
		Description:     fmt.Sprintf("Dev plan for $%.2f", amount),
		Type:            "subscription",
		Currency:        "USD",
		RecurringAmount: amount,
		BillingPeriod:   "monthly",
		Status:          "active",
	}, nil
}

func (m *mockHelcimClient) CreateSubscription(req SubscriptionRequest) (*SubscriptionResponse, error) {
	return &SubscriptionResponse{
		ID:             int(time.Now().Unix() % 1000000),
		CustomerID:     req.CustomerID,
		PaymentPlanID:  req.PaymentPlanID,
		Amount:         req.Amount,
		Status:         "active",
		ActivationDate: time.Now().Format("2006-01-02"),
		NextBillingDate: time.Now().AddDate(0, 1, 0),
		PaymentMethod:  req.PaymentMethod,
	}, nil
}

func (m *mockHelcimClient) GetSubscription(subscriptionID string) (*SubscriptionResponse, error) {
	// Return a simulated active subscription
	now := time.Now()
	return &SubscriptionResponse{
		ID:             123456,
		CustomerID:     "dev_customer",
		PaymentPlanID:  1111,
		Amount:         10.00,
		Status:         "active",
		ActivationDate: now.Format("2006-01-02"),
		NextBillingDate: now.AddDate(0, 1, 0),
		PaymentMethod:  "card",
	}, nil
}

func (m *mockHelcimClient) CancelSubscription(subscriptionID string) error {
	// Simulate success
	return nil
}

func (m *mockHelcimClient) UpdateSubscription(subscriptionID string, updates map[string]interface{}) (*SubscriptionResponse, error) {
	// Simulate returning an updated subscription
	sub := &SubscriptionResponse{
		ID:            123456,
		CustomerID:    "dev_customer",
		PaymentPlanID: 1111,
		Amount:        10.00,
		Status:        "active",
		ActivationDate: time.Now().Format("2006-01-02"),
		NextBillingDate: time.Now().AddDate(0, 1, 0),
		PaymentMethod: "card",
	}
	return sub, nil
}

func (m *mockHelcimClient) ListSubscriptionsByCustomer(customerID string) ([]SubscriptionResponse, error) {
	now := time.Now()
	return []SubscriptionResponse{
		{
			ID:             123456,
			CustomerID:     customerID,
			PaymentPlanID:  1111,
			Amount:         10.00,
			Status:         "active",
			ActivationDate: now.Format("2006-01-02"),
			NextBillingDate: now.AddDate(0, 1, 0),
			PaymentMethod:  "card",
		},
	}, nil
}

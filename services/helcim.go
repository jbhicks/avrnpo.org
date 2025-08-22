package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

// PaymentPlanCache provides simple in-memory caching for payment plans
type PaymentPlanCache struct {
	plans map[string]*CachedPaymentPlan
	mutex sync.RWMutex
}

type CachedPaymentPlan struct {
	Plan      *PaymentPlan
	ExpiresAt time.Time
}

var (
	planCache     *PaymentPlanCache
	planCacheOnce sync.Once
)

// getCurrency returns the configured currency with a fallback to USD
func getCurrency() string {
	currency := os.Getenv("HELCIM_CURRENCY")
	if currency == "" {
		return "USD" // Default fallback
	}
	return currency
}

// GetPaymentPlanCache returns the singleton payment plan cache
func GetPaymentPlanCache() *PaymentPlanCache {
	planCacheOnce.Do(func() {
		planCache = &PaymentPlanCache{
			plans: make(map[string]*CachedPaymentPlan),
		}
	})
	return planCache
}

// Get retrieves a cached payment plan
func (cache *PaymentPlanCache) Get(key string) (*PaymentPlan, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	cached, exists := cache.plans[key]
	if !exists || time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	return cached.Plan, true
}

// Set stores a payment plan in cache with 1 hour expiration
func (cache *PaymentPlanCache) Set(key string, plan *PaymentPlan) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.plans[key] = &CachedPaymentPlan{
		Plan:      plan,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
}

// Clear removes expired entries from cache
func (cache *PaymentPlanCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	now := time.Now()
	for key, cached := range cache.plans {
		if now.After(cached.ExpiresAt) {
			delete(cache.plans, key)
		}
	}
}

// HelcimAPI defines the methods used by the application
type HelcimAPI interface {
	ProcessPayment(req PaymentAPIRequest) (*PaymentAPIResponse, error)
	CreatePaymentPlan(amount float64, planName string) (*PaymentPlan, error)
	CreateSubscription(req SubscriptionRequest) (*SubscriptionResponse, error)
	GetSubscription(subscriptionID string) (*SubscriptionResponse, error)
	CancelSubscription(subscriptionID string) error
	UpdateSubscription(subscriptionID string, updates map[string]interface{}) (*SubscriptionResponse, error)
	ListSubscriptionsByCustomer(customerID string) ([]SubscriptionResponse, error)

	// Add-on management
	CreateAddOn(req AddOnRequest) (*AddOnResponse, error)
	GetAddOn(addOnID string) (*AddOnResponse, error)
	UpdateAddOn(addOnID string, updates map[string]interface{}) (*AddOnResponse, error)
	DeleteAddOn(addOnID string) error
	ListAddOns() ([]AddOnResponse, error)

	// Subscription add-on management
	LinkAddOnToSubscription(subscriptionID string, req SubscriptionAddOnRequest) (*SubscriptionAddOnResponse, error)
	UpdateSubscriptionAddOn(subscriptionID string, addOnID string, updates map[string]interface{}) (*SubscriptionAddOnResponse, error)
	DeleteSubscriptionAddOn(subscriptionID string, addOnID string) error

	// Payment method management
	SetCustomerCardDefault(customerID string, cardID string) error
	SetCustomerBankAccountDefault(customerID string, bankAccountID string) error

	// Procedures
	ProcessSubscriptionPayment(subscriptionID string) (*PaymentProcedureResponse, error)

	// Status sync
	SyncSubscriptionStatus(subscriptionID string) (*SubscriptionStatusSync, error)
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

// Add-on structures
type AddOnRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"` // "recurring" or "one_time"
	Quantity    bool    `json:"quantity"`
}

type AddOnResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Quantity    bool    `json:"quantity"`
	Status      string  `json:"status"`
}

type SubscriptionAddOnRequest struct {
	AddOnID  int `json:"addOnId"`
	Quantity int `json:"quantity"`
}

type SubscriptionAddOnResponse struct {
	ID             int     `json:"id"`
	SubscriptionID int     `json:"subscriptionId"`
	AddOnID        int     `json:"addOnId"`
	Quantity       int     `json:"quantity"`
	Amount         float64 `json:"amount"`
	Status         string  `json:"status"`
}

// Payment procedure structures
type PaymentProcedureResponse struct {
	TransactionID string  `json:"transactionId"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	ProcessedAt   string  `json:"processedAt"`
}

// Status sync structures
type SubscriptionStatusSync struct {
	SubscriptionID  string    `json:"subscriptionId"`
	Status          string    `json:"status"`
	NextBillingDate time.Time `json:"nextBillingDate"`
	LastSyncAt      time.Time `json:"lastSyncAt"`
	PaymentMethod   string    `json:"paymentMethod"`
	ActivationDate  string    `json:"activationDate"`
}

func NewHelcimClient() HelcimAPI {
	apiKey := os.Getenv("HELCIM_PRIVATE_API_KEY")
	goEnv := os.Getenv("GO_ENV")
	useLivePayments := os.Getenv("HELCIM_LIVE_TESTING") == "true"

	// In development, prefer a safe fallback rather than panicking (unless live testing enabled)
	if apiKey == "" {
		if goEnv == "development" && !useLivePayments {
			fmt.Printf("[Helcim] Development mode: HELCIM_PRIVATE_API_KEY not set — returning mockHelcimClient\n")
			return &mockHelcimClient{}
		}

		// Non-development environments must provide an API key
		panic("[Helcim] FATAL: HELCIM_PRIVATE_API_KEY is not set or empty! Test runner did not load .env or environment variable.")
	}

	// Check if we're in development with live testing enabled
	if goEnv == "development" && useLivePayments {
		fmt.Printf("[Helcim] Development mode with LIVE TESTING enabled - using real Helcim API\n")
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
	url := fmt.Sprintf("%s/payment-plans", h.BaseURL) // BaseURL already includes v2

	// Generate proper idempotency key as required by Helcim (25-36 chars, UUID recommended)
	idempotencyUUID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate idempotency key: %w", err)
	}
	idempotencyKey := idempotencyUUID.String()

	// Create payment plan request according to Helcim API docs
	request := map[string]interface{}{
		"paymentPlans": []map[string]interface{}{
			{
				"name":                    planName,
				"description":             fmt.Sprintf("Monthly donation plan for $%.2f", amount),
				"type":                    "subscription", // Bill on sign-up
				"currency":                getCurrency(),
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

	// Log the request for debugging
	requestJSON, _ := json.Marshal(request)
	fmt.Printf("[Helcim] Payment plan request to %s with idempotency key %s: %s\n", url, idempotencyKey, string(requestJSON))

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
	httpReq.Header.Set("Idempotency-Key", idempotencyKey) // Required by Helcim API

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - Helcim may return array or single object
	// Read raw response first to handle both formats
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Debug: Log the raw response to understand what Helcim is returning
	fmt.Printf("[CreatePaymentPlan] HTTP Status: %d\n", resp.StatusCode)
	fmt.Printf("[CreatePaymentPlan] Raw Helcim response: %s\n", string(body))

	// Force output to stderr which Buffalo should capture
	fmt.Fprintf(os.Stderr, "[HELCIM DEBUG] Payment plan response status %d: %s\n", resp.StatusCode, string(body))

	// Parse the Helcim response wrapper first
	type HelcimResponse struct {
		Status string        `json:"status"`
		Data   []PaymentPlan `json:"data"`
	}

	var helcimResponse HelcimResponse
	if err := json.Unmarshal(body, &helcimResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Helcim response wrapper: %w", err)
	}

	if helcimResponse.Status != "ok" {
		return nil, fmt.Errorf("Helcim API returned status: %s", helcimResponse.Status)
	}

	if len(helcimResponse.Data) == 0 {
		return nil, fmt.Errorf("no payment plan returned in Helcim response")
	}

	paymentPlan := &helcimResponse.Data[0]
	fmt.Printf("[CreatePaymentPlan] Successfully parsed payment plan ID: %d\n", paymentPlan.ID)
	return paymentPlan, nil
}

// CreateSubscription creates a new subscription using the Recurring API
func (h *HelcimClient) CreateSubscription(req SubscriptionRequest) (*SubscriptionResponse, error) {
	url := fmt.Sprintf("%s/subscriptions", h.BaseURL) // BaseURL already includes v2

	// Generate proper idempotency key as required by Helcim (25-36 chars, UUID recommended)
	idempotencyUUID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate idempotency key: %w", err)
	}
	idempotencyKey := idempotencyUUID.String() // This gives us a proper 36-character UUID

	// Create subscription request according to Helcim API docs
	request := map[string]interface{}{
		"subscriptions": []map[string]interface{}{
			{
				"customerCode":    req.CustomerID, // Use customerCode not customerId
				"paymentPlanId":   req.PaymentPlanID,
				"recurringAmount": req.Amount, // Use recurringAmount not amount
				"paymentMethod":   req.PaymentMethod,
				"dateActivated":   time.Now().Format("2006-01-02"), // Use dateActivated not activationDate
			},
		},
	}

	// Debug: Log the subscription request
	requestJSON, _ := json.Marshal(request)
	fmt.Printf("[CreateSubscription] Request to %s with idempotency key %s: %s\n", url, idempotencyKey, string(requestJSON))

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
	httpReq.Header.Set("Idempotency-Key", idempotencyKey) // Required by Helcim API

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the Helcim response wrapper (same format as payment plans)
	type HelcimSubscriptionResponse struct {
		Status string                 `json:"status"`
		Data   []SubscriptionResponse `json:"data"`
	}

	var helcimResponse HelcimSubscriptionResponse
	if err := json.Unmarshal(body, &helcimResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Helcim subscription response wrapper: %w", err)
	}

	if helcimResponse.Status != "ok" {
		return nil, fmt.Errorf("Helcim API returned status: %s", helcimResponse.Status)
	}

	if len(helcimResponse.Data) == 0 {
		return nil, fmt.Errorf("no subscription returned in Helcim response")
	}

	return &helcimResponse.Data[0], nil
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

// CreateAddOn creates a new add-on
func (h *HelcimClient) CreateAddOn(req AddOnRequest) (*AddOnResponse, error) {
	url := fmt.Sprintf("%s/add-ons", h.BaseURL)

	request := map[string]interface{}{
		"addOns": []map[string]interface{}{
			{
				"name":        req.Name,
				"description": req.Description,
				"amount":      req.Amount,
				"type":        req.Type,
				"quantity":    req.Quantity,
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

	var responseData []AddOnResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(responseData) == 0 {
		return nil, fmt.Errorf("no add-on returned in response")
	}

	return &responseData[0], nil
}

// GetAddOn retrieves an add-on by ID
func (h *HelcimClient) GetAddOn(addOnID string) (*AddOnResponse, error) {
	url := fmt.Sprintf("%s/add-ons/%s", h.BaseURL, addOnID)

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

	var result AddOnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateAddOn updates an add-on's details
func (h *HelcimClient) UpdateAddOn(addOnID string, updates map[string]interface{}) (*AddOnResponse, error) {
	url := fmt.Sprintf("%s/add-ons", h.BaseURL)

	addOnUpdate := make(map[string]interface{})
	for k, v := range updates {
		addOnUpdate[k] = v
	}
	addOnUpdate["id"] = addOnID

	request := map[string]interface{}{
		"addOns": []map[string]interface{}{addOnUpdate},
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

	var responseData []AddOnResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(responseData) == 0 {
		return nil, fmt.Errorf("no add-on returned in response")
	}

	return &responseData[0], nil
}

// DeleteAddOn deletes an add-on
func (h *HelcimClient) DeleteAddOn(addOnID string) error {
	url := fmt.Sprintf("%s/add-ons/%s", h.BaseURL, addOnID)

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

// ListAddOns retrieves all add-ons
func (h *HelcimClient) ListAddOns() ([]AddOnResponse, error) {
	url := fmt.Sprintf("%s/add-ons", h.BaseURL)

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

	var result []AddOnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// LinkAddOnToSubscription links an add-on to a subscription
func (h *HelcimClient) LinkAddOnToSubscription(subscriptionID string, req SubscriptionAddOnRequest) (*SubscriptionAddOnResponse, error) {
	url := fmt.Sprintf("%s/subscriptions/%s/add-ons", h.BaseURL, subscriptionID)

	request := map[string]interface{}{
		"addOnId":  req.AddOnID,
		"quantity": req.Quantity,
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

	var result SubscriptionAddOnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateSubscriptionAddOn updates a subscription add-on
func (h *HelcimClient) UpdateSubscriptionAddOn(subscriptionID string, addOnID string, updates map[string]interface{}) (*SubscriptionAddOnResponse, error) {
	url := fmt.Sprintf("%s/subscriptions/%s/add-ons/%s", h.BaseURL, subscriptionID, addOnID)

	jsonData, err := json.Marshal(updates)
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

	var result SubscriptionAddOnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteSubscriptionAddOn removes an add-on from a subscription
func (h *HelcimClient) DeleteSubscriptionAddOn(subscriptionID string, addOnID string) error {
	url := fmt.Sprintf("%s/subscriptions/%s/add-ons/%s", h.BaseURL, subscriptionID, addOnID)

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

// SetCustomerCardDefault sets a customer's default credit card
func (h *HelcimClient) SetCustomerCardDefault(customerID string, cardID string) error {
	url := fmt.Sprintf("%s/customers/%s/cards/%s/default", h.BaseURL, customerID, cardID)

	httpReq, err := http.NewRequest("PATCH", url, nil)
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SetCustomerBankAccountDefault sets a customer's default bank account
func (h *HelcimClient) SetCustomerBankAccountDefault(customerID string, bankAccountID string) error {
	url := fmt.Sprintf("%s/customers/%s/bank-accounts/%s/default", h.BaseURL, customerID, bankAccountID)

	httpReq, err := http.NewRequest("PATCH", url, nil)
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ProcessSubscriptionPayment retries a failed subscription payment
func (h *HelcimClient) ProcessSubscriptionPayment(subscriptionID string) (*PaymentProcedureResponse, error) {
	url := fmt.Sprintf("%s/subscriptions/%s/process-payment", h.BaseURL, subscriptionID)

	httpReq, err := http.NewRequest("POST", url, nil)
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

	var result PaymentProcedureResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SyncSubscriptionStatus retrieves and syncs subscription status from Helcim
func (h *HelcimClient) SyncSubscriptionStatus(subscriptionID string) (*SubscriptionStatusSync, error) {
	subscription, err := h.GetSubscription(subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	sync := &SubscriptionStatusSync{
		SubscriptionID:  subscriptionID,
		Status:          subscription.Status,
		NextBillingDate: subscription.NextBillingDate,
		LastSyncAt:      time.Now(),
		PaymentMethod:   subscription.PaymentMethod,
		ActivationDate:  subscription.ActivationDate,
	}

	return sync, nil
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
		Currency:        getCurrency(),
		RecurringAmount: amount,
		BillingPeriod:   "monthly",
		Status:          "active",
	}, nil
}

func (m *mockHelcimClient) CreateSubscription(req SubscriptionRequest) (*SubscriptionResponse, error) {
	return &SubscriptionResponse{
		ID:              int(time.Now().Unix() % 1000000),
		CustomerID:      req.CustomerID,
		PaymentPlanID:   req.PaymentPlanID,
		Amount:          req.Amount,
		Status:          "active",
		ActivationDate:  time.Now().Format("2006-01-02"),
		NextBillingDate: time.Now().AddDate(0, 1, 0),
		PaymentMethod:   req.PaymentMethod,
	}, nil
}

func (m *mockHelcimClient) GetSubscription(subscriptionID string) (*SubscriptionResponse, error) {
	// Return a simulated active subscription
	now := time.Now()
	return &SubscriptionResponse{
		ID:              123456,
		CustomerID:      "dev_customer",
		PaymentPlanID:   1111,
		Amount:          10.00,
		Status:          "active",
		ActivationDate:  now.Format("2006-01-02"),
		NextBillingDate: now.AddDate(0, 1, 0),
		PaymentMethod:   "card",
	}, nil
}

func (m *mockHelcimClient) CancelSubscription(subscriptionID string) error {
	// Simulate success
	return nil
}

func (m *mockHelcimClient) UpdateSubscription(subscriptionID string, updates map[string]interface{}) (*SubscriptionResponse, error) {
	// Simulate returning an updated subscription
	sub := &SubscriptionResponse{
		ID:              123456,
		CustomerID:      "dev_customer",
		PaymentPlanID:   1111,
		Amount:          10.00,
		Status:          "active",
		ActivationDate:  time.Now().Format("2006-01-02"),
		NextBillingDate: time.Now().AddDate(0, 1, 0),
		PaymentMethod:   "card",
	}
	return sub, nil
}

func (m *mockHelcimClient) ListSubscriptionsByCustomer(customerID string) ([]SubscriptionResponse, error) {
	now := time.Now()
	return []SubscriptionResponse{
		{
			ID:              123456,
			CustomerID:      customerID,
			PaymentPlanID:   1111,
			Amount:          10.00,
			Status:          "active",
			ActivationDate:  now.Format("2006-01-02"),
			NextBillingDate: now.AddDate(0, 1, 0),
			PaymentMethod:   "card",
		},
	}, nil
}

// Mock add-on methods
func (m *mockHelcimClient) CreateAddOn(req AddOnRequest) (*AddOnResponse, error) {
	return &AddOnResponse{
		ID:          int(time.Now().Unix() % 1000000),
		Name:        req.Name,
		Description: req.Description,
		Amount:      req.Amount,
		Type:        req.Type,
		Quantity:    req.Quantity,
		Status:      "active",
	}, nil
}

func (m *mockHelcimClient) GetAddOn(addOnID string) (*AddOnResponse, error) {
	return &AddOnResponse{
		ID:          123,
		Name:        "Dev Add-on",
		Description: "Development add-on",
		Amount:      5.00,
		Type:        "one_time",
		Quantity:    true,
		Status:      "active",
	}, nil
}

func (m *mockHelcimClient) UpdateAddOn(addOnID string, updates map[string]interface{}) (*AddOnResponse, error) {
	return &AddOnResponse{
		ID:          123,
		Name:        "Updated Dev Add-on",
		Description: "Updated development add-on",
		Amount:      7.50,
		Type:        "one_time",
		Quantity:    true,
		Status:      "active",
	}, nil
}

func (m *mockHelcimClient) DeleteAddOn(addOnID string) error {
	return nil
}

func (m *mockHelcimClient) ListAddOns() ([]AddOnResponse, error) {
	return []AddOnResponse{
		{
			ID:          123,
			Name:        "Dev Add-on",
			Description: "Development add-on",
			Amount:      5.00,
			Type:        "one_time",
			Quantity:    true,
			Status:      "active",
		},
	}, nil
}

// Mock subscription add-on methods
func (m *mockHelcimClient) LinkAddOnToSubscription(subscriptionID string, req SubscriptionAddOnRequest) (*SubscriptionAddOnResponse, error) {
	return &SubscriptionAddOnResponse{
		ID:             int(time.Now().Unix() % 1000000),
		SubscriptionID: 123456,
		AddOnID:        req.AddOnID,
		Quantity:       req.Quantity,
		Amount:         5.00,
		Status:         "active",
	}, nil
}

func (m *mockHelcimClient) UpdateSubscriptionAddOn(subscriptionID string, addOnID string, updates map[string]interface{}) (*SubscriptionAddOnResponse, error) {
	return &SubscriptionAddOnResponse{
		ID:             123,
		SubscriptionID: 123456,
		AddOnID:        456,
		Quantity:       2,
		Amount:         10.00,
		Status:         "active",
	}, nil
}

func (m *mockHelcimClient) DeleteSubscriptionAddOn(subscriptionID string, addOnID string) error {
	return nil
}

// Mock payment method management
func (m *mockHelcimClient) SetCustomerCardDefault(customerID string, cardID string) error {
	return nil
}

func (m *mockHelcimClient) SetCustomerBankAccountDefault(customerID string, bankAccountID string) error {
	return nil
}

// Mock procedures
func (m *mockHelcimClient) ProcessSubscriptionPayment(subscriptionID string) (*PaymentProcedureResponse, error) {
	return &PaymentProcedureResponse{
		TransactionID: fmt.Sprintf("dev_retry_txn_%d", time.Now().Unix()),
		Status:        "APPROVED",
		Amount:        10.00,
		ProcessedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

// Mock status sync
func (m *mockHelcimClient) SyncSubscriptionStatus(subscriptionID string) (*SubscriptionStatusSync, error) {
	now := time.Now()
	return &SubscriptionStatusSync{
		SubscriptionID:  subscriptionID,
		Status:          "active",
		NextBillingDate: now.AddDate(0, 1, 0),
		LastSyncAt:      now,
		PaymentMethod:   "card",
		ActivationDate:  now.Format("2006-01-02"),
	}, nil
}

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

// Helcim API client for Payment and Recurring APIs
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
func NewHelcimClient() *HelcimClient {
	apiKey := os.Getenv("HELCIM_PRIVATE_API_KEY")
	if apiKey == "" {
		panic("[Helcim] FATAL: HELCIM_PRIVATE_API_KEY is not set or empty! Test runner did not load .env or environment variable.")
	} else {
		fmt.Printf("[Helcim] API key loaded: %s\n", apiKey[:6]+"..."+apiKey[len(apiKey)-4:]) // Print only partial for safety
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

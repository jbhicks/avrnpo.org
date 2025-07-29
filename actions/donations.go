package actions

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"avrnpo.org/models"
	"avrnpo.org/services"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate"
)

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// HelcimPayRequest represents the request to initialize a Helcim payment
type HelcimPayRequest struct {
	PaymentType string  `json:"paymentType"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Customer    struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
	} `json:"customer"`
	CompanyName string `json:"companyName"`
}

// HelcimPayResponse represents the response from Helcim payment initialization
type HelcimPayResponse struct {
	CheckoutToken string `json:"checkoutToken"`
	SecretToken   string `json:"secretToken"`
}

// DonationRequest represents the donation form data
type DonationRequest struct {
	Amount       interface{} `json:"amount" form:"amount"`
	CustomAmount string      `json:"custom_amount" form:"custom_amount"`
	DonationType string      `json:"donation_type" form:"donation_type"`
	FirstName    string      `json:"first_name" form:"first-name"`
	LastName     string      `json:"last_name" form:"last-name"`
	DonorName    string      `json:"donor_name" form:"donor_name"`
	DonorEmail   string      `json:"donor_email" form:"donor_email"`
	DonorPhone   string      `json:"donor_phone" form:"donor_phone"`
	AddressLine1 string      `json:"address_line1" form:"address-line1"`
	AddressLine2 string      `json:"address_line2" form:"address-line2"`
	City         string      `json:"city" form:"city"`
	State        string      `json:"state" form:"state"`
	Zip          string      `json:"zip" form:"zip"`
	Comments     string      `json:"comments" form:"comments"`
}

// HelcimPayVerifyRequest represents a verify request to Helcim (unified approach)
type HelcimPayVerifyRequest struct {
	PaymentType     string                    `json:"paymentType"`
	Amount          float64                   `json:"amount"`
	Currency        string                    `json:"currency"`
	CustomerRequest *services.CustomerRequest `json:"customerRequest"`
}

// Webhook event structures
type HelcimWebhookEvent struct {
	ID   string            `json:"id"`
	Type string            `json:"type"`
	Data HelcimWebhookData `json:"data"`
}

type HelcimWebhookData struct {
	ID            string                `json:"id"`
	Amount        float64               `json:"amount"`
	Currency      string                `json:"currency"`
	Status        string                `json:"status"`
	TransactionID string                `json:"transactionId"`
	CardToken     string                `json:"cardToken"`
	CustomerCode  string                `json:"customerCode"`
	Customer      HelcimWebhookCustomer `json:"customer"`
	CreatedAt     string                `json:"createdAt"`
	ProcessedAt   string                `json:"processedAt"`
}

type HelcimWebhookCustomer struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// DonationInitializeHandler initializes a Helcim payment session (UNIFIED APPROACH)
func DonationInitializeHandler(c buffalo.Context) error {
	// Parse donation request
	var req DonationRequest
	if err := c.Bind(&req); err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Invalid request data",
		}))
	}

	// Use Buffalo's validate.Errors for field-specific error collection
	errors := validate.NewErrors()

	if strings.TrimSpace(req.FirstName) == "" {
		errors.Add("first-name", "First name is required")
	}
	if strings.TrimSpace(req.LastName) == "" {
		errors.Add("last-name", "Last name is required")
	}
	if strings.TrimSpace(req.DonorEmail) == "" {
		errors.Add("donor-email", "Email address is required")
	}
	// Basic email validation
	if req.DonorEmail != "" && (!strings.Contains(req.DonorEmail, "@") || !strings.Contains(req.DonorEmail, ".")) {
		errors.Add("donor-email", "Please enter a valid email address")
	}
	if strings.TrimSpace(req.AddressLine1) == "" {
		errors.Add("address-line1", "Address Line 1 is required")
	}
	if strings.TrimSpace(req.City) == "" {
		errors.Add("city", "City is required")
	}
	if strings.TrimSpace(req.State) == "" {
		errors.Add("state", "State is required")
	}
	if strings.TrimSpace(req.Zip) == "" {
		errors.Add("zip", "ZIP Code is required")
	}

	// Determine donation amount - simplified approach
	var amount float64
	var err error

	// Convert amount to string if it's numeric
	var amountStr string
	switch v := req.Amount.(type) {
	case string:
		amountStr = v
	case float64:
		amountStr = fmt.Sprintf("%.2f", v)
	case int:
		amountStr = fmt.Sprintf("%d", v)
	case nil:
		amountStr = ""
	default:
		errors.Add("amount", "Invalid amount format")
	}

	if strings.TrimSpace(amountStr) == "" {
		errors.Add("amount", "Donation amount is required")
	} else {
		amount, err = strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			errors.Add("amount", "Donation amount must be greater than zero")
		}
	}

	// If there are any errors, render the form with errors and user input
	if errors.HasAny() {
		c.Set("errors", errors)
		c.Set("hasAnyErrors", errors.HasAny())
		c.Set("hasCommentsError", errors.Get("comments") != nil)
		c.Set("hasAmountError", errors.Get("amount") != nil)
		c.Set("hasFirstNameError", errors.Get("first-name") != nil)
		c.Set("hasLastNameError", errors.Get("last-name") != nil)
		c.Set("hasDonorEmailError", errors.Get("donor-email") != nil)
		c.Set("hasDonorPhoneError", errors.Get("donor-phone") != nil)
		c.Set("hasAddressLine1Error", errors.Get("address-line1") != nil)
		c.Set("hasCityError", errors.Get("city") != nil)
		c.Set("hasStateError", errors.Get("state") != nil)
		c.Set("hasZipError", errors.Get("zip") != nil)
		c.Set("comments", req.Comments)
		c.Set("amount", req.Amount)
		c.Set("customAmount", req.CustomAmount)
		c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})
		c.Set("firstName", req.FirstName)
		c.Set("lastName", req.LastName)
		c.Set("donorEmail", req.DonorEmail)
		c.Set("donorPhone", req.DonorPhone)
		c.Set("addressLine1", req.AddressLine1)
		c.Set("addressLine2", req.AddressLine2)
		c.Set("city", req.City)
		c.Set("state", req.State)
		c.Set("zip", req.Zip)
		return c.Render(http.StatusOK, r.HTML("pages/donate.html"))
	}

	// UNIFIED APPROACH: Always use verify mode for payment collection
	// This creates a consistent flow for both one-time and recurring donations
	donorName := strings.TrimSpace(req.FirstName + " " + req.LastName)

	helcimReq := HelcimPayVerifyRequest{
		PaymentType: "verify", // Always verify first, charge later via API
		Amount:      0,        // Verify mode requires $0
		Currency:    "USD",
		CustomerRequest: &services.CustomerRequest{
			ContactName: donorName,
			Email:       req.DonorEmail,
			BillingAddress: services.BillingAddress{
				Name:       donorName,
				Street1:    req.AddressLine1,
				City:       req.City,
				Province:   req.State,
				Country:    "USA",
				PostalCode: req.Zip,
			},
		},
	}

	// Store donation details for later processing
	donation := &models.Donation{
		DonorName:    donorName,
		DonorEmail:   req.DonorEmail,
		DonorPhone:   stringPointer(req.DonorPhone),
		AddressLine1: stringPointer(req.AddressLine1),
		AddressLine2: stringPointer(req.AddressLine2),
		City:         stringPointer(req.City),
		State:        stringPointer(req.State),
		Zip:          stringPointer(req.Zip),
		Amount:       amount,
		Currency:     "USD",
		DonationType: req.DonationType, // "one-time" or "monthly"
		Status:       "pending",
		Comments:     stringPointer(req.Comments),
	}

	// Save to database
	tx := c.Value("tx").(*pop.Connection)
	if err := tx.Create(donation); err != nil {
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to create donation record",
		}))
	}

	// Call Helcim API with verify request
	helcimResponse, err := callHelcimVerifyAPI(helcimReq)
	if err != nil {
		// Log error for debugging
		c.Logger().Errorf("Helcim API error: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Payment system unavailable. Please try again later.",
		}))
	}

	// Update donation record with Helcim tokens
	donation.CheckoutToken = helcimResponse.CheckoutToken
	donation.SecretToken = helcimResponse.SecretToken

	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("Database error updating donation: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to update donation record",
		}))
	}

	// Return success with checkout token and donation ID
	return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
		"success":       true,
		"checkoutToken": helcimResponse.CheckoutToken,
		"donationId":    donation.ID.String(),
		"amount":        amount,
		"donorName":     req.DonorName,
	}))
}

// DonationCompleteHandler handles successful donation completion
func DonationCompleteHandler(c buffalo.Context) error {
	donationID := c.Param("donationId")

	// Parse completion data
	var completionData struct {
		TransactionID string `json:"transactionId"`
		Status        string `json:"status"`
	}

	if err := c.Bind(&completionData); err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Invalid completion data",
		}))
	}

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Database connection error",
		}))
	}

	// Find donation record
	donation := &models.Donation{}
	if err := tx.Find(donation, donationID); err != nil {
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{
			"error": "Donation not found",
		}))
	}
	// Update donation with transaction details
	donation.HelcimTransactionID = &completionData.TransactionID
	donation.Status = completionData.Status
	donation.UpdatedAt = time.Now()

	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("Error updating donation: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to update donation record",
		}))
	}

	// Send donation receipt email if payment was successful
	if completionData.Status == "APPROVED" {
		emailService := services.NewEmailService()
		// Prepare receipt data
		receiptData := services.DonationReceiptData{
			DonorName:           donation.DonorName,
			DonationAmount:      donation.Amount,
			DonationType:        donation.DonationType,
			TransactionID:       *donation.HelcimTransactionID, // Dereference pointer
			DonationDate:        donation.CreatedAt,
			TaxDeductibleAmount: donation.Amount, // Full amount is tax deductible
			OrganizationEIN:     os.Getenv("ORGANIZATION_EIN"),
			OrganizationName:    "American Veterans Rebuilding",
			OrganizationAddress: os.Getenv("ORGANIZATION_ADDRESS"),
			DonorAddressLine1:   stringOrEmpty(donation.AddressLine1),
			DonorAddressLine2:   stringOrEmpty(donation.AddressLine2),
			DonorCity:           stringOrEmpty(donation.City),
			DonorState:          stringOrEmpty(donation.State),
			DonorZip:            stringOrEmpty(donation.Zip),
		}

		// stringOrEmpty safely dereferences a *string, returning "" if nil

		// Send receipt email
		if err := emailService.SendDonationReceipt(donation.DonorEmail, receiptData); err != nil {
			// Log error but don't fail the request - donation was still successful
			c.Logger().Errorf("Failed to send donation receipt email: %v", err)
		} else {
			c.Logger().Infof("Donation receipt sent to %s for transaction %s", donation.DonorEmail, *donation.HelcimTransactionID)
		}
	}

	return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
		"success": true,
		"message": "Thank you for your donation!",
	}))
}

// DonationStatusHandler retrieves donation status
func DonationStatusHandler(c buffalo.Context) error {
	donationID := c.Param("donationId")

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Database connection error",
		}))
	}

	// Find donation record
	donation := &models.Donation{}
	if err := tx.Find(donation, donationID); err != nil {
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{
			"error": "Donation not found",
		}))
	}

	// Return donation status (without sensitive tokens)
	return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
		"id":           donation.ID,
		"amount":       donation.Amount,
		"currency":     donation.Currency,
		"donorName":    donation.DonorName,
		"donationType": donation.DonationType,
		"status":       donation.Status,
		"createdAt":    donation.CreatedAt,
	}))
}

// HelcimWebhookHandler processes webhook notifications from Helcim
func HelcimWebhookHandler(c buffalo.Context) error {
	// Get the raw body for signature verification
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Logger().Errorf("Failed to read webhook body: %v", err)
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid request body"}))
	}

	// Verify webhook signature
	signature := c.Request().Header.Get("X-Helcim-Signature")
	if !verifyWebhookSignature(body, signature) {
		c.Logger().Errorf("Invalid webhook signature")
		return c.Render(http.StatusUnauthorized, r.JSON(map[string]string{"error": "Invalid signature"}))
	}

	// Parse webhook event
	var event HelcimWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.Logger().Errorf("Failed to parse webhook event: %v", err)
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid JSON"}))
	}

	// Log the webhook event
	c.Logger().Infof("Received Helcim webhook: type=%s, id=%s, transactionId=%s",
		event.Type, event.ID, event.Data.TransactionID)

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		c.Logger().Errorf("No database transaction found")
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Database error"}))
	}

	// Process based on event type
	switch event.Type {
	case "payment_success", "payment.success":
		err = handlePaymentSuccess(tx, &event, c)
	case "payment_declined", "payment.declined":
		err = handlePaymentDeclined(tx, &event, c)
	case "payment_refunded", "payment.refunded":
		err = handlePaymentRefunded(tx, &event, c)
	case "payment_cancelled", "payment.cancelled":
		err = handlePaymentCancelled(tx, &event, c)
	default:
		c.Logger().Warnf("Unknown webhook event type: %s", event.Type)
		return c.Render(http.StatusOK, r.JSON(map[string]string{"status": "ignored", "reason": "unknown event type"}))
	}

	if err != nil {
		c.Logger().Errorf("Error processing webhook event: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Processing failed"}))
	}

	return c.Render(http.StatusOK, r.JSON(map[string]string{"status": "processed"}))
}

// verifyWebhookSignature verifies the webhook signature from Helcim
func verifyWebhookSignature(body []byte, signature string) bool {
	verifierToken := os.Getenv("HELCIM_WEBHOOK_VERIFIER_TOKEN")
	if verifierToken == "" {
		// In development, we might not have this configured yet
		if os.Getenv("GO_ENV") == "development" {
			return true
		}
		return false
	}

	// Helcim uses HMAC-SHA256 for signature verification
	// Format: sha256=<signature>
	expectedSig := "sha256=" + generateHMACSignature(body, verifierToken)
	return signature == expectedSig
}

// generateHMACSignature generates HMAC-SHA256 signature
func generateHMACSignature(body []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// handlePaymentSuccess processes successful payment webhooks
func handlePaymentSuccess(tx *pop.Connection, event *HelcimWebhookEvent, c buffalo.Context) error {
	// Find the donation record by transaction ID
	donation := &models.Donation{}
	err := tx.Where("helcim_transaction_id = ?", event.Data.TransactionID).First(donation)
	if err != nil {
		// If we can't find the donation, log it but don't fail the webhook
		c.Logger().Warnf("Could not find donation for transaction ID: %s", event.Data.TransactionID)
		return nil
	}

	// Update donation status
	donation.Status = "completed"
	donation.UpdatedAt = time.Now()

	if err := tx.Save(donation); err != nil {
		return fmt.Errorf("failed to update donation status: %v", err)
	}
	// Send thank you email
	emailService := services.NewEmailService()
	receiptData := services.DonationReceiptData{
		DonorName:           donation.DonorName,
		DonationAmount:      donation.Amount,
		DonationType:        donation.DonationType,
		TransactionID:       event.Data.TransactionID,
		DonationDate:        donation.CreatedAt,
		TaxDeductibleAmount: donation.Amount, // Full amount is tax deductible
		OrganizationEIN:     "XX-XXXXXXX",    // Replace with actual EIN
		OrganizationName:    "American Veterans Rebuilding",
		OrganizationAddress: "Your Organization Address", // Replace with actual address
	}

	if err := emailService.SendDonationReceipt(donation.DonorEmail, receiptData); err != nil {
		c.Logger().Errorf("Failed to send donation receipt: %v", err)
		// Don't fail the webhook if email fails
	}

	c.Logger().Infof("Payment completed for donation ID: %s, amount: $%.2f",
		donation.ID.String(), donation.Amount)

	return nil
}

// handlePaymentDeclined processes declined payment webhooks
func handlePaymentDeclined(tx *pop.Connection, event *HelcimWebhookEvent, c buffalo.Context) error {
	donation := &models.Donation{}
	err := tx.Where("helcim_transaction_id = ?", event.Data.TransactionID).First(donation)
	if err != nil {
		c.Logger().Warnf("Could not find donation for transaction ID: %s", event.Data.TransactionID)
		return nil
	}

	donation.Status = "failed"
	donation.UpdatedAt = time.Now()

	if err := tx.Save(donation); err != nil {
		return fmt.Errorf("failed to update donation status: %v", err)
	}

	c.Logger().Infof("Payment declined for donation ID: %s", donation.ID.String())
	return nil
}

// handlePaymentRefunded processes refunded payment webhooks
func handlePaymentRefunded(tx *pop.Connection, event *HelcimWebhookEvent, c buffalo.Context) error {
	donation := &models.Donation{}
	err := tx.Where("helcim_transaction_id = ?", event.Data.TransactionID).First(donation)
	if err != nil {
		c.Logger().Warnf("Could not find donation for transaction ID: %s", event.Data.TransactionID)
		return nil
	}

	donation.Status = "refunded"
	donation.UpdatedAt = time.Now()

	if err := tx.Save(donation); err != nil {
		return fmt.Errorf("failed to update donation status: %v", err)
	}

	c.Logger().Infof("Payment refunded for donation ID: %s", donation.ID.String())
	return nil
}

// handlePaymentCancelled processes cancelled payment webhooks
func handlePaymentCancelled(tx *pop.Connection, event *HelcimWebhookEvent, c buffalo.Context) error {
	donation := &models.Donation{}
	err := tx.Where("helcim_transaction_id = ?", event.Data.TransactionID).First(donation)
	if err != nil {
		c.Logger().Warnf("Could not find donation for transaction ID: %s", event.Data.TransactionID)
		return nil
	}

	donation.Status = "cancelled"
	donation.UpdatedAt = time.Now()

	if err := tx.Save(donation); err != nil {
		return fmt.Errorf("failed to update donation status: %v", err)
	}

	c.Logger().Infof("Payment cancelled for donation ID: %s", donation.ID.String())
	return nil
}

// callHelcimVerifyAPI calls the Helcim API with verify mode for unified payment collection
func callHelcimVerifyAPI(req HelcimPayVerifyRequest) (*HelcimPayResponse, error) {
	// Get API token from environment
	apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
	if apiToken == "" {
		return nil, fmt.Errorf("HELCIM_PRIVATE_API_KEY not set")
	}

	// Marshal request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "https://api.helcim.com/v2/helcim-pay/initialize", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-token", apiToken)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Helcim API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var helcimResp HelcimPayResponse
	if err := json.Unmarshal(body, &helcimResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &helcimResp, nil
}

// ProcessPaymentHandler handles payment processing after verification (UNIFIED APPROACH)
func ProcessPaymentHandler(c buffalo.Context) error {
	var req struct {
		CustomerCode string  `json:"customerCode"`
		CardToken    string  `json:"cardToken"`
		DonationID   string  `json:"donationId"`
		Amount       float64 `json:"amount"`
	}

	if err := c.Bind(&req); err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Invalid request data",
		}))
	}

	// Get donation record
	tx := c.Value("tx").(*pop.Connection)
	donation := &models.Donation{}
	if err := tx.Find(donation, req.DonationID); err != nil {
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{
			"error": "Donation not found",
		}))
	}

	if donation.DonationType == "monthly" {
		// RECURRING DONATION: Create subscription
		return handleRecurringPayment(c, req, donation)
	} else {
		// ONE-TIME DONATION: Process immediate payment
		return handleOneTimePayment(c, req, donation)
	}
}

// handleOneTimePayment processes a one-time donation using Payment API
func handleOneTimePayment(c buffalo.Context, req struct {
	CustomerCode string  `json:"customerCode"`
	CardToken    string  `json:"cardToken"`
	DonationID   string  `json:"donationId"`
	Amount       float64 `json:"amount"`
}, donation *models.Donation) error {

	// Create Helcim client
	helcimClient := services.NewHelcimClient()

	// Use Payment API to charge the card token
	paymentReq := services.PaymentAPIRequest{
		Amount:       donation.Amount,
		Currency:     "USD",
		CustomerCode: req.CustomerCode,
		CardData: services.CardData{
			CardToken: req.CardToken,
		},
	}

	transaction, err := helcimClient.ProcessPayment(paymentReq)
	if err != nil {
		c.Logger().Errorf("Payment processing failed: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Payment processing failed",
		}))
	}

	// Update donation record
	donation.TransactionID = &transaction.TransactionID
	donation.CustomerID = &req.CustomerCode
	donation.Status = "completed"

	tx := c.Value("tx").(*pop.Connection)
	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("Failed to update donation: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to update donation",
		}))
	}

	return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
		"success":       true,
		"transactionId": transaction.TransactionID,
		"type":          "one-time",
	}))
}

// handleRecurringPayment creates a subscription using Recurring API
func handleRecurringPayment(c buffalo.Context, req struct {
	CustomerCode string  `json:"customerCode"`
	CardToken    string  `json:"cardToken"`
	DonationID   string  `json:"donationId"`
	Amount       float64 `json:"amount"`
}, donation *models.Donation) error {

	// Create Helcim client
	helcimClient := services.NewHelcimClient()

	// Create or get payment plan
	paymentPlanID, err := getOrCreateMonthlyDonationPlan(helcimClient, donation.Amount)
	if err != nil {
		c.Logger().Errorf("Failed to setup payment plan: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to setup payment plan",
		}))
	}

	// Create subscription using Recurring API
	subscription, err := helcimClient.CreateSubscription(services.SubscriptionRequest{
		CustomerID:    req.CustomerCode,
		PaymentPlanID: paymentPlanID,
		Amount:        donation.Amount,
		PaymentMethod: "card",
	})
	if err != nil {
		c.Logger().Errorf("Failed to create subscription: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to create subscription",
		}))
	}

	// Update donation record - convert int IDs to strings for storage
	subscriptionIDStr := fmt.Sprintf("%d", subscription.ID)
	paymentPlanIDStr := fmt.Sprintf("%d", paymentPlanID)

	donation.SubscriptionID = &subscriptionIDStr
	donation.CustomerID = &req.CustomerCode
	donation.PaymentPlanID = &paymentPlanIDStr
	donation.Status = "active"

	tx := c.Value("tx").(*pop.Connection)
	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("Failed to update donation: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to update donation",
		}))
	}

	return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
		"success":        true,
		"subscriptionId": subscription.ID,
		"nextBilling":    subscription.NextBillingDate,
		"type":           "recurring",
	}))
}

// getOrCreateMonthlyDonationPlan creates a payment plan for monthly donations
func getOrCreateMonthlyDonationPlan(client *services.HelcimClient, amount float64) (int, error) {
	// Create a standardized plan name based on amount
	planName := fmt.Sprintf("Monthly Donation - $%.2f", amount)

	// For now, create a new plan each time
	// TODO: In production, implement plan caching/reuse logic
	plan, err := client.CreatePaymentPlan(amount, planName)
	if err != nil {
		return 0, fmt.Errorf("failed to create payment plan: %w", err)
	}

	return plan.ID, nil
}

// stringPointer is a helper function to convert string to string pointer
func stringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Helper function to call Helcim API
func callHelcimAPI(req HelcimPayRequest) (*HelcimPayResponse, error) {
	apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
	if apiToken == "" {
		return nil, fmt.Errorf("HELCIM_PRIVATE_API_KEY not configured")
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "https://api.helcim.com/v2/helcim-pay/initialize", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("api-token", apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Helcim API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var helcimResp HelcimPayResponse
	if err := json.Unmarshal(body, &helcimResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &helcimResp, nil
}

// Helper function to split full name into first and last name
func splitName(fullName string) (string, string) {
	// Simple name splitting - can be enhanced as needed
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")
	return firstName, lastName
}

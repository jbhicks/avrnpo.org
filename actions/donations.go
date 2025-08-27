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

// getCurrency returns the configured currency with a fallback to USD
func getCurrency() string {
	currency := os.Getenv("HELCIM_CURRENCY")
	if currency == "" {
		return "USD" // Default fallback
	}
	return currency
}

// safeString ensures a value is a string, converting or defaulting as needed
func safeString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case bool:
		return "" // Don't convert booleans to strings for form fields
	case int:
		if v != 0 {
			return fmt.Sprintf("%d", v)
		}
		return ""
	case float64:
		if v != 0 {
			return fmt.Sprintf("%.2f", v)
		}
		return ""
	default:
		return ""
	}
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// isAPIRequest determines if the request is an API request or form submission
func isAPIRequest(c buffalo.Context) bool {
	// Check Accept header for JSON preference
	accept := c.Request().Header.Get("Accept")
	contentType := c.Request().Header.Get("Content-Type")

	// If Accept header explicitly asks for JSON, it's an API request
	if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
		return true
	}

	// If Content-Type is JSON, it's an API request
	if strings.Contains(contentType, "application/json") {
		return true
	}

	// Otherwise, assume it's a form submission
	return false
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
	FirstName    string      `json:"first_name" form:"first_name"`
	LastName     string      `json:"last_name" form:"last_name"`
	DonorName    string      `json:"donor_name" form:"donor_name"`
	DonorEmail   string      `json:"donor_email" form:"donor_email"`
	DonorPhone   string      `json:"donor_phone" form:"donor_phone"`
	AddressLine1 string      `json:"address_line1" form:"address_line1"`
	AddressLine2 string      `json:"address_line2" form:"address_line2"`
	City         string      `json:"city" form:"city"`
	State        string      `json:"state" form:"state"`
	Zip          string      `json:"zip_code" form:"zip_code"`
	Comments     string      `json:"comments" form:"comments"`
}

// HelcimPayVerifyRequest represents a verify request to Helcim (unified approach)
type HelcimPayVerifyRequest struct {
	PaymentType     string                    `json:"paymentType"`
	Amount          float64                   `json:"amount"`
	Currency        string                    `json:"currency"`
	CustomerRequest *services.CustomerRequest `json:"customerRequest"`
}

// Webhook event structures for Helcim's actual format
type HelcimWebhookEvent struct {
	ID   string                 `json:"id"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"` // Optional - for terminalCancel events
}

type HelcimCardTransactionEvent struct {
	ID   string `json:"id"`   // Transaction ID
	Type string `json:"type"` // "cardTransaction"
}

type HelcimTerminalCancelEvent struct {
	Type string                 `json:"type"` // "terminalCancel"
	Data map[string]interface{} `json:"data"`
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
	// Subscription-specific fields
	SubscriptionID  string `json:"subscriptionId"`
	PaymentPlanID   string `json:"paymentPlanId"`
	PaymentNumber   int    `json:"paymentNumber"`
	NextBillingDate string `json:"nextBillingDate"`
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
		// Check if this is an API request or form submission
		if isAPIRequest(c) {
			return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
				"error": "Invalid request data",
			}))
		}
		// For form submissions, redirect back with error
		c.Flash().Add("error", "Invalid form data submitted")
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Use Buffalo's validate.Errors for field-specific error collection
	errors := validate.NewErrors()

	if strings.TrimSpace(req.FirstName) == "" {
		errors.Add("first_name", "First name is required")
	}
	if strings.TrimSpace(req.LastName) == "" {
		errors.Add("last_name", "Last name is required")
	}
	if strings.TrimSpace(req.DonorEmail) == "" {
		errors.Add("donor_email", "Email address is required")
	}
	// Basic email validation
	if req.DonorEmail != "" && (!strings.Contains(req.DonorEmail, "@") || !strings.Contains(req.DonorEmail, ".")) {
		errors.Add("donor_email", "Please enter a valid email address")
	}
	if strings.TrimSpace(req.AddressLine1) == "" {
		errors.Add("address_line1", "Address Line 1 is required")
	}
	if strings.TrimSpace(req.City) == "" {
		errors.Add("city", "City is required")
	}
	if strings.TrimSpace(req.State) == "" {
		errors.Add("state", "State is required")
	}
	if strings.TrimSpace(req.Zip) == "" {
		errors.Add("zip_code", "ZIP Code is required")
	}

	// Determine donation amount - check both form and session
	var amount float64
	var err error

	// First try to get amount from form submission
	amountStr := strings.TrimSpace(req.CustomAmount)

	// If no amount in form, check session (from preset button selections)
	if amountStr == "" {
		if sessionAmount := c.Session().Get("donation_amount"); sessionAmount != nil {
			amountStr = sessionAmount.(string)
		}
	}

	// Normalize money strings like "$25.00" or "25,00" -> "25.00"
	if amountStr != "" {
		// Remove currency symbols and commas
		amountStr = strings.ReplaceAll(amountStr, "$", "")
		amountStr = strings.ReplaceAll(amountStr, ",", "")
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
		// Check if this is an API request or form submission
		if isAPIRequest(c) {
			// Build a descriptive error message that includes required field info
			errorMsg := "Validation failed"
			for fieldName, fieldErrors := range errors.Errors {
				for _, err := range fieldErrors {
					if strings.Contains(err, "required") {
						errorMsg = "Required fields missing: " + fieldName + " - " + err
						break
					}
				}
			}
			return c.Render(http.StatusBadRequest, r.JSON(map[string]interface{}{
				"error":  errorMsg,
				"errors": errors,
			}))
		}

		// For form submissions, render the template with errors
		c.Set("errors", errors)
		c.Set("hasAnyErrors", errors.HasAny())
		c.Set("hasCommentsError", errors.Get("comments") != nil)
		c.Set("hasAmountError", errors.Get("amount") != nil)
		c.Set("hasFirstNameError", errors.Get("first_name") != nil)
		c.Set("hasLastNameError", errors.Get("last_name") != nil)
		c.Set("hasDonorEmailError", errors.Get("donor_email") != nil)
		c.Set("hasDonorPhoneError", errors.Get("donor_phone") != nil)
		c.Set("hasAddressLine1Error", errors.Get("address_line1") != nil)
		c.Set("hasCityError", errors.Get("city") != nil)
		c.Set("hasStateError", errors.Get("state") != nil)
		c.Set("hasZipError", errors.Get("zip_code") != nil)
		c.Set("comments", req.Comments)

		// Convert amount to string to avoid template rendering issues
		amountStr := ""
		if req.Amount != nil {
			switch v := req.Amount.(type) {
			case string:
				amountStr = v
			case float64:
				if v > 0 {
					amountStr = fmt.Sprintf("%.2f", v)
				}
			case int:
				if v > 0 {
					amountStr = fmt.Sprintf("%d", v)
				}
			}
		}
		c.Set("amount", amountStr)

		// Ensure customAmount is always a safe string
		c.Set("customAmount", safeString(req.CustomAmount))
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
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}

	// UNIFIED APPROACH: Always use verify mode for payment collection
	// This creates a consistent flow for both one-time and recurring donations
	donorName := strings.TrimSpace(req.FirstName + " " + req.LastName)

	helcimReq := HelcimPayVerifyRequest{
		PaymentType: "verify", // Always verify first, charge later via API
		Amount:      0,        // Verify mode requires $0
		Currency:    getCurrency(),
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
		Currency:     getCurrency(),
		DonationType: req.DonationType, // "one-time" or "monthly"
		Status:       "pending",
		Comments:     stringPointer(req.Comments),
	}

	// Link to user account if logged in
	if currentUser, ok := c.Value("current_user").(*models.User); ok && currentUser != nil {
		donation.UserID = &currentUser.ID
	}

	// Ensure amount is valid before saving - extra safeguard
	if amount <= 0 {
		if isAPIRequest(c) {
			return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid donation amount"}))
		}
		c.Flash().Add("error", "Invalid donation amount. Please try again.")
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Save to database
	tx := c.Value("tx").(*pop.Connection)
	if err := tx.Create(donation); err != nil {
		if isAPIRequest(c) {
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
				"error": "Failed to create donation record",
			}))
		}
		c.Flash().Add("error", "System error occurred. Please try again.")
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Call Helcim API with verify request
	helcimResponse, err := callHelcimVerifyAPI(helcimReq)
	if err != nil {
		// Log error for debugging
		c.Logger().Errorf("Helcim API error: %v", err)
		if isAPIRequest(c) {
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
				"error": "Payment system unavailable. Please try again later.",
			}))
		}
		c.Flash().Add("error", "Payment system unavailable. Please try again later.")
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Update donation record with Helcim tokens
	donation.CheckoutToken = helcimResponse.CheckoutToken
	donation.SecretToken = helcimResponse.SecretToken

	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("Database error updating donation: %v", err)
		if isAPIRequest(c) {
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
				"error": "Failed to update donation record",
			}))
		}
		c.Flash().Add("error", "System error occurred. Please try again.")
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Return success with checkout token and donation ID
	if isAPIRequest(c) {
		return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
			"success":       true,
			"checkoutToken": helcimResponse.CheckoutToken,
			"secretToken":   helcimResponse.SecretToken,
			"donationId":    donation.ID.String(),
			"amount":        amount,
			"donorName":     req.DonorName,
		}))
	}

	// For form submissions, redirect to payment processing page
	// Store checkout data in session for the payment page
	// Store amount as formatted string to avoid template rendering issues
	c.Session().Set("donation_id", donation.ID.String())
	c.Session().Set("checkout_token", helcimResponse.CheckoutToken)
	c.Session().Set("amount", fmt.Sprintf("%.2f", amount))
	c.Session().Set("donor_name", donorName)
	return c.Redirect(http.StatusSeeOther, "/donate/payment")
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
		// Map stored donation type to display label
		displayType := donation.DonationType
		if displayType == "monthly" {
			displayType = "Monthly"
		} else {
			displayType = "One-time"
		}

		receiptData := services.DonationReceiptData{
			DonorName:           donation.DonorName,
			DonationAmount:      donation.Amount,
			DonationType:        displayType,
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

	// Log the webhook event for debugging (signature verified)
	c.Logger().Infof("Received Helcim webhook: type=%s, id=%s", event.Type, event.ID)
	c.Logger().Debugf("Helcim webhook raw payload: %s", string(body))

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		c.Logger().Errorf("No database transaction found")
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Database error"}))
	}

	// Process based on event type - Helcim only sends cardTransaction and terminalCancel events
	switch event.Type {
	case "cardTransaction":
		// For cardTransaction events, event.ID contains the transaction ID
		err = handleCardTransaction(tx, event.ID, c)
	case "terminalCancel":
		// Terminal cancellation events - handle if needed
		c.Logger().Infof("Received terminal cancel event - ignoring for donation system")
		return c.Render(http.StatusOK, r.JSON(map[string]string{"status": "ignored", "reason": "terminal cancel not applicable"}))
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

// handleCardTransaction processes cardTransaction webhook events from Helcim
func handleCardTransaction(tx *pop.Connection, transactionID string, c buffalo.Context) error {
	c.Logger().Infof("Processing cardTransaction webhook for transaction ID: %s", transactionID)

	// Find the donation record by Helcim transaction ID
	donation := &models.Donation{}
	err := tx.Where("helcim_transaction_id = ?", transactionID).First(donation)
	if err != nil {
		// Try finding by transaction_id field as well
		err2 := tx.Where("transaction_id = ?", transactionID).First(donation)
		if err2 != nil {
			// If we can't find the donation, log it but don't fail the webhook
			c.Logger().Warnf("Could not find donation for transaction ID: %s - may be external transaction", transactionID)
			return nil
		}
	}

	c.Logger().Infof("Found donation record for transaction %s - donor: %s, amount: $%.2f, type: %s",
		transactionID, donation.DonorEmail, donation.Amount, donation.DonationType)

	// For webhook events, we need to fetch the full transaction details from Helcim API
	// to determine if it was successful, declined, etc.
	// For now, we'll assume success since Helcim sends webhooks for both success and failure

	// Update donation status to completed (webhooks typically indicate successful processing)
	donation.Status = "completed"
	if donation.HelcimTransactionID == nil {
		donation.HelcimTransactionID = &transactionID
	}
	if donation.TransactionID == nil {
		donation.TransactionID = &transactionID
	}
	donation.UpdatedAt = time.Now()

	if err := tx.Save(donation); err != nil {
		c.Logger().Errorf("Failed to update donation status for transaction %s: %v", transactionID, err)
		return fmt.Errorf("failed to update donation status: %v", err)
	}

	// Send receipt email for completed payments
	emailService := services.NewEmailService()

	// Determine display type for receipt
	displayType := donation.DonationType
	if displayType == "monthly" {
		displayType = "Monthly"
	} else {
		displayType = "One-time"
	}

	receiptData := services.DonationReceiptData{
		DonorName:           donation.DonorName,
		DonationAmount:      donation.Amount,
		DonationType:        displayType,
		TransactionID:       transactionID,
		DonationDate:        donation.CreatedAt,
		TaxDeductibleAmount: donation.Amount,
		OrganizationEIN:     os.Getenv("ORGANIZATION_EIN"),
		OrganizationName:    "American Veterans Rebuilding",
		OrganizationAddress: os.Getenv("ORGANIZATION_ADDRESS"),
		DonorAddressLine1:   stringOrEmpty(donation.AddressLine1),
		DonorAddressLine2:   stringOrEmpty(donation.AddressLine2),
		DonorCity:           stringOrEmpty(donation.City),
		DonorState:          stringOrEmpty(donation.State),
		DonorZip:            stringOrEmpty(donation.Zip),
	}

	// For recurring donations, add subscription details if available
	if donation.IsRecurring() && donation.SubscriptionID != nil {
		receiptData.SubscriptionID = *donation.SubscriptionID
	}

	if err := emailService.SendDonationReceipt(donation.DonorEmail, receiptData); err != nil {
		c.Logger().Errorf("Failed to send donation receipt for transaction %s: %v", transactionID, err)
		// Don't fail the webhook for email issues
	} else {
		c.Logger().Infof("Donation receipt sent successfully for transaction %s to %s", transactionID, donation.DonorEmail)
	}

	c.Logger().Infof("Successfully processed cardTransaction webhook for transaction %s", transactionID)
	return nil
}

// callHelcimVerifyAPI calls the Helcim API with verify mode for unified payment collection
func callHelcimVerifyAPI(req HelcimPayVerifyRequest) (*HelcimPayResponse, error) {
	// Check if we're in test environment - return mock data instead of calling real API
	if os.Getenv("GO_ENV") == "test" {
		// Return mock success response for tests
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
		return &HelcimPayResponse{
			CheckoutToken: "test_checkout_token_" + timestamp,
			SecretToken:   "test_secret_token_" + timestamp,
		}, nil
	}

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

	// Validate required fields for payment processing
	if req.CustomerCode == "" {
		c.Logger().Errorf("ProcessPaymentHandler: missing customerCode - req=%+v", req)
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Missing customer code - payment verification may have failed",
		}))
	}

	if req.DonationID == "" {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Missing donation ID",
		}))
	}

	c.Logger().Infof("ProcessPaymentHandler: customerCode=%s, donationId=%s, amount=%.2f",
		req.CustomerCode, req.DonationID, req.Amount)

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

	// Server-side sanity check: donation amount must be > 0
	if donation.Amount <= 0 {
		c.Logger().Errorf("Refusing to process payment: stored donation amount invalid (%.2f). req.Amount=%.2f donation.ID=%s", donation.Amount, req.Amount, donation.ID.String())
		return c.Render(http.StatusBadRequest, r.JSON(map[string]interface{}{
			"success": false,
			"error":   "Invalid donation amount on server",
		}))
	}

	// Log a mismatch if client-supplied amount differs significantly from stored amount
	if req.Amount > 0 && (fmt.Sprintf("%.2f", req.Amount) != fmt.Sprintf("%.2f", donation.Amount)) {
		c.Logger().Warnf("Client amount (%.2f) differs from stored donation amount (%.2f) for donation ID %s", req.Amount, donation.Amount, donation.ID.String())
	}

	// Check if we should use live payments even in development
	useLivePayments := os.Getenv("HELCIM_LIVE_TESTING") == "true"

	// TEMPORARY: For development, simulate successful payment (unless live testing enabled)
	if os.Getenv("GO_ENV") == "development" && !useLivePayments {
		c.Logger().Info("Development mode: Simulating successful payment")

		// Generate a fake transaction ID
		transactionID := fmt.Sprintf("dev_txn_%d", time.Now().Unix())

		// Update donation record
		donation.TransactionID = &transactionID
		donation.CustomerID = &req.CustomerCode
		donation.Status = "completed"

		tx := c.Value("tx").(*pop.Connection)
		if err := tx.Update(donation); err != nil {
			c.Logger().Errorf("Failed to update donation: %v", err)
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]interface{}{
				"success": false,
				"error":   "Failed to update donation",
			}))
		}

		// Send donation receipt email in development
		emailService := services.NewEmailService()
		// Map stored donation type to display label for receipt
		displayType := donation.DonationType
		if displayType == "monthly" {
			displayType = "Monthly"
		} else {
			displayType = "One-time"
		}

		receiptData := services.DonationReceiptData{
			DonorName:           donation.DonorName,
			DonationAmount:      donation.Amount,
			DonationType:        displayType,
			TransactionID:       transactionID,
			DonationDate:        donation.CreatedAt,
			TaxDeductibleAmount: donation.Amount,
			OrganizationEIN:     os.Getenv("ORGANIZATION_EIN"),
			OrganizationName:    "American Veterans Rebuilding",
			OrganizationAddress: os.Getenv("ORGANIZATION_ADDRESS"),
			DonorAddressLine1:   stringOrEmpty(donation.AddressLine1),
			DonorAddressLine2:   stringOrEmpty(donation.AddressLine2),
			DonorCity:           stringOrEmpty(donation.City),
			DonorState:          stringOrEmpty(donation.State),
			DonorZip:            stringOrEmpty(donation.Zip),
		}

		if err := emailService.SendDonationReceipt(donation.DonorEmail, receiptData); err != nil {
			c.Logger().Errorf("Failed to send donation receipt email: %v", err)
		} else {
			c.Logger().Infof("Development: Donation receipt sent to %s for transaction %s", donation.DonorEmail, transactionID)
		}

		return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
			"success":       true,
			"transactionId": transactionID,
			"type":          "one-time",
			"message":       "Payment processed successfully!",
		}))
	}

	// Production: Use real Helcim API
	helcimClient := services.NewHelcimClient()

	// Use Payment API to charge the card token
	paymentReq := services.PaymentAPIRequest{
		Amount:       donation.Amount,
		Currency:     getCurrency(),
		CustomerCode: req.CustomerCode,
		CardData: services.CardData{
			CardToken: req.CardToken,
		},
	}

	transaction, err := helcimClient.ProcessPayment(paymentReq)
	if err != nil {
		c.Logger().Errorf("Payment processing failed: %v", err)
		c.Logger().Errorf("Payment request data: %+v", paymentReq)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]interface{}{
			"success": false,
			"error":   "Payment processing failed: " + err.Error(),
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
	// Check if we should use live payments even in development
	useLivePayments := os.Getenv("HELCIM_LIVE_TESTING") == "true"

	// DEVELOPMENT-SAFE PATH: Simulate subscription creation when in development (unless live testing enabled)
	if os.Getenv("GO_ENV") == "development" && !useLivePayments {
		c.Logger().Infof("Development mode: Simulating recurring subscription creation - donation_id=%s, amount=%.2f, donor=%s",
			donation.ID.String(), donation.Amount, donation.DonorEmail)

		// Create fake IDs and next billing date
		subscriptionID := fmt.Sprintf("dev_sub_%d", time.Now().Unix())
		paymentPlanID := fmt.Sprintf("dev_plan_%.0f", donation.Amount)
		nextBilling := time.Now().AddDate(0, 1, 0)

		c.Logger().Infof("Generated development subscription: subscription_id=%s, plan_id=%s, next_billing=%s",
			subscriptionID, paymentPlanID, nextBilling.Format("2006-01-02"))

		// Update donation record
		donation.SubscriptionID = &subscriptionID
		donation.CustomerID = &req.CustomerCode
		donation.PaymentPlanID = &paymentPlanID
		donation.Status = "active"

		tx := c.Value("tx").(*pop.Connection)
		if err := tx.Update(donation); err != nil {
			c.Logger().Errorf("Failed to update donation: %v", err)
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
				"error": "Failed to update donation",
			}))
		}

		// Send simulated receipt email for subscription creation
		emailService := services.NewEmailService()
		receiptData := services.DonationReceiptData{
			DonorName:           donation.DonorName,
			DonationAmount:      donation.Amount,
			DonationType:        "Monthly",
			SubscriptionID:      subscriptionID,
			NextBillingDate:     &nextBilling,
			TransactionID:       "",
			DonationDate:        donation.CreatedAt,
			TaxDeductibleAmount: donation.Amount,
			OrganizationEIN:     os.Getenv("ORGANIZATION_EIN"),
			OrganizationName:    "American Veterans Rebuilding",
			OrganizationAddress: os.Getenv("ORGANIZATION_ADDRESS"),
			DonorAddressLine1:   stringOrEmpty(donation.AddressLine1),
			DonorAddressLine2:   stringOrEmpty(donation.AddressLine2),
			DonorCity:           stringOrEmpty(donation.City),
			DonorState:          stringOrEmpty(donation.State),
			DonorZip:            stringOrEmpty(donation.Zip),
		}

		if err := emailService.SendDonationReceipt(donation.DonorEmail, receiptData); err != nil {
			c.Logger().Errorf("Failed to send subscription receipt email: %v", err)
		} else {
			c.Logger().Infof("Development: Subscription receipt sent to %s for subscription %s", donation.DonorEmail, subscriptionID)
		}

		return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
			"success":        true,
			"subscriptionId": subscriptionID,
			"nextBilling":    nextBilling,
			"type":           "recurring",
		}))
	}

	// Create Helcim client
	helcimClient := services.NewHelcimClient()

	// Create or get payment plan
	c.Logger().Infof("Creating payment plan for recurring donation - donation_id=%s, amount=%.2f, donor=%s",
		donation.ID.String(), donation.Amount, donation.DonorEmail)
	paymentPlanID, err := getOrCreateMonthlyDonationPlan(helcimClient, donation.Amount)
	if err != nil {
		c.Logger().Errorf("Failed to setup payment plan for donation_id=%s, amount=%.2f: %v",
			donation.ID.String(), donation.Amount, err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to setup payment plan",
		}))
	}
	c.Logger().Infof("Payment plan created successfully - plan_id=%d, donation_id=%s", paymentPlanID, donation.ID.String())

	// Create subscription using Recurring API
	c.Logger().Infof("Creating Helcim subscription - customer_code=%s, plan_id=%d, amount=%.2f",
		req.CustomerCode, paymentPlanID, donation.Amount)
	subscription, err := helcimClient.CreateSubscription(services.SubscriptionRequest{
		CustomerID:    req.CustomerCode,
		PaymentPlanID: paymentPlanID,
		Amount:        donation.Amount, // Use actual donation amount for subscription
		PaymentMethod: "card",
	})
	if err != nil {
		c.Logger().Errorf("Failed to create Helcim subscription - donation_id=%s, customer_code=%s, plan_id=%d: %v",
			donation.ID.String(), req.CustomerCode, paymentPlanID, err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to create subscription",
		}))
	}
	c.Logger().Infof("Helcim subscription created successfully - subscription_id=%d, next_billing=%s, donation_id=%s",
		subscription.ID, subscription.NextBillingDate.Format("2006-01-02"), donation.ID.String())

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

	// Send receipt email for subscription creation (recurring donation)
	emailService := services.NewEmailService()
	receiptData := services.DonationReceiptData{
		DonorName:           donation.DonorName,
		DonationAmount:      donation.Amount,
		DonationType:        "Monthly",
		SubscriptionID:      subscriptionIDStr,
		NextBillingDate:     &subscription.NextBillingDate,
		TransactionID:       "", // No one-time transaction ID for subscriptions on create
		DonationDate:        donation.CreatedAt,
		TaxDeductibleAmount: donation.Amount,
		OrganizationEIN:     os.Getenv("ORGANIZATION_EIN"),
		OrganizationName:    "American Veterans Rebuilding",
		OrganizationAddress: os.Getenv("ORGANIZATION_ADDRESS"),
		DonorAddressLine1:   stringOrEmpty(donation.AddressLine1),
		DonorAddressLine2:   stringOrEmpty(donation.AddressLine2),
		DonorCity:           stringOrEmpty(donation.City),
		DonorState:          stringOrEmpty(donation.State),
		DonorZip:            stringOrEmpty(donation.Zip),
	}

	if err := emailService.SendDonationReceipt(donation.DonorEmail, receiptData); err != nil {
		c.Logger().Errorf("Failed to send subscription receipt email: %v", err)
	} else {
		c.Logger().Infof("Subscription receipt sent to %s for subscription %s", donation.DonorEmail, subscriptionIDStr)
	}

	return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
		"success":        true,
		"subscriptionId": subscription.ID,
		"nextBilling":    subscription.NextBillingDate,
		"type":           "recurring",
	}))
}

// getOrCreateMonthlyDonationPlan creates or reuses standardized payment plans for monthly donations
func getOrCreateMonthlyDonationPlan(client services.HelcimAPI, amount float64) (int, error) {
	// Standardized donation amounts to reduce plan proliferation
	// Note: The subscription amount can override the plan amount, so we can use standardized plans
	// while still charging the exact requested amount per Helcim documentation
	standardAmounts := []float64{5, 10, 25, 50, 100, 250, 500, 1000}

	// Find the closest standard amount or use exact amount for large donations
	var standardAmount float64
	if amount >= 1000 {
		// For large donations over $1000, create exact plans to maintain accuracy
		standardAmount = amount
	} else {
		// Find closest standard amount for common donation ranges
		standardAmount = findClosestStandardAmount(amount, standardAmounts)
	}

	// Create a standardized plan name and cache key
	planName := fmt.Sprintf("Monthly Donation - $%.0f", standardAmount)
	cacheKey := fmt.Sprintf("plan_%.0f_%s", standardAmount, getCurrency())

	// Check if we have a cached plan first
	if cachedPlan, found := services.GetPaymentPlanCache().Get(cacheKey); found {
		fmt.Printf("[PaymentPlan] Using cached plan ID %d for $%.2f\n", cachedPlan.ID, standardAmount)
		if standardAmount != amount {
			fmt.Printf("[PaymentPlan] Using standardized plan amount $%.2f instead of exact $%.2f\n", standardAmount, amount)
		}
		return cachedPlan.ID, nil
	}

	// Create new payment plan if not cached
	plan, err := client.CreatePaymentPlan(standardAmount, planName)
	if err != nil {
		return 0, fmt.Errorf("failed to create payment plan for $%.2f: %w", standardAmount, err)
	}

	// Cache the newly created plan
	services.GetPaymentPlanCache().Set(cacheKey, plan)
	fmt.Printf("[PaymentPlan] Created and cached new plan ID %d for $%.2f\n", plan.ID, standardAmount)

	// Log if we're using a different amount than requested (for monitoring)
	if standardAmount != amount {
		fmt.Printf("[PaymentPlan] Using standardized plan amount $%.2f instead of exact $%.2f\n", standardAmount, amount)
	}

	return plan.ID, nil
}

// findClosestStandardAmount finds the closest standard amount to the requested amount
func findClosestStandardAmount(amount float64, standardAmounts []float64) float64 {
	if len(standardAmounts) == 0 {
		return amount
	}

	closest := standardAmounts[0]
	minDiff := abs(amount - closest)

	for _, standard := range standardAmounts {
		diff := abs(amount - standard)
		if diff < minDiff {
			minDiff = diff
			closest = standard
		}
	}

	return closest
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
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

// DonateUpdateAmountHandler handles HTMX updates to donation amounts
func DonateUpdateAmountHandler(c buffalo.Context) error {
	// Add error handling wrapper
	defer func() {
		if r := recover(); r != nil {
			c.Logger().Errorf("Panic in DonateUpdateAmountHandler: %v", r)
			c.Error(http.StatusInternalServerError, fmt.Errorf("Internal server error"))
		}
	}()

	// Get form values
	amount := c.Param("amount")
	donationType := c.Param("donation_type")
	source := c.Param("source")

	c.Logger().Debugf("DonateUpdateAmountHandler called with amount=%s, donationType=%s, source=%s", amount, donationType, source)

	// Get current state from session with safe type checking
	sessionAmount := ""
	if sessionAmountInterface := c.Session().Get("donation_amount"); sessionAmountInterface != nil {
		if str, ok := sessionAmountInterface.(string); ok {
			sessionAmount = str
		}
	}

	sessionDonationType := "one-time" // Default
	if sessionDonationTypeInterface := c.Session().Get("donation_type"); sessionDonationTypeInterface != nil {
		if str, ok := sessionDonationTypeInterface.(string); ok {
			sessionDonationType = str
		}
	}

	// Update amount in session if a new amount is provided
	if amount != "" {
		c.Session().Set("donation_amount", amount)
		sessionAmount = amount
	}

	// Update donation type in session if provided
	if donationType != "" {
		c.Session().Set("donation_type", donationType)
		sessionDonationType = donationType
	}

	// Use session values if not provided in request (e.g., radio button clicks)
	if amount == "" {
		amount = sessionAmount
	}
	if donationType == "" {
		donationType = sessionDonationType
	}

	// Preserve existing form values from the request
	firstName := c.Param("first_name")
	lastName := c.Param("last_name")
	donorEmail := c.Param("donor_email")
	donorPhone := c.Param("donor_phone")
	addressLine1 := c.Param("address_line1")
	addressLine2 := c.Param("address_line2")
	city := c.Param("city")
	state := c.Param("state")
	zipCode := c.Param("zip_code")
	comments := c.Param("comments")

	// Set template variables for the donation form with defensive programming
	c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})
	c.Set("amount", amount)
	c.Set("donationType", donationType)
	c.Set("source", source)

	// Preserve form values with safe defaults
	c.Set("firstName", safeString(firstName))
	c.Set("lastName", safeString(lastName))
	c.Set("donorEmail", safeString(donorEmail))
	c.Set("donorPhone", safeString(donorPhone))
	c.Set("addressLine1", safeString(addressLine1))
	c.Set("addressLine2", safeString(addressLine2))
	c.Set("city", safeString(city))
	c.Set("state", safeString(state))
	c.Set("zip", safeString(zipCode))
	c.Set("comments", safeString(comments))

	// Set error flags to false (no errors in amount updates)
	c.Set("errors", nil)
	c.Set("hasAnyErrors", false)
	c.Set("hasCommentsError", false)
	c.Set("hasAmountError", false)
	c.Set("hasFirstNameError", false)
	c.Set("hasLastNameError", false)
	c.Set("hasDonorEmailError", false)
	c.Set("hasDonorPhoneError", false)
	c.Set("hasAddressLine1Error", false)
	c.Set("hasCityError", false)
	c.Set("hasStateError", false)
	c.Set("hasZipError", false)

	// Ensure CSRF token is available for template
	if token := c.Value("authenticity_token"); token != nil {
		c.Set("authenticity_token", token)
	} else {
		c.Logger().Warn("CSRF token not available in DonateUpdateAmountHandler context")
		c.Set("authenticity_token", "")
	}

	// Render appropriate content based on request type
	if c.Request().Header.Get("HX-Request") == "true" {
		// HTMX request - return only the form content
		return c.Render(http.StatusOK, r.HTML("pages/_donate_form.plush.html"))
	} else {
		// Regular request - return the complete page
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}
}

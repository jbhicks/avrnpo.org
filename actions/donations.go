package actions

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
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

// safePrefix safely returns the first n characters of a string, or the whole string if shorter
func safePrefix(s string, n int) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
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

// getClientIP extracts the client's IP address from the request
func getClientIP(c buffalo.Context) string {
	c.Logger().Debugf("[getClientIP] RemoteAddr: %s, X-Forwarded-For: %s, X-Real-IP: %s",
		c.Request().RemoteAddr,
		c.Request().Header.Get("X-Forwarded-For"),
		c.Request().Header.Get("X-Real-IP"))
	// Check X-Forwarded-For header first (for proxy/load balancer scenarios)
	xff := c.Request().Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		ip := strings.TrimSpace(ips[0])
		// Validate it's a valid IP
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Check X-Real-IP header (for nginx proxy)
	xri := c.Request().Header.Get("X-Real-IP")
	if xri != "" && net.ParseIP(xri) != nil {
		return xri
	}

	// Fall back to RemoteAddr
	ip := c.Request().RemoteAddr

	// Handle IPv6 addresses in brackets like [::1]:port
	if strings.HasPrefix(ip, "[") && strings.Contains(ip, "]:") {
		// IPv6 with port: [::1]:8080
		end := strings.Index(ip, "]:")
		if end != -1 {
			ip = ip[1:end] // Remove brackets
		}
	} else if strings.Contains(ip, ":") {
		// IPv4 with port: 127.0.0.1:8080
		ip, _, _ = strings.Cut(ip, ":")
	}

	// Validate the final IP
	if net.ParseIP(ip) != nil {
		c.Logger().Debugf("[getClientIP] Successfully parsed IP: %s", ip)
		return ip
	}

	// Fallback to a default IP if parsing fails
	c.Logger().Warnf("[getClientIP] Failed to parse valid IP from RemoteAddr: %s, X-Forwarded-For: %s, X-Real-IP: %s, using fallback 127.0.0.1",
		c.Request().RemoteAddr,
		c.Request().Header.Get("X-Forwarded-For"),
		c.Request().Header.Get("X-Real-IP"))
	return "127.0.0.1"
}

// HelcimPayRequest represents the request to initialize a Helcim payment
// This corresponds to the HelcimPay.js initialize API endpoint:
// POST https://api.helcim.com/v2/helcim-pay/initialize
// See: https://devdocs.helcim.com/docs/initialize-helcimpayjs
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
// Contains tokens required for frontend HelcimPay.js integration:
// - checkoutToken: Used by appendHelcimPayIframe() to render payment modal
// - secretToken: Used for transaction validation and webhook verification
// Tokens expire after 60 minutes per Helcim documentation
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
	c.Logger().Infof("[DonationInitialize] Starting donation initialization - Method: %s, Path: %s", c.Request().Method, c.Request().URL.Path)

	// Parse donation request
	var req DonationRequest
	if err := c.Bind(&req); err != nil {
		c.Logger().Errorf("[DonationInitialize] Failed to bind request: %v", err)
		// If CSRF token is missing/invalid, mw-csrf will already have returned 403 before reaching here.
		// For API requests with malformed JSON, return 400.
		if isAPIRequest(c) {
			return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
				"error": "Invalid request data",
			}))
		}
		// For form submissions, redirect back with error
		c.Flash().Add("error", "Invalid form data submitted")
		setDonateContext(c, nil)
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	c.Logger().Infof("[DonationInitialize] Request parsed - Type: %s, Amount: %s, Email: %s, Name: %s %s",
		req.DonationType, req.CustomAmount, req.DonorEmail, req.FirstName, req.LastName)

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

	// First try to get amount from form submission (HTMX or regular)
	amountStr := strings.TrimSpace(req.CustomAmount)

	// If no amount in form, check for amount_source (HTMX preset selection)
	if amountStr == "" {
		if sourceAmount := c.Param("custom_amount"); sourceAmount != "" {
			amountStr = sourceAmount
		}
	}

	// If still no amount, check session (fallback)
	if amountStr == "" {
		if sessionAmount := c.Session().Get("donation_amount"); sessionAmount != nil {
			if s, ok := sessionAmount.(string); ok {
				amountStr = s
			}
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
		c.Logger().Warnf("[DonationInitialize] Validation failed - Errors: %v", errors.Errors)
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
		setDonateContext(c, nil)

		// Check if this is an HTMX request
		isHTMX := c.Request().Header.Get("HX-Request") == "true"
		if isHTMX {
			c.Logger().Infof("[DonationInitialize] Returning HTMX form fragment due to validation errors")
			// For HTMX requests, return only the form content
			return c.Render(http.StatusOK, r.HTML("pages/_donate_form.plush.html"))
		}

		c.Logger().Infof("[DonationInitialize] Returning full donate page due to validation errors")
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}

	// Always process as full form submission, never partial
	c.Logger().Infof("[DonationInitialize] Validation passed - proceeding with donation creation")

	// UNIFIED APPROACH: Always use verify mode for payment collection
	// This creates a consistent flow for both one-time and recurring donations
	donorName := strings.TrimSpace(req.FirstName + " " + req.LastName)
	c.Logger().Infof("[DonationInitialize] Creating donation record - Name: %s, Amount: $%.2f, Type: %s",
		donorName, amount, req.DonationType)

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

	c.Logger().Debugf("[DonationInitialize] Helcim verify request prepared - Customer: %s, Email: %s, Address: %s, %s, %s %s",
		donorName, req.DonorEmail, req.AddressLine1, req.City, req.State, req.Zip)

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
		setDonateContext(c, nil)
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Save to database
	c.Logger().Infof("[DonationInitialize] Saving donation to database - ID will be generated")
	tx := c.Value("tx").(*pop.Connection)
	if err := tx.Create(donation); err != nil {
		c.Logger().Errorf("[DonationInitialize] Failed to create donation record: %v", err)
		if isAPIRequest(c) {
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
				"error": "Failed to create donation record",
			}))
		}
		c.Flash().Add("error", "System error occurred. Please try again.")
		ensureDonateContext(c)
		return c.Redirect(http.StatusSeeOther, "/donate")
	}
	c.Logger().Infof("[DonationInitialize] Donation record created successfully - ID: %s", donation.ID.String())

	// Call Helcim API with verify request
	c.Logger().Infof("[DonationInitialize] Calling Helcim verify API for donation %s", donation.ID.String())
	helcimResponse, err := callHelcimVerifyAPI(helcimReq)
	if err != nil {
		c.Logger().Errorf("[DonationInitialize] Helcim API error for donation %s: %v", donation.ID.String(), err)
		if isAPIRequest(c) {
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
				"error": "Payment system unavailable. Please try again later.",
			}))
		}
		c.Flash().Add("error", "Payment system unavailable. Please try again later.")
		ensureDonateContext(c)
		return c.Redirect(http.StatusSeeOther, "/donate")
	}
	c.Logger().Infof("[DonationInitialize] Helcim verify successful - CheckoutToken: %s, SecretToken: %s",
		helcimResponse.CheckoutToken[:8]+"...", helcimResponse.SecretToken[:8]+"...")

	// Update donation record with Helcim tokens
	c.Logger().Infof("[DonationInitialize] Updating donation %s with Helcim tokens", donation.ID.String())
	donation.CheckoutToken = helcimResponse.CheckoutToken
	donation.SecretToken = helcimResponse.SecretToken

	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("[DonationInitialize] Database error updating donation %s: %v", donation.ID.String(), err)
		if isAPIRequest(c) {
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
				"error": "Failed to update donation record",
			}))
		}
		c.Flash().Add("error", "System error occurred. Please try again.")
		ensureDonateContext(c)
		return c.Redirect(http.StatusSeeOther, "/donate")
	}
	c.Logger().Infof("[DonationInitialize] Donation %s updated successfully with tokens", donation.ID.String())

	// Return success with checkout token and donation ID
	if isAPIRequest(c) {
		c.Logger().Infof("[DonationInitialize] Returning API response for donation %s", donation.ID.String())
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
	c.Logger().Infof("[DonationInitialize] Redirecting to payment page for donation %s", donation.ID.String())
	// Store checkout data in session for the payment page
	// Store amount as formatted string to avoid template rendering issues
	c.Session().Set("donation_id", donation.ID.String())
	c.Session().Set("checkout_token", helcimResponse.CheckoutToken)
	c.Session().Set("amount", fmt.Sprintf("%.2f", amount))
	c.Session().Set("donor_name", donorName)
	ensureDonateContext(c)
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
	fmt.Printf("[WEBHOOK] Received Helcim webhook - Method: %s, Content-Type: %s\n",
		c.Request().Method, c.Request().Header.Get("Content-Type"))
	c.Logger().Infof("[Webhook] Received Helcim webhook - Method: %s, Content-Type: %s",
		c.Request().Method, c.Request().Header.Get("Content-Type"))

	// Get the raw body for signature verification
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Logger().Errorf("[Webhook] Failed to read webhook body: %v", err)
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid request body"}))
	}

	c.Logger().Debugf("[Webhook] Raw body length: %d bytes", len(body))

	// Verify webhook signature
	signature := c.Request().Header.Get("X-Helcim-Signature")
	signaturePreview := signature
	if len(signature) > 16 {
		signaturePreview = signature[:16] + "..."
	}
	c.Logger().Debugf("[Webhook] Verifying signature: %s", signaturePreview)

	if !verifyWebhookSignature(body, signature) {
		c.Logger().Errorf("[Webhook] Invalid webhook signature - rejecting request")
		return c.Render(http.StatusUnauthorized, r.JSON(map[string]string{"error": "Invalid signature"}))
	}

	c.Logger().Infof("[Webhook] Signature verification successful")

	// Parse webhook event
	var event HelcimWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.Logger().Errorf("[Webhook] Failed to parse webhook event: %v", err)
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid JSON"}))
	}

	// Log the webhook event for debugging (signature verified)
	c.Logger().Infof("[Webhook] Parsed webhook event - Type: %s, ID: %s", event.Type, event.ID)
	c.Logger().Debugf("[Webhook] Full event data: %+v", event)

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		c.Logger().Errorf("No database transaction found")
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Database error"}))
	}

	// Process based on event type - Helcim only sends cardTransaction and terminalCancel events
	switch event.Type {
	case "cardTransaction":
		// For cardTransaction events, parse the detailed data from the Data field
		var webhookData HelcimWebhookData
		if event.Data != nil {
			// Convert the map to JSON and then unmarshal to structured data
			dataJSON, err := json.Marshal(event.Data)
			if err != nil {
				c.Logger().Errorf("[Webhook] Failed to marshal event data: %v", err)
				return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid event data"}))
			}

			if err := json.Unmarshal(dataJSON, &webhookData); err != nil {
				c.Logger().Errorf("[Webhook] Failed to parse webhook data: %v", err)
				return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid webhook data format"}))
			}

			// Log subscription information for recurring payments
			if webhookData.SubscriptionID != "" {
				c.Logger().Infof("[Webhook] Recurring payment detected - SubscriptionID: %s, PaymentPlanID: %s, PaymentNumber: %d, NextBillingDate: %s",
					webhookData.SubscriptionID, webhookData.PaymentPlanID, webhookData.PaymentNumber, webhookData.NextBillingDate)
			} else {
				c.Logger().Infof("[Webhook] One-time payment detected - TransactionID: %s", webhookData.TransactionID)
			}

			// Log detailed transaction information
			c.Logger().Infof("[Webhook] Transaction details - Amount: $%.2f %s, Status: %s, Customer: %s %s (%s)",
				webhookData.Amount, webhookData.Currency, webhookData.Status,
				webhookData.Customer.FirstName, webhookData.Customer.LastName, webhookData.CustomerCode)
		}

		// Use transaction ID from webhook data if available, otherwise fall back to event.ID
		transactionID := event.ID
		if webhookData.TransactionID != "" {
			transactionID = webhookData.TransactionID
		}

		err = handleCardTransaction(tx, transactionID, c)
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
	// Test-only bypass: see AGENTS.md. Only active in test when HELCIM_TEST_BYPASS=="true".
	if os.Getenv("GO_ENV") == "test" && os.Getenv("HELCIM_TEST_BYPASS") == "true" {
		return true
	}

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
	c.Logger().Infof("[Webhook] Processing cardTransaction webhook for transaction ID: %s", transactionID)

	// Find the donation record by Helcim transaction ID
	donation := &models.Donation{}
	err := tx.Where("helcim_transaction_id = ?", transactionID).First(donation)
	if err != nil {
		c.Logger().Debugf("[Webhook] Donation not found by helcim_transaction_id, trying transaction_id")
		// Try finding by transaction_id field as well
		err2 := tx.Where("transaction_id = ?", transactionID).First(donation)
		if err2 != nil {
			// If we can't find the donation, log it but don't fail the webhook
			c.Logger().Warnf("[Webhook] Could not find donation for transaction ID: %s - may be external transaction", transactionID)
			return nil
		}
	}

	c.Logger().Infof("[Webhook] Found donation record for transaction %s - ID: %s, Donor: %s, Amount: $%.2f, Type: %s",
		transactionID, donation.ID.String(), donation.DonorEmail, donation.Amount, donation.DonationType)

	// Enhanced logging for recurring donations
	if donation.DonationType == "monthly" {
		if donation.SubscriptionID != nil {
			c.Logger().Infof("[Webhook] Recurring donation details - SubscriptionID: %s, PaymentPlanID: %s, Status: %s",
				*donation.SubscriptionID,
				stringOrEmpty(donation.PaymentPlanID),
				donation.Status)
		} else {
			c.Logger().Warnf("[Webhook] Recurring donation found but missing subscription data - DonationID: %s", donation.ID.String())
		}
	}

	// For webhook events, we need to fetch the full transaction details from Helcim API
	// to determine if it was successful, declined, etc.
	// For now, we'll assume success since Helcim sends webhooks for both success and failure

	// Update donation status to completed (webhooks typically indicate successful processing)
	c.Logger().Infof("[Webhook] Updating donation %s status to completed", donation.ID.String())
	donation.Status = "completed"
	if donation.HelcimTransactionID == nil {
		donation.HelcimTransactionID = &transactionID
	}
	if donation.TransactionID == nil {
		donation.TransactionID = &transactionID
	}
	donation.UpdatedAt = time.Now()

	if err := tx.Save(donation); err != nil {
		c.Logger().Errorf("[Webhook] Failed to update donation %s status for transaction %s: %v",
			donation.ID.String(), transactionID, err)
		return fmt.Errorf("failed to update donation status: %v", err)
	}
	c.Logger().Infof("[Webhook] Donation %s status updated successfully", donation.ID.String())

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
// Uses the official HelcimPay.js initialize endpoint:
// POST https://api.helcim.com/v2/helcim-pay/initialize
// See: docs/payment-system/helcim-integration.md for complete API documentation
func callHelcimVerifyAPI(req HelcimPayVerifyRequest) (*HelcimPayResponse, error) {
	// Check if we're in test environment - return mock data instead of calling real API
	if os.Getenv("GO_ENV") == "test" {
		// Return mock success response for tests
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
		fmt.Printf("[HelcimVerify] Test environment detected - returning mock tokens\n")
		return &HelcimPayResponse{
			CheckoutToken: "test_checkout_token_" + timestamp,
			SecretToken:   "test_secret_token_" + timestamp,
		}, nil
	}

	fmt.Printf("[HelcimVerify] Calling Helcim verify API - PaymentType: %s, Amount: %.2f, Currency: %s\n",
		req.PaymentType, req.Amount, req.Currency)

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
		fmt.Printf("[HelcimVerify] API error - Status: %d, Response: %s\n", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Helcim API error %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[HelcimVerify] API call successful - Status: %d\n", resp.StatusCode)

	// Parse response
	var helcimResp HelcimPayResponse
	if err := json.Unmarshal(body, &helcimResp); err != nil {
		fmt.Printf("[HelcimVerify] Failed to parse response: %v\n", err)
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	fmt.Printf("[HelcimVerify] Response parsed successfully - CheckoutToken: %s..., SecretToken: %s...\n",
		helcimResp.CheckoutToken[:8], helcimResp.SecretToken[:8])

	return &helcimResp, nil
}

// ProcessPaymentHandler handles payment processing after verification (UNIFIED APPROACH)
func ProcessPaymentHandler(c buffalo.Context) error {
	fmt.Printf("[ProcessPayment] Handler called - Method: %s\n", c.Request().Method)
	c.Logger().Infof("[ProcessPayment] Starting payment processing - Method: %s", c.Request().Method)

	var req struct {
		CustomerCode  string `json:"customerCode"`
		CardToken     string `json:"cardToken"`
		DonationID    string `json:"donationId"`
		TransactionID string `json:"transactionId"`
		Amount        string `json:"amount"` // Accept as string from JavaScript
	}

	if err := c.Bind(&req); err != nil {
		c.Logger().Errorf("[ProcessPayment] Failed to bind request: %v", err)
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Invalid request data",
		}))
	}

	// Parse amount string to float64
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		c.Logger().Errorf("[ProcessPayment] Failed to parse amount '%s': %v", req.Amount, err)
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Invalid amount format",
		}))
	}

	c.Logger().Infof("[ProcessPayment] Request parsed - CustomerCode: %s, DonationID: %s, Amount: $%.2f",
		req.CustomerCode, req.DonationID, amount)
	c.Logger().Debugf("[ProcessPayment] Full request data - CardToken: %s, TransactionID: %s",
		safePrefix(req.CardToken, 8)+"...", req.TransactionID)

	// Validate required fields for payment processing
	if req.CustomerCode == "" {
		// Generate a customer code if missing (fallback for HelcimPay.js response issues)
		req.CustomerCode = fmt.Sprintf("DON_%s_%d", req.DonationID, time.Now().Unix())
		c.Logger().Warnf("[ProcessPayment] Missing customerCode - generated fallback: %s for DonationID: %s",
			req.CustomerCode, req.DonationID)
	}

	if req.DonationID == "" {
		c.Logger().Errorf("[ProcessPayment] Missing donation ID")
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Missing donation ID",
		}))
	}

	c.Logger().Infof("[ProcessPayment] Validation passed - proceeding with payment for donation %s", req.DonationID)

	// Get donation record
	tx := c.Value("tx").(*pop.Connection)
	donation := &models.Donation{}
	if err := tx.Find(donation, req.DonationID); err != nil {
		c.Logger().Errorf("[ProcessPayment] Donation not found: %s - Error: %v", req.DonationID, err)
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{
			"error": "Donation not found",
		}))
	}

	c.Logger().Infof("[ProcessPayment] Donation found - ID: %s, Type: %s, Amount: $%.2f, Donor: %s",
		donation.ID.String(), donation.DonationType, donation.Amount, donation.DonorEmail)

	// Create payment request struct with parsed amount
	var paymentReq = struct {
		CustomerCode string  `json:"customerCode"`
		CardToken    string  `json:"cardToken"`
		DonationID   string  `json:"donationId"`
		Amount       float64 `json:"amount"`
	}{
		CustomerCode: req.CustomerCode,
		CardToken:    req.CardToken,
		DonationID:   req.DonationID,
		Amount:       amount,
	}

	if donation.DonationType == "monthly" {
		c.Logger().Infof("[ProcessPayment] Routing to recurring payment handler for donation %s", donation.ID.String())
		// RECURRING DONATION: Create subscription
		return handleRecurringPayment(c, paymentReq, donation)
	} else {
		c.Logger().Infof("[ProcessPayment] Routing to one-time payment handler for donation %s", donation.ID.String())
		// ONE-TIME DONATION: Process immediate payment
		return handleOneTimePayment(c, paymentReq, donation)
	}
}

// handleOneTimePayment processes a one-time donation using Payment API
func handleOneTimePayment(c buffalo.Context, req struct {
	CustomerCode string  `json:"customerCode"`
	CardToken    string  `json:"cardToken"`
	DonationID   string  `json:"donationId"`
	Amount       float64 `json:"amount"`
}, donation *models.Donation) error {

	c.Logger().Infof("[OneTimePayment] Starting one-time payment processing for donation %s", donation.ID.String())
	c.Logger().Debugf("[OneTimePayment] Payment details - CustomerCode: %s, CardToken: %s..., Amount: $%.2f",
		req.CustomerCode, safePrefix(req.CardToken, 8)+"...", req.Amount)

	// Server-side sanity check: donation amount must be > 0
	if donation.Amount <= 0 {
		c.Logger().Errorf("[OneTimePayment] Refusing to process payment: stored donation amount invalid (%.2f). req.Amount=%.2f donation.ID=%s",
			donation.Amount, req.Amount, donation.ID.String())
		return c.Render(http.StatusBadRequest, r.JSON(map[string]interface{}{
			"success": false,
			"error":   "Invalid donation amount on server",
		}))
	}

	// Log a mismatch if client-supplied amount differs significantly from stored amount
	if req.Amount > 0 && (fmt.Sprintf("%.2f", req.Amount) != fmt.Sprintf("%.2f", donation.Amount)) {
		c.Logger().Warnf("[OneTimePayment] Client amount (%.2f) differs from stored donation amount (%.2f) for donation ID %s",
			req.Amount, donation.Amount, donation.ID.String())
	}

	c.Logger().Infof("[OneTimePayment] Amount validation passed - proceeding with payment for $%.2f", donation.Amount)

	// Check if we should use live payments even in development
	useLivePayments := os.Getenv("HELCIM_LIVE_TESTING") == "true"
	c.Logger().Debugf("[OneTimePayment] Environment check - GO_ENV: %s, HELCIM_LIVE_TESTING: %s, useLivePayments: %t",
		os.Getenv("GO_ENV"), os.Getenv("HELCIM_LIVE_TESTING"), useLivePayments)

	// TEMPORARY: For development, simulate successful payment (unless live testing enabled)
	if os.Getenv("GO_ENV") == "development" && !useLivePayments {
		c.Logger().Infof("[OneTimePayment] Development mode: Simulating successful payment for donation %s", donation.ID.String())

		// Generate a fake transaction ID
		transactionID := fmt.Sprintf("dev_txn_%d", time.Now().Unix())
		c.Logger().Debugf("[OneTimePayment] Generated dev transaction ID: %s", transactionID)

		// Update donation record
		donation.TransactionID = &transactionID
		donation.CustomerID = &req.CustomerCode
		donation.Status = "completed"

		tx := c.Value("tx").(*pop.Connection)
		if err := tx.Update(donation); err != nil {
			c.Logger().Errorf("[OneTimePayment] Failed to update donation %s: %v", donation.ID.String(), err)
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]interface{}{
				"success": false,
				"error":   "Failed to update donation",
			}))
		}
		c.Logger().Infof("[OneTimePayment] Donation %s updated successfully with dev transaction", donation.ID.String())

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
			c.Logger().Errorf("[OneTimePayment] Failed to send donation receipt email for %s: %v", donation.DonorEmail, err)
		} else {
			c.Logger().Infof("[OneTimePayment] Development: Donation receipt sent to %s for transaction %s", donation.DonorEmail, transactionID)
		}

		response := map[string]interface{}{
			"success":       true,
			"transactionId": transactionID,
			"type":          "one-time",
			"message":       "Payment processed successfully!",
		}
		c.Logger().Infof("[OneTimePayment] Development simulation completed successfully for donation %s - Response: %+v", donation.ID.String(), response)
		return c.Render(http.StatusOK, r.JSON(response))
	}

	// Production: Use real Helcim API
	c.Logger().Infof("[OneTimePayment] Production mode: Calling Helcim Payment API for donation %s", donation.ID.String())
	helcimClient := services.NewHelcimClient()

	// Generate unique idempotency key for this payment (UUID format)
	// Use Payment API to charge the card token
	paymentReq := services.PaymentAPIRequest{
		PaymentType:  "purchase",
		Amount:       donation.Amount,
		Currency:     getCurrency(),
		CustomerCode: req.CustomerCode,
		CardData: services.CardData{
			CardToken: req.CardToken,
		},
		IPAddress:     getClientIP(c),
		Description:   "Donation to American Veterans Rebuilding",
		CustomerEmail: donation.DonorEmail,
		CustomerName:  donation.DonorName,
		BillingAddress: &services.BillingAddress{
			Name:       donation.DonorName,
			Street1:    stringOrEmpty(donation.AddressLine1),
			City:       stringOrEmpty(donation.City),
			Province:   stringOrEmpty(donation.State),
			Country:    "USA", // Helcim expects 3-letter country codes
			PostalCode: stringOrEmpty(donation.Zip),
		},
	}

	c.Logger().Debugf("[OneTimePayment] Payment request - Amount: $%.2f, Currency: %s, CustomerCode: %s, CardToken: %s",
		paymentReq.Amount, paymentReq.Currency, paymentReq.CustomerCode, safePrefix(paymentReq.CardData.CardToken, 8)+"...")

	transaction, err := helcimClient.ProcessPayment(paymentReq)
	if err != nil {
		c.Logger().Errorf("[OneTimePayment] Payment processing failed for donation %s: %v", donation.ID.String(), err)
		c.Logger().Errorf("[OneTimePayment] Payment request data: Amount=$%.2f, Currency=%s, CustomerCode=%s, CardToken=%s",
			paymentReq.Amount, paymentReq.Currency, paymentReq.CustomerCode, safePrefix(paymentReq.CardData.CardToken, 8)+"...")
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]interface{}{
			"success": false,
			"error":   "Payment processing failed: " + err.Error(),
		}))
	}

	transactionIDStr := fmt.Sprintf("%d", transaction.TransactionID)
	c.Logger().Infof("[OneTimePayment] Payment successful - TransactionID: %s, Status: %s",
		transactionIDStr, transaction.Status)

	// Update donation record
	donation.TransactionID = &transactionIDStr
	donation.CustomerID = &req.CustomerCode
	donation.Status = "completed"

	tx := c.Value("tx").(*pop.Connection)
	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("[OneTimePayment] Failed to update donation %s: %v", donation.ID.String(), err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to update donation",
		}))
	}

	c.Logger().Infof("[OneTimePayment] Donation %s completed successfully - TransactionID: %s",
		donation.ID.String(), transaction.TransactionID)

	// Send donation receipt email
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
		TransactionID:       transactionIDStr,
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
		c.Logger().Errorf("[OneTimePayment] Failed to send donation receipt email for %s: %v", donation.DonorEmail, err)
	} else {
		c.Logger().Infof("[OneTimePayment] Donation receipt sent to %s for transaction %s", donation.DonorEmail, transactionIDStr)
	}

	response := map[string]interface{}{
		"success":       true,
		"transactionId": transaction.TransactionID,
		"type":          "one-time",
		"message":       "Payment processed successfully!",
	}
	c.Logger().Infof("[OneTimePayment] Returning success response: %+v", response)
	return c.Render(http.StatusOK, r.JSON(response))
}

// handleRecurringPayment creates a subscription using Recurring API
func handleRecurringPayment(c buffalo.Context, req struct {
	CustomerCode string  `json:"customerCode"`
	CardToken    string  `json:"cardToken"`
	DonationID   string  `json:"donationId"`
	Amount       float64 `json:"amount"`
}, donation *models.Donation) error {

	c.Logger().Infof("[RecurringPayment] Starting recurring payment processing for donation %s", donation.ID.String())
	c.Logger().Debugf("[RecurringPayment] Payment details - CustomerCode: %s, CardToken: %s..., Amount: $%.2f",
		req.CustomerCode, safePrefix(req.CardToken, 8)+"...", req.Amount)

	// Check if we should use live payments even in development
	useLivePayments := os.Getenv("HELCIM_LIVE_TESTING") == "true"
	c.Logger().Debugf("[RecurringPayment] Environment check - GO_ENV: %s, HELCIM_LIVE_TESTING: %s, useLivePayments: %t",
		os.Getenv("GO_ENV"), os.Getenv("HELCIM_LIVE_TESTING"), useLivePayments)

	// DEVELOPMENT-SAFE PATH: Simulate subscription creation when in development (unless live testing enabled)
	if os.Getenv("GO_ENV") == "development" && !useLivePayments {
		c.Logger().Infof("[RecurringPayment] Development mode: Simulating recurring subscription creation - donation_id=%s, amount=%.2f, donor=%s",
			donation.ID.String(), donation.Amount, donation.DonorEmail)

		// Create fake IDs and next billing date
		subscriptionID := fmt.Sprintf("dev_sub_%d", time.Now().Unix())
		paymentPlanID := fmt.Sprintf("dev_plan_%.0f", donation.Amount)
		nextBilling := time.Now().AddDate(0, 1, 0)

		c.Logger().Infof("[RecurringPayment] Generated development subscription: subscription_id=%s, plan_id=%s, next_billing=%s",
			subscriptionID, paymentPlanID, nextBilling.Format("2006-01-02"))

		// Update donation record
		donation.SubscriptionID = &subscriptionID
		donation.CustomerID = &req.CustomerCode
		donation.PaymentPlanID = &paymentPlanID
		donation.Status = "active"

		tx := c.Value("tx").(*pop.Connection)
		if err := tx.Update(donation); err != nil {
			c.Logger().Errorf("[RecurringPayment] Failed to update donation %s: %v", donation.ID.String(), err)
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
				"error": "Failed to update donation",
			}))
		}
		c.Logger().Infof("[RecurringPayment] Donation %s updated successfully with dev subscription", donation.ID.String())

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
			c.Logger().Errorf("[RecurringPayment] Failed to send subscription receipt email for %s: %v", donation.DonorEmail, err)
		} else {
			c.Logger().Infof("[RecurringPayment] Development: Subscription receipt sent to %s for subscription %s", donation.DonorEmail, subscriptionID)
		}

		c.Logger().Infof("[RecurringPayment] Development simulation completed successfully for donation %s", donation.ID.String())
		return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
			"success":        true,
			"subscriptionId": subscriptionID,
			"nextBilling":    nextBilling,
			"type":           "recurring",
		}))
	}

	// Create Helcim client
	c.Logger().Infof("[RecurringPayment] Production mode: Creating Helcim client for donation %s", donation.ID.String())
	helcimClient := services.NewHelcimClient()

	// Create or get payment plan
	c.Logger().Infof("[RecurringPayment] Creating payment plan for recurring donation - donation_id=%s, amount=%.2f, donor=%s",
		donation.ID.String(), donation.Amount, donation.DonorEmail)
	paymentPlanID, err := getOrCreateMonthlyDonationPlan(helcimClient, donation.Amount)
	if err != nil {
		c.Logger().Errorf("[RecurringPayment] Failed to setup payment plan for donation_id=%s, amount=%.2f: %v",
			donation.ID.String(), donation.Amount, err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to setup payment plan",
		}))
	}
	c.Logger().Infof("[RecurringPayment] Payment plan created successfully - plan_id=%d, donation_id=%s", paymentPlanID, donation.ID.String())

	// Create subscription using Recurring API
	c.Logger().Infof("[RecurringPayment] Creating Helcim subscription - customer_code=%s, plan_id=%d, amount=%.2f",
		req.CustomerCode, paymentPlanID, donation.Amount)
	subscription, err := helcimClient.CreateSubscription(services.SubscriptionRequest{
		CustomerID:    req.CustomerCode,
		PaymentPlanID: paymentPlanID,
		Amount:        donation.Amount, // Use actual donation amount for subscription
		PaymentMethod: "card",
	})
	if err != nil {
		c.Logger().Errorf("[RecurringPayment] Failed to create Helcim subscription - donation_id=%s, customer_code=%s, plan_id=%d: %v",
			donation.ID.String(), req.CustomerCode, paymentPlanID, err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to create subscription",
		}))
	}
	c.Logger().Infof("[RecurringPayment] Helcim subscription created successfully - subscription_id=%d, next_billing=%s, donation_id=%s",
		subscription.ID, subscription.NextBillingDate.Format("2006-01-02"), donation.ID.String())

	// Update donation record - convert int IDs to strings for storage
	subscriptionIDStr := fmt.Sprintf("%d", subscription.ID)
	paymentPlanIDStr := fmt.Sprintf("%d", paymentPlanID)

	c.Logger().Infof("[RecurringPayment] Updating donation %s with subscription details - SubscriptionID: %s, PaymentPlanID: %s",
		donation.ID.String(), subscriptionIDStr, paymentPlanIDStr)

	donation.SubscriptionID = &subscriptionIDStr
	donation.CustomerID = &req.CustomerCode
	donation.PaymentPlanID = &paymentPlanIDStr
	donation.Status = "active"

	tx := c.Value("tx").(*pop.Connection)
	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("[RecurringPayment] Failed to update donation %s: %v", donation.ID.String(), err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to update donation",
		}))
	}
	c.Logger().Infof("[RecurringPayment] Donation %s updated successfully with subscription details", donation.ID.String())

	// Send receipt email for subscription creation (recurring donation)
	c.Logger().Infof("[RecurringPayment] Sending subscription receipt email to %s", donation.DonorEmail)
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
		c.Logger().Errorf("[RecurringPayment] Failed to send subscription receipt email to %s: %v", donation.DonorEmail, err)
	} else {
		c.Logger().Infof("[RecurringPayment] Subscription receipt sent successfully to %s for subscription %s", donation.DonorEmail, subscriptionIDStr)
	}

	c.Logger().Infof("[RecurringPayment] Recurring payment processing completed successfully for donation %s - SubscriptionID: %s",
		donation.ID.String(), subscriptionIDStr)

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

// DonateUpdateSubmitHandler handles HTMX requests to update only the submit button text
func DonateUpdateSubmitHandler(c buffalo.Context) error {
	defer func() {
		if r := recover(); r != nil {
			c.Logger().Errorf("Panic in DonateUpdateSubmitHandler: %v", r)
			// For HTMX requests, redirect rather than injecting a full page into a fragment
			if c.Request().Header.Get("HX-Request") == "true" {
				c.Response().Header().Set("HX-Redirect", "/donate")
				c.Response().WriteHeader(http.StatusOK)
				return
			}
			c.Error(http.StatusInternalServerError, fmt.Errorf("Internal server error"))
		}
	}()

	ensureDonateContext(c)
	req := c.Request()
	_ = req.ParseForm()

	normalizeMoney := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.ReplaceAll(s, "$", "")
		s = strings.ReplaceAll(s, ",", "")
		return s
	}
	firstNonEmpty := func(vals ...string) string {
		for _, v := range vals {
			if strings.TrimSpace(v) != "" {
				return v
			}
		}
		return ""
	}

	// Determine selected amount and donation type from form/session
	amountRaw := firstNonEmpty(req.FormValue("amount"), req.PostFormValue("amount"), req.FormValue("custom_amount"), req.PostFormValue("custom_amount"), c.Param("amount"))
	amountRaw = normalizeMoney(amountRaw)
	if amountRaw == "" {
		if v := c.Session().Get("donation_amount"); v != nil {
			if str, ok := v.(string); ok && strings.TrimSpace(str) != "" {
				amountRaw = str
			}
		}
	}

	donationType := firstNonEmpty(req.FormValue("donation_type"))
	if donationType == "" {
		if v := c.Session().Get("donation_type"); v != nil {
			if str, ok := v.(string); ok && strings.TrimSpace(str) != "" {
				donationType = str
			}
		}
	}
	if donationType == "" {
		donationType = "one-time"
	}

	// Persist session updates if any
	if amountRaw != "" {
		c.Session().Set("donation_amount", amountRaw)
	}
	if donationType != "" {
		c.Session().Set("donation_type", donationType)
	}

	// Compute button text
	buttonText := "Donate Now"
	if amountRaw != "" {
		if parsedAmount, err := strconv.ParseFloat(amountRaw, 64); err == nil && parsedAmount > 0 {
			formattedAmount := fmt.Sprintf("%.2f", parsedAmount)
			if donationType == "monthly" {
				buttonText = fmt.Sprintf("Donate $%s Monthly", formattedAmount)
			} else {
				buttonText = fmt.Sprintf("Donate $%s Now", formattedAmount)
			}
		}
	} else if donationType == "monthly" {
		buttonText = "Donate Monthly"
	}

	c.Set("buttonText", buttonText)

	// Render only the submit button fragment
	return c.Render(http.StatusOK, rFrag.HTML("pages/_submit_button.plush.html"))
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

	// Ensure minimal donate context is present
	ensureDonateContext(c)

	// Parse form data defensively (HTMX sends form-encoded)
	req := c.Request()
	_ = req.ParseForm()

	firstNonEmpty := func(vals ...string) string {
		for _, v := range vals {
			if strings.TrimSpace(v) != "" {
				return v
			}
		}
		return ""
	}
	normalizeMoney := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.ReplaceAll(s, "$", "")
		s = strings.ReplaceAll(s, ",", "")
		return s
	}

	// Read inputs
	source := firstNonEmpty(req.FormValue("source"), c.Param("source"))
	formDonationType := firstNonEmpty(req.FormValue("donation_type"), c.Param("donation_type"))
	// Default donation type: one-time; fallback to session if present
	sessionDonationType := "one-time"
	if v := c.Session().Get("donation_type"); v != nil {
		if str, ok := v.(string); ok && strings.TrimSpace(str) != "" {
			sessionDonationType = str
		}
	}
	donationType := firstNonEmpty(formDonationType, sessionDonationType, "one-time")

	// HTMX may send values via hx-vals JSON; Buffalo exposes them as form values
	// Try multiple locations for amount/custom amount, including hx-vals and included hidden inputs
	// Also consider query string (HTMX sometimes sends hx-vals merged but tests may send raw body)
	q := req.URL.Query()
	// Accept amount from multiple potential sources; tests post form-encoded body like "amount=25&source=preset&donation_type=one-time"
	amountRaw := firstNonEmpty(
		req.FormValue("amount"), // Buffalo merges PostForm and Form
		req.PostFormValue("amount"),
		req.Form.Get("amount"),
		q.Get("amount"),
		c.Param("amount"),
	)
	customRaw := firstNonEmpty(
		req.PostFormValue("custom_amount"),
		req.FormValue("custom_amount"),
		req.Form.Get("custom_amount"),
		q.Get("custom_amount"),
		c.Param("custom_amount"),
	)
	// Some test harnesses may pass amount via URL params
	paramAmount := firstNonEmpty(c.Param("amount"))

	// Selection logic by source
	var selected string
	switch source {
	case "preset":
		selected = amountRaw
	case "custom":
		selected = customRaw
	default:
		selected = firstNonEmpty(amountRaw, customRaw, paramAmount)
	}

	// As a final fallback, some tests only send amount in the body; ensure we don't drop it
	if selected == "" {
		selected = req.PostFormValue("amount")
	}

	selected = normalizeMoney(selected)
	// If none provided but source=preset and amountRaw exists, use it
	if selected == "" && source == "preset" {
		selected = normalizeMoney(amountRaw)
	}
	// Fallback to session if still empty
	if selected == "" {
		if v := c.Session().Get("donation_amount"); v != nil {
			if str, ok := v.(string); ok && strings.TrimSpace(str) != "" {
				selected = str
			}
		}
	}

	// Update session if we have values
	if selected != "" {
		c.Session().Set("donation_amount", selected)
	}
	if donationType != "" {
		c.Session().Set("donation_type", donationType)
	}

	// Preserve existing form values from the request
	firstName := req.FormValue("first_name")
	lastName := req.FormValue("last_name")
	donorEmail := req.FormValue("donor_email")
	donorPhone := req.FormValue("donor_phone")
	addressLine1 := req.FormValue("address_line1")
	addressLine2 := req.FormValue("address_line2")
	city := req.FormValue("city")
	state := req.FormValue("state")
	zipCode := req.FormValue("zip_code")
	comments := req.FormValue("comments")

	// Set template variables
	c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})
	c.Set("amount", selected)
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

	// HTMX / wants HTML?
	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	accept := c.Request().Header.Get("Accept")
	wantsHTML := strings.Contains(accept, "text/html") && !strings.Contains(accept, "application/json")

	if isHTMX || wantsHTML {
		c.Response().Header().Set("HX-Trigger", "donation-amount-updated")

		buttonText := "Donate Now"
		if selected != "" {
			if parsedAmount, err := strconv.ParseFloat(selected, 64); err == nil && parsedAmount > 0 {
				formattedAmount := fmt.Sprintf("%.2f", parsedAmount)
				if donationType == "monthly" {
					buttonText = fmt.Sprintf("Donate $%s Monthly", formattedAmount)
				} else {
					buttonText = fmt.Sprintf("Donate $%s Now", formattedAmount)
				}
			}
		} else if donationType == "monthly" {
			buttonText = "Donate Monthly"
		}
		c.Set("buttonText", buttonText)

		// Ensure hidden input reflects current selection for tests looking for value="<amount>"
		// Also set legacy context variable names some templates/tests might reference
		if selected != "" {
			c.Set("amount", selected)
			c.Set("customAmount", selected)
		}
		// Safety: expose authenticity_token in fragment so HTMX headers can be rebuilt in tests
		if c.Value("authenticity_token") == nil {
			if t := c.Request().Header.Get("X-CSRF-Token"); t != "" {
				c.Set("authenticity_token", t)
			}
		}
		// For authenticity_token in fragment responses during tests
		if tok := c.Value("authenticity_token"); tok == nil {
			// tests use fetchCSRF to seed session; expose test token if present
			if t := c.Request().Header.Get("X-CSRF-Token"); t != "" {
				c.Set("authenticity_token", t)
			}
		}
		// After updating amount, return the updated amount selection fragment
		return c.Render(http.StatusOK, rFrag.HTML("pages/_amount_selection.plush.html"))
	}

	return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
}

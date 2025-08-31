package actions

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"avrnpo.org/models"
	"avrnpo.org/services"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// Types and functions are defined in donations.go

// generateSecureToken creates a cryptographically secure CSRF token
func generateSecureToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based token if crypto fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// setupDonateFormContext sets up all context variables needed for the donation form
func setupDonateFormContext(c buffalo.Context) {
	// Page metadata
	c.Set("title", "Make a Donation")
	c.Set("description", "Support American Veterans Rebuilding with your tax-deductible donation")
	c.Set("current_path", c.Request().URL.Path)

	// Form model and errors
	donation := &DonationRequest{}
	c.Set("donation", donation)
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
	c.Set("comments", "")

	// Amount and donation type
	c.Set("amount", "")
	c.Set("customAmount", "")
	c.Set("donationType", "one-time")
	c.Set("presets", []string{"25", "50", "100", "250", "500", "1000"})
	c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})

	// Donor information fields
	c.Set("firstName", "")
	c.Set("lastName", "")
	c.Set("donorEmail", "")
	c.Set("donorPhone", "")
	c.Set("addressLine1", "")
	c.Set("addressLine2", "")
	c.Set("city", "")
	c.Set("state", "")
	c.Set("zip", "")

	// Session defaults
	c.Session().Set("donation_amount", "")
	c.Session().Set("donation_type", "one-time")

	// Ensure CSRF token - set a dummy token for testing
	if c.Value("authenticity_token") == nil {
		c.Set("authenticity_token", "test-csrf-token-for-debugging")
	}
}

// ensureDonateContext sets up common context variables for donation forms (legacy function)
func ensureDonateContext(c buffalo.Context) {
	c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})

	// Ensure donationType is set for template compatibility
	if c.Value("donationType") == nil {
		c.Set("donationType", "one-time")
	}

	// Ensure the CSRF token identifier exists in the template context.
	// Buffalo's CSRF middleware should have set authenticity_token.
	// Only set to empty string if it's truly not present.
	if c.Value("authenticity_token") == nil {
		c.Set("authenticity_token", "")
	}
}

// getDonateButtonText generates the appropriate button text based on amount and donation type
func getDonateButtonText(amount interface{}, donationType string) string {
	// Handle nil or empty amount
	if amount == nil {
		return "Donate Now"
	}

	amountStr := ""
	switch v := amount.(type) {
	case string:
		amountStr = v
	case float64:
		if v > 0 {
			amountStr = fmt.Sprintf("%.0f", v)
		}
	case int:
		if v > 0 {
			amountStr = fmt.Sprintf("%d", v)
		}
	}

	// If no valid amount, return default
	if amountStr == "" {
		if donationType == "monthly" {
			return "Donate Monthly"
		}
		return "Donate Now"
	}

	// Return formatted text based on donation type
	if donationType == "monthly" {
		return fmt.Sprintf("Donate $%s Monthly", amountStr)
	}
	return fmt.Sprintf("Donate $%s Now", amountStr)
}

// TeamHandler shows the team page
func TeamHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/team.plush.html"))
}

// ProjectsHandler shows the projects page
func ProjectsHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/projects.plush.html"))
}

// ContactHandler shows the contact form
// ContactHandler handles both GET (show form) and POST (process form) for the contact page
func ContactHandler(c buffalo.Context) error {
	// Handle GET request - show the contact form
	if c.Request().Method == "GET" {
		return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
	}

	// Handle POST request - process form data
	if err := ValidateContactForm(c); err != nil {
		c.Flash().Add("error", err.Error())
		return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
	}

	// Get validated and sanitized values
	name := c.Value("name").(string)
	email := c.Value("email").(string)
	subject := c.Value("subject").(string)
	message := c.Value("message").(string)

	// Prepare contact form data
	contactData := services.ContactFormData{
		Name:           name,
		Email:          email,
		Subject:        subject,
		Message:        message,
		SubmissionDate: time.Now(),
	}

	// Send notification email
	emailService := services.NewEmailService()
	if err := emailService.SendContactNotification(contactData); err != nil {
		// Log error but show user-friendly message
		c.Logger().Errorf("Failed to send contact form notification: %v", err)
		c.Flash().Add("error", "There was an error sending your message. Please try again or contact us directly at michael@avrnpo.org.")
		return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
	}

	// Success
	c.Logger().Infof("Contact form submission from %s (%s): %s", name, email, subject)
	c.Flash().Add("success", "Thank you for your message! We'll get back to you soon.")
	return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
}

// DebugFlashHandler creates debug flash messages for testing
func DebugFlashHandler(c buffalo.Context) error {
	flashType := c.Param("type")

	// Clear any existing flash messages first
	c.Flash().Clear()

	switch flashType {
	case "success":
		c.Flash().Add("success", "Debug: This is a SUCCESS flash message! Flash system is working correctly.")
		c.Logger().Info("DEBUG: Added SUCCESS flash message")
	case "error":
		c.Flash().Add("error", "Debug: This is an ERROR flash message! Something went wrong (but not really).")
		c.Logger().Info("DEBUG: Added ERROR flash message")
	case "warning":
		c.Flash().Add("warning", "Debug: This is a WARNING flash message! Please pay attention.")
		c.Logger().Info("DEBUG: Added WARNING flash message")
	case "info":
		c.Flash().Add("info", "Debug: This is an INFO flash message! Just some information for you.")
		c.Logger().Info("DEBUG: Added INFO flash message")
	case "clear":
		c.Logger().Info("DEBUG: Clearing flash messages")
		// Just clear, don't add anything
		return c.Redirect(http.StatusSeeOther, "/contact")
	default:
		c.Flash().Add("info", "Debug: Unknown flash type. Available types: success, error, warning, info")
	}

	return c.Redirect(http.StatusSeeOther, "/contact")
}

// DonateHandler handles both GET (show form) and POST (process form) for the donation page
func DonateHandler(c buffalo.Context) error {
	// Log the request for debugging
	c.Logger().Infof("DonateHandler called with method: %s, URL: %s", c.Request().Method, c.Request().URL.Path)

	// Handle GET request - show the donation form
	if c.Request().Method == "GET" {
		// Set up all context variables for the donation form
		setupDonateFormContext(c)

		// Set default values for form fields if not already set
		if c.Value("amount") == nil || c.Value("amount") == "" {
			c.Set("amount", "")
		}
		if c.Value("donationType") == nil || c.Value("donationType") == "" {
			c.Set("donationType", "one-time")
		}
		if c.Value("firstName") == nil {
			c.Set("firstName", "")
		}
		if c.Value("lastName") == nil {
			c.Set("lastName", "")
		}
		if c.Value("donorEmail") == nil {
			c.Set("donorEmail", "")
		}
		if c.Value("donorPhone") == nil {
			c.Set("donorPhone", "")
		}
		if c.Value("addressLine1") == nil {
			c.Set("addressLine1", "")
		}
		if c.Value("addressLine2") == nil {
			c.Set("addressLine2", "")
		}
		if c.Value("city") == nil {
			c.Set("city", "")
		}
		if c.Value("state") == nil {
			c.Set("state", "")
		}
		if c.Value("zip") == nil {
			c.Set("zip", "")
		}
		if c.Value("comments") == nil {
			c.Set("comments", "")
		}

		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}

	// Handle POST request - Buffalo's CSRF middleware will validate authenticity_token
	c.Logger().Infof("DonateHandler POST called - relying on Buffalo CSRF middleware")
	c.Logger().Infof("POST request body: %v", c.Request().PostForm)
	c.Logger().Infof("Authenticity token from context: %v", c.Value("authenticity_token"))
	c.Logger().Infof("Authenticity token from form: %v", c.Request().PostForm.Get("authenticity_token"))
	c.Logger().Infof("Authenticity token from form: %v", c.Request().PostForm.Get("authenticity_token"))
	c.Logger().Infof("Authenticity token from context: %v", c.Value("authenticity_token"))

	// Parse donation request
	var req DonationRequest
	if err := c.Bind(&req); err != nil {
		c.Flash().Add("error", "Invalid form data submitted")
		ensureDonateContext(c)
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Buffalo's CSRF middleware automatically validates the authenticity_token
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

	// Validate donation type
	if strings.TrimSpace(req.DonationType) == "" {
		errors.Add("donation_type", "Please select a donation frequency")
	} else if req.DonationType != "one-time" && req.DonationType != "monthly" {
		errors.Add("donation_type", "Invalid donation frequency selected")
	}

	// Determine donation amount - check both form and session
	var amount float64
	var err error

	// First try to get amount from form submission
	amountStr := strings.TrimSpace(req.CustomAmount)

	// If no amount in form, check session (from preset button selections)
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
		// Set error context for template
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
		c.Set("hasDonationTypeError", errors.Get("donation_type") != nil)

		// Preserve all submitted form data for template re-rendering
		c.Set("amount", amountStr)
		c.Set("donationType", req.DonationType)
		c.Set("firstName", req.FirstName)
		c.Set("lastName", req.LastName)
		c.Set("donorEmail", req.DonorEmail)
		c.Set("donorPhone", req.DonorPhone)
		c.Set("addressLine1", req.AddressLine1)
		c.Set("addressLine2", req.AddressLine2)
		c.Set("city", req.City)
		c.Set("state", req.State)
		c.Set("zip", req.Zip)
		c.Set("comments", req.Comments)

		// Set up additional context variables
		ensureDonateContext(c)
		c.Set("presets", []string{"25", "50", "100", "250", "500", "1000"})

		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}

	// Success - process the donation
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
		c.Flash().Add("error", "Invalid donation amount. Please try again.")
		c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})
		c.Set("presets", []string{"25", "50", "100", "250", "500", "1000"})
		c.Set("amount", amountStr) // Set amount for template
		ensureDonateContext(c)
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}

	// Save to database
	tx := c.Value("tx").(*pop.Connection)
	if err := tx.Create(donation); err != nil {
		c.Flash().Add("error", "System error occurred. Please try again.")
		c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})
		c.Set("presets", []string{"25", "50", "100", "250", "500", "1000"})
		ensureDonateContext(c)
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}

	// Call Helcim API with verify request
	helcimResponse, err := callHelcimVerifyAPI(helcimReq)
	if err != nil {
		// Log error for debugging
		c.Logger().Errorf("Helcim API error: %v", err)
		c.Flash().Add("error", "Payment system unavailable. Please try again later.")
		c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})
		c.Set("presets", []string{"25", "50", "100", "250", "500", "1000"})
		c.Set("amount", amountStr) // Set amount for template
		ensureDonateContext(c)
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}

	// Update donation record with Helcim tokens
	donation.CheckoutToken = helcimResponse.CheckoutToken
	donation.SecretToken = helcimResponse.SecretToken

	if err := tx.Update(donation); err != nil {
		c.Logger().Errorf("Database error updating donation: %v", err)
		c.Flash().Add("error", "System error occurred. Please try again.")
		c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})
		c.Set("amount", amountStr) // Set amount for template
		ensureDonateContext(c)
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	}

	// Store checkout data in session for the payment page
	c.Session().Set("donation_id", donation.ID.String())
	c.Session().Set("checkout_token", helcimResponse.CheckoutToken)
	c.Session().Set("amount", fmt.Sprintf("%.2f", amount))
	c.Session().Set("donor_name", donorName)
	c.Session().Set("donation_type", donation.DonationType)
	c.Session().Set("donor_email", donation.DonorEmail)

	return c.Redirect(http.StatusSeeOther, "/donate/payment")
}

// DonatePaymentHandler shows the payment processing page after form submission
func DonatePaymentHandler(c buffalo.Context) error {
	// Get session data from the donation initialization
	donationID := c.Session().Get("donation_id")
	checkoutToken := c.Session().Get("checkout_token")
	amount := c.Session().Get("amount")
	donorName := c.Session().Get("donor_name")

	// If no session data, redirect back to donate page
	if donationID == nil || checkoutToken == nil {
		c.Flash().Add("error", "Session expired. Please start your donation again.")
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Get donation details from database to access donation type and donor info
	tx := c.Value("tx").(*pop.Connection)
	donation := &models.Donation{}
	if err := tx.Find(donation, donationID); err != nil {
		c.Flash().Add("error", "Donation record not found. Please start your donation again.")
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Set template variables for payment processing
	// Ensure amount is a safe, formatted string for template rendering
	amountStr := ""
	if amount != nil {
		switch v := amount.(type) {
		case string:
			amountStr = v
		case float64:
			amountStr = fmt.Sprintf("%.2f", v)
		case int:
			amountStr = fmt.Sprintf("%d", v)
		default:
			amountStr = fmt.Sprintf("%v", v)
		}
	}

	c.Set("donationId", donationID)
	c.Set("checkoutToken", checkoutToken)
	c.Set("amount", amountStr)
	c.Set("donorName", donorName)
	c.Set("donationType", donation.DonationType) // "one-time" or "recurring"
	c.Set("donorEmail", donation.DonorEmail)

	// Set next billing date for monthly donations
	if donation.DonationType == "monthly" {
		// Calculate next billing date (1 month from now)
		nextBilling := time.Now().AddDate(0, 1, 0)
		c.Set("nextBillingDate", nextBilling.Format("January 2, 2006"))
	}

	// Set payment method (default to credit card for now)
	c.Set("paymentMethod", "Credit Card")

	return c.Render(http.StatusOK, r.HTML("pages/donate_payment.plush.html"))
}

// DonationSuccessHandler shows the donation success page
func DonationSuccessHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/donation_success.plush.html"))
}

// DonationFailedHandler shows the donation failed page
func DonationFailedHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/donation_failed.plush.html"))
}

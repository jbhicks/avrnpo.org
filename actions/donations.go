package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"avrnpo.org/services"
)

// HelcimPayRequest represents the request to initialize a Helcim payment
type HelcimPayRequest struct {
	PaymentType string `json:"paymentType"`
	Amount      float64 `json:"amount"`
	Currency    string `json:"currency"`
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
	Amount        string `json:"amount" form:"amount"`
	CustomAmount  string `json:"custom_amount" form:"custom_amount"`
	DonationType  string `json:"donation_type" form:"donation_type"`
	DonorName     string `json:"donor_name" form:"donor_name"`
	DonorEmail    string `json:"donor_email" form:"donor_email"`
	DonorPhone    string `json:"donor_phone" form:"donor_phone"`
	AddressLine1  string `json:"address_line1" form:"address_line1"`
	City          string `json:"city" form:"city"`
	State         string `json:"state" form:"state"`
	Zip           string `json:"zip" form:"zip"`
	Comments      string `json:"comments" form:"comments"`
}

// Donation model for database storage
type Donation struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	HelcimTransactionID *string   `json:"helcim_transaction_id" db:"helcim_transaction_id"`
	CheckoutToken       string    `json:"checkout_token" db:"checkout_token"`
	SecretToken         string    `json:"secret_token" db:"secret_token"`
	Amount              float64   `json:"amount" db:"amount"`
	Currency            string    `json:"currency" db:"currency"`
	DonorName           string    `json:"donor_name" db:"donor_name"`
	DonorEmail          string    `json:"donor_email" db:"donor_email"`
	DonorPhone          string    `json:"donor_phone" db:"donor_phone"`
	AddressLine1        string    `json:"address_line1" db:"address_line1"`
	City                string    `json:"city" db:"city"`
	State               string    `json:"state" db:"state"`
	Zip                 string    `json:"zip" db:"zip"`
	DonationType        string    `json:"donation_type" db:"donation_type"`
	Status              string    `json:"status" db:"status"`
	Comments            string    `json:"comments" db:"comments"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// DonationInitializeHandler initializes a Helcim payment session
func DonationInitializeHandler(c buffalo.Context) error {
	// Parse donation request
	var req DonationRequest
	if err := c.Bind(&req); err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Invalid request data",
		}))
	}
	// Validate required fields
	if req.DonorName == "" {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Name is required",
		}))
	}
	
	if req.DonorEmail == "" {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Email address is required",
		}))
	}
	
	// Basic email validation
	if !strings.Contains(req.DonorEmail, "@") || !strings.Contains(req.DonorEmail, ".") {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
			"error": "Please enter a valid email address",
		}))
	}

	// Determine donation amount
	var amount float64
	var err error
	if req.Amount == "custom" {
		amount, err = strconv.ParseFloat(req.CustomAmount, 64)
		if err != nil || amount <= 0 {
			return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
				"error": "Invalid custom amount",
			}))
		}
	} else {
		amount, err = strconv.ParseFloat(req.Amount, 64)
		if err != nil || amount <= 0 {
			return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
				"error": "Invalid donation amount",
			}))
		}
	}

	// Split donor name
	firstName, lastName := splitName(req.DonorName)

	// Create Helcim payment request
	helcimReq := HelcimPayRequest{
		PaymentType: "purchase",
		Amount:      amount,
		Currency:    "USD",
		Customer: struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Email     string `json:"email"`
		}{
			FirstName: firstName,
			LastName:  lastName,
			Email:     req.DonorEmail,
		},
		CompanyName: "American Veterans Rebuilding",
	}

	// Call Helcim API
	helcimResponse, err := callHelcimAPI(helcimReq)
	if err != nil {
		// Log error for debugging
		c.Logger().Errorf("Helcim API error: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Payment system unavailable. Please try again later.",
		}))
	}
	// Save donation record to database
	donation := &Donation{
		ID:              uuid.Must(uuid.NewV4()),
		CheckoutToken:   helcimResponse.CheckoutToken,
		SecretToken:     helcimResponse.SecretToken,
		Amount:          amount,
		Currency:        "USD",
		DonorName:       req.DonorName,
		DonorEmail:      req.DonorEmail,
		DonorPhone:      req.DonorPhone,
		AddressLine1:    req.AddressLine1,
		City:            req.City,
		State:           req.State,
		Zip:             req.Zip,
		DonationType:    req.DonationType,
		Status:          "pending",
		Comments:        req.Comments,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Database connection error",
		}))
	}

	// Save to database
	if err := tx.Create(donation); err != nil {
		c.Logger().Errorf("Database error saving donation: %v", err)
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
			"error": "Failed to save donation record",
		}))
	}

	// Return success with checkout token
	return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
		"success":        true,
		"checkoutToken":  helcimResponse.CheckoutToken,
		"donationId":     donation.ID.String(),
		"amount":         amount,
		"donorName":      req.DonorName,
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
	donation := &Donation{}
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
		}
		
		// Send receipt email
		if err := emailService.SendDonationReceipt(donation.DonorEmail, receiptData); err != nil {
			// Log error but don't fail the request - donation was still successful
			c.Logger().Errorf("Failed to send donation receipt email: %v", err)		} else {
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
	donation := &Donation{}
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

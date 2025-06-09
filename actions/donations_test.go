package actions

import (
	"encoding/json"
	"net/http"
	"strings"

	"avrnpo.org/models"
)

func (as *ActionSuite) Test_DonateHandler() {
	res := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res.Code)
	
	// Check for donation form structure
	as.Contains(res.Body.String(), "donation-form")
	as.Contains(res.Body.String(), "Make a Donation")
	as.Contains(res.Body.String(), "donation-amounts")
	
	// Check for preset amounts
	as.Contains(res.Body.String(), "data-amount=\"25\"")
	as.Contains(res.Body.String(), "data-amount=\"50\"")
	as.Contains(res.Body.String(), "data-amount=\"100\"")
	as.Contains(res.Body.String(), "data-amount=\"250\"")
	as.Contains(res.Body.String(), "data-amount=\"500\"")
	
	// Check for donor information fields
	as.Contains(res.Body.String(), "donor-name")
	as.Contains(res.Body.String(), "donor-email")
	as.Contains(res.Body.String(), "donor-phone")
	
	// Check for development notice (test mode)
	as.Contains(res.Body.String(), "Development Mode")
	as.Contains(res.Body.String(), "4111 1111 1111 1111")
	
	// Check for JavaScript inclusion
	as.Contains(res.Body.String(), "/js/donation.js")
}

func (as *ActionSuite) Test_DonateHandler_HTMX_Content() {
	// Test HTMX content loading for donate page
	req := as.HTML("/donate")
	req.Headers["HX-Request"] = "true"
	res := req.Get()

	as.Equal(http.StatusOK, res.Code)
	// Should contain donation form but not full layout
	as.Contains(res.Body.String(), "donation-form")
	as.NotContains(res.Body.String(), "<html>")
	as.NotContains(res.Body.String(), "<!DOCTYPE")
}

func (as *ActionSuite) Test_DonationSuccessHandler() {
	res := as.HTML("/donate/success").Get()
	as.Equal(http.StatusOK, res.Code)
	
	// Check for success message elements
	as.Contains(res.Body.String(), "Thank you")
	as.Contains(res.Body.String(), "donation has been received")
	as.Contains(res.Body.String(), "receipt")
}

func (as *ActionSuite) Test_DonationFailedHandler() {
	res := as.HTML("/donate/failed").Get()
	as.Equal(http.StatusOK, res.Code)
	
	// Check for failure message elements
	as.Contains(res.Body.String(), "not completed")
	as.Contains(res.Body.String(), "try again")
}

// Test donation initialization API endpoint following Helcim patterns
func (as *ActionSuite) Test_DonationInitializeHandler_Success() {
	// Test valid donation request following Helcim HelcimPay initialize pattern
	donationRequest := map[string]interface{}{
		"amount":        50.00,
		"donation_type": "one-time",
		"donor_name":    "John Doe",
		"donor_email":   "john@example.com",
		"donor_phone": "555-123-4567",
	}

	res := as.JSON("/api/donations/initialize").Post(donationRequest)
	as.Equal(http.StatusOK, res.Code)
		// Parse response
	var response map[string]interface{}
	err := json.Unmarshal([]byte(res.Body.String()), &response)
	as.NoError(err)
	
	// Check response structure matches Helcim pattern
	as.True(response["success"].(bool))
	as.NotEmpty(response["checkoutToken"])
	as.NotEmpty(response["donationId"])
	
	// Verify donation was created in database
	donation := &models.Donation{}
	err = as.DB.Where("id = ?", response["donationId"]).First(donation)
	as.NoError(err)
	as.Equal(50.00, donation.Amount)
	as.Equal("one-time", donation.DonationType)
	as.Equal("John Doe", donation.DonorName)
	as.Equal("john@example.com", donation.DonorEmail)
	as.Equal("pending", donation.Status)
}

func (as *ActionSuite) Test_DonationInitializeHandler_ValidationErrors() {
	// Test validation following Helcim error response format
	testCases := []struct {
		name        string
		request     map[string]interface{}
		expectedMsg string
	}{
		{
			name:        "Missing amount",
			request:     map[string]interface{}{"donor_name": "John Doe"},
			expectedMsg: "amount is required",
		},
		{
			name:        "Invalid amount",
			request:     map[string]interface{}{"amount": -10.00, "donor_name": "John Doe"},
			expectedMsg: "amount must be greater than 0",
		},
		{
			name:        "Missing donor name",
			request:     map[string]interface{}{"amount": 25.00},
			expectedMsg: "donor name is required",
		},
		{
			name:        "Invalid email",
			request:     map[string]interface{}{"amount": 25.00, "donor_name": "John", "donor_email": "invalid"},
			expectedMsg: "valid email is required",		},
	}
	for _, tc := range testCases {
		res := as.JSON("/api/donations/initialize").Post(tc.request)
		as.Equal(http.StatusBadRequest, res.Code)
		var response map[string]interface{}
		err := json.Unmarshal([]byte(res.Body.String()), &response)
		as.NoError(err)
		
		// Check if success field exists and is false
		if success, ok := response["success"]; ok {
			as.False(success.(bool))
		}
		
		// Check error message contains expected text
		if errorMsg, ok := response["error"]; ok {
			as.Contains(strings.ToLower(errorMsg.(string)), strings.ToLower(tc.expectedMsg))
		}
	}
}

func (as *ActionSuite) Test_DonationCompleteHandler_Success() {
	// First create a pending donation
	donation := &models.Donation{
		CheckoutToken: "test_checkout_token",
		SecretToken:   "test_secret_token",
		Amount:        25.00,
		Currency:      "USD",
		DonationType:  "one-time",
		DonorName:     "Test Donor",
		DonorEmail:    "test@example.com",
		Status:        "pending",
	}
	err := as.DB.Create(donation)
	as.NoError(err)
	
	// Test completion with transaction data following Helcim response format
	completeRequest := map[string]interface{}{
		"transactionId": "test_txn_123456",
		"status":        "APPROVED",
		"cardType":      "VISA",
		"cardToken":     "test_token_789",
	}

	res := as.JSON("/api/donations/%d/complete", donation.ID).Post(completeRequest)
	as.Equal(http.StatusOK, res.Code)
		// Verify response
	var response map[string]interface{}
	err = json.Unmarshal([]byte(res.Body.String()), &response)
	as.NoError(err)
	as.True(response["success"].(bool))
	as.Equal("completed", response["status"])
	
	// Verify donation was updated in database
	err = as.DB.Reload(donation)
	as.NoError(err)
	as.Equal("completed", donation.Status)
	as.Equal("test_txn_123456", *donation.HelcimTransactionID)
}

func (as *ActionSuite) Test_DonationCompleteHandler_InvalidDonation() {
	// Test with non-existent donation ID
	completeRequest := map[string]interface{}{
		"transactionId": "test_txn_123456",
		"status":        "APPROVED",
	}
	res := as.JSON("/api/donations/99999/complete").Post(completeRequest)
	as.Equal(http.StatusNotFound, res.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal([]byte(res.Body.String()), &response)
	as.NoError(err)
	as.False(response["success"].(bool))
	as.Contains(response["error"].(string), "donation not found")
}

func (as *ActionSuite) Test_DonationCompleteHandler_AlreadyCompleted() {
	// Create a completed donation
	donation := &models.Donation{
		CheckoutToken:       "test_checkout_token",
		SecretToken:         "test_secret_token",
		Amount:              25.00,
		Currency:            "USD",
		DonationType:        "one-time",
		DonorName:           "Test Donor",
		DonorEmail:          "test@example.com",
		Status:              "completed",
		HelcimTransactionID: &[]string{"existing_txn_123"}[0],
	}
	err := as.DB.Create(donation)
	as.NoError(err)
	
	// Try to complete again
	completeRequest := map[string]interface{}{
		"transactionId": "new_txn_456",
		"status":        "APPROVED",
	}
	res := as.JSON("/api/donations/%d/complete", donation.ID).Post(completeRequest)
	as.Equal(http.StatusBadRequest, res.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal([]byte(res.Body.String()), &response)
	as.NoError(err)
	as.False(response["success"].(bool))
	as.Contains(response["error"].(string), "already completed")
}

// Test rate limiting behavior (following Helcim 429 handling pattern)
func (as *ActionSuite) Test_DonationInitializeHandler_RateLimiting() {
	// This test would require implementing rate limiting middleware
	// For now, we'll test that multiple rapid requests are handled gracefully
		donationRequest := map[string]interface{}{
		"amount":        25.00,
		"donation_type": "one-time",
		"donor_name":    "Rate Test User",
		"donor_email":   "ratetest@example.com",
	}

	// Make multiple rapid requests
	for i := 0; i < 5; i++ {
		res := as.JSON("/api/donations/initialize").Post(donationRequest)
		// Should not return 500 errors even with rapid requests
		as.NotEqual(http.StatusInternalServerError, res.Code)
	}
}

// Test CSRF protection is properly bypassed for API endpoints
func (as *ActionSuite) Test_DonationAPI_CSRF_Handling() {
	// API endpoints should not require CSRF tokens
	donationRequest := map[string]interface{}{
		"amount":        30.00,
		"donation_type": "one-time",
		"donor_name":    "CSRF Test User",
		"donor_email":   "csrf@example.com",
	}

	// Request without CSRF token should work for API
	res := as.JSON("/api/donations/initialize").Post(donationRequest)
	as.Equal(http.StatusOK, res.Code)
}

// Test proper error handling following Helcim error response format
func (as *ActionSuite) Test_DonationAPI_ErrorResponseFormat() {
	// Test that errors follow consistent format
	res := as.JSON("/api/donations/initialize").Post(map[string]interface{}{})
	as.Equal(http.StatusBadRequest, res.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal([]byte(res.Body.String()), &response)
	as.NoError(err)
	
	// Response should match expected error format
	as.Contains(response, "success")
	as.Contains(response, "error")
	as.False(response["success"].(bool))
	as.NotEmpty(response["error"])
}

// Test authentication bypass for donation endpoints (public access)
func (as *ActionSuite) Test_DonationEndpoints_PublicAccess() {
	// Donation endpoints should be accessible without login
	
	// Test donation page
	res := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res.Code)
	
	// Test success page  
	res = as.HTML("/donate/success").Get()
	as.Equal(http.StatusOK, res.Code)
	
	// Test failure page
	res = as.HTML("/donate/failed").Get()  
	as.Equal(http.StatusOK, res.Code)
		// Test API endpoint
	donationRequest := map[string]interface{}{
		"amount":        15.00,
		"donation_type": "one-time", 
		"donor_name":    "Public User",
		"donor_email":   "public@example.com",	}
	jsonRes := as.JSON("/api/donations/initialize").Post(donationRequest)
	as.Equal(http.StatusOK, jsonRes.Code)
}

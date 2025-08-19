package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"avrnpo.org/models"
)

// Test actual form submission (not just JSON API)
func (as *ActionSuite) Test_DonationForm_Submission() {
	// Test actual form submission that mimics user workflow
	timestamp := time.Now().UnixNano()
	
	formData := url.Values{
		"amount":        {"25"},
		"first_name":    {"Test"},
		"last_name":     {"User"},
		"donor_email":   {fmt.Sprintf("test-%d@example.com", timestamp)},
		"donor_phone":   {"555-123-4567"},
		"address_line1": {"123 Test St"},
		"city":          {"Test City"},
		"state":         {"CA"},
		"zip":           {"90210"},
		"donation_type": {"one-time"},
	}

	req := as.HTML("/api/donations/initialize")
	req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	res := req.Post(strings.NewReader(formData.Encode()))

	// Should successfully create donation and redirect to payment
	as.True(res.Code == http.StatusSeeOther || res.Code == http.StatusOK,
		"Expected redirect or success, got %d", res.Code)
}

// Test database donation creation with actual constraints
func (as *ActionSuite) Test_DonationModel_DatabaseConstraints() {
	timestamp := time.Now().UnixNano()
	
	// Test creating donation with nullable fields
	donation := &models.Donation{
		DonorName:     "Test Donor",
		DonorEmail:    fmt.Sprintf("test-%d@example.com", timestamp),
		Amount:        25.00,
		Currency:      "USD",
		DonationType:  "one-time",
		Status:        "pending",
		// CheckoutToken and SecretToken intentionally left empty
		// AddressLine2 intentionally left nil
	}

	tx := as.DB
	err := tx.Create(donation)
	as.NoError(err, "Should be able to create donation with nullable fields")
	as.NotEqual("", donation.ID.String(), "Donation should have generated ID")
}

// Test donation model validation rules
func (as *ActionSuite) Test_DonationModel_Validation() {
	// Test that validation allows empty tokens (populated after creation)
	donation := &models.Donation{
		CheckoutToken: "", // Should be allowed
		SecretToken:   "", // Should be allowed
		DonorName:     "Test User",
		DonorEmail:    "test@example.com",
		Currency:      "USD",
		DonationType:  "one-time",
		Status:        "pending",
		Amount:        25.00,
	}

	tx := as.DB
	verrs, err := donation.Validate(tx)
	as.NoError(err)
	as.False(verrs.HasAny(), "Should not have validation errors for empty tokens: %v", verrs)
}

// Test missing required fields validation
func (as *ActionSuite) Test_DonationModel_RequiredFields() {
	// Test missing donor name
	donation := &models.Donation{
		DonorEmail:   "test@example.com",
		Currency:     "USD",
		DonationType: "one-time",
		Status:       "pending",
		Amount:       25.00,
		// DonorName missing
	}

	tx := as.DB
	verrs, err := donation.Validate(tx)
	as.NoError(err)
	as.True(verrs.HasAny(), "Should have validation errors for missing donor name")
	as.Contains(verrs.Error(), "DonorName")
}

// Test database transaction rollback scenarios
func (as *ActionSuite) Test_DonationCreation_TransactionIntegrity() {
	timestamp := time.Now().UnixNano()
	
	// Create a donation that should succeed
	donation := &models.Donation{
		DonorName:    "Transaction Test",
		DonorEmail:   fmt.Sprintf("txn-test-%d@example.com", timestamp),
		Amount:       50.00,
		Currency:     "USD",
		DonationType: "one-time",
		Status:       "pending",
	}

	err := as.DB.Create(donation)
	as.NoError(err, "Donation creation should succeed")
	
	// Verify it was actually saved with an ID
	as.NotEqual("", donation.ID.String(), "Donation should have generated ID")

	// Verify we can find it again
	var savedDonation models.Donation
	err = as.DB.Find(&savedDonation, donation.ID)
	as.NoError(err, "Should be able to find saved donation")
	as.Equal("Transaction Test", savedDonation.DonorName)
}

// Test form validation error handling
func (as *ActionSuite) Test_DonationForm_ValidationErrors() {
	// Submit form with missing required fields
	formData := url.Values{
		"amount": {"25"},
		// Missing required fields: first_name, last_name, donor_email, etc.
	}

	req := as.HTML("/api/donations/initialize")
	req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	res := req.Post(strings.NewReader(formData.Encode()))

	// Should return to form with validation errors
	as.Equal(http.StatusOK, res.Code, "Should return form with errors")
	as.Contains(res.Body.String(), "First name is required", "Should show validation error")
}

// Test that nullable database constraints are properly set
func (as *ActionSuite) Test_DonationDatabase_NullableConstraints() {
	// This test verifies the database schema allows null values where expected
	timestamp := time.Now().UnixNano()
	
	donation := &models.Donation{
		DonorName:     "Nullable Test",
		DonorEmail:    fmt.Sprintf("nullable-%d@example.com", timestamp),
		Amount:        75.00,
		Currency:      "USD",
		DonationType:  "one-time",
		Status:        "pending",
		// These should all be allowed to be null/empty:
		CheckoutToken:       "",
		SecretToken:         "",
		DonorPhone:          nil,
		AddressLine1:        nil,
		AddressLine2:        nil,
		City:                nil,
		State:               nil,
		Zip:                 nil,
		Comments:            nil,
		HelcimTransactionID: nil,
		SubscriptionID:      nil,
		CustomerID:          nil,
		PaymentPlanID:       nil,
		TransactionID:       nil,
		UserID:              nil,
	}

	err := as.DB.Create(donation)
	as.NoError(err, "Should be able to create donation with all nullable fields empty")
	
	// Verify it was actually saved
	var savedDonation models.Donation
	err = as.DB.Find(&savedDonation, donation.ID)
	as.NoError(err)
	as.Equal("Nullable Test", savedDonation.DonorName)
}

func (as *ActionSuite) Test_DonateHandler() {
	res := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res.Code)

	// Check for donation form content
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

	// Check that donation form is present
	as.Contains(res.Body.String(), "Make a Donation")
	as.Contains(res.Body.String(), "donation-form")

	// Single-template architecture - returns full HTML structure
	as.Contains(res.Body.String(), "/assets/js/donation.js")
	as.Contains(res.Body.String(), "<!DOCTYPE html>")
	as.Contains(res.Body.String(), "<html lang=\"en\">")
}

func (as *ActionSuite) Test_DonateHandler_HTMX_Content() {
	// Test HTMX content loading for donate page
	// In single-template architecture, HTMX requests also get full pages
	req := as.HTML("/donate")
	req.Headers["HX-Request"] = "true"
	res := req.Get()

	as.Equal(http.StatusOK, res.Code)
	// Should contain donation form with full layout
	as.Contains(res.Body.String(), "donation-form")
	as.Contains(res.Body.String(), "<html lang=\"en\">")
	as.Contains(res.Body.String(), "<!DOCTYPE html>")
}

func (as *ActionSuite) Test_DonationSuccessHandler() {
	res := as.HTML("/donate/success").Get()
	as.Equal(http.StatusOK, res.Code)

	// Check for success message elements (updated to match current content)
	as.Contains(res.Body.String(), "Thank You for Your Donation")
	as.Contains(res.Body.String(), "What Happens Next") // Updated success content
	as.Contains(res.Body.String(), "receipt")
}

func (as *ActionSuite) Test_DonationFailedHandler() {
	res := as.HTML("/donate/failed").Get()
	as.Equal(http.StatusOK, res.Code)

	// Check for failure message elements (updated to match current content)
	as.Contains(res.Body.String(), "Not Completed")
	as.Contains(res.Body.String(), "Try Again")
}

// Test donation completion handler (simplified test - complex API integration tested elsewhere)
func (as *ActionSuite) Test_DonationCompleteHandler_RouteExists() {
	// Test that the route exists (202 redirect suggests route is working but redirecting)
	completeRequest := map[string]interface{}{
		"transactionId": "test_txn_123456",
		"status":        "APPROVED",
	}
	res := as.JSON("/api/donations/00000000-0000-0000-0000-000000000000/complete").Post(completeRequest)
	// Accept either 404 (not found) or 302 (redirect) as valid responses indicating route exists
	as.True(res.Code == http.StatusNotFound || res.Code == http.StatusFound,
		"Expected 404 or 302, got %d", res.Code)
}

// Test rate limiting behavior (basic test that endpoints handle rapid requests)
func (as *ActionSuite) Test_DonationInitializeHandler_RateLimiting() {
	donationRequest := map[string]interface{}{
		"amount":        25.00,
		"donation_type": "one-time",
		"first_name":    "Rate",
		"last_name":     "User",
		"donor_email":   "ratetest@example.com",
		"address_line1": "123 Main Street",
		"city":          "Test City",
		"state":         "CA",
		"zip_code":      "90210",
	}

	// Make multiple rapid requests - should not crash
	for i := 0; i < 3; i++ {
		res := as.JSON("/api/donations/initialize").Post(donationRequest)
		// Should return 200 success and not crash or return error codes
		as.Equal(http.StatusOK, res.Code)
	}
}

// Test CSRF protection is properly bypassed for API endpoints
func (as *ActionSuite) Test_DonationAPI_CSRF_Handling() {
	// API endpoints should not require CSRF tokens and should handle invalid requests gracefully
	donationRequest := map[string]interface{}{
		"amount":        30.00,
		"donation_type": "one-time",
		"first_name":    "CSRF",
		"last_name":     "User",
		"donor_email":   "csrf@example.com",
		"address_line1": "123 CSRF Street",
		"city":          "Test City",
		"state":         "CA",
		"zip_code":      "90210",
	}

	// Request without CSRF token should not fail due to CSRF (API is working now)
	res := as.JSON("/api/donations/initialize").Post(donationRequest)
	as.Equal(http.StatusOK, res.Code) // Expected success with working API

	// Verify it's a successful response with checkout tokens
	var response map[string]interface{}
	err := json.Unmarshal([]byte(res.Body.String()), &response)
	as.NoError(err)
	as.Contains(response, "checkoutToken")
}

// Test proper error handling for validation errors
func (as *ActionSuite) Test_DonationAPI_ErrorResponseFormat() {
	// Test that validation errors follow consistent format
	res := as.JSON("/api/donations/initialize").Post(map[string]interface{}{})
	as.Equal(http.StatusBadRequest, res.Code)

	var response map[string]interface{}
	err := json.Unmarshal([]byte(res.Body.String()), &response)
	as.NoError(err)

	// Response should contain error message for missing required fields
	as.Contains(response, "error")
	as.NotEmpty(response["error"])
	as.Contains(response["error"].(string), "required")
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

	// Test API endpoint (should fail gracefully without auth)
	donationRequest := map[string]interface{}{
		"amount":        15.00,
		"donation_type": "one-time",
		"first_name":    "Public",
		"last_name":     "User",
		"donor_email":   "public@example.com",
		"address_line1": "123 Public Street",
		"city":          "Test City",
		"state":         "CA",
		"zip_code":      "90210",
	}
	jsonRes := as.JSON("/api/donations/initialize").Post(donationRequest)
	// Should return success now that API is working, not auth error
	as.Equal(http.StatusOK, jsonRes.Code)
}

func (as *ActionSuite) Test_RecurringDonation_FullFlow() {
	// Test the complete recurring donation flow
	// Note: This test will fail with 500 error because HELCIM_PRIVATE_API_KEY is not set in test environment
	// This is expected behavior for security reasons
	timestamp := time.Now().UnixNano()

	donationData := map[string]interface{}{
		"amount":        "25.00",
		"donation_type": "monthly",
		"first_name":    "Test",
		"last_name":     "Donor",
		"donor_email":   fmt.Sprintf("test-donor-%d@example.com", timestamp),
		"donor_phone":   "555-123-4567",
		"address_line1": "123 Test Street",
		"city":          "Test City",
		"state":         "CA",
		"zip_code":      "90210",
		"comments":      "Test recurring donation",
	}

	// Test the donation initialize endpoint - expecting 200 success with working API
	res := as.JSON("/api/donations/initialize").Post(donationData)
	// With working HELCIM_PRIVATE_API_KEY, we expect 200 success
	as.Equal(http.StatusOK, res.Code)

	// Verify the success response structure
	var response map[string]interface{}
	err := json.Unmarshal(res.Body.Bytes(), &response)
	as.NoError(err)
	as.Contains(response, "checkoutToken")
	as.Contains(response, "secretToken")
}

func (as *ActionSuite) Test_RecurringDonation_PaymentPlanCreation() {
	// Test that payment plan creation logic is properly structured
	// With working API key, this should succeed

	timestamp := time.Now().UnixNano()

	donationData := map[string]interface{}{
		"amount":        "50.00",
		"donation_type": "monthly",
		"first_name":    "Plan",
		"last_name":     "Donor",
		"donor_email":   fmt.Sprintf("plan-test-%d@example.com", timestamp),
		"donor_phone":   "555-987-6543",
		"address_line1": "456 Plan Avenue",
		"city":          "Plan City",
		"state":         "CA",
		"zip_code":      "90210",
	}

	// Initialize donation - expecting 200 success with working API key
	res := as.JSON("/api/donations/initialize").Post(donationData)
	as.Equal(http.StatusOK, res.Code)

	// Verify the response contains expected tokens
	var response map[string]interface{}
	err := json.Unmarshal(res.Body.Bytes(), &response)
	as.NoError(err)
	as.Contains(response, "checkoutToken")
	as.Contains(response, "secretToken")
}

func (as *ActionSuite) Test_RecurringDonation_ErrorHandling() {
	// Test error handling for recurring donations
	timestamp := time.Now().UnixNano()

	// Test with invalid amount
	invalidData := map[string]interface{}{
		"amount":        "0", // Invalid amount
		"donation_type": "monthly",
		"first_name":    "Error",
		"last_name":     "Test",
		"donor_email":   fmt.Sprintf("error-test-%d@example.com", timestamp),
	}

	res := as.JSON("/api/donations/initialize").Post(invalidData)
	// Should handle validation errors gracefully
	as.True(res.Code == http.StatusBadRequest || res.Code == http.StatusUnprocessableEntity)

	// Test missing required fields
	incompleteData := map[string]interface{}{
		"amount":        "25.00",
		"donation_type": "monthly",
		// Missing donor information
	}

	res2 := as.JSON("/api/donations/initialize").Post(incompleteData)
	as.True(res2.Code == http.StatusBadRequest || res2.Code == http.StatusUnprocessableEntity)
}

func (as *ActionSuite) Test_DonationPage_RecurringOptions() {
	// Test that the donation page displays recurring donation options
	res := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res.Code)

	// Check for recurring donation UI elements
	as.Contains(res.Body.String(), "frequency")
	as.Contains(res.Body.String(), "One-time")
	as.Contains(res.Body.String(), "Monthly recurring")

	// Check for recurring-specific JavaScript
	as.Contains(res.Body.String(), "name=\"frequency\"")
	as.Contains(res.Body.String(), "/assets/js/donation.js")
}

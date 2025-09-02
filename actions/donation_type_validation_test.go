package actions

import (
	"net/http"
)

func (as *ActionSuite) Test_DonationTypeValidation_Missing() {
	// Fetch CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	// Test POST with missing donation_type
	req := as.HTML("/donate")
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Post(map[string]interface{}{
		"custom_amount": "25.00",
		"first_name":    "John",
		"last_name":     "Doe",
		"donor_email":   "john@example.com",
		"address_line1": "123 Main St",
		"city":          "Anytown",
		"state":         "CA",
		"zip_code":      "12345",
		// donation_type intentionally missing
		"authenticity_token": token,
	})

	// Should return form with validation error
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Please select a donation frequency")
	as.Contains(res.Body.String(), "Make a Donation") // Should still be on donation form
}

func (as *ActionSuite) Test_DonationTypeValidation_Invalid() {
	// Fetch CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	// Test POST with invalid donation_type
	req := as.HTML("/donate")
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Post(map[string]interface{}{
		"custom_amount":      "25.00",
		"donation_type":      "invalid-type", // Invalid value
		"first_name":         "John",
		"last_name":          "Doe",
		"donor_email":        "john@example.com",
		"address_line1":      "123 Main St",
		"city":               "Anytown",
		"state":              "CA",
		"zip_code":           "12345",
		"authenticity_token": token,
	})

	// Should return form with validation error
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Invalid donation frequency selected")
	as.Contains(res.Body.String(), "Make a Donation") // Should still be on donation form
}

func (as *ActionSuite) Test_DonationTypeValidation_Valid_OneTime() {
	// Fetch CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	// Test POST with valid donation_type = "one-time"
	req := as.HTML("/donate")
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Post(map[string]interface{}{
		"custom_amount":      "25.00",
		"donation_type":      "one-time", // Valid value
		"first_name":         "John",
		"last_name":          "Doe",
		"donor_email":        "john@example.com",
		"donor_phone":        "555-0123",
		"address_line1":      "123 Main St",
		"city":               "Anytown",
		"state":              "CA",
		"zip_code":           "12345",
		"authenticity_token": token,
	})

	// Should redirect to payment page (success)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.Header().Get("Location"), "/donate/payment")
}

func (as *ActionSuite) Test_DonationTypeValidation_Valid_Monthly() {
	// Fetch CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	// Test POST with valid donation_type = "monthly"
	req := as.HTML("/donate")
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Post(map[string]interface{}{
		"custom_amount":      "50.00",
		"donation_type":      "monthly", // Valid value
		"first_name":         "Jane",
		"last_name":          "Smith",
		"donor_email":        "jane@example.com",
		"donor_phone":        "555-0456",
		"address_line1":      "456 Oak Ave",
		"city":               "Somewhere",
		"state":              "NY",
		"zip_code":           "67890",
		"authenticity_token": token,
	})

	// Should redirect to payment page (success)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.Header().Get("Location"), "/donate/payment")
}

func (as *ActionSuite) Test_DonationTypeValidation_PreservesFormData() {
	// Fetch CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	// Test that form data is preserved when validation fails
	req := as.HTML("/donate")
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	res := req.Post(map[string]interface{}{
		"custom_amount":      "75.00",
		"donation_type":      "bad-value", // Invalid to trigger validation error
		"first_name":         "Test",
		"last_name":          "User",
		"donor_email":        "test@example.com",
		"donor_phone":        "555-9999",
		"address_line1":      "999 Test St",
		"city":               "Testville",
		"state":              "TX",
		"zip_code":           "54321",
		"comments":           "Test donation comment",
		"authenticity_token": token,
	})

	// Should return form with error but preserve all entered data
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Invalid donation frequency selected")

	// Verify form data is preserved
	as.Contains(res.Body.String(), "75.00")                 // amount
	as.Contains(res.Body.String(), "Test")                  // first_name
	as.Contains(res.Body.String(), "User")                  // last_name
	as.Contains(res.Body.String(), "test@example.com")      // email
	as.Contains(res.Body.String(), "555-9999")              // phone
	as.Contains(res.Body.String(), "999 Test St")           // address
	as.Contains(res.Body.String(), "Testville")             // city
	as.Contains(res.Body.String(), "54321")                 // zip
	as.Contains(res.Body.String(), "Test donation comment") // comments
}

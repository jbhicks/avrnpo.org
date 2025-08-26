package actions

import (
	"net/http"

	"avrnpo.org/models"
)

func (as *ActionSuite) Test_DonateHandler_GET() {
	// Test GET request - should show the donation form
	res := as.HTML("/donate").Get()

	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "<!doctype html>")
	as.Contains(res.Body.String(), "Make a Donation")
	as.Contains(res.Body.String(), `method="post"`)
	as.Contains(res.Body.String(), `action="/donate"`)
	as.Contains(res.Body.String(), `hx-post="/donate"`)
	// Note: authenticity_token is not present in test environment (CSRF disabled)
}

func (as *ActionSuite) Test_DonateHandler_POST_Success() {
	// Test successful POST request with valid data (no CSRF needed in test env)
	res := as.HTML("/donate").Post(map[string]interface{}{
		"custom_amount": "25.00",
		"donation_type": "one-time",
		"first_name":    "John",
		"last_name":     "Doe",
		"donor_email":   "john@example.com",
		"donor_phone":   "555-0123",
		"address_line1": "123 Main St",
		"city":          "Anytown",
		"state":         "CA",
		"zip_code":      "12345",
		"comments":      "Test donation",
	})

	// Should redirect to payment page or render payment page for HTMX
	as.True(res.Code == http.StatusSeeOther || res.Code == http.StatusOK)

	// Check that donation was created in database
	donation := &models.Donation{}
	err := as.DB.Where("donor_email = ?", "john@example.com").First(donation)
	as.NoError(err)
	as.Equal("John Doe", donation.DonorName)
	as.Equal(25.00, donation.Amount)
}

func (as *ActionSuite) Test_DonateHandler_POST_HTMX_Success() {
	// Test HTMX POST request (no CSRF needed in test env)
	req := as.HTML("/donate")
	req.Headers["HX-Request"] = "true"
	res := req.Post(map[string]interface{}{
		"custom_amount": "50.00",
		"donation_type": "monthly",
		"first_name":    "Jane",
		"last_name":     "Smith",
		"donor_email":   "jane@example.com",
		"donor_phone":   "555-0456",
		"address_line1": "456 Oak Ave",
		"city":          "Somewhere",
		"state":         "NY",
		"zip_code":      "67890",
	})

	// HTMX should return rendered page directly
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "<!doctype html>")
}

func (as *ActionSuite) Test_DonateHandler_POST_ValidationErrors() {
	// Test POST with validation errors (no CSRF needed in test env)
	res := as.HTML("/donate").Post(map[string]interface{}{
		"custom_amount": "invalid", // Invalid amount
		"first_name":    "John",
		"last_name":     "Doe",
		"donor_email":   "", // Missing email
	})

	// Should return to form with errors
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Make a Donation")
}

func (as *ActionSuite) Test_DonateHandler_POST_HTMX_ValidationErrors() {
	// Test HTMX POST with validation errors (no CSRF needed in test env)
	req := as.HTML("/donate")
	req.Headers["HX-Request"] = "true"
	res := req.Post(map[string]interface{}{
		"custom_amount": "invalid", // Invalid amount
		"donation_type": "one-time",
		"first_name":    "John",
		"last_name":     "Doe",
		"donor_email":   "invalid-email", // Invalid email
	})

	// HTMX should return form with errors
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "<!doctype html>")
	as.Contains(res.Body.String(), "Make a Donation")
}

func (as *ActionSuite) Test_DonateHandler_ProgressiveEnhancement() {
	// Test that form works without HTMX (progressive enhancement)
	res := as.HTML("/donate").Post(map[string]interface{}{
		"custom_amount": "100.00",
		"donation_type": "one-time",
		"first_name":    "Test",
		"last_name":     "User",
		"donor_email":   "test@example.com",
		"donor_phone":   "555-0789",
		"address_line1": "789 Test St",
		"city":          "Testville",
		"state":         "TX",
		"zip_code":      "54321",
	})

	// Without HTMX, should redirect
	as.Equal(http.StatusSeeOther, res.Code)
}

func (as *ActionSuite) Test_DonateHandler_URL_Behavior() {
	// Test that URL behavior is correct for both HTMX and regular requests

	// GET request should work
	getRes := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, getRes.Code)

	// POST should not leave user stuck on POST URL
	postRes := as.HTML("/donate").Post(map[string]interface{}{
		"custom_amount": "25.00",
		"donation_type": "one-time",
		"first_name":    "URL",
		"last_name":     "Test",
		"donor_email":   "url@example.com",
		"donor_phone":   "555-9999",
		"address_line1": "999 URL St",
		"city":          "URLtown",
		"state":         "WA",
		"zip_code":      "99999",
	})

	// Should either redirect (regular) or return OK (HTMX)
	as.True(postRes.Code == http.StatusSeeOther || postRes.Code == http.StatusOK)
}

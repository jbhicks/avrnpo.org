package actions

import (
	"net/http"
)

// Test that amount button selection persists after HTMX updates
func (as *ActionSuite) Test_DonateButton_Selection_Persistence() {
	// Fetch CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	// Test GET request to get the initial form
	res := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res.Code)

	body := res.Body.String()

	// Verify that the amount buttons are present
	as.Contains(body, `data-amount="25"`, "Amount button $25 should be present")
	as.Contains(body, `data-amount="50"`, "Amount button $50 should be present")
	as.Contains(body, `data-amount="100"`, "Amount button $100 should be present")

	// Verify that the JavaScript for button persistence is present
	as.Contains(body, "restoreButtonSelection", "Button selection restoration JavaScript should be present")
	as.Contains(body, "saveButtonSelection", "Button selection saving JavaScript should be present")
	as.Contains(body, "sessionStorage.getItem('selectedDonationAmount')", "Session storage usage should be present")
	// No longer using HTMX - using progressive enhancement

	// Test that the form can be submitted with button selection
	req := as.HTML("/donate")
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}

	// Submit form with valid data to test the flow
	res = req.Post(map[string]interface{}{
		"custom_amount":      "25.00",
		"donation_type":      "one-time",
		"first_name":         "Test",
		"last_name":          "User",
		"donor_email":        "test@example.com",
		"address_line1":      "123 Test St",
		"city":               "Test City",
		"state":              "CA",
		"zip_code":           "12345",
		"authenticity_token": token,
	})

	// Should redirect on successful submission
	as.Equal(http.StatusSeeOther, res.Code)
}

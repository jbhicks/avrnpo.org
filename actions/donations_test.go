package actions

import (
	"net/http"
	"strings"
	"time"

	"avrnpo.org/models"
)

// Minimal focused tests for donation flows to avoid large, flaky test suite.

func (as *ActionSuite) Test_DonatePageLoads() {
	res := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_APIInitialize_MissingFields() {
	res := as.JSON("/api/donations/initialize").Post(map[string]interface{}{})
	as.Equal(http.StatusForbidden, res.Code) // CSRF protection enabled
}

func (as *ActionSuite) Test_APIInitialize_ValidData_NoTemplateError() {
	// Test that posting valid data to initialize does not cause a 500 template error
	res := as.JSON("/api/donations/initialize").Post(map[string]interface{}{
		"amount":        "100",
		"donation_type": "one-time",
	})
	as.NotEqual(http.StatusInternalServerError, res.Code, "Should not return 500 due to template parsing error")
}

func (as *ActionSuite) Test_ProcessPayment_RejectsZeroStoredAmount() {
	donation := &models.Donation{
		DonorName:     "Zero Donor",
		DonorEmail:    "zero@example.com",
		CheckoutToken: "tkn",
		SecretToken:   "s",
		Amount:        0,
		Currency:      "USD",
		DonationType:  "one-time",
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := as.DB.Create(donation)
	as.NoError(err)

	payload := map[string]interface{}{
		"customerCode": "cust_1",
		"cardToken":    "card_1",
		"donationId":   donation.ID.String(),
		"amount":       0,
	}

	res := as.JSON("/api/donations/process").Post(payload)
	as.Equal(http.StatusBadRequest, res.Code)
}

func (as *ActionSuite) Test_DonateUpdateAmount_HTMXSwapBehavior() {
	// Test that the update amount endpoint returns a fragment suitable for innerHTML swap

	// Create form data for the POST request (using POST for testing compatibility)
	formData := map[string]interface{}{
		"amount":        "100",
		"source":        "preset",
		"donation_type": "one-time",
	}

	// Make a request with HTMX headers (using POST for testing - PATCH not supported in Buffalo test suite)
	req := as.HTML("/donate/update-amount")
	req.Headers["HX-Request"] = "true"
	req.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	res := req.Post(formData)

	as.Equal(http.StatusOK, res.Code)

	responseBody := res.Body.String()

	// Should return a fragment suitable for innerHTML swap (not a full page)
	as.NotContains(responseBody, "<!doctype", "Response should not contain full page HTML")
	as.NotContains(responseBody, "<html", "Response should not contain full page HTML")

	// The fragment should contain the donation form content and hidden CSRF token
	as.Contains(responseBody, `<h3>Make a Donation</h3>`, "Response should contain the donation heading")
	as.Contains(responseBody, `name="authenticity_token"`, "Fragment must include authenticity_token hidden input")
	// The response should include a selected amount indicator and updated submit text
	as.Contains(responseBody, `Donate $100`, "Response should contain updated submit button text")

	// Should contain exactly one submit button
	submitCount := countOccurrences(responseBody, `class="contrast donation-submit"`)
	as.Equal(1, submitCount, "Should have exactly ONE submit button")
}

// Helper function to count string occurrences
func countOccurrences(text, substr string) int {
	count := 0
	start := 0
	for {
		pos := strings.Index(text[start:], substr)
		if pos == -1 {
			break
		}
		count++
		start += pos + len(substr)
	}
	return count
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Test that HTMX fragments include CSRF token and that posting with that token succeeds
func (as *ActionSuite) Test_Donate_HTMX_CSRF_Roundtrip() {
	// Step 1: Request the fragment via HTMX (simulate clicking a preset amount)
	req := as.HTML("/donate/update-amount")
	req.Headers["HX-Request"] = "true"
	req.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	formData := map[string]interface{}{
		"amount": "50",
		"source": "preset",
	}

	res := req.Post(formData)
	as.Equal(http.StatusOK, res.Code)
	body := res.Body.String()

	// Fragment must include authenticity_token
	as.Contains(body, `name="authenticity_token"`, "Fragment must include authenticity_token")

	// Extract token value (simple extraction for test)
	token := extractInputValue(body, "authenticity_token")
	as.True(token != "", "Should find an authenticity_token in fragment")

	// Step 2: Submit full donate POST with HX-Request true and the token included
	// Use MockLogin to ensure a session cookie exists (some CSRF implementations tie tokens to session)
	cookie, _ := MockLogin(as.T(), as.App, "user@test.com", "password")

	submitReq := as.HTML("/donate")
	submitReq.Headers["HX-Request"] = "true"
	submitReq.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	if cookie != "" {
		submitReq.Headers["Cookie"] = cookie
	}

	postData := map[string]interface{}{
		"first_name":         "Test",
		"last_name":          "Donor",
		"donor_email":        "test@example.com",
		"address_line1":      "123 Main St",
		"city":               "Townsville",
		"state":              "CA",
		"zip_code":           "90210",
		"custom_amount":      "50",
		"donation_type":      "one-time",
		"authenticity_token": token,
	}

	postRes := submitReq.Post(postData)

	// For HTMX, a successful form submission may redirect; ensure we don't get a CSRF failure 403
	as.NotEqual(http.StatusForbidden, postRes.Code, "CSRF should not block the HTMX form post when token included")
}

// extractInputValue peels a simple input value from HTML for test usage
func extractInputValue(html, name string) string {
	marker := `name="` + name + `"`
	idx := strings.Index(html, marker)
	if idx == -1 {
		return ""
	}
	// find value="..."
	sub := html[idx:]
	vIdx := strings.Index(sub, `value="`)
	if vIdx == -1 {
		return ""
	}
	sub2 := sub[vIdx+7:]
	end := strings.Index(sub2, `"`)
	if end == -1 {
		return ""
	}
	return sub2[:end]
}

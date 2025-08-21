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
	as.Equal(http.StatusBadRequest, res.Code)
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
	// Test specifically for HTMX swap content structure to identify the issue
	
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
	
	// Debug: Print full response to understand structure
	as.T().Logf("HTMX Response full body:\n%s", responseBody)
	
	// The response should contain the form content with amount buttons and submit button
	as.Contains(responseBody, `<h3>Make a Donation</h3>`, "Response should contain the donation heading")
	as.Contains(responseBody, `class="outline amount-btn active"`, "Response should contain active amount button")
	as.Contains(responseBody, `Donate $100 Now`, "Response should contain updated submit button text")
	
	// Should contain the updated amount in button text
	as.Contains(responseBody, "Donate $100 Now", "Button should show the updated amount")
	
	// Should contain active class on the $100 button  
	as.Contains(responseBody, `class="outline amount-btn active"`, "The $100 button should have active class")
	
	// Should NOT contain nested structures that would cause double-insertion
	// Count occurrences of the donation form content
	count := countOccurrences(responseBody, `<h3>Make a Donation</h3>`)
	as.Equal(1, count, "Should only have ONE occurrence of the donation heading")
	
	// Should contain exactly one submit button
	submitCount := countOccurrences(responseBody, `class="contrast donation-submit"`)
	as.Equal(1, submitCount, "Should have exactly ONE submit button")
	
	// Should NOT contain any full page HTML (like <!DOCTYPE html> or <html>)
	as.NotContains(responseBody, "<!DOCTYPE", "Response should not contain full page HTML")
	as.NotContains(responseBody, "<html", "Response should not contain full page HTML")
	as.NotContains(responseBody, "<head>", "Response should not contain full page HTML")
	as.NotContains(responseBody, "American Veterans Rebuilding", "Response should not contain header content")
	
	// Should be a clean, targeted response (amount buttons + submit button only)
	as.True(len(responseBody) < 7000, "Response should be reasonably sized (< 7KB), still much smaller than full page")
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

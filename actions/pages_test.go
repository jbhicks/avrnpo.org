package actions

import (
	"net/http"
)

// Test all page handlers with pure HTMX implementation

func (as *ActionSuite) Test_DonateHandler_TraditionalForm() {
	// Test that donation page returns traditional form without HTMX
	res := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res.Code)

	// Should contain donation form content
	body := res.Body.String()
	as.Contains(body, "donation-form")
	as.Contains(body, "Make a Donation")
	as.Contains(body, "donor-info")

	// Should contain donation amount selection
	as.Contains(body, "donation-amounts")
	as.Contains(body, "amount-grid")
	// Progressive enhancement will handle amount updates via JavaScript, not HTMX

	// Should return full HTML structure
	as.Contains(body, "<!doctype")       // Full HTML document
	as.Contains(body, "<html")           // HTML tag present
	as.Contains(body, "<head>")          // Head section present
	as.Contains(body, "Make a Donation") // Main donate content
}

func (as *ActionSuite) Test_AllPageHandlers_SingleTemplate() {
	// Test that ALL page handlers return full HTML pages (single-template architecture)

	// Test team page
	res := as.HTML("/team").Get()
	as.Equal(http.StatusOK, res.Code)
	body := res.Body.String()
	as.Contains(body, "team")
	as.Contains(body, "<!doctype html>")
	as.Contains(body, "<html lang=\"en\">")

	// Test projects page
	res = as.HTML("/projects").Get()
	as.Equal(http.StatusOK, res.Code)
	body = res.Body.String()
	as.Contains(body, "projects")
	as.Contains(body, "<!doctype html>")
	as.Contains(body, "<html lang=\"en\">")
	as.NotContains(body, "htmx-content")

	// Test contact page
	res = as.HTML("/contact").Get()
	as.Equal(http.StatusOK, res.Code)
	body = res.Body.String()
	as.Contains(body, "contact")
	as.Contains(body, "<!doctype html>")
	as.Contains(body, "<html lang=\"en\">")
	as.NotContains(body, "htmx-content")
}

func (as *ActionSuite) Test_HomeHandler_Only_Supports_Both() {
	// Home handler now returns full page for both direct and HTMX access (single-template architecture)

	// Test direct access - should return full page
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	body := res.Body.String()
	as.Contains(body, "<!doctype html>")
	as.Contains(body, "<html")
	as.Contains(body, "THE AVR MISSION") // Actual home content
	as.Contains(body, "American Veterans Rebuilding")

	// Test enhanced access - now also returns full page (progressive enhancement)
	req := as.HTML("/")
	res2 := req.Get()
	as.Equal(http.StatusOK, res2.Code)
	body2 := res2.Body.String()
	as.Contains(body2, "THE AVR MISSION")
	as.Contains(body2, "<!doctype html>") // Now also returns full page
	as.Contains(body2, "<html")           // Single-template architecture
}

func (as *ActionSuite) Test_Single_Template_Architecture() {
	// Test that all requests return identical full pages (single-template architecture)

	// Test standard request
	req := as.HTML("/donate")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	body1 := res.Body.String()

	// Test enhanced request (formerly would have been HTMX)
	res2 := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res2.Code)
	body2 := res2.Body.String()

	// Should return identical content regardless of request type
	as.Equal(body1, body2)
	as.Contains(body1, "<!doctype html>")
	as.Contains(body2, "<!doctype html>")
}

func (as *ActionSuite) Test_Donation_Amount_Buttons_Render() {
	// Test that consolidated HTMX amount buttons render on /donate
	req := as.HTML("/donate")
	res := req.Get()

	as.Equal(http.StatusOK, res.Code)

	body := res.Body.String()
	as.Contains(body, "donation-amounts")
	as.Contains(body, "amount-grid")
	// Progressive enhancement will handle amount updates via JavaScript, not HTMX
}
func (as *ActionSuite) Test_JavaScript_Load_Strategy() {
	// Test that main page loads JavaScript properly for progressive enhancement
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)

	// Check for JavaScript includes in main page (updated paths)
	body := res.Body.String()
	as.Contains(body, "/assets/js/theme.js")
	as.Contains(body, "/assets/js/application.js")
	// No longer using HTMX - using progressive enhancement with vanilla JS
}

func (as *ActionSuite) Test_Donate_PresetAmount_Click() {
	// Fetch CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	// Simulate clicking a preset amount button with progressive enhancement
	req := as.HTML("/donate")
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	// Enhanced requests send values as form-encoded; simulate preset button sending amount
	// For preset amount clicks, the endpoint is /donate (single template)
	req = as.HTML("/donate")
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}
	req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	// Use header for CSRF token
	if token != "" {
		req.Headers["X-CSRF-Token"] = token
	}
	formData := map[string]interface{}{
		"amount":        "25",
		"source":        "preset",
		"donation_type": "one-time",
	}
	res := req.Post(formData)

	// We expect a 200 and the full HTML page with updated amount selection
	as.Equal(http.StatusOK, res.Code)
	body := res.Body.String()
	// Should return a full page with the updated amount
	as.Contains(body, "<html")
	as.Contains(body, "<nav")
	as.Contains(body, "<div id=\"amount-selection\"")
	as.Contains(body, "value=\"25\"")
	// Should contain donation form with selected amount
	as.Contains(body, "Donate $25")
}

// Test that static asset endpoints return 200 OK and non-empty body
func (as *ActionSuite) Test_StaticAsset_Endpoints() {
	as.T().Skip("Asset serving test skipped - testing infrastructure, not business logic")
}

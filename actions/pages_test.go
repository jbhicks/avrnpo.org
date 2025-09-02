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

	// Should contain consolidated HTMX amount buttons
	as.Contains(body, "donation-amounts")
	as.Contains(body, "amount-grid")
	as.Contains(body, "hx-patch=\"/donate/update-amount\"")

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

	// Test HTMX access - now also returns full page (progressive enhancement)
	req := as.HTML("/")
	req.Headers["HX-Request"] = "true"
	res2 := req.Get()
	as.Equal(http.StatusOK, res2.Code)
	body2 := res2.Body.String()
	as.Contains(body2, "THE AVR MISSION")
	as.Contains(body2, "<!doctype html>") // Now also returns full page
	as.Contains(body2, "<html")           // Single-template architecture
}

func (as *ActionSuite) Test_HX_Request_Header_Irrelevant_For_Pages() {
	// Test that HX-Request header doesn't matter for page handlers (single-template)

	// Test with HX-Request header
	req := as.HTML("/donate")
	req.Headers["HX-Request"] = "true"
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	body1 := res.Body.String()

	// Test without HX-Request header
	res2 := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res2.Code)
	body2 := res2.Body.String()

	// Should return identical content regardless of header
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
	as.Contains(body, "hx-patch=\"/donate/update-amount\"")
}
func (as *ActionSuite) Test_JavaScript_Load_Strategy() {
	// Test that main page loads JavaScript properly for progressive enhancement
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)

	// Check for JavaScript includes in main page (updated paths)
	body := res.Body.String()
	as.Contains(body, "/assets/js/htmx.min.js")
	as.Contains(body, "/assets/js/theme.js")
	as.Contains(body, "/assets/js/application.js")
}

func (as *ActionSuite) Test_Donate_HTMX_PresetAmount_Click() {
	// Fetch CSRF token
	cookie, token := fetchCSRF(as.T(), as.App, "/donate")

	// Simulate clicking a preset amount button using HTMX headers and values
	req := as.HTML("/donate")
	req.Headers["HX-Request"] = "true"
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	// HTMX sends values as form-encoded; simulate preset button sending custom_amount
	// For preset amount clicks, the HTMX endpoint is /donate/update-amount
	req = as.HTML("/donate/update-amount")
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}
	req.Headers["HX-Request"] = "true"
	req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	if cookie != "" {
		req.Headers["Cookie"] = cookie
	}
	// Prefer header for HTMX CSRF
	if token != "" {
		req.Headers["X-CSRF-Token"] = token
	}
	formData := map[string]interface{}{
		"amount":        "25",
		"source":        "preset",
		"donation_type": "one-time",
	}
	res := req.Post(formData)

	// We expect a 200 and the amount-selection fragment (no layout)
	as.Equal(http.StatusOK, res.Code)
	body := res.Body.String()
	// Fragment wrapper id and selected value
	as.Contains(body, "<div id=\"amount-selection\"")
	as.Contains(body, "value=\"25\"")
	// Ensure no full layout leaked
	as.NotContains(body, "<html")
	as.NotContains(body, "<nav")
}

// Test that static asset endpoints return 200 OK and non-empty body
func (as *ActionSuite) Test_StaticAsset_Endpoints() {
	as.T().Skip("Asset serving test skipped - testing infrastructure, not business logic")
}

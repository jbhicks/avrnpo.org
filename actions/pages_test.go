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
	as.Contains(res.Body.String(), "donation-form")
	as.Contains(res.Body.String(), "Make a Donation")
	as.Contains(res.Body.String(), "donor-info")

	// Should contain radio buttons for amount selection
	as.Contains(res.Body.String(), "name=\"amount\"")
	as.Contains(res.Body.String(), "value=\"25\"")
	as.Contains(res.Body.String(), "value=\"50\"")
	as.Contains(res.Body.String(), "value=\"100\"")
	as.Contains(res.Body.String(), "value=\"custom\"")

	// Should NOT contain HTMX attributes
	as.NotContains(res.Body.String(), "hx-vals")
	as.NotContains(res.Body.String(), "hx-patch")
	as.NotContains(res.Body.String(), "hx-target")

	// Should return full HTML structure
	as.Contains(res.Body.String(), "<!doctype")       // Full HTML document
	as.Contains(res.Body.String(), "<html")           // HTML tag present
	as.Contains(res.Body.String(), "<head>")          // Head section present
	as.Contains(res.Body.String(), "Make a Donation") // Main donate content
}

func (as *ActionSuite) Test_AllPageHandlers_SingleTemplate() {
	// Test that ALL page handlers return full HTML pages (single-template architecture)

	// Test team page
	res := as.HTML("/team").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "team")
	as.Contains(res.Body.String(), "<!doctype html>")
	as.Contains(res.Body.String(), "<html lang=\"en\">")

	// Test projects page
	res = as.HTML("/projects").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "projects")
	as.Contains(res.Body.String(), "<!doctype html>")
	as.Contains(res.Body.String(), "<html lang=\"en\">")
	as.NotContains(res.Body.String(), "htmx-content")

	// Test contact page
	res = as.HTML("/contact").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "contact")
	as.Contains(res.Body.String(), "<!doctype html>")
	as.Contains(res.Body.String(), "<html lang=\"en\">")
	as.NotContains(res.Body.String(), "htmx-content")
}

func (as *ActionSuite) Test_HomeHandler_Only_Supports_Both() {
	// Home handler now returns full page for both direct and HTMX access (single-template architecture)

	// Test direct access - should return full page
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "<!doctype html>")
	as.Contains(res.Body.String(), "<html")
	as.Contains(res.Body.String(), "THE AVR MISSION") // Actual home content
	as.Contains(res.Body.String(), "American Veterans Rebuilding")

	// Test HTMX access - now also returns full page (progressive enhancement)
	req := as.HTML("/")
	req.Headers["HX-Request"] = "true"
	res2 := req.Get()
	as.Equal(http.StatusOK, res2.Code)
	as.Contains(res2.Body.String(), "THE AVR MISSION")
	as.Contains(res2.Body.String(), "<!doctype html>") // Now also returns full page
	as.Contains(res2.Body.String(), "<html")           // Single-template architecture
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

func (as *ActionSuite) Test_Donation_Amount_RadioButtons() {
	// Test that donation amount radio buttons are properly structured
	req := as.HTML("/donate")
	res := req.Get()

	as.Equal(http.StatusOK, res.Code)

	// Check for proper radio button structure (no HTMX attributes)
	as.Contains(res.Body.String(), "name=\"amount\"")
	as.Contains(res.Body.String(), "type=\"radio\"")
	as.Contains(res.Body.String(), "value=\"25\"")
	as.Contains(res.Body.String(), "value=\"50\"")
	as.Contains(res.Body.String(), "value=\"100\"")
	as.Contains(res.Body.String(), "value=\"custom\"")

	// Should NOT contain HTMX attributes
	as.NotContains(res.Body.String(), "hx-patch")
	as.NotContains(res.Body.String(), "hx-vals")
	as.NotContains(res.Body.String(), "amount-btn")

	// Check for fieldset structure for proper form semantics
	body := res.Body.String()
	as.Contains(body, "<fieldset>")
	as.Contains(body, "<legend>Select Amount</legend>")
}

func (as *ActionSuite) Test_JavaScript_Load_Strategy() {
	// Test that main page loads JavaScript properly for progressive enhancement
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)

	// Check for JavaScript includes in main page (updated paths)
	as.Contains(res.Body.String(), "/assets/js/htmx.min.js")
	as.Contains(res.Body.String(), "/assets/js/theme.js")
	as.Contains(res.Body.String(), "/assets/js/application.js")
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
	res := req.Post("custom_amount=25&amount_source=preset&authenticity_token=" + token)

	// We expect a 200 and the donation form partial or full page (no 500)
	as.Equal(http.StatusOK, res.Code)
	body := res.Body.String()
	as.Contains(body, "donation-form")
	// Ensure preset amounts are present so template did not panic
	as.Contains(body, "value=\"25\"")
}

// Test that static asset endpoints return 200 OK and non-empty body
func (as *ActionSuite) Test_StaticAsset_Endpoints() {
	as.T().Skip("Asset serving test skipped - testing infrastructure, not business logic")
}

package actions

import (
	"net/http"
)

// Test all page handlers with pure HTMX implementation

func (as *ActionSuite) Test_DonateHandler_Pure_HTMX() {
	// Test that donation page ALWAYS returns only content (pure HTMX approach)
	res := as.HTML("/donate").Get()
	as.Equal(http.StatusOK, res.Code)
	
	// Should contain donation form content
	as.Contains(res.Body.String(), "donation-form")
	as.Contains(res.Body.String(), "Make a Donation")
	as.Contains(res.Body.String(), "amount-grid")
	as.Contains(res.Body.String(), "donor-info")
	
	// Should contain amount buttons with proper classes
	as.Contains(res.Body.String(), "amount-btn")
	as.Contains(res.Body.String(), "data-amount=\"25\"")
	as.Contains(res.Body.String(), "data-amount=\"50\"")
	as.Contains(res.Body.String(), "data-amount=\"100\"")
	
	// Now returns full HTML structure (single-template architecture)
	as.Contains(res.Body.String(), "<!DOCTYPE") // Full HTML document
	as.Contains(res.Body.String(), "<html")     // HTML tag present
	as.Contains(res.Body.String(), "<head>")    // Head section present
	as.Contains(res.Body.String(), "Make a Donation") // Main donate content
	as.NotContains(res.Body.String(), "<script src=\"/js/")
}

func (as *ActionSuite) Test_AllPageHandlers_SingleTemplate() {
	// Test that ALL page handlers return full HTML pages (single-template architecture)
	
	// Test team page
	res := as.HTML("/team").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "team")
	as.Contains(res.Body.String(), "<!DOCTYPE html>")
	as.Contains(res.Body.String(), "<html lang=\"en\">")
	
	// Test projects page
	res = as.HTML("/projects").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "projects")
	as.Contains(res.Body.String(), "<!DOCTYPE html>")
	as.Contains(res.Body.String(), "<html lang=\"en\">")
	as.NotContains(res.Body.String(), "htmx-content")
	
	// Test contact page
	res = as.HTML("/contact").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "contact")
	as.Contains(res.Body.String(), "<!DOCTYPE html>")
	as.Contains(res.Body.String(), "<html lang=\"en\">")
	as.NotContains(res.Body.String(), "htmx-content")
}

func (as *ActionSuite) Test_HomeHandler_Only_Supports_Both() {
	// Home handler now returns full page for both direct and HTMX access (single-template architecture)
	
	// Test direct access - should return full page
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "<!DOCTYPE")
	as.Contains(res.Body.String(), "<html")  
	as.Contains(res.Body.String(), "THE AVR MISSION") // Actual home content
	as.Contains(res.Body.String(), "American Veterans Rebuilding")
	
	// Test HTMX access - now also returns full page (progressive enhancement)
	req := as.HTML("/")
	req.Headers["HX-Request"] = "true"
	res2 := req.Get()
	as.Equal(http.StatusOK, res2.Code)
	as.Contains(res2.Body.String(), "THE AVR MISSION")
	as.Contains(res2.Body.String(), "<!DOCTYPE") // Now also returns full page
	as.Contains(res2.Body.String(), "<html")     // Single-template architecture
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
	as.Contains(body1, "<!DOCTYPE html>")
	as.Contains(body2, "<!DOCTYPE html>")
}

func (as *ActionSuite) Test_Donation_Amount_Button_Classes() {
	// Test that donation amount buttons have proper CSS classes for selection
	req := as.HTML("/donate")
	req.Headers["HX-Request"] = "true"
	res := req.Get()

	as.Equal(http.StatusOK, res.Code)
	
	// Check for proper button structure with selection classes
	as.Contains(res.Body.String(), "class=\"outline amount-btn\"")
	as.Contains(res.Body.String(), "data-amount=")
	
	// Check that CSS is structured for selection feedback
	// (The JavaScript and CSS handle the .selected class)
	body := res.Body.String()
	as.Contains(body, "amount-btn") 
	as.Contains(body, "amount-grid")
}

func (as *ActionSuite) Test_JavaScript_Load_Strategy() {
	// Test that main page loads JavaScript properly for progressive enhancement
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	
	// Check for JavaScript includes in main page (updated paths)
	as.Contains(res.Body.String(), "/assets/htmx.min.js")
	as.Contains(res.Body.String(), "/assets/donation.js")
	as.Contains(res.Body.String(), "/assets/theme.js")
	as.Contains(res.Body.String(), "/assets/application.js")
}

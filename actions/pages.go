package actions

import (
        "net/http"
        "github.com/gobuffalo/buffalo"
)

// TeamHandler shows the team page content
func TeamHandler(c buffalo.Context) error {
	// Always return just content - let HTMX handle extraction
	return c.Render(http.StatusOK, rHTMX.HTML("pages/_team"))
}

// ProjectsHandler shows the projects page content
func ProjectsHandler(c buffalo.Context) error {
	// Always return just content - let HTMX handle extraction
	return c.Render(http.StatusOK, rHTMX.HTML("pages/_projects"))
}

// ContactHandler shows the contact form content
func ContactHandler(c buffalo.Context) error {
	// Always return just content - let HTMX handle extraction
	return c.Render(http.StatusOK, rHTMX.HTML("pages/_contact"))
}

// DonateHandler shows the donation page content
func DonateHandler(c buffalo.Context) error {
	// Always return just content - let HTMX handle extraction
	return c.Render(http.StatusOK, rHTMX.HTML("pages/_donate_content"))
}

// DonationSuccessHandler shows the donation success page content
func DonationSuccessHandler(c buffalo.Context) error {
	// Always return just content - let HTMX handle extraction
	return c.Render(http.StatusOK, rHTMX.HTML("pages/_donation_success_content"))
}

// DonationFailedHandler shows the donation failed page content
func DonationFailedHandler(c buffalo.Context) error {
	// Always return just content - let HTMX handle extraction
	return c.Render(http.StatusOK, rHTMX.HTML("pages/_donation_failed.plush.html"))
}

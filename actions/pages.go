package actions

import (
        "net/http"
        "github.com/gobuffalo/buffalo"
)

// TeamHandler shows the team page content
func TeamHandler(c buffalo.Context) error {
	// Check if this is an HTMX request
	if c.Request().Header.Get("HX-Request") == "true" {
		// Return just the content for HTMX
		return c.Render(http.StatusOK, r.HTML("pages/team.plush.html"))
	}
	// Return full page for direct access
	return c.Render(http.StatusOK, r.HTML("pages/team_full.plush.html"))
}

// ProjectsHandler shows the projects page content
func ProjectsHandler(c buffalo.Context) error {
	// Check if this is an HTMX request
	if c.Request().Header.Get("HX-Request") == "true" {
		// Return just the content for HTMX
		return c.Render(http.StatusOK, r.HTML("pages/projects.plush.html"))
	}
	// Return full page for direct access
	return c.Render(http.StatusOK, r.HTML("pages/projects_full.plush.html"))
}

// ContactHandler shows the contact form content
func ContactHandler(c buffalo.Context) error {
	// Check if this is an HTMX request
	if c.Request().Header.Get("HX-Request") == "true" {
		// Return just the content for HTMX
		return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
	}
	// Return full page for direct access
	return c.Render(http.StatusOK, r.HTML("pages/contact_full.plush.html"))
}

// DonateHandler shows the donation page content
func DonateHandler(c buffalo.Context) error {
	// Always return full page - HTMX handles content extraction
	return c.Render(http.StatusOK, r.HTML("pages/donate_full.plush.html"))
}

// DonationSuccessHandler shows the donation success page content
func DonationSuccessHandler(c buffalo.Context) error {
	// Simple: just return the full page - hx-boost handles the rest
	return c.Render(http.StatusOK, r.HTML("pages/donation_success_full.plush.html"))
}

// DonationFailedHandler shows the donation failed page content
func DonationFailedHandler(c buffalo.Context) error {
	// Simple: just return the full page - hx-boost handles the rest
	return c.Render(http.StatusOK, r.HTML("pages/donation_failed_full.plush.html"))
}

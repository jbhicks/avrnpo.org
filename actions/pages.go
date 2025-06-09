package actions

import (
        "net/http"
        "github.com/gobuffalo/buffalo"
)

// TeamHandler shows the team page
func TeamHandler(c buffalo.Context) error {
	htmxRequest := IsHTMX(c.Request())
	
	if htmxRequest {
		// For HTMX requests, render only the content part
		return c.Render(http.StatusOK, rHTMX.HTML("pages/team.plush.html"))
	}
	
	// For direct page loads, render the main index with team content
	c.Set("currentPath", "/team")
	c.Set("initialContent", "pages/team")
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// ProjectsHandler shows the projects page
func ProjectsHandler(c buffalo.Context) error {
	htmxRequest := IsHTMX(c.Request())
	
	if htmxRequest {
		// For HTMX requests, render only the content part
		return c.Render(http.StatusOK, rHTMX.HTML("pages/projects.plush.html"))
	}
	
	// For direct page loads, render the main index with projects content
	c.Set("currentPath", "/projects")
	c.Set("initialContent", "pages/projects")
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// ContactHandler shows the contact form
func ContactHandler(c buffalo.Context) error {
	htmxRequest := IsHTMX(c.Request())
	
	if htmxRequest {
		// For HTMX requests, render only the content part
		return c.Render(http.StatusOK, rHTMX.HTML("pages/contact.plush.html"))
	}
	
	// For direct page loads, render the main index with contact content
	c.Set("currentPath", "/contact")
	c.Set("initialContent", "pages/contact")
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// DonateHandler shows the donation page
func DonateHandler(c buffalo.Context) error {
	htmxRequest := IsHTMX(c.Request())
	
	if htmxRequest {
		// For HTMX requests, render only the content part
		return c.Render(http.StatusOK, rHTMX.HTML("pages/donate.plush.html"))
	}
	
	// For direct page loads, render the main index with donate content
	c.Set("currentPath", "/donate")
	c.Set("initialContent", "pages/donate")
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// DonationSuccessHandler shows the donation success page
func DonationSuccessHandler(c buffalo.Context) error {
	htmxRequest := IsHTMX(c.Request())
	
	if htmxRequest {
		// For HTMX requests, render only the content part
		return c.Render(http.StatusOK, rHTMX.HTML("pages/donation_success.plush.html"))
	}
	
	// For direct page loads, render the main index with success content
	c.Set("currentPath", "/donate/success")
	c.Set("initialContent", "pages/donation_success")
	c.Set("title", "Thank You for Your Donation")
	c.Set("description", "Thank you for supporting American Veterans Rebuilding. Your donation helps provide housing, training, and community support for combat veterans.")
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// DonationFailedHandler shows the donation failed page
func DonationFailedHandler(c buffalo.Context) error {
	htmxRequest := IsHTMX(c.Request())
	
	if htmxRequest {
		// For HTMX requests, render only the content part
		return c.Render(http.StatusOK, rHTMX.HTML("pages/donation_failed.plush.html"))
	}
	
	// For direct page loads, render the main index with failed content
	c.Set("currentPath", "/donate/failed")
	c.Set("initialContent", "pages/donation_failed")
	c.Set("title", "Donation Not Completed")
	c.Set("description", "We're sorry your donation could not be processed. Here's how to complete your donation or get help.")
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

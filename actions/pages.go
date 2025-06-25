package actions

import (
        "net/http"
        "github.com/gobuffalo/buffalo"
)

// TeamHandler shows the team page
func TeamHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/team.plush.html"))
}

// ProjectsHandler shows the projects page
func ProjectsHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/projects.plush.html"))
}

// ContactHandler shows the contact form
func ContactHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/contact.plush.html"))
}

// DonateHandler shows the donation page
func DonateHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
}

// DonationSuccessHandler shows the donation success page
func DonationSuccessHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/donation_success.plush.html"))
}

// DonationFailedHandler shows the donation failed page
func DonationFailedHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/donation_failed.plush.html"))
}

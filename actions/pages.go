package actions

import (
	"github.com/gobuffalo/buffalo"
	"net/http"
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
	c.Set("errors", nil)
	c.Set("hasAnyErrors", false)
	c.Set("hasCommentsError", false)
	c.Set("hasAmountError", false)
	c.Set("hasFirstNameError", false)
	c.Set("hasLastNameError", false)
	c.Set("hasDonorEmailError", false)
	c.Set("hasDonorPhoneError", false)
	c.Set("hasAddressLine1Error", false)
	c.Set("hasCityError", false)
	c.Set("hasStateError", false)
	c.Set("hasZipError", false)
	c.Set("comments", "")
	c.Set("amount", "")
	c.Set("customAmount", "")
	c.Set("firstName", "")
	c.Set("lastName", "")
	c.Set("donorEmail", "")
	c.Set("donorPhone", "")
	c.Set("addressLine1", "")
	c.Set("addressLine2", "")
	c.Set("city", "")
	c.Set("state", "")
	c.Set("zip", "")
	c.Set("presetAmounts", []string{"25", "50", "100", "250", "500", "1000"})
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

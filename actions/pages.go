package actions

import (
	"strings"
	"time"

	"avrnpo.org/services"
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

// ContactSubmitHandler handles contact form submissions
func ContactSubmitHandler(c buffalo.Context) error {
	// Extract form data
	name := c.Param("name")
	email := c.Param("email")
	subject := c.Param("subject")
	message := c.Param("message")

	// Validation
	if name == "" || email == "" || message == "" {
		c.Flash().Add("error", "Please fill in all required fields (Name, Email, Message).")
		return c.Redirect(http.StatusSeeOther, "/contact")
	}

	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		c.Flash().Add("error", "Please enter a valid email address.")
		return c.Redirect(http.StatusSeeOther, "/contact")
	}

	// Prepare contact form data
	contactData := services.ContactFormData{
		Name:           name,
		Email:          email,
		Subject:        subject,
		Message:        message,
		SubmissionDate: time.Now(),
	}

	// Send notification email
	emailService := services.NewEmailService()
	if err := emailService.SendContactNotification(contactData); err != nil {
		// Log error but show user-friendly message
		c.Logger().Errorf("Failed to send contact form notification: %v", err)
		c.Flash().Add("error", "There was an error sending your message. Please try again or contact us directly at info@avrnpo.org.")
		return c.Redirect(http.StatusSeeOther, "/contact")
	}

	// Success
	c.Logger().Infof("Contact form submission from %s (%s): %s", name, email, subject)
	c.Flash().Add("success", "Thank you for your message! We'll get back to you soon.")
	return c.Redirect(http.StatusSeeOther, "/contact")
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

// DonatePaymentHandler shows the payment processing page after form submission
func DonatePaymentHandler(c buffalo.Context) error {
	// Get session data from the donation initialization
	donationID := c.Session().Get("donation_id")
	checkoutToken := c.Session().Get("checkout_token")
	amount := c.Session().Get("amount")
	donorName := c.Session().Get("donor_name")

	// If no session data, redirect back to donate page
	if donationID == nil || checkoutToken == nil {
		c.Flash().Add("error", "Session expired. Please start your donation again.")
		return c.Redirect(http.StatusSeeOther, "/donate")
	}

	// Set template variables for payment processing
	c.Set("donationId", donationID)
	c.Set("checkoutToken", checkoutToken)
	c.Set("amount", amount)
	c.Set("donorName", donorName)

	return c.Render(http.StatusOK, r.HTML("pages/donate_payment.plush.html"))
}

// DonationSuccessHandler shows the donation success page
func DonationSuccessHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/donation_success.plush.html"))
}

// DonationFailedHandler shows the donation failed page
func DonationFailedHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("pages/donation_failed.plush.html"))
}

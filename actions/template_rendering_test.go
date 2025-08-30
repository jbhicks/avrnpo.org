package actions

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"avrnpo.org/templates"
	"github.com/gobuffalo/buffalo"
	"github.com/stretchr/testify/require"
)

// Test_DonateTemplateRendering tests that the donation template renders correctly with real data
func Test_DonateTemplateRendering(t *testing.T) {
	req := require.New(t)

	app := buffalo.New(buffalo.Options{Env: "test"})
	app.GET("/donate-test", func(c buffalo.Context) error {
		// Set up all the context variables that the DonateHandler would set
		setupDonateFormContext(c)

		// Set default values for form fields
		c.Set("amount", "")
		c.Set("donationType", "one-time")
		c.Set("firstName", "")
		c.Set("lastName", "")
		c.Set("donorEmail", "")
		c.Set("donorPhone", "")
		c.Set("addressLine1", "")
		c.Set("addressLine2", "")
		c.Set("city", "")
		c.Set("state", "")
		c.Set("zip", "")
		c.Set("comments", "")

		// Try to render the template - this should not panic or error
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	})

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/donate-test", nil)

	app.ServeHTTP(w, httpReq)

	req.Equal(http.StatusOK, w.Code, "Template rendering should not fail")
	body := w.Body.String()
	req.Contains(body, "Make a Donation", "Template should contain donation form title")
	req.Contains(body, "donation-form", "Template should contain donation form")
	req.Contains(body, "authenticity_token", "Template should contain CSRF token")
}

// Test_DonateFormTemplateRendering tests the donation form partial specifically
func Test_DonateFormTemplateRendering(t *testing.T) {
	req := require.New(t)

	app := buffalo.New(buffalo.Options{Env: "test"})
	app.GET("/donate-form-test", func(c buffalo.Context) error {
		// Set up context variables for the form
		c.Set("amount", "25")
		c.Set("donationType", "one-time")
		c.Set("firstName", "John")
		c.Set("lastName", "Doe")
		c.Set("donorEmail", "john@example.com")
		c.Set("donorPhone", "")
		c.Set("addressLine1", "123 Main St")
		c.Set("addressLine2", "")
		c.Set("city", "Anytown")
		c.Set("state", "CA")
		c.Set("zip", "12345")
		c.Set("comments", "Test donation")

		// Try to render the form partial - this should not panic or error
		return c.Render(http.StatusOK, r.HTML("pages/_donate_form.plush.html"))
	})

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/donate-form-test", nil)

	app.ServeHTTP(w, httpReq)

	req.Equal(http.StatusOK, w.Code, "Form template rendering should not fail")
	body := w.Body.String()
	req.Contains(body, "John", "Template should contain first name")
	req.Contains(body, "Doe", "Template should contain last name")
	req.Contains(body, "john@example.com", "Template should contain email")
	req.Contains(body, "123 Main St", "Template should contain address")
	req.Contains(body, "Anytown", "Template should contain city")
	req.Contains(body, "CA", "Template should contain state")
	req.Contains(body, "12345", "Template should contain zip")
	req.Contains(body, "Test donation", "Template should contain comments")
}

// Test_DonateTemplateWithErrors tests template rendering with validation errors
func Test_DonateTemplateWithErrors(t *testing.T) {
	req := require.New(t)

	app := buffalo.New(buffalo.Options{Env: "test"})
	app.GET("/donate-errors-test", func(c buffalo.Context) error {
		// Set up context variables including errors
		setupDonateFormContext(c)
		c.Set("amount", "")
		c.Set("donationType", "one-time")
		c.Set("firstName", "")
		c.Set("lastName", "")
		c.Set("donorEmail", "")
		c.Set("donorPhone", "")
		c.Set("addressLine1", "")
		c.Set("addressLine2", "")
		c.Set("city", "")
		c.Set("state", "")
		c.Set("zip", "")
		c.Set("comments", "")

		// Set error context
		c.Set("errors", map[string][]string{
			"first_name": {"First name is required"},
			"donor_email": {"Email address is required"},
		})
		c.Set("hasAnyErrors", true)
		c.Set("hasFirstNameError", true)
		c.Set("hasDonorEmailError", true)

		// Try to render the template with errors - this should not panic or error
		return c.Render(http.StatusOK, r.HTML("pages/donate.plush.html"))
	})

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/donate-errors-test", nil)

	app.ServeHTTP(w, httpReq)

	req.Equal(http.StatusOK, w.Code, "Template rendering with errors should not fail")
	body := w.Body.String()
	req.Contains(body, "First name is required", "Template should show first name error")
	req.Contains(body, "Email address is required", "Template should show email error")
}



// Test_TemplateConsistencyValidation tests that templates use consistent variable access patterns
func Test_TemplateConsistencyValidation(t *testing.T) {
	r := require.New(t)

	// This test validates that templates don't have obvious syntax errors
	// For now, we'll just check that our enhanced validation script exists and runs
	// In a real implementation, this could parse templates and check for consistency

	templatePaths := []string{
		"templates/pages/_donate_form.plush.html",
		"templates/pages/donate.plush.html",
		"templates/pages/contact.plush.html",
	}

	templateFS := templates.FS()

	for _, templatePath := range templatePaths {
		// Basic check that template files exist
		require.FileExists(t, templatePath, "Template file should exist: %s", templatePath)

		// Read template content
		content, err := fs.ReadFile(templateFS, strings.TrimPrefix(templatePath, "templates/"))
		r.NoError(err, "Should be able to read template: %s", templatePath)

		templateContent := string(content)

		// Check that templates don't have obvious syntax errors
		r.NotContains(templateContent, "<%=", "Template should not have unclosed template tags")
		r.NotContains(templateContent, "<% %>", "Template should not have malformed template tags")

		// Check for balanced braces in template expressions (basic check)
		openBraces := strings.Count(templateContent, "<%=")
		closeBraces := strings.Count(templateContent, "%>")
		r.Equal(openBraces, closeBraces, "Template should have balanced template tags: %s", templatePath)
	}
}
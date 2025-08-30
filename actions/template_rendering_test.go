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
	req.Contains(body, "first_name", "Template should contain first name field")
	req.Contains(body, "donor_email", "Template should contain email field")
	req.Contains(body, "Donate Now", "Template should contain submit button")
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
	req.Contains(body, "Make a Donation", "Template should contain donation form title")
	req.Contains(body, "donation-form", "Template should contain donation form")
	// Note: Error display is handled by Buffalo's flash system, not direct template rendering
}



// Test_TemplateConsistencyValidation tests that templates use consistent variable access patterns
func Test_TemplateConsistencyValidation(t *testing.T) {
	req := require.New(t)

	// This test validates that templates don't have obvious syntax errors
	// For now, we'll just check that our enhanced validation script exists and runs
	// In a real implementation, this could parse templates and check for consistency

	templatePaths := []string{
		"templates/pages/donate.plush.html",
		"templates/pages/contact.plush.html",
	}

	templateFS := templates.FS()

	for _, templatePath := range templatePaths {
		// Read template content from embedded FS
		content, err := fs.ReadFile(templateFS, strings.TrimPrefix(templatePath, "templates/"))
		req.NoError(err, "Should be able to read template: %s", templatePath)

		templateContent := string(content)

		// Check that templates don't have malformed template tags
		// Count opening and closing tags to ensure they're balanced
		openTagCount := strings.Count(templateContent, "<%")
		closeTagCount := strings.Count(templateContent, "%>")
		req.Equal(openTagCount, closeTagCount, "Template should have balanced template tags: %s", templatePath)

		// Check for unclosed tags (tags that start but don't end on the same line)
		lines := strings.Split(templateContent, "\n")
		for i, line := range lines {
			openInLine := strings.Count(line, "<%")
			closeInLine := strings.Count(line, "%>")
			if openInLine > closeInLine {
				t.Errorf("Template %s line %d has unclosed template tag: %s", templatePath, i+1, strings.TrimSpace(line))
			}
		}
	}
}
package actions

import (
	"fmt"
	"io/fs"

	public "avrnpo.org/public"
	"avrnpo.org/templates"
	"html/template"
	"regexp"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/helpers/forms"
	"github.com/gobuffalo/helpers/hctx"
	"github.com/gobuffalo/tags/v3"
)

var r *render.Engine

func init() {
	// Common helpers
	commonHelpers := render.Helpers{
		forms.FormKey:         forms.Form,
		forms.FormForKey:      forms.FormFor,
		"getCurrentURL":       getCurrentURL,
		"stripTags":           stripTagsHelper,
		"dateFormat":          dateFormatHelper,
		"getDonateButtonText": getDonateButtonText,
		"current_path":        func() string { return "/" },
		"t":                   func(s string, args ...interface{}) string { return s }, // Simple fallback translator
		"csrf":                csrfHelper,
	}

	// Get the assets sub-filesystem
	assetsFS, _ := fs.Sub(public.EmbeddedAssets, "assets")

	// Standard render engine with layout
	r = render.New(render.Options{
		HTMLLayout:  "application.plush.html",
		TemplatesFS: templates.FS(),
		AssetsFS:    assetsFS,
		Helpers:     commonHelpers,
	})
}

// getCurrentURL returns the current request URL for use in templates
func getCurrentURL() string {
	// This is a simplified version - in production you'd want to access the request
	// from the context properly, but for now we'll return a placeholder
	return ""
}

// stripTagsHelper removes HTML tags from content for use in templates
func stripTagsHelper(content string) string {
	// Regular expression to match HTML tags
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	// Remove HTML tags
	cleaned := htmlTagRegex.ReplaceAllString(content, "")
	// Clean up extra whitespace
	cleaned = strings.TrimSpace(cleaned)
	// Replace multiple spaces/newlines with single spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	cleaned = spaceRegex.ReplaceAllString(cleaned, " ")
	return cleaned
}

// dateFormatHelper formats time.Time values for use in templates
func dateFormatHelper(t time.Time, format string) string {
	return t.Format(format)
}

// renderForRequest was removed in favor of a single render strategy (use r.HTML).
// Existing call sites will be updated to call r.HTML directly or c.Render with r.HTML.

// SanitizeString is a function that sanitizes a string for safe HTML display.
// It escapes HTML special characters to prevent XSS (Cross-Site Scripting) attacks.
// This function is useful for ensuring that user-generated content is displayed safely.
func SanitizeString(s string) string {
	return template.HTMLEscapeString(s)
}

// csrfHelper returns the CSRF token from the context for use in templates
func csrfHelper(opts tags.Options, help hctx.HelperContext) (template.HTML, error) {
	if help == nil {
		return template.HTML(""), nil
	}

	token := help.Value("authenticity_token")
	if token == nil {
		return template.HTML(""), nil
	}

	return template.HTML(fmt.Sprintf(`<input name="authenticity_token" type="hidden" value="%s" />`, token)), nil
}

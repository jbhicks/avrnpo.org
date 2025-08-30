package actions

import (
	avrnpo "avrnpo.org"
	"avrnpo.org/templates"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/helpers/forms"
)

var r *render.Engine
var rNoLayout *render.Engine

func init() {
	// Common helpers
	commonHelpers := render.Helpers{
		forms.FormKey:    forms.Form,
		forms.FormForKey: forms.FormFor,
		"getCurrentURL":  getCurrentURL,
		"stripTags":      stripTagsHelper,
		"dateFormat":     dateFormatHelper,
	}

	// Standard render engine with layout
	r = render.New(render.Options{
		HTMLLayout:  "application.plush.html",
		TemplatesFS: templates.FS(),
		AssetsFS:    avrnpo.FS(),
		Helpers:     commonHelpers,
	})

	// No-layout render engine for standalone pages like home
	rNoLayout = render.New(render.Options{
		TemplatesFS: templates.FS(),
		AssetsFS:    avrnpo.FS(),
		Helpers:     commonHelpers,
	})

	// Note: HTMX-specific minimal layout removed — HTMX requests will receive the
	// standard application layout to simplify rendering and keep a single template
	// strategy.
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

// IsHTMX detects if the request is from HTMX. It's retained for optional
// use elsewhere but we will no longer branch rendering engines on it.
func IsHTMX(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}

// renderForRequest was removed in favor of a single render strategy (use r.HTML).
// Existing call sites will be updated to call r.HTML directly or c.Render with r.HTML.

// SanitizeString is a function that sanitizes a string for safe HTML display.
// It escapes HTML special characters to prevent XSS (Cross-Site Scripting) attacks.
// This function is useful for ensuring that user-generated content is displayed safely.
func SanitizeString(s string) string {
	return template.HTMLEscapeString(s)
}

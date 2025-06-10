package actions

import (
	"avrnpo.org/public"
	"avrnpo.org/templates"
	"net/http"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/helpers/forms"
)

var r *render.Engine
var rHTMX *render.Engine // New engine for HTMX requests

func init() {
	// Common helpers for both render engines
	commonHelpers := render.Helpers{
		forms.FormKey:    forms.Form,
		forms.FormForKey: forms.FormFor,
		"getCurrentURL":  getCurrentURL,
		// You can add other common helpers here
	}

	// Standard render engine
	r = render.New(render.Options{
		HTMLLayout:  "application.plush.html",
		TemplatesFS: templates.FS(),
		AssetsFS:    public.FS(),
		Helpers:     commonHelpers,
	})

	// Render engine for HTMX requests
	rHTMX = render.New(render.Options{
		HTMLLayout:  "htmx.plush.html", // Use the minimal layout for HTMX
		TemplatesFS: templates.FS(),
		AssetsFS:    public.FS(),
		Helpers:     commonHelpers, // Share the same helpers
	})
}

// IsHTMX checks if the current request is an HTMX request.
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// getCurrentURL returns the current request URL for use in templates
func getCurrentURL(c interface{}) string {
	if ctx, ok := c.(interface {
		Request() *http.Request
	}); ok {
		req := ctx.Request()
		scheme := "http"
		if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		return scheme + "://" + req.Host + req.RequestURI
	}
	return ""
}

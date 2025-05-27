package actions

import (
	"my_go_saas_template/public"
	"my_go_saas_template/templates"

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

package actions

import (
        "net/http"
        "github.com/gobuffalo/buffalo"
)

// TeamHandler shows the team page
func TeamHandler(c buffalo.Context) error {
        // Always render partial content - navigation is handled by the home page
        return c.Render(http.StatusOK, rHTMX.HTML("pages/team.plush.html"))
}

// ProjectsHandler shows the projects page
func ProjectsHandler(c buffalo.Context) error {
        // Always render partial content - navigation is handled by the home page
        return c.Render(http.StatusOK, rHTMX.HTML("pages/projects.plush.html"))
}

// ContactHandler shows the contact form
func ContactHandler(c buffalo.Context) error {
        // Always render partial content - navigation is handled by the home page
        return c.Render(http.StatusOK, rHTMX.HTML("pages/contact.plush.html"))
}

// DonateHandler shows the donation page
func DonateHandler(c buffalo.Context) error {
        // Always render partial content - navigation is handled by the home page
        return c.Render(http.StatusOK, rHTMX.HTML("pages/donate.plush.html"))
}

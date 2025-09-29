package actions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/mw-csrf"
	"github.com/stretchr/testify/require"
)

// Minimal CSRF integration test to ensure middleware works with forms
func TestCSRFIntegrationCycle(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "development"})
	app.Use(csrf.New)

	// GET returns a one-time token in a form
	app.GET("/csrf-test", func(c buffalo.Context) error {
		responseToken := c.Value("authenticity_token")
		if responseToken == nil {
			return c.Render(500, r.String("No CSRF token generated"))
		}
		html := fmt.Sprintf(`<form method="post" action="/csrf-test"><input type="hidden" name="authenticity_token" value="%s" /></form>`, responseToken)
		return c.Render(200, r.String(html))
	})

	// POST accepts the token and returns ok
	app.POST("/csrf-test", func(c buffalo.Context) error {
		token := c.Param("authenticity_token")
		if token == "" {
			return c.Render(400, r.String("missing token"))
		}
		return c.Render(200, r.String("ok"))
	})

	// Request to get token (use GET to avoid CSRF middleware rejecting POST)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/csrf-test", nil)
	app.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "authenticity_token")

	bodyString := w.Body.String()
	tokenStart := strings.Index(bodyString, `value="`) + 7
	tokenEnd := strings.Index(bodyString[tokenStart:], `"`) + tokenStart
	token := bodyString[tokenStart:tokenEnd]
	require.NotEmpty(t, token)

	// Submit using token. Include cookies from the GET response so the
	// CSRF middleware can validate the one-time token against the session.
	formData := url.Values{"authenticity_token": {token}, "message": {"x"}}
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/csrf-test", strings.NewReader(formData.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// copy cookies from the GET response
	if res := w.Result(); res != nil {
		for _, c := range res.Cookies() {
			req2.AddCookie(c)
		}
	}

	app.ServeHTTP(w2, req2)
	require.Equal(t, 200, w2.Code)
}

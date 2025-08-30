package actions

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// fetchCSRF performs a GET to the given path using the app and returns the session cookie and authenticity_token value found in the response body.
func fetchCSRF(t *testing.T, app http.Handler, path string) (string, string) {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	rw := httptest.NewRecorder()
	app.ServeHTTP(rw, req)
	res := rw.Result()
	defer res.Body.Close()

	// capture cookie
	cookie := ""
	for _, c := range res.Cookies() {
		// Keep any session-like cookie (test suite uses _avrnpo.org_session)
		if strings.Contains(c.Name, "session") || c.Name == "_avrnpo.org_session" {
			cookie = c.String()
			break
		}
	}

	// read body
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)

	// find authenticity_token in a hidden input
	re := regexp.MustCompile(`<input[^>]+name="authenticity_token"[^>]+value="([^"]+)"`)
	m := re.FindStringSubmatch(body)
	if len(m) >= 2 {
		return cookie, m[1]
	}

	// fallback: try meta tag
	re2 := regexp.MustCompile(`<meta[^>]+name="csrf-token"[^>]+content="([^"]+)"`)
	m2 := re2.FindStringSubmatch(body)
	if len(m2) >= 2 {
		return cookie, m2[1]
	}

	// no token found
	return cookie, ""
}

// MockLogin performs a login POST to the application's auth endpoint and returns the session cookie and CSRF token for subsequent requests.
func MockLogin(t *testing.T, app http.Handler, email, password string) (string, string) {
	t.Helper()
	// First fetch login page to get initial cookie and token
	cookie, token := fetchCSRF(t, app, "/auth/new")

	form := url.Values{}
	form.Set("email", email)
	form.Set("password", password)
	form.Set("authenticity_token", token)

	req := httptest.NewRequest("POST", "/auth", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	rw := httptest.NewRecorder()
	app.ServeHTTP(rw, req)
	res := rw.Result()
	defer res.Body.Close()

	// Grab session cookie from response
	sessCookie := ""
	for _, c := range res.Cookies() {
		if strings.Contains(c.Name, "session") || c.Name == "_avrnpo.org_session" {
			sessCookie = c.String()
			break
		}
	}

	// After login, fetch root to get a fresh CSRF token tied to session
	finalCookie, finalToken := "", ""
	if sessCookie != "" {
		// combine cookies if initial cookie exists
		combined := sessCookie
		if cookie != "" && !strings.Contains(sessCookie, cookie) {
			combined = cookie + "; " + sessCookie
		}
		// fetch home to extract token
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.Header.Set("Cookie", combined)
		rw2 := httptest.NewRecorder()
		app.ServeHTTP(rw2, req2)
		res2 := rw2.Result()
		defer res2.Body.Close()

		for _, c := range res2.Cookies() {
			if strings.Contains(c.Name, "session") || c.Name == "_avrnpo.org_session" {
				finalCookie = c.String()
				break
			}
		}

		bodyBytes, _ := io.ReadAll(res2.Body)
		body := string(bodyBytes)
		re3 := regexp.MustCompile(`<input[^>]+name="authenticity_token"[^>]+value="([^"]+)"`)
		m3 := re3.FindStringSubmatch(body)
		if len(m3) >= 2 {
			finalToken = m3[1]
		} else {
			re4 := regexp.MustCompile(`<meta[^>]+name="csrf-token"[^>]+content="([^"]+)"`)
			m4 := re4.FindStringSubmatch(body)
			if len(m4) >= 2 {
				finalToken = m4[1]
			}
		}
	}

	if finalCookie == "" {
		finalCookie = sessCookie
	}
	return finalCookie, finalToken
}

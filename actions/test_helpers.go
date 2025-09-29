package actions

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"avrnpo.org/models"
)

// CreatePostForTest creates and saves a published post with the given attributes and returns it.
func CreatePostForTest(db *pop.Connection, title, slug, content string, authorID uuid.UUID) (*models.Post, error) {
	post := &models.Post{
		Title:     title,
		Slug:      slug,
		Content:   content,
		Excerpt:   content,
		Published: true,
		AuthorID:  authorID,
	}
	verrs, err := db.ValidateAndCreate(post)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if verrs.HasAny() {
		return nil, errors.New("validation errors creating post")
	}
	return post, nil
}

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
	t.Logf("🔍 MockLogin: Initial fetchCSRF - Cookie: '%s', Token: '%s'", cookie, token)

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

	// Debug the login response
	t.Logf("🔍 MockLogin: POST /auth - Status: %d, Location: %s", res.StatusCode, res.Header.Get("Location"))

	if res.StatusCode == 401 {
		bodyBytes, _ := io.ReadAll(res.Body)
		bodyExcerpt := string(bodyBytes)
		if len(bodyExcerpt) > 200 {
			bodyExcerpt = bodyExcerpt[:200] + "..."
		}
		t.Logf("🔍 MockLogin: LOGIN FAILED - Status: 401, Body excerpt: %s", bodyExcerpt)
		return "", ""
	}

	// Extract session cookie from login response headers
	sessCookie := ""
	t.Logf("🔍 MockLogin: Checking login response cookies (%d total)", len(res.Cookies()))
	for i, c := range res.Cookies() {
		t.Logf("🔍 MockLogin: Cookie %d: %s = %s", i, c.Name, c.Value)
		if strings.Contains(c.Name, "session") || c.Name == "_avrnpo.org_session" {
			sessCookie = c.String()
			t.Logf("🔍 MockLogin: Found session cookie: %s", sessCookie)
			break
		}
	}

	// If no session cookie in response, try to find it in request headers (Buffalo might reuse existing session)
	if sessCookie == "" && cookie != "" {
		t.Logf("🔍 MockLogin: No new session cookie, testing if original cookie works by accessing /account")
		// Buffalo often reuses the existing session cookie, so check if login was successful by trying to access a protected page
		testReq := httptest.NewRequest("GET", "/account", nil)
		testReq.Header.Set("Cookie", cookie)
		testRw := httptest.NewRecorder()
		app.ServeHTTP(testRw, testReq)
		testRes := testRw.Result()
		defer testRes.Body.Close()

		t.Logf("🔍 MockLogin: /account test - Status: %d", testRes.StatusCode)
		// If we can access /account (which requires auth), then the original cookie is valid
		if testRes.StatusCode == 200 {
			sessCookie = cookie
			t.Logf("🔍 MockLogin: Reusing initial session cookie after successful login")
		} else {
			t.Logf("🔍 MockLogin: /account access failed, trying /profile instead")
			// Try /profile as alternative
			testReq2 := httptest.NewRequest("GET", "/profile", nil)
			testReq2.Header.Set("Cookie", cookie)
			testRw2 := httptest.NewRecorder()
			app.ServeHTTP(testRw2, testReq2)
			testRes2 := testRw2.Result()
			defer testRes2.Body.Close()

			t.Logf("🔍 MockLogin: /profile test - Status: %d", testRes2.StatusCode)
			if testRes2.StatusCode == 200 {
				sessCookie = cookie
				t.Logf("🔍 MockLogin: Reusing initial session cookie (verified with /profile)")
			}
		}
	}

	t.Logf("🔍 MockLogin: Session cookie from response: '%s'", sessCookie)

	// After login, fetch a page to get a fresh CSRF token tied to the authenticated session
	finalCookie, finalToken := "", ""
	if sessCookie != "" {
		// Use the session cookie to fetch a fresh CSRF token
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.Header.Set("Cookie", sessCookie)
		rw2 := httptest.NewRecorder()
		app.ServeHTTP(rw2, req2)
		res2 := rw2.Result()
		defer res2.Body.Close()

		// Check for updated session cookie
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

// includeCSRF adds the CSRF token to a request, either as a form field or header, and sets the cookie.
func includeCSRF(req *http.Request, token, cookie string) {
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if token != "" {
		// For enhanced requests (formerly HTMX/AJAX), use header
		if req.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			req.Header.Set("X-CSRF-Token", token)
		} else {
			// For form requests, add to form data if it's a POST with form encoding
			if req.Method == "POST" && req.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				body := req.Body
				if body != nil {
					data, _ := io.ReadAll(body)
					form := string(data)
					if form != "" && !strings.Contains(form, "authenticity_token=") {
						form += "&authenticity_token=" + url.QueryEscape(token)
						req.Body = io.NopCloser(strings.NewReader(form))
						req.ContentLength = int64(len(form))
					}
				}
			}
		}
	}
}

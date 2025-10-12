package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
)

const testBaseURL = "http://localhost:8090"

func init() {
	_ = godotenv.Load()
}

func TestE2E_FullUserJourney(t *testing.T) {
	if os.Getenv("E2E_TESTS") != "1" {
		t.Skip("Skipping E2E tests. Set E2E_TESTS=1 to run")
	}

	if !isServerRunning(testBaseURL) {
		t.Fatal("Server is not running. Start the app with: ./avrnpo serve")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	defer timeoutCancel()
	ctx = timeoutCtx

	t.Run("Homepage loads with recent posts", func(t *testing.T) {
		var pageTitle string
		var postsCount int

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
			chromedp.Title(&pageTitle),
			chromedp.Evaluate(`document.querySelectorAll('.blog-post-preview').length`, &postsCount),
		)

		if err != nil {
			t.Fatalf("Homepage test failed: %v", err)
		}

		if pageTitle == "" {
			t.Error("Page title is empty")
		}

		t.Logf("✓ Homepage loaded: '%s'", pageTitle)
		t.Logf("✓ Found %d blog post previews", postsCount)
	})

	t.Run("Blog listing page works", func(t *testing.T) {
		var heading string

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/blog"),
			chromedp.WaitVisible(`h1`, chromedp.ByQuery),
			chromedp.Text(`h1`, &heading, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Blog listing test failed: %v", err)
		}

		t.Logf("✓ Blog listing loaded: '%s'", heading)
	})

	t.Run("Blog post 'Read More' button works", func(t *testing.T) {
		var postTitle string
		var currentURL string

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/blog"),
			chromedp.WaitVisible(`article`, chromedp.ByQuery),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Click(`a[href^="/blog/"][role="button"]`, chromedp.ByQuery),
			chromedp.WaitVisible(`h1`, chromedp.ByQuery),
			chromedp.Location(&currentURL),
			chromedp.Text(`h1`, &postTitle, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Blog post navigation test failed: %v", err)
		}

		if postTitle == "" {
			t.Error("Post title is empty")
		}

		if !strings.Contains(currentURL, "/blog/") || currentURL == testBaseURL+"/blog" {
			t.Errorf("Expected to navigate to a blog post URL, got: %s", currentURL)
		}

		t.Logf("✓ 'Read More' button navigated to: %s - '%s'", currentURL, postTitle)
	})

	t.Run("Contact form submission", func(t *testing.T) {
		var csrfToken string
		var successMessage string

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/contact"),
			chromedp.WaitVisible(`form`, chromedp.ByQuery),
			chromedp.AttributeValue(`input[name="csrf_token"]`, "value", &csrfToken, nil, chromedp.ByQuery),
			chromedp.SendKeys(`input[name="name"]`, "Test User", chromedp.ByQuery),
			chromedp.SendKeys(`input[name="email"]`, "test@example.com", chromedp.ByQuery),
			chromedp.SendKeys(`input[name="phone"]`, "555-1234", chromedp.ByQuery),
			chromedp.SendKeys(`textarea[name="message"]`, "This is a test message from E2E tests", chromedp.ByQuery),
			chromedp.Submit(`form`, chromedp.ByQuery),
			chromedp.WaitVisible(`.success-message, .alert-success, h1`, chromedp.ByQuery),
			chromedp.Text(`body`, &successMessage, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Contact form test failed: %v", err)
		}

		if csrfToken == "" {
			t.Error("CSRF token not found in contact form")
		}

		t.Logf("✓ Contact form submitted successfully with CSRF token")
	})

	t.Run("Donate page loads", func(t *testing.T) {
		var donateHeading string

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/donate"),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
			chromedp.WaitReady(`#donationForm`, chromedp.ByQuery),
			chromedp.Text(`h1`, &donateHeading, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Donate page test failed: %v", err)
		}

		t.Logf("✓ Donate page loaded: '%s'", donateHeading)
	})

	t.Run("Donate form submission with CSRF", func(t *testing.T) {
		var formCsrfToken string
		var metaCsrfToken string

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/donate"),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
			chromedp.WaitReady(`#donationForm`, chromedp.ByQuery),
			chromedp.AttributeValue(`input[name="csrf_token"]`, "value", &formCsrfToken, nil, chromedp.ByQuery),
			chromedp.AttributeValue(`meta[name="csrf-token"]`, "content", &metaCsrfToken, nil, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Failed to load donate form: %v", err)
		}

		if formCsrfToken == "" {
			t.Error("CSRF token not found in donate form")
		}

		if metaCsrfToken == "" {
			t.Error("CSRF token not found in meta tag")
		}

		if formCsrfToken != metaCsrfToken {
			t.Errorf("Form CSRF token (%s) does not match meta tag (%s)", formCsrfToken, metaCsrfToken)
		}

		t.Logf("Form CSRF: %s, Meta CSRF: %s", formCsrfToken, metaCsrfToken)

		err = chromedp.Run(ctx,
			chromedp.SendKeys(`input#amount`, "25", chromedp.ByQuery),
			chromedp.SendKeys(`input#name`, "E2E Test Donor", chromedp.ByQuery),
			chromedp.SendKeys(`input#email`, "test-donor@example.com", chromedp.ByQuery),
			chromedp.SendKeys(`input#address_line1`, "123 Test St", chromedp.ByQuery),
			chromedp.SendKeys(`input#city`, "Testville", chromedp.ByQuery),
			chromedp.SendKeys(`input#province`, "TS", chromedp.ByQuery),
			chromedp.SendKeys(`input#postal_code`, "12345", chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Failed to fill form: %v", err)
		}

		var httpStatus int
		var responseBody string
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`
				fetch('/donate', {
					method: 'POST',
					body: new FormData(document.getElementById('donationForm'))
				}).then(r => { window._testStatus = r.status; return r.text(); }).then(t => { window._testBody = t; })
			`, nil),
			chromedp.Sleep(1*time.Second),
			chromedp.Evaluate(`window._testStatus`, &httpStatus),
			chromedp.Evaluate(`window._testBody`, &responseBody),
		)

		if err != nil {
			t.Fatalf("Failed to test fetch: %v", err)
		}

		t.Logf("POST /donate response: Status=%d, Body=%s", httpStatus, responseBody)

		if httpStatus == 403 {
			t.Errorf("Donate POST failed with 403 Forbidden - CSRF validation failed!")
		} else if httpStatus == 200 {
			t.Logf("✓ Donate form submitted successfully with CSRF token (HTTP 200)")
		} else {
			t.Logf("Unexpected status: %d", httpStatus)
		}
	})

	t.Run("About page loads", func(t *testing.T) {
		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/about"),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("About page test failed: %v", err)
		}

		t.Logf("✓ About page loaded")
	})

	t.Run("Team page loads", func(t *testing.T) {
		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/team"),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Team page test failed: %v", err)
		}

		t.Logf("✓ Team page loaded")
	})

	t.Run("Projects page loads", func(t *testing.T) {
		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/projects"),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Projects page test failed: %v", err)
		}

		t.Logf("✓ Projects page loaded")
	})
}

func TestE2E_AdminWorkflow(t *testing.T) {
	if os.Getenv("E2E_TESTS") != "1" {
		t.Skip("Skipping E2E tests. Set E2E_TESTS=1 to run")
	}

	if !isServerRunning(testBaseURL) {
		t.Fatal("Server is not running. Start the app with: ./avrnpo serve")
	}

	adminEmail := os.Getenv("PB_ADMIN_EMAIL")
	adminPassword := os.Getenv("PB_ADMIN_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		t.Skip("Admin credentials not set. Set PB_ADMIN_EMAIL and PB_ADMIN_PASSWORD")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
	defer timeoutCancel()
	ctx = timeoutCtx

	t.Run("Admin login", func(t *testing.T) {
		var csrfToken string

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/auth/login"),
			chromedp.WaitVisible(`input[name="email"]`, chromedp.ByQuery),
			chromedp.AttributeValue(`input[name="csrf_token"]`, "value", &csrfToken, nil, chromedp.ByQuery),
			chromedp.SendKeys(`input[name="email"]`, adminEmail, chromedp.ByQuery),
			chromedp.SendKeys(`input[name="password"]`, adminPassword, chromedp.ByQuery),
			chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Admin login form submission failed: %v", err)
		}

		if csrfToken == "" {
			t.Error("CSRF token not found in login form")
		}

		var currentURL string
		timeout := time.After(5 * time.Second)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				t.Fatal("Login timed out - still on login page after 5s")
			case <-ticker.C:
				chromedp.Run(ctx, chromedp.Location(&currentURL))
				if currentURL != testBaseURL+"/auth/login" && !strings.Contains(currentURL, "email=") {
					goto loginSuccess
				}
			}
		}

	loginSuccess:
		t.Logf("✓ Admin logged in successfully - Current URL: %s", currentURL)
	})

	t.Run("Access admin post list", func(t *testing.T) {
		var heading string

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/cms/posts"),
			chromedp.WaitVisible(`h1, .admin-header`, chromedp.ByQuery),
			chromedp.Text(`h1`, &heading, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Admin post list access failed: %v", err)
		}

		t.Logf("✓ Admin post list accessible: '%s'", heading)
	})

	t.Run("Create new post", func(t *testing.T) {
		testTitle := fmt.Sprintf("E2E Test Post %d", time.Now().Unix())
		testContent := "This is test content from E2E tests"
		var csrfToken string

		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/cms/posts/new"),
			chromedp.WaitVisible(`input#title`, chromedp.ByQuery),
			chromedp.AttributeValue(`input[name="csrf_token"]`, "value", &csrfToken, nil, chromedp.ByQuery),
			chromedp.SendKeys(`input#title`, testTitle, chromedp.ByQuery),
			chromedp.SendKeys(`textarea#excerpt`, "Test excerpt for E2E testing", chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Failed to fill form fields: %v", err)
		}

		if csrfToken == "" {
			t.Error("CSRF token not found in post creation form")
		}

		err = chromedp.Run(ctx,
			chromedp.Evaluate(`if (window.easyMDE) { window.easyMDE.value('`+testContent+`'); }`, nil),
			chromedp.Click(`input#published`, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatalf("Failed to set content and publish: %v", err)
		}

		var finalURL string
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`
				const form = document.querySelector('form');
				const formData = new FormData(form);
				htmx.ajax('POST', '/cms/posts', {values: Object.fromEntries(formData), target: 'body', swap: 'innerHTML'});
			`, nil),
			chromedp.WaitReady(`body`, chromedp.ByQuery),
			chromedp.Location(&finalURL),
		)
		if err != nil {
			t.Fatalf("Failed to submit form: %v", err)
		}

		t.Logf("✓ Created new post: '%s' - Final URL: %s", testTitle, finalURL)
	})

	t.Run("Edit existing post", func(t *testing.T) {
		var editURL string

		// Navigate to post list and find Edit link
		err := chromedp.Run(ctx,
			chromedp.Navigate(testBaseURL+"/cms/posts"),
			chromedp.WaitVisible(`a[href*="/cms/posts/"][href*="/edit"]`, chromedp.ByQuery),
			chromedp.AttributeValue(`a[href*="/cms/posts/"][href*="/edit"]`, "href", &editURL, nil, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Failed to find edit link: %v", err)
		}

		if editURL == "" {
			t.Fatal("Edit URL is empty")
		}

		t.Logf("Found edit URL: %s", editURL)

		// Click the edit link
		var currentURL string
		var pageTitle string
		err = chromedp.Run(ctx,
			chromedp.Click(`a[href*="/cms/posts/"][href*="/edit"]`, chromedp.ByQuery),
			chromedp.WaitVisible(`h1`, chromedp.ByQuery),
			chromedp.Location(&currentURL),
			chromedp.Text(`h1`, &pageTitle, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Failed to navigate to edit page: %v", err)
		}

		if !strings.Contains(currentURL, "/cms/posts/") || !strings.Contains(currentURL, "/edit") {
			t.Errorf("Expected to be on edit page, but got URL: %s", currentURL)
		}

		if !strings.Contains(pageTitle, "Edit") {
			t.Errorf("Expected page title to contain 'Edit', got: %s", pageTitle)
		}

		t.Logf("✓ Edit page loaded successfully: '%s' at %s", pageTitle, currentURL)

		// Fill out the edit form and submit
		var updatedTitle = "Updated E2E Test Post"
		var updatedContent = "This is updated test content from E2E tests"
		var updatedExcerpt = "Updated excerpt for E2E testing"

		err = chromedp.Run(ctx,
			chromedp.WaitVisible(`input[name="title"]`, chromedp.ByQuery),
			chromedp.Clear(`input[name="title"]`, chromedp.ByQuery),
			chromedp.SendKeys(`input[name="title"]`, updatedTitle, chromedp.ByQuery),
			chromedp.Clear(`textarea[name="content"]`, chromedp.ByQuery),
			chromedp.SendKeys(`textarea[name="content"]`, updatedContent, chromedp.ByQuery),
			chromedp.Clear(`textarea[name="excerpt"]`, chromedp.ByQuery),
			chromedp.SendKeys(`textarea[name="excerpt"]`, updatedExcerpt, chromedp.ByQuery),
			chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
			chromedp.WaitVisible(`h1`, chromedp.ByQuery),
		)

		if err != nil {
			t.Fatalf("Failed to submit edit form: %v", err)
		}

		// Verify we're back on the admin posts page
		var finalURL string
		err = chromedp.Run(ctx, chromedp.Location(&finalURL))
		if err != nil {
			t.Fatalf("Failed to get final URL: %v", err)
		}
		if !strings.Contains(finalURL, "/admin/posts") {
			t.Errorf("Expected to be redirected to admin posts page, got: %s", finalURL)
		}

		t.Logf("✓ Edit form submitted successfully and redirected to admin posts")

	})

	t.Run("Delete post", func(t *testing.T) {

		// Navigate to post list and find Delete button

		err := chromedp.Run(ctx,

			chromedp.Navigate(testBaseURL+"/admin/posts"),

			chromedp.WaitVisible(`button[title="Delete post"]`, chromedp.ByQuery),
		)

		if err != nil {

			t.Fatalf("Failed to navigate to admin posts page: %v", err)

		}

		// Click the delete button (this should trigger a confirmation dialog)

		err = chromedp.Run(ctx,

			chromedp.Click(`button[title="Delete post"]`, chromedp.ByQuery),

			chromedp.WaitVisible(`.modal`, chromedp.ByQuery),
		)

		if err != nil {

			t.Fatalf("Failed to click delete button: %v", err)

		}

		// Confirm deletion in the modal

		err = chromedp.Run(ctx,

			chromedp.Click(`.modal button[class*="primary"]`, chromedp.ByQuery),

			chromedp.WaitVisible(`h1`, chromedp.ByQuery),
		)

		if err != nil {

			t.Fatalf("Failed to confirm deletion: %v", err)

		}

		t.Logf("✓ Post deleted successfully")

	})

	if !isServerRunning(testBaseURL) {
		t.Fatal("Server is not running")
	}

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{"Homepage", "/", http.StatusOK},
		{"Blog List", "/blog", http.StatusOK},
		{"Donate Page", "/donate", http.StatusOK},
		{"Contact Page", "/contact", http.StatusOK},
		{"About Page", "/about", http.StatusOK},
		{"Team Page", "/team", http.StatusOK},
		{"Projects Page", "/projects", http.StatusOK},
		{"Login Page", "/auth/login", http.StatusOK},
		{"Nav State API", "/api/nav-state", http.StatusOK},
		{"PocketBase Admin", "/_/", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(testBaseURL + tt.url)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d for %s", tt.expectedStatus, resp.StatusCode, tt.url)
			} else {
				t.Logf("✓ %s returned %d", tt.url, resp.StatusCode)
			}
		})

		t.Run("Delete post", func(t *testing.T) {
			// Navigate to post list and find Delete button
			err := chromedp.Run(ctx,
				chromedp.Navigate(testBaseURL+"/admin/posts"),
				chromedp.WaitVisible(`button[title="Delete post"]`, chromedp.ByQuery),
			)

			if err != nil {
				t.Fatalf("Failed to navigate to admin posts page: %v", err)
			}

			// Click the delete button (this should trigger a confirmation dialog)
			err = chromedp.Run(ctx,
				chromedp.Click(`button[title="Delete post"]`, chromedp.ByQuery),
				chromedp.WaitVisible(`.modal`, chromedp.ByQuery),
			)

			if err != nil {
				t.Fatalf("Failed to click delete button: %v", err)
			}

			// Confirm deletion in the modal
			err = chromedp.Run(ctx,
				chromedp.Click(`.modal button[class*="primary"]`, chromedp.ByQuery),
				chromedp.WaitVisible(`h1`, chromedp.ByQuery),
			)

			if err != nil {
				t.Fatalf("Failed to confirm deletion: %v", err)
			}

			t.Logf("✓ Post deleted successfully")
		})
	}
}

func isServerRunning(baseURL string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}

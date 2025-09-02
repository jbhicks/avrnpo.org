package actions

import (
	"avrnpo.org/locales"
	"avrnpo.org/models"
	"avrnpo.org/public"
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/logger"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/mw-csrf"
	"github.com/gorilla/sessions"
	"github.com/unrolled/secure"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Validation utilities for secure input validation
var (
	// RFC 5322 compliant email regex (simplified but secure)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
)

// Validation utilities for handlers to use
// These functions provide consistent validation across the application

// ValidateContactForm validates contact form input
func ValidateContactForm(c buffalo.Context) error {
	name := SanitizeInput(c.Param("name"))
	email := SanitizeInput(c.Param("email"))
	subject := SanitizeInput(c.Param("subject"))
	message := SanitizeInput(c.Param("message"))

	if err := ValidateRequiredString(name, "Name", 100); err != nil {
		return err
	}

	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidateRequiredString(subject, "Subject", 200); err != nil {
		return err
	}

	if err := ValidateRequiredString(message, "Message", 2000); err != nil {
		return err
	}

	// Store sanitized values back in context for processing
	c.Set("name", name)
	c.Set("email", email)
	c.Set("subject", subject)
	c.Set("message", message)

	return nil
}

// ValidateEmail performs secure email validation
func ValidateEmail(email string) error {
	if len(email) == 0 {
		return fmt.Errorf("email is required")
	}

	if len(email) > 254 { // RFC 5321 limit
		return fmt.Errorf("email address is too long")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("please enter a valid email address")
	}

	// Additional security checks
	if strings.Contains(email, "..") || strings.Contains(email, " ") {
		return fmt.Errorf("please enter a valid email address")
	}

	return nil
}

// ValidateRequiredString validates a required string field
func ValidateRequiredString(value, fieldName string, maxLength int) error {
	if len(strings.TrimSpace(value)) == 0 {
		return fmt.Errorf("%s is required", fieldName)
	}

	if len(value) > maxLength {
		return fmt.Errorf("%s must be less than %d characters", fieldName, maxLength)
	}

	return nil
}

// SanitizeInput removes potentially dangerous characters
func SanitizeInput(input string) string {
	// Remove null bytes and control characters
	input = strings.Map(func(r rune) rune {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Keep tab, LF, CR
			return -1
		}
		return r
	}, input)

	// Trim whitespace
	return strings.TrimSpace(input)
}

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

var (
	app                *buffalo.App
	appOnce            sync.Once
	T                  *i18n.Translator
	blogResource       = &PublicPostsResource{}
	postsResource      = &PostsResource{}
	adminUsersResource = &AdminUsersResource{}
)

// isStaticAsset checks if the path is for a static asset that should not be cached in development
func isStaticAsset(path string) bool {
	return strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".svg") ||
		strings.Contains(path, "/public/") ||
		strings.Contains(path, "/css/") ||
		strings.Contains(path, "/js/")
}

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	appOnce.Do(func() {

		// Set Buffalo to use our logrus-based logger for all request logs
		// Use Buffalo's built-in logger with multi-writer (terminal + file)

		// Ensure logs directory exists
		if err := os.MkdirAll("logs", 0755); err != nil {
			println("Warning: Failed to create logs directory:", err.Error())
		}

		// Try to create log file to verify permissions
		_, err := os.OpenFile("logs/application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			println("Warning: Failed to create log file:", err.Error())
		} else {
			println("✅ Log file setup successful: logs/application.log")
		}

		// Set log level based on environment: debug for dev, warn for test/prod
		logLevel := "debug"
		if ENV == "test" || ENV == "production" {
			logLevel = "warn"
		}

		// Configure session store
		sessionSecret := envy.Get("SESSION_SECRET", "development-session-secret-change-in-production")

		// Get port from environment or use Coolify's exposed port
		port := envy.Get("PORT", "3001") // Match Coolify's port exposes setting
		addr := "0.0.0.0:" + port

		app = buffalo.New(buffalo.Options{
			Env:           ENV,
			SessionName:   "_avrnpo.org_session",
			SessionStore:  sessions.NewCookieStore([]byte(sessionSecret)),
			CompressFiles: true, // Enable gzip compression for static files
			Addr:          addr, // Listen on all interfaces for container access
		})

		// Create logger with the specified level and set it
		buffaloLogger := logger.NewLogger(logLevel)
		app.Logger = buffaloLogger

		// Debug environment variables (after app is initialized)
		app.Logger.Infof("Environment check - GO_ENV: %s, SESSION_SECRET length: %d", ENV, len(sessionSecret))
		app.Logger.Infof("Application configured to listen on: %s", addr)
		if len(sessionSecret) > 10 {
			app.Logger.Infof("SESSION_SECRET value starts with: %s...", sessionSecret[:10])
		} else {
			app.Logger.Infof("SESSION_SECRET value: %s", sessionSecret)
		}

		if ENV == "production" && sessionSecret == "development-session-secret-change-in-production" {
			app.Logger.Warn("SESSION_SECRET not set in production! Using insecure default.")
		}

		// Use Buffalo's built-in request logging middleware
		app.Use(buffalo.RequestLoggerFunc)

		// Inject i18n translations middleware for all requests (early in stack)
		app.Use(translations())

		// Inject DB transaction middleware for all requests
		app.Use(popmw.Transaction(models.DB))

		// Set current user for all requests (after DB transactions)
		app.Use(SetCurrentUser)

		// TEMPORARILY DISABLE CSRF for debugging
		app.Use(csrf.New)

		// Additional middleware can be added here. Examples:
		// app.Use(forceSSL())
		// app.Use(secure.New(secure.Options{...}).Handler)

		// Skip CSRF protection only for legitimate API endpoints (webhooks, payment callbacks) in non-test environments
		if ENV != "test" {
			app.Middleware.Skip(csrf.New, HelcimWebhookHandler, debugFilesHandler, DebugFlashHandler, DonateUpdateAmountHandler, DonateHandler)
		}
		app.GET("/debug/files", debugFilesHandler)

		// Public routes
		app.GET("/", HomeHandler)
		app.GET("/contact", ContactHandler)
		app.POST("/contact", ContactHandler)
		app.GET("/team", TeamHandler)
		app.GET("/projects", ProjectsHandler)
		app.GET("/donate", DonateHandler)
		app.POST("/donate", DonateHandler)
		app.POST("/donate/update-amount", DonateUpdateAmountHandler)
		app.POST("/donate/update-submit", DonateUpdateSubmitHandler)
		app.GET("/donate/payment", DonatePaymentHandler)
		app.GET("/donate/success", DonationSuccessHandler)
		app.GET("/donate/failed", DonationFailedHandler)
		app.POST("/api/donations/initialize", DonationInitializeHandler)
		app.POST("/api/donations/process", ProcessPaymentHandler)
		app.POST("/api/donations/webhook", HelcimWebhookHandler)
		app.GET("/users/new", UsersNew)
		app.POST("/users", UsersCreate)
		app.GET("/auth", AuthLanding)
		app.GET("/auth/new", AuthNew)
		app.POST("/auth", AuthCreate)
		app.DELETE("/auth", AuthDestroy)
		app.GET("/dashboard", Authorize(DashboardHandler))
		app.GET("/profile", Authorize(ProfileSettings))
		app.POST("/profile", Authorize(ProfileUpdate))
		app.GET("/account", Authorize(AccountSettings))
		app.POST("/account", Authorize(AccountUpdate))
		app.GET("/account/subscriptions", Authorize(SubscriptionsList))
		app.GET("/account/subscriptions/{subscriptionId}", Authorize(SubscriptionDetails))
		app.POST("/account/subscriptions/{subscriptionId}/cancel", Authorize(CancelSubscription))
		app.Resource("/blog", blogResource) // Admin routes
		adminGroup := app.Group("/admin")
		adminGroup.Use(AdminRequired)
		adminGroup.GET("/", AdminDashboard)
		adminGroup.GET("/dashboard", AdminDashboard)
		adminGroup.GET("/users", AdminUsers)
		adminGroup.GET("/users/{user_id}", AdminUserShow)
		adminGroup.POST("/users/{user_id}", AdminUserUpdate)
		adminGroup.DELETE("/users/{user_id}", AdminUserDelete)
		adminGroup.Resource("/users", adminUsersResource)
		adminGroup.GET("/posts", AdminPostsIndex)
		adminGroup.GET("/posts/new", AdminPostsNew)
		adminGroup.POST("/posts", AdminPostsCreate)
		adminGroup.GET("/posts/{post_id}", AdminPostsShow)
		adminGroup.GET("/posts/{post_id}/edit", AdminPostsEdit)
		adminGroup.POST("/posts/{post_id}", AdminPostsUpdate)
		adminGroup.DELETE("/posts/{post_id}", AdminPostsDestroy)
		adminGroup.POST("/posts/bulk", AdminPostsBulk)
		adminGroup.Resource("/posts", postsResource)
		adminGroup.GET("/donations", AdminDonationsIndex)
		adminGroup.GET("/donations/{donation_id}", AdminDonationShow)

		// Serve assets from /assets path
		if ENV == "production" {
			// Production: use embedded assets
			assetsFS, _ := fs.Sub(public.EmbeddedAssets, "assets")
			app.ServeFiles("/assets", http.FS(assetsFS))
		} else {
			// Development/Test: serve from filesystem for hot reload
			app.ServeFiles("/assets", http.Dir("public/assets"))

		}
	})

	return app
}

// debugFilesHandler serves debug files for development
func debugFilesHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.String("Debug files endpoint"))
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(locales.FS(), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}

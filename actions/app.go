package actions

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"sync"

	avrnpo "avrnpo.org"
	"avrnpo.org/locales"
	"avrnpo.org/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/logger"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/mw-csrf"
	"github.com/unrolled/secure"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
	T       *i18n.Translator
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
		logFile, err := os.OpenFile("logs/application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			println("Warning: Failed to open log file:", err.Error())
		}
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		buffaloLogger := logger.NewLogger("debug")
		if outableLogger, ok := buffaloLogger.(logger.Outable); ok {
			outableLogger.SetOutput(multiWriter)
		}

		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_avrnpo.org_session",
		})
		app.Logger = buffaloLogger

		// Inject DB transaction middleware for all requests
		app.Use(popmw.Transaction(models.DB))

		// Inject i18n translations middleware for all requests
		app.Use(translations())

		// Enable CSRF protection for non-test environments
		if os.Getenv("GO_ENV") != "test" {
			app.Use(csrf.New)
		}

		blogResource := PublicPostsResource{}

		// Main route declarations
		app.GET("/", SetCurrentUser(HomeHandler))
		app.GET("/dashboard", SetCurrentUser(Authorize(DashboardHandler)))
		app.GET("/donate", SetCurrentUser(DonateHandler))
		app.POST("/donate", SetCurrentUser(DonateHandler))
		app.PATCH("/donate/update-amount", DonateUpdateAmountHandler)
		app.POST("/donate/update-amount", DonateUpdateAmountHandler) // For testing - Buffalo test suite doesn't support PATCH
		app.GET("/donate/payment", DonatePaymentHandler)
		app.GET("/donate/success", DonationSuccessHandler)
		app.GET("/donate/failed", DonationFailedHandler)
		app.GET("/team", SetCurrentUser(TeamHandler))
		app.GET("/projects", SetCurrentUser(ProjectsHandler))
		app.GET("/contact", SetCurrentUser(ContactHandler))
		app.POST("/contact", SetCurrentUser(ContactSubmitHandler))
		app.GET("/blog", SetCurrentUser(blogResource.List))
		app.GET("/blog/{slug}", SetCurrentUser(blogResource.Show))
		app.GET("/users/new", SetCurrentUser(UsersNew))
		app.GET("/users/new/", func(c buffalo.Context) error {
			return c.Redirect(http.StatusFound, "/users/new")
		})
		app.POST("/users", SetCurrentUser(UsersCreate))
		app.GET("/auth/new", AuthNew)
		app.POST("/auth", AuthCreate)
		app.GET("/auth/", func(c buffalo.Context) error {
			return c.Redirect(http.StatusFound, "/auth/new")
		})
		app.GET("/account", SetCurrentUser(Authorize(AccountSettings)))
		app.POST("/account", SetCurrentUser(Authorize(AccountUpdate)))
		app.GET("/account/subscriptions", SetCurrentUser(Authorize(SubscriptionsList)))
		app.GET("/account/subscriptions/{subscriptionId}", SetCurrentUser(Authorize(SubscriptionDetails)))
		app.POST("/account/subscriptions/{subscriptionId}/cancel", SetCurrentUser(Authorize(CancelSubscription)))
		app.GET("/profile", SetCurrentUser(Authorize(ProfileSettings)))
		app.POST("/profile", SetCurrentUser(Authorize(ProfileUpdate)))

		// Admin group with required middleware
		adminGroup := app.Group("/admin")
		adminGroup.Use(SetCurrentUser)
		adminGroup.Use(AdminRequired)
		adminGroup.GET("/", func(c buffalo.Context) error {
			return c.Redirect(http.StatusFound, "/admin/dashboard")
		})
		adminGroup.GET("/dashboard", AdminDashboard)
		postsResource := PostsResource{}
		adminGroup.Resource("/posts", postsResource)
		adminUsersResource := AdminUsersResource{}
		adminGroup.Resource("/users", adminUsersResource)

		// Donation API endpoints
		app.POST("/api/donations/initialize", DonationInitializeHandler)
		app.POST("/api/donations/{donationId}/complete", DonationCompleteHandler)
		app.GET("/api/donations/{donationId}/status", DonationStatusHandler)
		app.POST("/api/donations/process", ProcessPaymentHandler)
		app.POST("/api/donations/webhook", HelcimWebhookHandler)

		// Add more as needed for your app
		// Debug route: list embedded files
		var debugFilesHandler buffalo.Handler
		debugFilesHandler = func(c buffalo.Context) error {
			var out string
			err := fs.WalkDir(avrnpo.FS(), ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				out += path + "\n"
				return nil
			})
			if err != nil {
				return c.Render(500, r.String(err.Error()))
			}
			return c.Render(200, r.String(out))
		}

		// Debug routes
		app.GET("/debug/flash/{type}", DebugFlashHandler)

		// Skip CSRF protection for specific routes
		if ENV != "test" {
			app.Middleware.Skip(csrf.New, DonationInitializeHandler, DonationCompleteHandler, DonationStatusHandler, ProcessPaymentHandler, HelcimWebhookHandler, debugFilesHandler, DebugFlashHandler, UsersCreate)
		}
		app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthLanding, AuthNew, AuthCreate, blogResource.List, blogResource.Show, TeamHandler, ProjectsHandler, ContactHandler, DonateHandler, DonateUpdateAmountHandler, DonatePaymentHandler, DonationSuccessHandler, DonationFailedHandler, DonationInitializeHandler, ProcessPaymentHandler, HelcimWebhookHandler, debugFilesHandler, DebugFlashHandler)
		app.GET("/debug/files", debugFilesHandler)

		// Serve assets using Buffalo best practices
		// ServeFiles should be LAST as it's a catch-all route
		if ENV == "production" {
			// Production: use embedded assets
			app.ServeFiles("/", http.FS(avrnpo.FS()))
		} else {
			// Development/Test: serve from filesystem for hot reload
			app.ServeFiles("/", http.Dir("public"))
		}
	})

	return app
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

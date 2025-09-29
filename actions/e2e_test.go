package actions

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gobuffalo/suite/v4"
)

type E2ESuite struct {
	*suite.Action
}

func Test_E2ESuite(t *testing.T) {
	// Ensure test environment
	os.Setenv("GO_ENV", "test")

	// Reset app instance
	appOnce = sync.Once{}
	app = nil

	// Create app instance
	testApp := App()

	as := &E2ESuite{
		Action: suite.NewAction(testApp),
	}

	suite.Run(t, as)
}

func (as *E2ESuite) Test_HomePageLoad() {
	as.T().Run("home page loads successfully", func(t *testing.T) {
		// Create headless Chrome context
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			chromedp.NoDefaultBrowserCheck,
			chromedp.Flag("headless", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)
		allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancelAlloc()
		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		// Set timeout
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		// Navigate to home page and check title (assuming test suite starts server)
		var title string
		err := chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:3001/"),
			chromedp.Sleep(2*time.Second),
			chromedp.Title(&title),
		)
		as.NoError(err)
		as.Contains(title, "American Veterans Rebuilding") // Adjust based on actual page title
	})
}

func (as *E2ESuite) Test_UserAuthFlow() {
	as.T().Run("auth page loads successfully", func(t *testing.T) {
		// Create headless Chrome context
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			chromedp.NoDefaultBrowserCheck,
			chromedp.Flag("headless", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)
		allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancelAlloc()
		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		// Navigate to signup page and check for form
		err := chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:3001/auth/new"),
			chromedp.Sleep(2*time.Second),
			chromedp.WaitVisible(`input[name="user[email]"]`),
		)
		as.NoError(err)
	})
}

func (as *E2ESuite) Test_DonationFlow() {
	as.T().Run("donation page loads successfully", func(t *testing.T) {
		// Create headless Chrome context
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			chromedp.NoDefaultBrowserCheck,
			chromedp.Flag("headless", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)
		allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancelAlloc()
		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		// Navigate to donation page and check title
		var title string
		err := chromedp.Run(ctx,
			chromedp.Navigate("http://localhost:3001/donate"),
			chromedp.Sleep(2*time.Second),
			chromedp.Title(&title),
		)
		as.NoError(err)
		as.Contains(title, "Make a Donation") // Adjust based on actual page title
	})
}

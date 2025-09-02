package actions

import (
	"os"
	"testing"
)

// TestMain sets environment variables for the test suite to prevent
// real emails and ensure test mode. This runs before all tests in package actions.
func TestMain(m *testing.M) {
	// ensure test environment
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "test")
	}
	// disable real email sending during tests
	os.Setenv("EMAIL_ENABLED", "false")
	// set a default helcim verifier token for webhook signing tests
	if os.Getenv("HELCIM_WEBHOOK_VERIFIER_TOKEN") == "" {
		os.Setenv("HELCIM_WEBHOOK_VERIFIER_TOKEN", "test_verifier_token")
	}

	os.Exit(m.Run())
}

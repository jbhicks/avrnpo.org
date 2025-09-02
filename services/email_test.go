package services

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

// TestEmailService_SendDonationReceipt_Gmail is an optional integration test that
// is intentionally skipped by default to prevent sending real emails during
// development or CI. To run it locally you must set SMTP_USERNAME and
// SMTP_PASSWORD in a local .env and enable EMAIL_INTEGRATION_TESTS=true.
func TestEmailService_SendDonationReceipt_Gmail(t *testing.T) {
	if os.Getenv("EMAIL_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping external SMTP integration test. Set EMAIL_INTEGRATION_TESTS=true to enable locally.")
	}

	// Load environment variables from .env file
	_ = godotenv.Load("../.env")

	// Get Gmail credentials from environment (now loaded from .env)
	gmailUser := os.Getenv("SMTP_USERNAME")
	gmailAppPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("FROM_EMAIL")
	fromName := os.Getenv("FROM_NAME")

	if gmailUser == "" || gmailAppPassword == "" {
		t.Skip("Skipping Gmail email test - SMTP_USERNAME and SMTP_PASSWORD environment variables not set")
	}

	t.Logf("Using SMTP config - Host: %s, Port: %s, Username: %s, FromEmail: %s",
		os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"), gmailUser, fromEmail)
	// Create email service with Gmail configuration from .env
	emailService := &EmailService{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: gmailUser,
		SMTPPassword: gmailAppPassword,
		FromEmail:    fromEmail,
		FromName:     fromName,
		EmailEnabled: true, // explicitly enable for this integration test
	}

	// Create test donation receipt data
	testData := DonationReceiptData{
		DonorName:           "Test Donor",
		DonationAmount:      100.00,
		DonationType:        "One-time",
		TransactionID:       "TEST-" + time.Now().Format("20060102-150405"),
		DonationDate:        time.Now(),
		TaxDeductibleAmount: 100.00,
		OrganizationEIN:     "12-3456789",
		OrganizationName:    "American Veterans Real Estate Network",
		OrganizationAddress: "123 Test Street, Test City, TX 12345",
		DonorAddressLine1:   "456 Donor Lane",
		DonorAddressLine2:   "Apt 2B",
		DonorCity:           "Test City",
		DonorState:          "TX",
		DonorZip:            "12345",
	}

	// Send test email
	err := emailService.SendDonationReceipt(os.Getenv("TEST_EMAIL_RECIPIENT"), testData)
	require.NoError(t, err, "Failed to send test email")

	t.Logf("Test email sent successfully to %s with transaction ID: %s", os.Getenv("TEST_EMAIL_RECIPIENT"), testData.TransactionID)
}

func TestEmailService_isConfigured(t *testing.T) {
	tests := []struct {
		name     string
		service  *EmailService
		expected bool
	}{
		{
			name: "fully configured",
			service: &EmailService{
				SMTPHost:     "smtp.gmail.com",
				SMTPPort:     "587",
				SMTPUsername: "test@gmail.com",
				SMTPPassword: "password",
				FromEmail:    "test@gmail.com",
				FromName:     "Test",
			},
			expected: true,
		},
		{
			name: "missing SMTP host",
			service: &EmailService{
				SMTPHost:     "",
				SMTPPort:     "587",
				SMTPUsername: "test@gmail.com",
				SMTPPassword: "password",
				FromEmail:    "test@gmail.com",
				FromName:     "Test",
			},
			expected: false,
		},
		{
			name: "missing from email",
			service: &EmailService{
				SMTPHost:     "smtp.gmail.com",
				SMTPPort:     "587",
				SMTPUsername: "test@gmail.com",
				SMTPPassword: "password",
				FromEmail:    "",
				FromName:     "Test",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.isConfigured()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestEmailService_generateReceiptHTML(t *testing.T) {
	emailService := &EmailService{}

	testData := DonationReceiptData{
		DonorName:           "John Doe",
		DonationAmount:      250.00,
		DonationType:        "Monthly",
		TransactionID:       "TXN-123456",
		DonationDate:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		TaxDeductibleAmount: 250.00,
		OrganizationEIN:     "12-3456789",
		OrganizationName:    "Test Organization",
		OrganizationAddress: "123 Main St, City, ST 12345",
		DonorAddressLine1:   "456 Oak Ave",
		DonorAddressLine2:   "Suite 100",
		DonorCity:           "Anytown",
		DonorState:          "CA",
		DonorZip:            "90210",
	}

	html, err := emailService.generateReceiptHTML(testData)
	require.NoError(t, err)
	require.NotEmpty(t, html)

	// Check that key data is present in the HTML
	require.Contains(t, html, "John Doe")
	require.Contains(t, html, "$250.00")
	require.Contains(t, html, "TXN-123456")
	require.Contains(t, html, "Test Organization")
	require.Contains(t, html, "456 Oak Ave")
	require.Contains(t, html, "Suite 100")
	require.Contains(t, html, "Anytown, CA 90210")
}

func TestEmailService_generateReceiptText(t *testing.T) {
	emailService := &EmailService{}

	testData := DonationReceiptData{
		DonorName:           "Jane Smith",
		DonationAmount:      75.50,
		DonationType:        "One-time",
		TransactionID:       "TXN-789012",
		DonationDate:        time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC),
		TaxDeductibleAmount: 75.50,
		OrganizationEIN:     "98-7654321",
		OrganizationName:    "Test Charity",
		OrganizationAddress: "789 Elm St, Town, ST 54321",
		DonorAddressLine1:   "321 Pine St",
		DonorAddressLine2:   "",
		DonorCity:           "Somewhere",
		DonorState:          "NY",
		DonorZip:            "10001",
	}

	text := emailService.generateReceiptText(testData)
	require.NotEmpty(t, text)

	// Check that key data is present in the text
	require.Contains(t, text, "Jane Smith")
	require.Contains(t, text, "$75.50")
	require.Contains(t, text, "TXN-789012")
	require.Contains(t, text, "Test Charity")
	require.Contains(t, text, "321 Pine St")
	require.Contains(t, text, "Somewhere, NY 10001")
}

func TestEmailService_generateReceipt_IncludesSubscription(t *testing.T) {
	emailService := &EmailService{}

	next := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)

	testData := DonationReceiptData{
		DonorName:           "Subscription Donor",
		DonationAmount:      20.00,
		DonationType:        "Monthly",
		TransactionID:       "SUB-123",
		DonationDate:        time.Now(),
		SubscriptionID:      "SUB-ABC-123",
		NextBillingDate:     &next,
		OrganizationEIN:     "12-3456789",
		OrganizationName:    "Test Organization",
		OrganizationAddress: "123 Main St",
		DonorAddressLine1:   "456 Donor Rd",
		DonorCity:           "City",
		DonorState:          "ST",
		DonorZip:            "00000",
	}

	html, err := emailService.generateReceiptHTML(testData)
	require.NoError(t, err)
	require.NotEmpty(t, html)

	// Ensure subscription fields are present in HTML
	require.Contains(t, html, testData.SubscriptionID)
	require.Contains(t, html, next.Format("January 2, 2006"))

	// Log HTML for inspection (sanitized data used)
	t.Logf("Generated HTML:\n%s", html)

	text := emailService.generateReceiptText(testData)
	require.NotEmpty(t, text)
	require.Contains(t, text, testData.SubscriptionID)
	require.Contains(t, text, next.Format("January 2, 2006"))
	t.Logf("Generated Text:\n%s", text)
}

func TestEmailService_generateReceipt_ZeroNextBillingDate(t *testing.T) {
	emailService := &EmailService{}

	// Test the issue where NextBillingDate is zero value
	var zeroTime time.Time
	testData := DonationReceiptData{
		DonorName:           "Zero Date Donor",
		DonationAmount:      50.00,
		DonationType:        "Monthly",
		TransactionID:       "ZERO-123",
		DonationDate:        time.Now(),
		SubscriptionID:      "SUB-ZERO-123",
		NextBillingDate:     &zeroTime, // This reproduces the issue
		OrganizationEIN:     "12-3456789",
		OrganizationName:    "Test Organization",
		OrganizationAddress: "123 Main St",
		DonorAddressLine1:   "456 Donor Rd",
		DonorCity:           "City",
		DonorState:          "ST",
		DonorZip:            "00000",
	}

	html, err := emailService.generateReceiptHTML(testData)
	require.NoError(t, err)
	require.NotEmpty(t, html)

	// This should NOT contain "January 1, 0001" - that's the bug we're fixing
	require.NotContains(t, html, "January 1, 0001")

	// Instead it should show a reasonable fallback message
	require.Contains(t, html, "To be determined") // Or whatever fallback we implement

	text := emailService.generateReceiptText(testData)
	require.NotEmpty(t, text)
	require.NotContains(t, text, "January 1, 0001")
	require.Contains(t, text, "To be determined")

	t.Logf("Generated HTML with zero date:\n%s", html)
	t.Logf("Generated Text with zero date:\n%s", text)
}

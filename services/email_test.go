package services

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestEmailService_SendDonationReceipt_Gmail(t *testing.T) {
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
	err := emailService.SendDonationReceipt("joshua.brock.hicks@gmail.com", testData)
	require.NoError(t, err, "Failed to send test email")

	t.Logf("Test email sent successfully to joshua.brock.hicks@gmail.com with transaction ID: %s", testData.TransactionID)
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

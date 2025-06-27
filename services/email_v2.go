package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// EmailProvider represents different email sending methods
type EmailProvider string

const (
	ProviderSMTP     EmailProvider = "smtp"
	ProviderGmailAPI EmailProvider = "gmail_api"
)

// EmailService handles sending emails with multiple providers
type EmailService struct {
	Provider EmailProvider
	
	// SMTP Configuration
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	
	// Gmail API Configuration
	ServiceAccountFile string
	ClientID          string
	ClientSecret      string
	RefreshToken      string
	
	// Common Configuration
	FromEmail string
	FromName  string
}

// DonationReceiptData contains data for donation receipt emails
type DonationReceiptData struct {
	DonorName           string
	DonationAmount      float64
	DonationType        string
	TransactionID       string
	DonationDate        time.Time
	TaxDeductibleAmount float64
	OrganizationEIN     string
	OrganizationName    string
	OrganizationAddress string
}

// NewEmailService creates a new email service instance with auto-detection of provider
func NewEmailService() *EmailService {
	service := &EmailService{
		// SMTP Configuration
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		
		// Gmail API Configuration
		ServiceAccountFile: os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE"),
		ClientID:          os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret:      os.Getenv("GOOGLE_CLIENT_SECRET"),
		RefreshToken:      os.Getenv("GOOGLE_REFRESH_TOKEN"),
		
		// Common Configuration
		FromEmail: os.Getenv("FROM_EMAIL"),
		FromName:  os.Getenv("FROM_NAME"),
	}
	
	// Auto-detect provider based on available configuration
	if service.ServiceAccountFile != "" || (service.ClientID != "" && service.RefreshToken != "") {
		service.Provider = ProviderGmailAPI
	} else if service.SMTPHost != "" {
		service.Provider = ProviderSMTP
	}
	
	return service
}

// SendDonationReceipt sends a donation receipt email to the donor
func (e *EmailService) SendDonationReceipt(toEmail string, data DonationReceiptData) error {
	if !e.isConfigured() {
		return fmt.Errorf("email service not configured - missing environment variables")
	}
	
	switch e.Provider {
	case ProviderGmailAPI:
		return e.sendViaGmailAPI(toEmail, data)
	case ProviderSMTP:
		return e.sendViaSMTP(toEmail, data)
	default:
		return fmt.Errorf("no email provider configured")
	}
}

// isConfigured checks if email service is properly configured
func (e *EmailService) isConfigured() bool {
	switch e.Provider {
	case ProviderGmailAPI:
		return e.FromEmail != "" && (e.ServiceAccountFile != "" || 
			(e.ClientID != "" && e.ClientSecret != "" && e.RefreshToken != ""))
	case ProviderSMTP:
		return e.SMTPHost != "" && e.SMTPPort != "" && e.SMTPUsername != "" && 
			e.SMTPPassword != "" && e.FromEmail != ""
	default:
		return false
	}
}

// sendViaGmailAPI sends email using Gmail API
func (e *EmailService) sendViaGmailAPI(toEmail string, data DonationReceiptData) error {
	ctx := context.Background()
	
	var service *gmail.Service
	var err error
	
	if e.ServiceAccountFile != "" {
		// Service Account authentication
		service, err = e.createGmailServiceWithServiceAccount(ctx)
	} else {
		// OAuth2 authentication
		service, err = e.createGmailServiceWithOAuth2(ctx)
	}
	
	if err != nil {
		return fmt.Errorf("failed to create Gmail service: %v", err)
	}
	
	// Create email content
	htmlContent, textContent, err := e.generateEmailContent(data)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %v", err)
	}
	
	// Create MIME message
	subject := "Donation Receipt - American Veterans Rebuilding"
	mimeMessage := e.createMimeMessage(e.FromEmail, toEmail, subject, htmlContent, textContent)
	
	// Send via Gmail API
	message := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(mimeMessage)),
	}
	
	_, err = service.Users.Messages.Send("me", message).Do()
	if err != nil {
		return fmt.Errorf("failed to send email via Gmail API: %v", err)
	}
	
	return nil
}

// createGmailServiceWithServiceAccount creates Gmail service using service account
func (e *EmailService) createGmailServiceWithServiceAccount(ctx context.Context) (*gmail.Service, error) {
	credentials, err := os.ReadFile(e.ServiceAccountFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account file: %v", err)
	}
	
	config, err := google.JWTConfigFromJSON(credentials, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %v", err)
	}
	
	// For domain-wide delegation, set the subject to the email address
	config.Subject = e.FromEmail
	
	service, err := gmail.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %v", err)
	}
	
	return service, nil
}

// createGmailServiceWithOAuth2 creates Gmail service using OAuth2
func (e *EmailService) createGmailServiceWithOAuth2(ctx context.Context) (*gmail.Service, error) {
	config := &oauth2.Config{
		ClientID:     e.ClientID,
		ClientSecret: e.ClientSecret,
		Scopes:       []string{gmail.GmailSendScope},
		Endpoint:     google.Endpoint,
	}
	
	token := &oauth2.Token{
		RefreshToken: e.RefreshToken,
	}
	
	client := config.Client(ctx, token)
	
	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %v", err)
	}
	
	return service, nil
}

// sendViaSMTP sends email using traditional SMTP (existing implementation)
func (e *EmailService) sendViaSMTP(toEmail string, data DonationReceiptData) error {
	// Generate email content
	htmlContent, textContent, err := e.generateEmailContent(data)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %v", err)
	}
	
	// SMTP authentication
	auth := smtp.PlainAuth("", e.SMTPUsername, e.SMTPPassword, e.SMTPHost)
	
	// Email headers and content
	subject := "Donation Receipt - American Veterans Rebuilding"
	mimeMessage := e.createMimeMessage(e.FromEmail, toEmail, subject, htmlContent, textContent)
	
	// Send email
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	err = smtp.SendMail(addr, auth, e.FromEmail, []string{toEmail}, []byte(mimeMessage))
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %v", err)
	}
	
	return nil
}

// generateEmailContent creates both HTML and text versions of the receipt
func (e *EmailService) generateEmailContent(data DonationReceiptData) (string, string, error) {
	// HTML template (existing implementation)
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Donation Receipt</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; border-bottom: 2px solid #2563eb; padding-bottom: 20px; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #2563eb; }
        .receipt-details { background-color: #f8fafc; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .amount { font-size: 24px; font-weight: bold; color: #059669; text-align: center; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; font-size: 14px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">🇺🇸 American Veterans Rebuilding</div>
        <p>Rebuilding the American Veteran's Self, Family and Community</p>
    </div>
    
    <h2>Thank You for Your Donation!</h2>
    
    <p>Dear {{.DonorName}},</p>
    
    <p>Thank you for your generous donation to American Veterans Rebuilding. Your support helps us continue our mission of improving the lives of American Veterans through Technical Training, Occupational Licensing, Home Ownership Options, and Professional Networking.</p>
    
    <div class="receipt-details">
        <h3>Donation Details</h3>
        <p><strong>Donation Amount:</strong> <span class="amount">${{printf "%.2f" .DonationAmount}}</span></p>
        <p><strong>Donation Type:</strong> {{.DonationType}}</p>
        <p><strong>Transaction ID:</strong> {{.TransactionID}}</p>
        <p><strong>Date:</strong> {{.DonationDate.Format "January 2, 2006"}}</p>
        {{if .TaxDeductibleAmount}}
        <p><strong>Tax Deductible Amount:</strong> ${{printf "%.2f" .TaxDeductibleAmount}}</p>
        {{end}}
    </div>
    
    <div class="footer">
        <p><strong>Tax Information:</strong> American Veterans Rebuilding is a 501(c)(3) non-profit organization. {{if .OrganizationEIN}}EIN: {{.OrganizationEIN}}.{{end}} Your donation is tax-deductible to the full extent allowed by law.</p>
        
        {{if .OrganizationAddress}}
        <p><strong>Organization Address:</strong><br>{{.OrganizationAddress}}</p>
        {{end}}
        
        <p>This receipt serves as your official record of donation. Please keep it for your tax records.</p>
        
        <p><em>Together, we rebuild lives and strengthen communities. Thank you for supporting our veterans.</em></p>
    </div>
</body>
</html>`
	
	// Text template (simple version)
	textTemplate := `American Veterans Rebuilding - Donation Receipt

Dear {{.DonorName}},

Thank you for your generous donation to American Veterans Rebuilding.

Donation Details:
- Amount: ${{printf "%.2f" .DonationAmount}}
- Type: {{.DonationType}}
- Transaction ID: {{.TransactionID}}
- Date: {{.DonationDate.Format "January 2, 2006"}}
{{if .TaxDeductibleAmount}}- Tax Deductible Amount: ${{printf "%.2f" .TaxDeductibleAmount}}{{end}}

Tax Information: American Veterans Rebuilding is a 501(c)(3) non-profit organization. {{if .OrganizationEIN}}EIN: {{.OrganizationEIN}}.{{end}} Your donation is tax-deductible to the full extent allowed by law.

{{if .OrganizationAddress}}Organization Address:
{{.OrganizationAddress}}{{end}}

This receipt serves as your official record of donation. Please keep it for your tax records.

Together, we rebuild lives and strengthen communities. Thank you for supporting our veterans.

American Veterans Rebuilding
Rebuilding the American Veteran's Self, Family and Community`
	
	// Parse and execute HTML template
	htmlTmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse HTML template: %v", err)
	}
	
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute HTML template: %v", err)
	}
	
	// Parse and execute text template
	textTmpl, err := template.New("text").Parse(textTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse text template: %v", err)
	}
	
	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute text template: %v", err)
	}
	
	return htmlBuf.String(), textBuf.String(), nil
}

// createMimeMessage creates a MIME multipart message with both HTML and text
func (e *EmailService) createMimeMessage(from, to, subject, htmlContent, textContent string) string {
	boundary := "boundary-avr-receipt-email"
	
	mime := fmt.Sprintf(`From: %s <%s>
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="%s"

--%s
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 8bit

%s

--%s
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: 8bit

%s

--%s--`, e.FromName, from, to, subject, boundary, boundary, textContent, boundary, htmlContent, boundary)
	
	return mime
}

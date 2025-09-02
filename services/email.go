package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"time"
)

// SMTPClient defines an interface for sending mail. This allows injecting
// a mock client in tests to prevent network calls.
type SMTPClient interface {
	SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// realSMTPClient wraps the standard library smtp.SendMail
type realSMTPClient struct{}

func (r *realSMTPClient) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, a, from, to, msg)
}

// EmailService handles sending emails
type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	EmailEnabled bool // controls whether emails are actually sent
	client       SMTPClient
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	// In test mode, return a service with EmailEnabled=false and a nil client to avoid network calls
	if os.Getenv("GO_ENV") == "test" {
		return &EmailService{EmailEnabled: false, FromEmail: "test@example.com", FromName: "AVRNPO Test", client: nil}
	}
	// Determine default for EMAIL_ENABLED based on GO_ENV
	enabledStr := os.Getenv("EMAIL_ENABLED")
	if enabledStr == "" {
		goEnv := os.Getenv("GO_ENV")
		if goEnv == "production" {
			enabledStr = "true"
		} else {
			enabledStr = "false"
		}
	}
	emailEnabled := enabledStr == "true" || enabledStr == "1" || enabledStr == "yes"

	svc := &EmailService{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		FromName:     os.Getenv("FROM_NAME"),
		EmailEnabled: emailEnabled,
		client:       &realSMTPClient{},
	}
	return svc
}

// DonationReceiptData contains data for donation receipt emails
type DonationReceiptData struct {
	DonorName           string
	DonationAmount      float64
	DonationType        string
	SubscriptionID      string
	CustomerID          string // Helcim Customer ID for subscription management
	NextBillingDate     *time.Time
	TransactionID       string
	DonationDate        time.Time
	TaxDeductibleAmount float64
	OrganizationEIN     string
	OrganizationName    string
	OrganizationAddress string
	DonorAddressLine1   string
	DonorAddressLine2   string
	DonorCity           string
	DonorState          string
	DonorZip            string
}

// ContactFormData contains data for contact form submissions
type ContactFormData struct {
	Name           string
	Email          string
	Subject        string
	Message        string
	SubmissionDate time.Time
}

// SendDonationReceipt sends a donation receipt email to the donor
func (e *EmailService) SendDonationReceipt(toEmail string, data DonationReceiptData) error {
	if !e.isConfigured() {
		fmt.Printf("[EMAIL] Service not configured - missing SMTP environment variables\n")
		return fmt.Errorf("email service not configured - missing environment variables")
	}

	// Generate email content
	subject := fmt.Sprintf("Thank you for your donation to %s", data.OrganizationName)
	htmlBody, err := e.generateReceiptHTML(data)
	if err != nil {
		return fmt.Errorf("error generating email HTML: %v", err)
	}

	textBody := e.generateReceiptText(data)

	// Send email with BCC to michael@avrnpo.org
	bccEmails := []string{"michael@avrnpo.org"}
	return e.sendEmailWithBCC(toEmail, subject, htmlBody, textBody, bccEmails)
}

// SendContactNotification sends a contact form notification to the organization
func (e *EmailService) SendContactNotification(contactData ContactFormData) error {
	if !e.isConfigured() {
		return fmt.Errorf("email service not configured - missing environment variables")
	}

	// Send to organization email
	toEmail := "michael@avrnpo.org"
	subject := fmt.Sprintf("New Contact Form Submission: %s", contactData.Subject)
	htmlBody, err := e.generateContactNotificationHTML(contactData)
	if err != nil {
		return fmt.Errorf("error generating email HTML: %v", err)
	}

	textBody := e.generateContactNotificationText(contactData)

	// Send email
	return e.sendEmail(toEmail, subject, htmlBody, textBody)
}

// isConfigured checks if the email service has all required configuration
func (e *EmailService) isConfigured() bool {
	return e.SMTPHost != "" &&
		e.SMTPPort != "" &&
		e.SMTPUsername != "" &&
		e.SMTPPassword != "" &&
		e.FromEmail != ""
}

// isValidNextBillingDate checks if NextBillingDate is not nil and not zero time
func (e *EmailService) isValidNextBillingDate(data DonationReceiptData) bool {
	return data.NextBillingDate != nil && !data.NextBillingDate.IsZero()
}

// formatNextBillingDate formats NextBillingDate with fallback for zero dates
func (e *EmailService) formatNextBillingDate(data DonationReceiptData) string {
	if e.isValidNextBillingDate(data) {
		return data.NextBillingDate.Format("January 2, 2006")
	}
	return "To be determined"
}

// generateReceiptHTML creates HTML email content for donation receipt
func (e *EmailService) generateReceiptHTML(data DonationReceiptData) (string, error) {
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Donation Receipt</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #fff; color: #333; padding: 20px; text-align: center; border-bottom: 1px solid #ddd; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .receipt-details { background-color: #fff; padding: 15px; border: 1px solid #ddd; margin: 20px 0; }
        .amount { font-size: 24px; font-weight: bold; color: #dc2626; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .logo { max-width: 150px; height: auto; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <img src="data:image/avif;base64,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" alt="AVR NPO Logo" class="logo" style="max-width: 150px; height: auto; margin-bottom: 10px;" />
            <h1>{{.OrganizationName}}</h1>
            <p>Thank you for your generous donation!</p>
        </div>
        
        <div class="content">
            <h2>Dear {{.DonorName}},</h2>
            <p>
                Thank you for your generous donation to {{.OrganizationName}}. 
            </p>
            <div class="donor-address">
                <strong>Donor Address:</strong><br>
                {{.DonorAddressLine1}}
                {{if .DonorAddressLine2}}, {{.DonorAddressLine2}}{{end}}<br>
                {{.DonorCity}}, {{.DonorState}} {{.DonorZip}}
            </div>
                Your support helps us continue our mission of supporting combat veterans 
                through housing projects, skills training, and community building programs.
            </p>
            
            <div class="receipt-details">
                <h3>Donation Receipt</h3>
                <p><strong>Transaction ID:</strong> {{.TransactionID}}</p>
                <p><strong>Date:</strong> {{.DonationDate.Format "January 2, 2006"}}</p>
				<p><strong>Donation Type:</strong> {{.DonationType}}</p>
				<p><strong>Amount:</strong> <span class="amount">${{printf "%.2f" .DonationAmount}}</span></p>
				{{if .SubscriptionID}}
				<p><strong>Subscription ID:</strong> {{.SubscriptionID}}</p>
				{{end}}
				{{if .CustomerID}}
				<p><strong>Customer ID:</strong> {{.CustomerID}}</p>
				{{end}}
				{{if .NextBillingDate}}
				{{if .NextBillingDate.IsZero}}
				<p><strong>Next Billing Date:</strong> To be determined</p>
				{{else}}
				<p><strong>Next Billing Date:</strong> {{.NextBillingDate.Format "January 2, 2006"}}</p>
				{{end}}
				{{end}}
				{{if ne .TaxDeductibleAmount .DonationAmount}}
				<p><strong>Tax Deductible Amount:</strong> ${{printf "%.2f" .TaxDeductibleAmount}}</p>
				{{end}}
            </div>
            
            {{if eq .DonationType "Monthly"}}
            <h3>Subscription Management</h3>
            <p>
                Your monthly recurring donation will automatically process on the same day each month. 
                To modify the amount, change frequency, or cancel your subscription, please contact us at 
                <strong>michael@avrnpo.org</strong> and reference your <strong>Customer ID: {{.CustomerID}}</strong> 
                in your message.
            </p>
            {{end}}
            
            <h3>Tax Information</h3>
            <p>
                {{.OrganizationName}} is a registered 501(c)(3) non-profit organization. 
                Your donation is tax-deductible to the full extent allowed by law. 
                No goods or services were provided in exchange for this donation.
            </p>
            {{if .OrganizationEIN}}
            <p><strong>Tax ID (EIN):</strong> {{.OrganizationEIN}}</p>
            {{end}}
            
            <h3>How Your Donation Helps</h3>
            <p>
                Your contribution directly supports:
            </p>
            <ul>
                <li>Housing projects providing affordable homeownership for veteran families</li>
                <li>Technical training programs for professional certifications</li>
                <li>Community building and networking opportunities</li>
                <li>Program operations and veteran support services</li>
            </ul>
            
            <p>
                We'll keep you updated on the impact your donation is making. 
                If you have any questions about your donation or our programs, 
                please don't hesitate to contact us.
            </p>
        </div>
        
        <div class="footer">
            <p>{{.OrganizationName}}</p>
            {{if .OrganizationAddress}}
            <p>{{.OrganizationAddress}}</p>
            {{end}}
            <p>This is an automated receipt. Please save this for your tax records.</p>
        </div>
    </div>
</body>
</html>
`

	tmpl, err := template.New("receipt").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateReceiptHTMLForTool is an exported wrapper used by small tools
// to produce receipt HTML without sending email. It reuses the internal
// template generation logic.
func (e *EmailService) GenerateReceiptHTMLForTool(data DonationReceiptData) (string, error) {
	return e.generateReceiptHTML(data)
}

// generateReceiptText creates plain text email content for donation receipt
func (e *EmailService) generateReceiptText(data DonationReceiptData) string {
	return fmt.Sprintf(`
Dear %s,

Thank you for your generous donation to %s!

DONATION RECEIPT
Transaction ID: %s
Date: %s
Donation Type: %s
Amount: $%.2f

Subscription ID: %s
Customer ID: %s
Next Billing Date: %s

RECURRING SUBSCRIPTION MANAGEMENT
Your subscription Customer ID is %s. Please reference this ID when 
contacting us to cancel or modify your recurring donation.
Email: michael@avrnpo.org

Donor Address:
%s
%s
%s, %s %s

TAX INFORMATION
%s is a registered 501(c)(3) non-profit organization. 
Your donation is tax-deductible to the full extent allowed by law. 
No goods or services were provided in exchange for this donation.
%s

HOW YOUR DONATION HELPS
Your contribution directly supports:
- Housing projects providing affordable homeownership for veteran families
- Technical training programs for professional certifications  
- Community building and networking opportunities
- Program operations and veteran support services

We'll keep you updated on the impact your donation is making. 
If you have any questions, please contact us at michael@avrnpo.org.

Thank you for supporting our mission!

%s
%s

This is an automated receipt. Please save this for your tax records.
`,
		data.DonorName,
		data.OrganizationName,
		data.TransactionID,
		data.DonationDate.Format("January 2, 2006"),
		data.DonationType,
		data.DonationAmount,
		data.SubscriptionID,
		data.CustomerID,
		func() string {
			if data.NextBillingDate != nil && !data.NextBillingDate.IsZero() {
				return data.NextBillingDate.Format("January 2, 2006")
			}
			return "To be determined"
		}(),
		data.CustomerID,
		data.DonorAddressLine1,
		data.DonorAddressLine2,
		data.DonorCity,
		data.DonorState,
		data.DonorZip,
		data.OrganizationName,
		func() string {
			if data.OrganizationEIN != "" {
				return fmt.Sprintf("Tax ID (EIN): %s", data.OrganizationEIN)
			}
			return ""
		}(),
		data.OrganizationName,
		data.OrganizationAddress,
	)
}

// sendEmail sends an email using SMTP
func (e *EmailService) sendEmail(toEmail, subject, htmlBody, textBody string) error {
	return e.sendEmailWithBCC(toEmail, subject, htmlBody, textBody, nil)
}

// sendEmailWithBCC sends an email using SMTP with BCC recipients
func (e *EmailService) sendEmailWithBCC(toEmail, subject, htmlBody, textBody string, bccEmails []string) error {
	// Create message with both HTML and text parts
	message := fmt.Sprintf(`To: %s
From: %s <%s>
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain; charset=UTF-8

%s

--boundary123
Content-Type: text/html; charset=UTF-8

%s

--boundary123--
`, toEmail, e.FromName, e.FromEmail, subject, textBody, htmlBody)

	// If email sending is disabled, log and return without sending
	if !e.EmailEnabled {
		// Write a short log to stdout so developers can inspect locally
		fmt.Printf("[EMAIL_DISABLED] To: %s Subject: %s\nPreview: %.200s\n", toEmail, subject, textBody)
		return nil
	}

	fmt.Printf("[EMAIL] Attempting to send email to %s: %s\n", toEmail, subject)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", e.SMTPUsername, e.SMTPPassword, e.SMTPHost)
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)

	// Build recipient list (TO + BCC)
	recipients := []string{toEmail}
	if bccEmails != nil {
		recipients = append(recipients, bccEmails...)
	}

	// Send email using injected client
	if e.client == nil {
		// fallback to real client
		e.client = &realSMTPClient{}
	}
	err := e.client.SendMail(addr, auth, e.FromEmail, recipients, []byte(message))
	if err != nil {
		fmt.Printf("[EMAIL] Failed to send email to %s: %v\n", toEmail, err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	fmt.Printf("[EMAIL] Successfully sent email to %s: %s\n", toEmail, subject)
	return nil
}

// generateContactNotificationHTML creates HTML email content for contact notifications
func (e *EmailService) generateContactNotificationHTML(data ContactFormData) (string, error) {
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Contact Form Submission</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #fff; color: #333; padding: 20px; text-align: center; border-bottom: 1px solid #ddd; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .form-details { background-color: #fff; padding: 15px; border: 1px solid #ddd; margin: 20px 0; }
        .message-content { background-color: #f5f5f5; padding: 15px; border-left: 4px solid #ffb627; margin: 15px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>New Contact Form Submission</h1>
            <p>American Veterans Rebuilding</p>
        </div>
        
        <div class="content">
            <div class="form-details">
                <h3>Contact Details</h3>
                <p><strong>Name:</strong> {{.Name}}</p>
                <p><strong>Email:</strong> {{.Email}}</p>
                <p><strong>Subject:</strong> {{.Subject}}</p>
                <p><strong>Submitted:</strong> {{.SubmissionDate.Format "January 2, 2006 at 3:04 PM"}}</p>
            </div>
            
            <div class="message-content">
                <h3>Message</h3>
                <p>{{.Message}}</p>
            </div>
            
            <p><strong>Reply to:</strong> <a href="mailto:{{.Email}}">{{.Email}}</a></p>
        </div>
        
        <div class="footer">
            <p>This message was sent from the AVRNPO.org contact form</p>
        </div>
    </div>
</body>
</html>
`

	tmpl, err := template.New("contact").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateContactNotificationText creates plain text email content for contact notifications
func (e *EmailService) generateContactNotificationText(data ContactFormData) string {
	return fmt.Sprintf(`
New Contact Form Submission
American Veterans Rebuilding

CONTACT DETAILS
Name: %s
Email: %s
Subject: %s
Submitted: %s

MESSAGE
%s

Reply to: %s

This message was sent from the AVRNPO.org contact form.
`,
		data.Name,
		data.Email,
		data.Subject,
		data.SubmissionDate.Format("January 2, 2006 at 3:04 PM"),
		data.Message,
		data.Email,
	)
}

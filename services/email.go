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
	ContactEmail string // configurable contact form recipient email
	EmailEnabled bool   // controls whether emails are actually sent
	client       SMTPClient
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	// In test mode, return a service with EmailEnabled=false and a nil client to avoid network calls
	if os.Getenv("GO_ENV") == "test" {
		return &EmailService{
			EmailEnabled: false,
			FromEmail:    "test@example.com",
			FromName:     "AVRNPO Test",
			ContactEmail: "test-contact@example.com",
			client:       nil,
		}
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

	// Get contact email from environment with fallback
	contactEmail := os.Getenv("CONTACT_EMAIL")
	if contactEmail == "" {
		contactEmail = "AmericanVeteransRebuilding@avrnpo.org" // Default to match displayed email
	}

	svc := &EmailService{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		FromName:     os.Getenv("FROM_NAME"),
		ContactEmail: contactEmail,
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
	ContactEmail        string // configurable contact email for support
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
	fmt.Printf("[EMAIL_SERVICE] Starting donation receipt for %s - Amount: $%.2f, Type: %s\n",
		toEmail, data.DonationAmount, data.DonationType)

	if !e.isConfigured() {
		fmt.Printf("[EMAIL_SERVICE] Configuration check failed - missing SMTP environment variables\n")
		return fmt.Errorf("email service not configured - missing environment variables")
	}
	fmt.Printf("[EMAIL_SERVICE] Configuration validated for donation receipt\n")

	// Inject the contact email into the data
	data.ContactEmail = e.ContactEmail
	fmt.Printf("[EMAIL_SERVICE] Contact email injected: %s\n", data.ContactEmail)

	// Generate email content with timing
	subject := fmt.Sprintf("Thank you for your donation to %s", data.OrganizationName)
	fmt.Printf("[EMAIL_SERVICE] Generated donation receipt subject: %s\n", subject)

	htmlBody, err := e.generateReceiptHTML(data)
	if err != nil {
		fmt.Printf("[EMAIL_SERVICE] Failed to generate donation receipt HTML: %v\n", err)
		return fmt.Errorf("error generating email HTML: %v", err)
	}

	textBody := e.generateReceiptText(data)

	// Log content metrics
	htmlSize := len(htmlBody)
	textSize := len(textBody)
	fmt.Printf("[EMAIL_SERVICE] Generated receipt content - HTML: %d bytes, Text: %d bytes\n",
		htmlSize, textSize)

	// Send email with BCC to michael@avrnpo.org (keep this for now for receipt tracking)
	bccEmails := []string{"michael@avrnpo.org"}
	fmt.Printf("[EMAIL_SERVICE] Donation receipt BCC recipients: %v\n", bccEmails)

	return e.sendEmailWithBCC(toEmail, subject, htmlBody, textBody, bccEmails)
}

// SendContactNotification sends a contact form notification to the organization
func (e *EmailService) SendContactNotification(contactData ContactFormData) error {
	fmt.Printf("[EMAIL_SERVICE] Starting contact notification for submission from %s (%s)\n",
		contactData.Name, contactData.Email)

	if !e.isConfigured() {
		fmt.Printf("[EMAIL_SERVICE] Configuration check failed - missing SMTP environment variables\n")
		return fmt.Errorf("email service not configured - missing environment variables")
	}
	fmt.Printf("[EMAIL_SERVICE] Configuration validated successfully\n")

	// Send to configured contact email
	toEmail := e.ContactEmail
	fmt.Printf("[EMAIL_SERVICE] Contact notification recipient: %s\n", toEmail)

	subject := fmt.Sprintf("New Contact Form Submission: %s", contactData.Subject)
	fmt.Printf("[EMAIL_SERVICE] Generated subject: %s\n", subject)

	// Generate email content with timing
	htmlBody, err := e.generateContactNotificationHTML(contactData)
	if err != nil {
		fmt.Printf("[EMAIL_SERVICE] Failed to generate HTML content: %v\n", err)
		return fmt.Errorf("error generating email HTML: %v", err)
	}

	textBody := e.generateContactNotificationText(contactData)

	// Log content metrics
	htmlSize := len(htmlBody)
	textSize := len(textBody)
	fmt.Printf("[EMAIL_SERVICE] Generated email content - HTML: %d bytes, Text: %d bytes\n",
		htmlSize, textSize)

	// Send email
	fmt.Printf("[EMAIL_SERVICE] Initiating email send for contact notification\n")
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
				<strong>{{.ContactEmail}}</strong> and reference your <strong>Customer ID: {{.CustomerID}}</strong> 
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
Email: %s

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
If you have any questions, please contact us at %s.

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
		data.ContactEmail,
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
		data.ContactEmail,
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
	startTime := time.Now()
	fmt.Printf("[EMAIL_SMTP] Starting email send operation at %s\n", startTime.Format("2006-01-02 15:04:05"))

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

	// Log message statistics
	messageSize := len(message)
	fmt.Printf("[EMAIL_SMTP] Message composed - Size: %d bytes, To: %s, From: %s <%s>\n",
		messageSize, toEmail, e.FromName, e.FromEmail)

	// Build and log recipient list
	recipients := []string{toEmail}
	if bccEmails != nil {
		recipients = append(recipients, bccEmails...)
		fmt.Printf("[EMAIL_SMTP] Recipients: Primary=%s, BCC=%v, Total=%d\n",
			toEmail, bccEmails, len(recipients))
	} else {
		fmt.Printf("[EMAIL_SMTP] Recipients: Primary=%s, BCC=none, Total=%d\n",
			toEmail, len(recipients))
	}

	// If email sending is disabled, log and return without sending
	if !e.EmailEnabled {
		elapsed := time.Since(startTime)
		fmt.Printf("[EMAIL_DISABLED] Email sending disabled - Duration: %v\n", elapsed)
		fmt.Printf("[EMAIL_DISABLED] To: %s Subject: %s\nPreview: %.200s\n", toEmail, subject, textBody)
		return nil
	}

	// Log SMTP connection attempt
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	fmt.Printf("[EMAIL_SMTP] Attempting SMTP connection to %s with user %s\n", addr, e.SMTPUsername)

	// Connect to SMTP server with timing
	authStart := time.Now()
	auth := smtp.PlainAuth("", e.SMTPUsername, e.SMTPPassword, e.SMTPHost)
	authDuration := time.Since(authStart)
	fmt.Printf("[EMAIL_SMTP] SMTP auth prepared in %v\n", authDuration)

	// Send email using injected client
	if e.client == nil {
		fmt.Printf("[EMAIL_SMTP] No injected client, using real SMTP client\n")
		e.client = &realSMTPClient{}
	}

	fmt.Printf("[EMAIL_SMTP] Initiating SMTP send to %d recipients\n", len(recipients))
	sendStart := time.Now()
	err := e.client.SendMail(addr, auth, e.FromEmail, recipients, []byte(message))
	sendDuration := time.Since(sendStart)
	totalDuration := time.Since(startTime)

	if err != nil {
		fmt.Printf("[EMAIL_SMTP] SEND FAILED - Duration: %v, Total: %v, Error: %v\n",
			sendDuration, totalDuration, err)
		fmt.Printf("[EMAIL_SMTP] Failed details - To: %s, Size: %d bytes, Recipients: %d\n",
			toEmail, messageSize, len(recipients))
		return fmt.Errorf("failed to send email: %v", err)
	}

	fmt.Printf("[EMAIL_SMTP] SEND SUCCESS - Send: %v, Total: %v, Size: %d bytes\n",
		sendDuration, totalDuration, messageSize)
	fmt.Printf("[EMAIL_SMTP] Successfully delivered to %s (+ %d BCC recipients)\n",
		toEmail, len(recipients)-1)

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

package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"time"
)

// EmailService handles sending emails
type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		FromName:     os.Getenv("FROM_NAME"),
	}
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
	DonorAddressLine1   string
	DonorAddressLine2   string
	DonorCity           string
	DonorState          string
	DonorZip            string
}

// SendDonationReceipt sends a donation receipt email to the donor
func (e *EmailService) SendDonationReceipt(toEmail string, data DonationReceiptData) error {
	if !e.isConfigured() {
		return fmt.Errorf("email service not configured - missing environment variables")
	}

	// Generate email content
	subject := fmt.Sprintf("Thank you for your donation to %s", data.OrganizationName)
	htmlBody, err := e.generateReceiptHTML(data)
	if err != nil {
		return fmt.Errorf("error generating email HTML: %v", err)
	}

	textBody := e.generateReceiptText(data)

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
        .header { background-color: #ffb627; color: #fff; padding: 20px; text-align: center; }
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
                {{if ne .TaxDeductibleAmount .DonationAmount}}
                <p><strong>Tax Deductible Amount:</strong> ${{printf "%.2f" .TaxDeductibleAmount}}</p>
                {{end}}
            </div>
            
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
If you have any questions, please contact us at info@avrnpo.org.

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

	// Connect to SMTP server
	auth := smtp.PlainAuth("", e.SMTPUsername, e.SMTPPassword, e.SMTPHost)
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)

	// Send email
	err := smtp.SendMail(addr, auth, e.FromEmail, []string{toEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

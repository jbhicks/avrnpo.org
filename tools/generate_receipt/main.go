package main

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"time"

	"avrnpo.org/services"
)

func main() {
	next := time.Now().AddDate(0, 1, 0)
	data := services.DonationReceiptData{
		DonorName:           "Test Donor",
		DonationAmount:      50.00,
		DonationType:        "Monthly",
		SubscriptionID:      "SUB-EXAMPLE-123",
		NextBillingDate:     &next,
		TransactionID:       "TEST-RECEIPT-001",
		DonationDate:        time.Now(),
		TaxDeductibleAmount: 50.00,
		OrganizationEIN:     "12-3456789",
		OrganizationName:    "American Veterans Rebuilding",
		OrganizationAddress: "123 Test St, Test City, ST 00000",
		DonorAddressLine1:   "456 Donor Lane",
		DonorAddressLine2:   "",
		DonorCity:           "Testville",
		DonorState:          "TS",
		DonorZip:            "00000",
	}

	htmlTemplate := `<!DOCTYPE html>
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
				{{if .SubscriptionID}}
				<p><strong>Subscription ID:</strong> {{.SubscriptionID}}</p>
				{{end}}
				{{if .NextBillingDate}}
				<p><strong>Next Billing Date:</strong> {{.NextBillingDate.Format "January 2, 2006"}}</p>
				{{end}}
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
</html>`

	tmpl, err := template.New("receipt").Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("parse template error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Fatalf("execute template error: %v", err)
	}

	out := "/tmp/donation_receipt.html"
	if err := os.WriteFile(out, buf.Bytes(), 0644); err != nil {
		log.Fatalf("write file error: %v", err)
	}

	log.Printf("Wrote receipt HTML to %s", out)
}

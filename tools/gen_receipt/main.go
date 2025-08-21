package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"avrnpo.org/services"
)

func main() {
	emailSvc := services.NewEmailService()
	next := time.Now().AddDate(0, 1, 0)
	data := services.DonationReceiptData{
		DonorName:           "Demo Donor",
		DonationAmount:      50.00,
		DonationType:        "Monthly",
		SubscriptionID:      "SUB-EXAMPLE-001",
		NextBillingDate:     &next,
		TransactionID:       "DEMO-0001",
		DonationDate:        time.Now(),
		TaxDeductibleAmount: 50.00,
		OrganizationEIN:     "12-3456789",
		OrganizationName:    "American Veterans Rebuilding",
		OrganizationAddress: "123 Demo St, Demo City, ST 00000",
		DonorAddressLine1:   "456 Donor Rd",
		DonorAddressLine2:   "",
		DonorCity:           "DemoCity",
		DonorState:          "ST",
		DonorZip:            "00000",
	}

	html, err := emailSvc.GenerateReceiptHTMLForTool(data)
	if err != nil {
		log.Fatalf("failed to generate html: %v", err)
	}

	outPath := "/tmp/donation_receipt.html"
	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(html)
	if err != nil {
		log.Fatalf("failed to write file: %v", err)
	}

	fmt.Printf("Wrote HTML to %s\n", outPath)
}

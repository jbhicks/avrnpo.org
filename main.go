package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp" // Add this import for the standard library SMTP client
	"os"
	"time"

	gosmtp "github.com/emersion/go-smtp" // Rename this import to avoid conflicts
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Update contact form struct to include referral source
type ContactForm struct {
	FirstName      string `form:"fname"`
	LastName       string `form:"lname"`
	Email          string `form:"email"`
	ReferralSource string `form:"referralSource"`
	Message        string `form:"message"`
}

type HelcimInitializeResponse struct {
	CheckoutToken string `json:"checkoutToken"`
	SecretToken   string `json:"secretToken"` // Add SecretToken field
}

// DonationInfo holds information about a donor and their donation
type DonationInfo struct {
	FirstName string
	LastName  string
	Email     string
	Purpose   string
	Referral  string
}

// Define our internal mail message structure
type Message struct {
	From    string
	To      []string
	Subject string
	Body    string
}

// Queue to store messages
var messageQueue = make(chan Message, 100)

// Simple backend that implements emersion/go-smtp Backend interface
type Backend struct{}

func (bkd *Backend) NewSession(_ *gosmtp.Conn) (gosmtp.Session, error) {
	return &Session{}, nil
}

// Session that implements emersion/go-smtp Session interface
type Session struct {
	From        string
	To          []string
	MessageData []byte // Renamed from Data to MessageData
}

func (s *Session) AuthPlain(username, password string) error {
	return nil // No auth needed for internal usage
}

func (s *Session) Mail(from string, opts *gosmtp.MailOptions) error {
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *gosmtp.RcptOptions) error {
	s.To = append(s.To, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := io.ReadAll(r); err != nil {
		return err
	} else {
		s.MessageData = b // Updated to use MessageData instead of Data
		return nil
	}
}

func (s *Session) Reset() {
	s.From = ""
	s.To = []string{}
	s.MessageData = nil // Updated to use MessageData instead of Data
}

func (s *Session) Logout() error {
	return nil
}

// Start our internal SMTP server
func startSMTPServer() {
	be := &Backend{}

	s := gosmtp.NewServer(be)
	s.Addr = "localhost:1025" // Use a non-standard port
	s.Domain = "localhost"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	log.Println("Starting internal SMTP server at", s.Addr)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Process messages in the queue
	go processMessageQueue()
}

// Process messages in the queue
func processMessageQueue() {
	for msg := range messageQueue {
		log.Printf("Sending email to: %v\nSubject: %s\n", msg.To, msg.Subject)

		// Create email message
		message := fmt.Sprintf("From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s\r\n", msg.From, msg.To[0], msg.Subject, msg.Body)

		// Connect to our local SMTP server
		// No authentication needed for local server
		err := smtp.SendMail(
			"localhost:1025",
			nil, // No auth needed for local server
			msg.From,
			msg.To,
			[]byte(message),
		)

		if err != nil {
			log.Printf("Error sending email: %s\n", err)
		} else {
			log.Println("Email sent successfully")
		}
	}
}

// Send email function now simply queues the message
func sendEmail(to []string, subject string, body string, fromEmail string) error {
	messageQueue <- Message{
		From:    fromEmail,
		To:      to,
		Subject: subject,
		Body:    body,
	}
	return nil
}

func setupRouter() *gin.Engine {
	// Start the internal SMTP server
	startSMTPServer()

	// Set Gin mode based on environment
	ginMode := os.Getenv("GIN_MODE")
	apiKey := os.Getenv("HELCIM_PRIVATE_API_KEY")

	// Safely log API key (first few chars only)
	if len(apiKey) >= 4 {
		log.Println("HELCIM_PRIVATE_API_KEY:", apiKey[:4]+"****")
	} else {
		log.Println("HELCIM_PRIVATE_API_KEY: <not set or too short>")
	}

	isDevMode := ginMode != "release"
	if isDevMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Trust all proxies (for development/testing, NOT recommended for production)
	// In production, specify trusted proxies using:
	// r.SetTrustedProxies([]string{"127.0.0.1"})
	r.SetTrustedProxies(nil)

	r.LoadHTMLGlob("templates/*")

	// Define a middleware to prevent caching
	r.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, max-age=0, must-revalidate, proxy-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/team", func(c *gin.Context) {
		c.HTML(http.StatusOK, "team.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/projects", func(c *gin.Context) {
		c.HTML(http.StatusOK, "projects.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/join", func(c *gin.Context) {
		c.HTML(http.StatusOK, "join.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/donate", func(c *gin.Context) {
		c.HTML(http.StatusOK, "donate.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/contact", func(c *gin.Context) {
		// Check if it's an HTMX request
		if c.GetHeader("HX-Request") == "true" {
			// Return just the form
			c.HTML(http.StatusOK, "contact_form.html", gin.H{
				"title": "Contact Form",
			})
		} else {
			// Return the full page
			c.HTML(http.StatusOK, "contact.html", gin.H{
				"title": "Main website",
			})
		}
	})

	// Add POST handler for contact form
	r.POST("/contact", func(c *gin.Context) {
		var form ContactForm
		if err := c.ShouldBind(&form); err != nil {
			if isDevMode {
				log.Println("Error binding form:", err)
			}
			c.HTML(http.StatusBadRequest, "contact.html", gin.H{
				"title": "Error",
				"error": "Invalid form submission. Please try again.",
			})
			return
		}

		// Compose and send email
		to := []string{"michael@avrnpo.org"}
		subject := "New Contact Form Submission from AVR Website"
		body := fmt.Sprintf(
			"Name: %s %s\nEmail: %s\nReferral Source: %s\n\nMessage:\n%s",
			form.FirstName,
			form.LastName,
			form.Email,
			form.ReferralSource,
			form.Message,
		)

		err := sendEmail(to, subject, body, form.Email)
		if err != nil {
			if isDevMode {
				log.Println("Error sending email:", err)
			}
		}

		// Return HTML response that will replace the form
		c.HTML(http.StatusOK, "contact_success.html", gin.H{
			"firstName": form.FirstName,
		})
	})

	// Add a route to view queued messages (admin use only)
	r.GET("/admin/messages", func(c *gin.Context) {
		// In a real implementation, you would add authentication here
		// For now, we just return a placeholder
		c.JSON(http.StatusOK, gin.H{
			"message": "This would show queued messages in a real implementation",
		})
	})

	r.GET("/footer", func(c *gin.Context) {
		c.HTML(http.StatusOK, "footer.html", gin.H{
			"title": "footer",
		})
	})

	r.GET("/api/checkout_token", func(c *gin.Context) {
		// Environment variables are now loaded at startup

		apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
		if isDevMode {
			log.Println("Using API token:", apiToken[:4]+"****") // Log first 4 chars for identification
		}

		// Helcim API endpoint
		helcimAPIURL := "https://api.helcim.com/v2/helcim-pay/initialize"
		if isDevMode {
			log.Println("Making request to:", helcimAPIURL)
		}

		// Get amount from query parameters
		amountStr := c.Query("amount")
		if amountStr == "" {
			if isDevMode {
				log.Println("Error: Amount parameter is missing")
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Amount is required"})
			return
		}

		var amount float64
		var err error
		_, err = fmt.Sscan(amountStr, &amount)
		if err != nil {
			if isDevMode {
				log.Println("Error parsing amount:", err)
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
			return
		}

		// Log all query parameters in dev mode
		if isDevMode {
			log.Println("Received parameters:")
			for k, v := range c.Request.URL.Query() {
				log.Printf("  %s: %v\n", k, v)
			}
		}

		// Get the additional donor information
		donorInfo := DonationInfo{
			FirstName: c.Query("firstName"),
			LastName:  c.Query("lastName"),
			Email:     c.Query("email"),
			Purpose:   c.Query("purpose"),
			Referral:  c.Query("referral"),
		}

		if isDevMode {
			log.Printf("Donor info: %+v\n", donorInfo)
		}

		// Create a payment request with customer info
		requestData := map[string]interface{}{
			"paymentType": "purchase",
			"amount":      amount,
			"currency":    "USD",
			"companyName": "American Veterans Rebuilding",
		}

		// Add customer information if available
		if donorInfo.FirstName != "" || donorInfo.LastName != "" || donorInfo.Email != "" {
			requestData["customer"] = map[string]interface{}{
				"firstName": donorInfo.FirstName,
				"lastName":  donorInfo.LastName,
				"email":     donorInfo.Email,
			}
		}

		// Add custom fields for donation purpose and referral source
		if donorInfo.Purpose != "" || donorInfo.Referral != "" {
			customFields := []map[string]string{}

			if donorInfo.Purpose != "" {
				customFields = append(customFields, map[string]string{
					"label": "Donation Purpose",
					"value": donorInfo.Purpose,
				})
			}

			if donorInfo.Referral != "" {
				customFields = append(customFields, map[string]string{
					"label": "Referral Source",
					"value": donorInfo.Referral,
				})
			}

			if len(customFields) > 0 {
				requestData["customFields"] = customFields
			}
		}

		// Marshal request body
		requestBody, err := json.Marshal(requestData)
		if err != nil {
			if isDevMode {
				log.Println("Error marshaling request body:", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request body"})
			return
		}

		// Log the request being sent to Helcim
		if isDevMode {
			log.Println("Sending to Helcim:", string(requestBody))
		}

		// Create request
		req, err := http.NewRequest("POST", helcimAPIURL, bytes.NewBuffer(requestBody))
		if err != nil {
			if isDevMode {
				log.Println("Error creating HTTP request:", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-token", apiToken)
		req.Header.Set("accept", "application/json")

		if isDevMode {
			log.Println("Request headers set")
			for k, v := range req.Header {
				if k != "api-token" {
					log.Printf("  %s: %v\n", k, v)
				} else {
					log.Printf("  %s: %s****\n", k, v[0][:4])
				}
			}
		}

		// Make the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			if isDevMode {
				log.Println("Error making HTTP request:", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make request"})
			return
		}
		defer resp.Body.Close()

		if isDevMode {
			log.Println("Response status:", resp.Status)
		}

		// Read response for debugging in dev mode
		if isDevMode {
			responseBody, _ := io.ReadAll(resp.Body)
			log.Println("Response body:", string(responseBody))

			// Create a new reader with the same data for the JSON decoder
			resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
		}

		// Read response
		var helcimResponse HelcimInitializeResponse
		err = json.NewDecoder(resp.Body).Decode(&helcimResponse)
		if err != nil {
			if isDevMode {
				log.Println("Error decoding response:", err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
			return
		}

		// Log the checkout token and secret token
		if isDevMode {
			log.Println("Checkout Token:", helcimResponse.CheckoutToken)
			log.Println("Secret Token:", helcimResponse.SecretToken[:4]+"****") // Only log part of the secret token
		}

		// Return the checkout token
		c.JSON(http.StatusOK, gin.H{
			"checkoutToken": helcimResponse.CheckoutToken,
			"secretToken":   helcimResponse.SecretToken,
		})
	})

	r.Static("/static", "./static")
	r.Static("/templates", "./templates")

	return r
}

func main() {
	// Load environment variables at startup
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	r := setupRouter()
	// Listen and Server in
	r.Run(":3000")
}

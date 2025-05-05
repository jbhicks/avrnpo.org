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
	"path/filepath"
	"strings"
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

	// Configure trusted proxies based on environment
	if isDevMode {
		// In development, trust no proxies for direct access
		r.SetTrustedProxies(nil)
	} else {
		// In production, trust all proxies since we're behind Traefik in Coolify
		// This ensures X-Forwarded-For headers are processed correctly
		r.SetTrustedProxies([]string{"0.0.0.0/0"})
		log.Println("Configured to trust proxies for production environment")
	}

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
			log.Println("Error binding form:", err)
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
			log.Println("Error sending email:", err)
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
		// Get API token and validate it's not empty
		apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
		if apiToken == "" {
			log.Println("ERROR: HELCIM_PRIVATE_API_KEY is not set or empty")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment gateway configuration error"})
			return
		}

		// Safely log API token
		if len(apiToken) >= 4 {
			log.Println("Using API token:", apiToken[:4]+"****")
		}

		// Helcim API endpoint
		helcimAPIURL := "https://api.helcim.com/v2/helcim-pay/initialize"
		log.Println("Making request to:", helcimAPIURL)

		// Get amount from query parameters
		amountStr := c.Query("amount")
		if amountStr == "" {
			log.Println("Error: Amount parameter is missing")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Amount is required"})
			return
		}

		var amount float64
		var err error
		_, err = fmt.Sscan(amountStr, &amount)
		if err != nil {
			log.Println("Error parsing amount:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
			return
		}

		// Log all query parameters
		log.Println("Received parameters:")
		for k, v := range c.Request.URL.Query() {
			log.Printf("  %s: %v\n", k, v)
		}

		// Get the additional donor information
		donorInfo := DonationInfo{
			FirstName: c.Query("firstName"),
			LastName:  c.Query("lastName"),
			Email:     c.Query("email"),
			Purpose:   c.Query("purpose"),
			Referral:  c.Query("referral"),
		}

		log.Printf("Donor info: %+v\n", donorInfo)

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
			log.Println("Error marshaling request body:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request body"})
			return
		}

		// Log the request being sent to Helcim
		log.Println("Sending to Helcim:", string(requestBody))

		// Create request
		req, err := http.NewRequest("POST", helcimAPIURL, bytes.NewBuffer(requestBody))
		if err != nil {
			log.Println("Error creating HTTP request:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-token", apiToken)
		req.Header.Set("accept", "application/json")

		log.Println("Request headers set")
		for k, v := range req.Header {
			if k != "api-token" {
				log.Printf("  %s: %v\n", k, v)
			} else {
				log.Printf("  %s: %s****\n", k, v[0][:4])
			}
		}

		// Make the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error making HTTP request:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make request"})
			return
		}
		defer resp.Body.Close()

		log.Println("Response status:", resp.Status)

		// Read response body for debugging
		responseBody, _ := io.ReadAll(resp.Body)
		log.Println("Response body:", string(responseBody))

		// Create a new reader with the same data for the JSON decoder
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

		// Read response
		var helcimResponse HelcimInitializeResponse
		err = json.NewDecoder(resp.Body).Decode(&helcimResponse)
		if err != nil {
			log.Println("Error decoding response:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
			return
		}

		// Check if the response contains valid tokens
		if helcimResponse.CheckoutToken == "" {
			log.Println("ERROR: Helcim API returned empty checkout token")

			// Try to read the response body again for error details
			respBody, _ := json.Marshal(helcimResponse)
			log.Println("Raw response:", string(respBody))

			// Check if the actual HTTP status from Helcim indicates an error
			if resp.StatusCode >= 400 {
				log.Println("Helcim API returned error status:", resp.Status)
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Payment gateway returned an invalid response",
				"details": "The payment service did not authorize this transaction. Check API credentials.",
			})
			return
		}

		// Log the checkout token and secret token
		log.Println("Checkout Token received successfully:", helcimResponse.CheckoutToken[:4]+"****")
		if len(helcimResponse.SecretToken) > 0 {
			log.Println("Secret Token received successfully:", helcimResponse.SecretToken[:4]+"****")
		} else {
			log.Println("WARNING: Secret token is empty")
		}

		// Return the checkout token
		c.JSON(http.StatusOK, gin.H{
			"checkoutToken": helcimResponse.CheckoutToken,
			"secretToken":   helcimResponse.SecretToken,
		})
	})

	// Add a diagnostic endpoint to check headers
	r.GET("/debug/headers", func(c *gin.Context) {
		// Only available in development mode or with a special header for security
		if isDevMode || c.GetHeader("X-Debug-Access") == os.Getenv("DEBUG_ACCESS_KEY") {
			headers := map[string]string{}
			for k, v := range c.Request.Header {
				headers[k] = strings.Join(v, ", ")
			}

			// Also include important request information
			info := map[string]string{
				"RemoteAddr":   c.Request.RemoteAddr,
				"RequestURI":   c.Request.RequestURI,
				"Method":       c.Request.Method,
				"Host":         c.Request.Host,
				"TLS":          fmt.Sprintf("%v", c.Request.TLS != nil),
				"ProtoMajor":   fmt.Sprintf("%d", c.Request.ProtoMajor),
				"ProtoMinor":   fmt.Sprintf("%d", c.Request.ProtoMinor),
				"ClientIP":     c.ClientIP(),
				"TrustedProxy": fmt.Sprintf("%v", c.Request.Header.Get("X-Forwarded-For") != ""),
			}

			c.JSON(http.StatusOK, gin.H{
				"headers": headers,
				"info":    info,
			})
		} else {
			c.Status(http.StatusNotFound)
		}
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.Static("/static", "./static")
	r.Static("/templates", "./templates")

	return r
}

func main() {
	// Load environment variables from .env file at startup
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file, using system environment variables")
	} else {
		log.Println("Loaded environment variables from .env file")
	}

	// Print important environment variables for debugging
	log.Println("Environment variables:")
	log.Println("PORT:", os.Getenv("PORT"))
	log.Println("GIN_MODE:", os.Getenv("GIN_MODE"))

	// Safely print API key
	apiKey := os.Getenv("HELCIM_PRIVATE_API_KEY")
	if len(apiKey) >= 4 {
		log.Println("HELCIM_PRIVATE_API_KEY:", apiKey[:4]+"****")
	} else {
		log.Println("HELCIM_PRIVATE_API_KEY: <not set or too short>")
	}

	// Download static assets
	downloadStaticAssets()

	r := setupRouter()
	// Listen and Server in
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	r.Run(":" + port)
}

// downloadStaticAssets fetches the latest CSS libraries at startup
func downloadStaticAssets() {
	// Create the static directory if it doesn't exist
	staticDir := "./static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.Mkdir(staticDir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create static directory: %v", err)
		}
	}

	// Download Tailwind CSS
	tailwindURL := "https://cdn.jsdelivr.net/npm/tailwindcss@latest/dist/tailwind.min.css"
	downloadAsset(tailwindURL, filepath.Join(staticDir, "tailwind.min.css"), "Tailwind CSS")

	// Download DaisyUI
	daisyUIURL := "https://cdn.jsdelivr.net/npm/daisyui@latest/dist/full.min.css"
	downloadAsset(daisyUIURL, filepath.Join(staticDir, "daisyui.min.css"), "DaisyUI")
}

// downloadAsset downloads a file from url and saves it to the specified path
func downloadAsset(url, filePath, name string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download %s: %v", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to download %s: received status code %d", name, resp.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create %s file: %v", name, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Fatalf("Failed to write %s to file: %v", name, err)
	}

	log.Printf("Successfully downloaded %s", name)
}

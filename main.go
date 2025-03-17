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

// Add this struct to define the contact form data
type ContactForm struct {
	FirstName string `form:"fname"`
	LastName  string `form:"lname"`
	Email     string `form:"email"`
	Message   string `form:"message"`
}

type HelcimInitializeResponse struct {
	CheckoutToken string `json:"checkoutToken"`
	SecretToken   string `json:"secretToken"` // Add SecretToken field
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
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
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
			"Name: %s %s\nEmail: %s\n\nMessage:\n%s",
			form.FirstName,
			form.LastName,
			form.Email,
			form.Message,
		)

		err := sendEmail(to, subject, body, form.Email)
		if err != nil {
			fmt.Println("Error sending email:", err)
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
		err := godotenv.Load()
		if err != nil {
			fmt.Println("Error loading .env file")
		}

		apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
		fmt.Println(apiToken)

		// Helcim API endpoint
		helcimAPIURL := "https://api.helcim.com/v2/helcim-pay/initialize"

		// Get amount from query parameters
		amountStr := c.Query("amount")
		if amountStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Amount is required"})
			return
		}

		var amount float64
		_, err = fmt.Sscan(amountStr, &amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
			return
		}

		// Request body
		requestBody, err := json.Marshal(map[string]interface{}{
			"paymentType": "purchase",
			"amount":      amount,
			"currency":    "USD",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request body"})
			return
		}

		// Create request
		req, err := http.NewRequest("POST", helcimAPIURL, bytes.NewBuffer(requestBody))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-token", apiToken)
		req.Header.Set("accept", "application/json")

		// Make the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make request"})
			return
		}
		defer resp.Body.Close()

		// Read response
		var helcimResponse HelcimInitializeResponse
		err = json.NewDecoder(resp.Body).Decode(&helcimResponse)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
			return
		}

		// Log the checkout token and secret token
		fmt.Println("Checkout Token:", helcimResponse.CheckoutToken)
		fmt.Println("Secret Token:", helcimResponse.SecretToken)

		// Return the checkout token
		c.JSON(http.StatusOK, gin.H{
			"checkoutToken": helcimResponse.CheckoutToken,
			"secretToken":   helcimResponse.SecretToken, // Include secretToken in the response
		})
	})

	r.Static("/static", "./static")
	r.Static("/templates", "./templates")

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in
	r.Run(":3000")
}

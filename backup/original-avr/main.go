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
	"regexp"
	"strings"
	"sync"
	"time"

	gosmtp "github.com/emersion/go-smtp" // Rename this import to avoid conflicts
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Rate limiting structure
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	// Clean up old requests
	if times, exists := rl.requests[ip]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.window {
				validTimes = append(validTimes, t)
			}
		}
		rl.requests[ip] = validTimes
	}

	// Check if under limit
	if len(rl.requests[ip]) >= rl.limit {
		return false
	}

	// Add this request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

// Global rate limiter for donation endpoint
var donationRateLimiter = NewRateLimiter(5, time.Minute) // 5 requests per minute per IP

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

	r.POST("/api/checkout_token", func(c *gin.Context) {
		// Rate limiting check
		clientIP := c.ClientIP()
		if !donationRateLimiter.Allow(clientIP) {
			log.Printf("Rate limit exceeded for IP: %s", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please wait before trying again."})
			return
		}

		// Define request structure for JSON body
		var request struct {
			Amount    float64 `json:"amount" binding:"required"`
			FirstName string  `json:"firstName"`
			LastName  string  `json:"lastName"`
			Email     string  `json:"email"`
			Purpose   string  `json:"purpose"`
			Referral  string  `json:"referral"`
		}

		// Bind JSON request body
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("Error binding JSON request from IP %s: %v", clientIP, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Validate and sanitize amount
		if err := validateAmount(request.Amount); err != nil {
			log.Printf("Invalid amount from IP %s: %v", clientIP, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Sanitize and validate input fields
		request.FirstName = sanitizeString(request.FirstName)
		request.LastName = sanitizeString(request.LastName)
		request.Email = sanitizeString(request.Email)
		request.Purpose = sanitizeString(request.Purpose)
		request.Referral = sanitizeString(request.Referral)

		// Validate email format
		if !validateEmail(request.Email) {
			log.Printf("Invalid email format from IP %s: %s", clientIP, request.Email)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}

		// Validate names
		if !validateName(request.FirstName) {
			log.Printf("Invalid first name from IP %s: %s", clientIP, request.FirstName)
			c.JSON(http.StatusBadRequest, gin.H{"error": "First name contains invalid characters"})
			return
		}

		if !validateName(request.LastName) {
			log.Printf("Invalid last name from IP %s: %s", clientIP, request.LastName)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Last name contains invalid characters"})
			return
		}

		// Validate purpose
		if !validatePurpose(request.Purpose) {
			log.Printf("Invalid purpose from IP %s: %s", clientIP, request.Purpose)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid donation purpose"})
			return
		}

		// Get API token and validate it's not empty
		apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
		if apiToken == "" {
			log.Println("ERROR: HELCIM_PRIVATE_API_KEY is not set or empty")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment gateway configuration error"})
			return
		}

		// Check if API token is potentially truncated
		const expectedMinLength = 30
		if len(apiToken) < expectedMinLength {
			log.Printf("ERROR: HELCIM_PRIVATE_API_KEY appears to be truncated (length: %d, expected at least: %d)",
				len(apiToken), expectedMinLength)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment gateway misconfiguration"})
			return
		}

		// Helcim API URL
		helcimAPIURL := "https://api.helcim.com/v2/helcim-pay/initialize"
		log.Println("Making request to:", helcimAPIURL)

		// Log received data (without sensitive info, after validation)
		log.Printf("Validated donation request: Amount=%.2f, FirstName=%s, LastName=%s, Email=%s, Purpose=%s, Referral=%s\n",
			request.Amount, request.FirstName, request.LastName, request.Email, request.Purpose, request.Referral)

		// Get the additional donor information
		donorInfo := DonationInfo{
			FirstName: request.FirstName,
			LastName:  request.LastName,
			Email:     request.Email,
			Purpose:   request.Purpose,
			Referral:  request.Referral,
		}

		log.Printf("Donor info: %+v\n", donorInfo)

		// Create a payment request with customer info
		requestData := map[string]interface{}{
			"paymentType": "purchase",
			"amount":      request.Amount,
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

		// Set headers - explicitly using the string, not using Header.Set
		req.Header = http.Header{
			"Content-Type": []string{"application/json"},
			"Api-Token":    []string{apiToken},
			"Accept":       []string{"application/json"},
		}

		// Debug: Log the full token
		log.Println("DEBUG - FULL API TOKEN:", apiToken)
		log.Printf("API Token length: %d characters", len(apiToken))

		log.Println("Request headers set")
		for k, v := range req.Header {
			if k != "Api-Token" {
				log.Printf("  %s: %v\n", k, v)
			} else {
				// Also log the full header value for debugging
				log.Printf("  %s: %s\n", k, v[0])
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

	// Add dedicated endpoint for payment debugging
	r.GET("/api/payment/debug", func(c *gin.Context) {
		// Only available in development mode or with a special header for security
		if isDevMode || c.GetHeader("X-Debug-Access") == os.Getenv("DEBUG_ACCESS_KEY") {
			// Test Helcim API key validity without making a full transaction
			apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
			maskedToken := "not set"
			if apiToken != "" && len(apiToken) >= 4 {
				maskedToken = apiToken[:4] + "****"
			}

			// Create debug info
			debugInfo := map[string]interface{}{
				"apiKeySet":    apiToken != "",
				"apiKeyMasked": maskedToken,
				"mode":         os.Getenv("GIN_MODE"),
				"environment": map[string]string{
					"GIN_MODE": os.Getenv("GIN_MODE"),
					"PORT":     os.Getenv("PORT"),
				},
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "Payment system debugging information",
				"info":   debugInfo,
			})
		} else {
			c.Status(http.StatusNotFound)
		}
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

	// Add a diagnostic endpoint specifically for Helcim API testing
	r.GET("/api/payment/test", func(c *gin.Context) {
		if isDevMode || c.GetHeader("X-Debug-Access") == os.Getenv("DEBUG_ACCESS_KEY") {
			// Get API token and validate it's not empty
			apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
			if apiToken == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "API token not set"})
				return
			}

			// Check if API token is potentially truncated
			const expectedMinLength = 30
			if len(apiToken) < expectedMinLength {
				log.Printf("ERROR: HELCIM_PRIVATE_API_KEY appears to be truncated (length: %d, expected at least: %d)",
					len(apiToken), expectedMinLength)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment gateway misconfiguration"})
				return
			}

			// Create a test request to Helcim API
			helcimAPIURL := "https://api.helcim.com/v2/helcim-pay/initialize"

			// Create a minimal test request
			testRequest := map[string]interface{}{
				"paymentType": "test",
				"amount":      1,
				"currency":    "USD",
			}

			requestBody, _ := json.Marshal(testRequest)

			// Create two requests - one with canonicalized headers and one with lowercase
			req1, _ := http.NewRequest("POST", helcimAPIURL, bytes.NewBuffer(requestBody))
			req1.Header.Set("Content-Type", "application/json")
			req1.Header.Set("api-token", apiToken) // Standard way
			req1.Header.Set("Accept", "application/json")

			// Second request with custom header setting to prevent canonicalization
			req2, _ := http.NewRequest("POST", helcimAPIURL, bytes.NewBuffer(requestBody))
			req2.Header = make(http.Header)
			req2.Header["content-type"] = []string{"application/json"}
			req2.Header["api-token"] = []string{apiToken} // Lowercase forced
			req2.Header["accept"] = []string{"application/json"}

			// Make both requests
			client := &http.Client{Timeout: 10 * time.Second}

			// First request (standard headers)
			resp1, err1 := client.Do(req1)
			result1 := map[string]interface{}{
				"success": false,
				"error":   "Request failed",
				"headers": req1.Header,
			}

			if err1 == nil {
				defer resp1.Body.Close()
				result1["status"] = resp1.Status
				result1["success"] = resp1.StatusCode < 400
				body1, _ := io.ReadAll(resp1.Body)
				result1["body"] = string(body1)
			} else {
				result1["error"] = err1.Error()
			}

			// Second request (forced lowercase headers)
			resp2, err2 := client.Do(req2)
			result2 := map[string]interface{}{
				"success": false,
				"error":   "Request failed",
				"headers": req2.Header,
			}

			if err2 == nil {
				defer resp2.Body.Close()
				result2["status"] = resp2.Status
				result2["success"] = resp2.StatusCode < 400
				body2, _ := io.ReadAll(resp2.Body)
				result2["body"] = string(body2)
			} else {
				result2["error"] = err2.Error()
			}

			// Return results for comparison
			c.JSON(http.StatusOK, gin.H{
				"standardHeaders":  result1,
				"lowercaseHeaders": result2,
				"apiTokenPrefix":   apiToken[:5] + "...",
			})
		} else {
			c.Status(http.StatusNotFound)
		}
	})

	// Add a Helcim API relay endpoint for detailed diagnostics
	r.GET("/api/diagnostics/helcim", func(c *gin.Context) {
		// Secure this endpoint - only available with debug access key
		if !isDevMode && c.GetHeader("X-Debug-Access") != os.Getenv("DEBUG_ACCESS_KEY") {
			c.Status(http.StatusNotFound)
			return
		}

		// Get API token from environment
		apiToken := os.Getenv("HELCIM_PRIVATE_API_KEY")
		if apiToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "HELCIM_PRIVATE_API_KEY not set",
			})
			return
		}

		// Diagnostic data to collect
		diagnostics := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"tokenInfo": map[string]interface{}{
				"length": len(apiToken),
				"prefix": apiToken[:min(4, len(apiToken))],
				"suffix": apiToken[max(0, len(apiToken)-4):],
				"containsSpecialChars": map[string]bool{
					"$":  strings.Contains(apiToken, "$"),
					"*":  strings.Contains(apiToken, "*"),
					"\\": strings.Contains(apiToken, "\\"),
				},
			},
			"environment": map[string]string{
				"GIN_MODE": os.Getenv("GIN_MODE"),
				"mode":     gin.Mode(),
			},
			"requestTests": make(map[string]interface{}),
		}

		// Test API endpoint
		helcimAPIURL := "https://api.helcim.com/v2/helcim-pay/initialize"

		// Create a minimal test request body
		requestData := map[string]interface{}{
			"paymentType": "purchase",
			"amount":      1,
			"currency":    "USD",
			"companyName": "American Veterans Rebuilding",
		}
		requestBody, _ := json.Marshal(requestData)

		// Create a relay function we'll use for different header combinations
		makeRelayRequest := func(headerName, headerValue string) map[string]interface{} {
			result := map[string]interface{}{
				"success": false,
				"requestInfo": map[string]interface{}{
					"url":        helcimAPIURL,
					"method":     "POST",
					"headerName": headerName,
					"headerMask": headerValue[:min(4, len(headerValue))] + "..." + headerValue[max(0, len(headerValue)-4):],
				},
			}

			// Create request with specified header
			req, err := http.NewRequest("POST", helcimAPIURL, bytes.NewBuffer(requestBody))
			if err != nil {
				result["error"] = "Failed to create request: " + err.Error()
				return result
			}

			// Set request headers
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(headerName, headerValue)
			req.Header.Set("Accept", "application/json")

			// Log all request headers for diagnostics
			requestHeaders := make(map[string]string)
			for k, v := range req.Header {
				requestHeaders[k] = strings.Join(v, ", ")
			}
			result["requestHeaders"] = requestHeaders

			// Make the request
			client := &http.Client{Timeout: 10 * time.Second}
			startTime := time.Now()
			resp, err := client.Do(req)
			requestDuration := time.Since(startTime)

			result["requestDuration"] = requestDuration.String()

			if err != nil {
				result["error"] = "Request failed: " + err.Error()
				return result
			}
			defer resp.Body.Close()

			// Read response details
			responseData, _ := io.ReadAll(resp.Body)

			// Add response info to result
			result["statusCode"] = resp.StatusCode
			result["success"] = resp.StatusCode < 400
			result["responseHeaders"] = resp.Header
			result["responseBody"] = string(responseData)

			// Try to parse the response if it's JSON
			var parsedJSON interface{}
			if err := json.Unmarshal(responseData, &parsedJSON); err == nil {
				result["parsedResponse"] = parsedJSON
			}

			return result
		}

		// Test different authentication headers
		diagnostics["requestTests"] = map[string]interface{}{
			"apiToken":   makeRelayRequest("api-token", apiToken),
			"ApiToken":   makeRelayRequest("Api-Token", apiToken),
			"xAuthToken": makeRelayRequest("x-auth-token", apiToken),
			"XAuthToken": makeRelayRequest("X-Auth-Token", apiToken),
		}

		// Return all diagnostic information
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"diagnostics": diagnostics,
		})
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.Static("/static", "./static")
	r.Static("/templates", "./templates")

	return r
}

// Helper functions for min and max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Validation functions
func sanitizeString(input string) string {
	// Remove any HTML tags and dangerous characters
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "<", "")
	input = strings.ReplaceAll(input, ">", "")
	input = strings.ReplaceAll(input, "\"", "")
	input = strings.ReplaceAll(input, "'", "")
	return input
}

func validateEmail(email string) bool {
	if email == "" {
		return true // Email is optional
	}
	// Basic email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func validateName(name string) bool {
	if name == "" {
		return true // Names are optional
	}
	// Only allow letters, spaces, hyphens, and apostrophes
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']{1,50}$`)
	return nameRegex.MatchString(name)
}

func validateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if amount > 10000 {
		return fmt.Errorf("amount exceeds maximum allowed ($10,000)")
	}
	if amount < 1 {
		return fmt.Errorf("minimum donation amount is $1")
	}
	return nil
}

func validatePurpose(purpose string) bool {
	if purpose == "" {
		return true // Purpose is optional
	}
	validPurposes := []string{"general", "veterans", "housing", "emergency", "education", "other"}
	for _, valid := range validPurposes {
		if purpose == valid {
			return true
		}
	}
	return false
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

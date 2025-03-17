package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type HelcimInitializeResponse struct {
	CheckoutToken string `json:"checkoutToken"`
	SecretToken   string `json:"secretToken"` // Add SecretToken field
}

func setupRouter() *gin.Engine {
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
		c.HTML(http.StatusOK, "contact.html", gin.H{
			"title": "Main website",
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

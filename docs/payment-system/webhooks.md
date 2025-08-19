# Helcim Webhooks Implementation Guide

This guide provides step-by-step instructions for implementing Helcim webhooks in the AVR donation system (Phase 2 of the improvement plan).

## Overview

Webhooks allow Helcim to notify our system when payment events occur, enabling real-time updates on donation status without polling the API.

## Webhook Events

Helcim sends webhooks for these card transaction events:
- **Payment Success** - Donation processed successfully
- **Payment Declined** - Donation was declined
- **Payment Refunded** - Donation was refunded
- **Payment Cancelled** - Donation was cancelled

## Implementation Steps

### Step 1: Configure Webhook URL in Helcim

1. Log into your Helcim account
2. Navigate to **All Tools** → **Integrations** → **Webhooks**
3. Toggle **Webhooks ON**
4. Set **Deliver URL** to: `https://yourdomain.com/api/webhooks/helcim`
5. Ensure **Notify events for Transactions** is checked
6. Click **Save**
7. Copy the **Verifier Token** for signature verification

### Step 2: Environment Configuration

Add to your `.env` file:
```bash
# Webhook verification token from Helcim dashboard
HELCIM_WEBHOOK_VERIFIER_TOKEN=your-verifier-token-here

# Your webhook endpoint URL (for reference)
HELCIM_WEBHOOK_URL=https://yourdomain.com/api/webhooks/helcim
```

### Step 3: Go Implementation

#### 3.1 Add Webhook Structures

```go
// Webhook event structure
type HelcimWebhookEvent struct {
    ID   string `json:"id"`
    Type string `json:"type"`
}

// Webhook headers structure
type WebhookHeaders struct {
    Signature string
    Timestamp string
    ID        string
}
```

#### 3.2 Add Signature Verification

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "strconv"
    "strings"
    "time"
)

func verifyWebhookSignature(signature, timestamp, webhookID string, body []byte) bool {
    // Get verifier token from environment
    verifierToken := os.Getenv("HELCIM_WEBHOOK_VERIFIER_TOKEN")
    if verifierToken == "" {
        log.Println("ERROR: HELCIM_WEBHOOK_VERIFIER_TOKEN not set")
        return false
    }

    // Decode the verifier token from base64
    verifierTokenBytes, err := base64.StdEncoding.DecodeString(verifierToken)
    if err != nil {
        log.Printf("ERROR: Failed to decode verifier token: %v", err)
        return false
    }

    // Create signed content: webhook-id.webhook-timestamp.body
    signedContent := fmt.Sprintf("%s.%s.%s", webhookID, timestamp, string(body))

    // Generate HMAC signature
    mac := hmac.New(sha256.New, verifierTokenBytes)
    mac.Write([]byte(signedContent))
    expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

    // Parse the signature header (format: "v1,signature v2,signature")
    signatures := strings.Split(signature, " ")
    for _, sig := range signatures {
        if strings.HasPrefix(sig, "v1,") {
            receivedSignature := strings.TrimPrefix(sig, "v1,")
            if hmac.Equal([]byte(expectedSignature), []byte(receivedSignature)) {
                return true
            }
        }
    }

    return false
}
```

#### 3.3 Add Timestamp Validation

```go
func isValidWebhookTimestamp(timestampStr string) bool {
    timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
    if err != nil {
        return false
    }

    webhookTime := time.Unix(timestamp, 0)
    now := time.Now()
    
    // Reject webhooks older than 5 minutes (300 seconds)
    return now.Sub(webhookTime) <= 5*time.Minute
}
```

#### 3.4 Add Webhook Handler Endpoint

```go
func setupWebhookRoutes(r *gin.Engine) {
    r.POST("/api/webhooks/helcim", handleHelcimWebhook)
}

func handleHelcimWebhook(c *gin.Context) {
    // Extract webhook headers
    signature := c.GetHeader("webhook-signature")
    timestamp := c.GetHeader("webhook-timestamp")
    webhookID := c.GetHeader("webhook-id")

    if signature == "" || timestamp == "" || webhookID == "" {
        log.Printf("Missing webhook headers from IP %s", c.ClientIP())
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing webhook headers"})
        return
    }

    // Validate timestamp
    if !isValidWebhookTimestamp(timestamp) {
        log.Printf("Invalid webhook timestamp from IP %s: %s", c.ClientIP(), timestamp)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timestamp"})
        return
    }

    // Read and verify body
    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        log.Printf("Failed to read webhook body from IP %s: %v", c.ClientIP(), err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
        return
    }

    // Verify signature
    if !verifyWebhookSignature(signature, timestamp, webhookID, body) {
        log.Printf("Invalid webhook signature from IP %s", c.ClientIP())
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
        return
    }

    // Parse webhook event
    var event HelcimWebhookEvent
    if err := json.Unmarshal(body, &event); err != nil {
        log.Printf("Failed to parse webhook event from IP %s: %v", c.ClientIP(), err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
        return
    }

    // Process the webhook event
    if err := processWebhookEvent(event); err != nil {
        log.Printf("Failed to process webhook event %s: %v", event.ID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
        return
    }

    // Log successful processing
    log.Printf("Successfully processed webhook event: ID=%s, Type=%s", event.ID, event.Type)
    
    // Return success response
    c.JSON(http.StatusOK, gin.H{"message": "Webhook processed"})
}
```

#### 3.5 Add Event Processing Logic

```go
func processWebhookEvent(event HelcimWebhookEvent) error {
    switch event.Type {
    case "cardTransaction":
        return processTransactionEvent(event.ID)
    default:
        log.Printf("Unknown webhook event type: %s", event.Type)
        return nil // Don't treat unknown events as errors
    }
}

func processTransactionEvent(transactionID string) error {
    // Fetch transaction details from Helcim API
    transaction, err := fetchTransactionDetails(transactionID)
    if err != nil {
        return fmt.Errorf("failed to fetch transaction %s: %w", transactionID, err)
    }

    // Log transaction details
    log.Printf("Transaction %s: Status=%s, Amount=%.2f, Customer=%s %s", 
        transactionID, transaction.Status, transaction.Amount, 
        transaction.Customer.FirstName, transaction.Customer.LastName)

    // Handle based on transaction status
    switch transaction.Status {
    case "APPROVED":
        return handleSuccessfulDonation(transaction)
    case "DECLINED":
        return handleDeclinedDonation(transaction)
    case "REFUNDED":
        return handleRefundedDonation(transaction)
    default:
        log.Printf("Unhandled transaction status: %s", transaction.Status)
        return nil
    }
}

func handleSuccessfulDonation(transaction *TransactionDetails) error {
    // TODO: Future implementation
    // - Store donation in database
    // - Send thank you email
    // - Generate receipt
    log.Printf("✅ Successful donation: $%.2f from %s %s", 
        transaction.Amount, transaction.Customer.FirstName, transaction.Customer.LastName)
    return nil
}

func handleDeclinedDonation(transaction *TransactionDetails) error {
    // TODO: Future implementation
    // - Log declined attempt
    // - Optionally notify user of decline
    log.Printf("❌ Declined donation: $%.2f from %s %s", 
        transaction.Amount, transaction.Customer.FirstName, transaction.Customer.LastName)
    return nil
}

func handleRefundedDonation(transaction *TransactionDetails) error {
    // TODO: Future implementation
    // - Update donation status in database
    // - Send refund confirmation email
    log.Printf("🔄 Refunded donation: $%.2f to %s %s", 
        transaction.Amount, transaction.Customer.FirstName, transaction.Customer.LastName)
    return nil
}
```

#### 3.6 Add Transaction Fetching (Future Phase 3)

```go
type TransactionDetails struct {
    ID       string  `json:"id"`
    Status   string  `json:"status"`
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
    Customer struct {
        FirstName string `json:"firstName"`
        LastName  string `json:"lastName"`
        Email     string `json:"email"`
    } `json:"customer"`
    CreatedAt string `json:"createdAt"`
}

func fetchTransactionDetails(transactionID string) (*TransactionDetails, error) {
    // This will be implemented in Phase 3 when we add database integration
    // For now, return minimal transaction info
    return &TransactionDetails{
        ID:     transactionID,
        Status: "APPROVED", // Placeholder
        Amount: 0.0,        // Placeholder
    }, nil
}
```

## Security Considerations

### 1. Signature Verification
- **Always verify** webhook signatures before processing
- Use constant-time comparison to prevent timing attacks
- Log failed verification attempts

### 2. Timestamp Validation
- Reject webhooks older than 5 minutes
- Prevents replay attacks
- Accounts for reasonable network delays

### 3. Rate Limiting
- Apply rate limiting to webhook endpoint
- Separate limits from regular API endpoints
- Consider IP-based limiting

### 4. Error Handling
- Return appropriate HTTP status codes
- Log all webhook events for debugging
- Don't expose internal errors in responses

### 5. Idempotency
- Handle duplicate webhook deliveries gracefully
- Helcim may send the same webhook multiple times
- Store webhook IDs to detect duplicates

## Testing Webhooks

### 1. Local Testing with ngrok
```bash
# Install ngrok and expose local server
ngrok http 3001

# Use the HTTPS URL in Helcim webhook settings
# Example: https://abc123.ngrok.io/api/webhooks/helcim
```

### 2. Webhook Testing Endpoint
Add a test endpoint for development:
```go
r.POST("/api/webhooks/test", func(c *gin.Context) {
    // Simulate a webhook event for testing
    testEvent := HelcimWebhookEvent{
        ID:   "test_" + time.Now().Format("20060102150405"),
        Type: "cardTransaction",
    }
    
    if err := processWebhookEvent(testEvent); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Test webhook processed"})
})
```

### 3. Manual Testing with curl
```bash
# Test webhook endpoint manually
curl -X POST http://localhost:3001/api/webhooks/test \
  -H "Content-Type: application/json"
```

## Monitoring and Logging

### 1. Webhook Logs
```go
// Add structured logging for webhooks
log.Printf("[WEBHOOK] Event received: ID=%s, Type=%s, IP=%s", 
    event.ID, event.Type, c.ClientIP())

log.Printf("[WEBHOOK] Processing result: ID=%s, Success=%t, Duration=%v", 
    event.ID, success, duration)
```

### 2. Error Tracking
- Log all webhook failures with details
- Track signature verification failures
- Monitor processing errors by event type

### 3. Performance Monitoring
- Track webhook processing times
- Monitor endpoint availability
- Alert on processing failures

## Deployment Considerations

### 1. HTTPS Requirements
- Helcim requires HTTPS webhook URLs
- Ensure SSL certificate is valid
- Test SSL configuration

### 2. DNS and Routing
- Webhook URL must be publicly accessible
- Configure firewall rules if needed
- Test connectivity from external services

### 3. High Availability
- Consider webhook endpoint redundancy
- Implement health checks
- Plan for server maintenance windows

## Error Recovery

### 1. Webhook Retry Handling
Helcim automatically retries failed webhooks:
- Immediate retry
- 5 seconds
- 5 minutes
- 30 minutes
- 2 hours
- 5 hours
- 10 hours (twice)

### 2. Manual Recovery
```go
// Add endpoint to manually process missed transactions
r.POST("/api/admin/process-transaction/:id", func(c *gin.Context) {
    transactionID := c.Param("id")
    // Manually fetch and process transaction
    // Useful for webhook failures or missed events
})
```

## Phase 3 Integration Notes

When implementing database integration:
1. Store webhook events for audit trail
2. Implement idempotency using webhook IDs
3. Add transaction status tracking
4. Enable transaction lookup by webhook event

---

*This guide should be followed for Phase 2 implementation. Update as webhook functionality is added.*

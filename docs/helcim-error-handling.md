# Helcim Error Handling Reference

This document provides comprehensive error handling patterns for the AVR Helcim integration.

## Common HTTP Status Codes

| Code | Type | Description | Action Required |
|------|------|-------------|-----------------|
| 200 | Success | Request completed successfully | Continue processing |
| 400 | Bad Request | Invalid request data or missing required fields | Fix request format/data |
| 401 | Unauthorized | Invalid or missing API token | Check API token |
| 403 | Forbidden | API token lacks required permissions | Update token permissions |
| 404 | Not Found | Endpoint or resource doesn't exist | Check URL/endpoint |
| 429 | Too Many Requests | Rate limit exceeded | Implement backoff/retry |
| 500 | Internal Server Error | Helcim system error | Retry after delay |

## Error Response Formats

### Single Error
```json
{
  "errors": "Unauthorized"
}
```

### Multiple Errors
```json
{
  "errors": {
    "billingAddress[name]": "Missing required data billing Address.name",
    "billingAddress[postalCode]": "Missing required data billing Address.postal Code"
  }
}
```

## Authentication Errors

### 401 - Unauthorized
**Cause:** Invalid or inactive API token
```json
{
    "errors": "Unauthorized"
}
```
**Resolution:**
1. Verify API token in `.env` file
2. Check token is active in Helcim dashboard
3. Ensure token is properly formatted (no extra spaces/characters)

### 403 - No Access Permission
**Cause:** API token lacks required permissions
```json
{
    "errors": "No access permission"
}
```
**Resolution:**
1. Log into Helcim account
2. Go to All Tools → Integrations
3. Edit API Access Configuration
4. Enable required permissions (e.g., HelcimPay.js, Payment API)

## Payment API Errors

### Missing Card Data
```json
{
  "errors": {
    "cardData[cardNumber]": "Missing required data card Data.card Number",
    "cardData[cardExpiry]": "Missing required data card Data.card Expiry",
    "cardData[cardCVV]": "Missing required data card Data.card CVV"
  }
}
```
**Resolution:** Use HelcimPay.js for card tokenization instead of raw card data

### Not Allowed to Send Full Card Number
```json
{
  "errors": "Not allowed to send full card number"
}
```
**Resolution:** 
1. Use HelcimPay.js for PCI-compliant card handling
2. Send `cardToken` instead of raw card details
3. Review PCI compliance scope

### Missing Required Data
```json
{
  "errors": {
    "billingAddress": "Missing required data billing Address"
  }
}
```
**Resolution:** Include all required fields in request

### Missing Idempotency Key
```json
{
  "errors": "invalid idempotencyKey"
}
```
**Resolution:** Include unique `idempotency-key` header for payment requests

## HelcimPay.js Specific Errors

### CORS Errors
**Cause:** Making API calls from frontend/client-side
**Resolution:** Move all Helcim API calls to backend server

### Token Expiration
**Cause:** Checkout tokens expire after 60 minutes
**Resolution:** Generate new tokens for each payment session

### Invalid Checkout Token
**Cause:** Using expired or invalid checkout token
**Resolution:** Call initialize endpoint to get fresh tokens

## Go Error Handling Patterns

### Basic Error Structure
```go
type HelcimError struct {
    StatusCode int                    `json:"status_code"`
    Errors     map[string]interface{} `json:"errors"`
    Message    string                 `json:"message"`
}

func (e *HelcimError) Error() string {
    return fmt.Sprintf("Helcim API error %d: %s", e.StatusCode, e.Message)
}
```

### Response Parsing
```go
func parseHelcimResponse(resp *http.Response) (*HelcimResponse, error) {
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    if resp.StatusCode >= 400 {
        var helcimErr HelcimError
        if err := json.Unmarshal(body, &helcimErr); err != nil {
            return nil, fmt.Errorf("failed to parse error response: %w", err)
        }
        helcimErr.StatusCode = resp.StatusCode
        return nil, &helcimErr
    }

    var response HelcimResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return nil, fmt.Errorf("failed to parse success response: %w", err)
    }

    return &response, nil
}
```

### Retry Logic with Exponential Backoff
```go
func retryWithBackoff(operation func() error, maxRetries int) error {
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }

        // Check if error is retryable
        if !isRetryableError(err) {
            return err
        }

        // Calculate backoff delay
        delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        log.Printf("Attempt %d failed, retrying in %v: %v", attempt+1, delay, err)
        time.Sleep(delay)
    }

    return fmt.Errorf("operation failed after %d attempts", maxRetries)
}

func isRetryableError(err error) bool {
    if helcimErr, ok := err.(*HelcimError); ok {
        // Retry on server errors and rate limits
        return helcimErr.StatusCode >= 500 || helcimErr.StatusCode == 429
    }
    return false
}
```

### Validation Helpers
```go
func validateHelcimResponse(resp *HelcimInitializeResponse) error {
    if resp.CheckoutToken == "" {
        return fmt.Errorf("missing checkout token in response")
    }
    if resp.SecretToken == "" {
        return fmt.Errorf("missing secret token in response")
    }
    return nil
}

func logHelcimError(err error, context string) {
    if helcimErr, ok := err.(*HelcimError); ok {
        log.Printf("Helcim API error in %s: Status=%d, Errors=%+v", 
            context, helcimErr.StatusCode, helcimErr.Errors)
    } else {
        log.Printf("Error in %s: %v", context, err)
    }
}
```

## Specific Error Scenarios for AVR

### 1. Donation Amount Validation
```go
func validateDonationAmount(amount float64) error {
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
```

### 2. API Token Validation
```go
func validateAPIToken(token string) error {
    if token == "" {
        return fmt.Errorf("HELCIM_PRIVATE_API_KEY is not set")
    }
    if len(token) < 30 {
        return fmt.Errorf("API token appears to be truncated (length: %d, expected at least: 30)", len(token))
    }
    return nil
}
```

### 3. Rate Limiting Handler
```go
func handleRateLimit(c *gin.Context, clientIP string) {
    log.Printf("Rate limit exceeded for IP: %s", clientIP)
    c.Header("Retry-After", "60") // Suggest retry after 60 seconds
    c.JSON(http.StatusTooManyRequests, gin.H{
        "error": "Too many requests. Please wait before trying again.",
        "retry_after": 60,
    })
}
```

### 4. Webhook Signature Validation
```go
func validateWebhookSignature(signature, timestamp, webhookID string, body []byte) error {
    if signature == "" {
        return fmt.Errorf("missing webhook signature")
    }
    if timestamp == "" {
        return fmt.Errorf("missing webhook timestamp")
    }
    if webhookID == "" {
        return fmt.Errorf("missing webhook ID")
    }

    // Validate timestamp age
    if !isValidWebhookTimestamp(timestamp) {
        return fmt.Errorf("webhook timestamp is too old or invalid")
    }

    // Verify signature
    if !verifyWebhookSignature(signature, timestamp, webhookID, body) {
        return fmt.Errorf("webhook signature verification failed")
    }

    return nil
}
```

## Frontend Error Handling

### JavaScript Error Display
```javascript
function displayDonationError(error) {
    const errorElement = document.getElementById('transactionDetails');
    
    let errorMessage = 'An error occurred processing your donation.';
    
    // Handle specific error types
    if (error.includes('amount')) {
        errorMessage = 'Please check the donation amount and try again.';
    } else if (error.includes('email')) {
        errorMessage = 'Please check your email address and try again.';
    } else if (error.includes('rate limit')) {
        errorMessage = 'Too many requests. Please wait a moment and try again.';
    }
    
    errorElement.innerHTML = `
        <div class="alert alert-error">
            <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>${errorMessage}</span>
        </div>
    `;
}
```

## Monitoring and Alerting

### Error Metrics to Track
1. **Authentication Errors**
   - Invalid API token frequency
   - Permission errors

2. **Payment Errors**
   - Declined transaction rates
   - API timeout frequency
   - Rate limit hits

3. **Webhook Errors**
   - Signature verification failures
   - Processing errors
   - Timeout/retry counts

### Logging Patterns
```go
// Structured error logging
func logAPIError(operation string, err error, metadata map[string]interface{}) {
    logEntry := map[string]interface{}{
        "timestamp": time.Now().UTC(),
        "operation": operation,
        "error":     err.Error(),
        "level":     "error",
    }
    
    // Add metadata
    for k, v := range metadata {
        logEntry[k] = v
    }
    
    logJSON, _ := json.Marshal(logEntry)
    log.Println(string(logJSON))
}

// Usage example
logAPIError("helcim_initialize", err, map[string]interface{}{
    "amount":     request.Amount,
    "client_ip":  clientIP,
    "user_agent": userAgent,
})
```

## Recovery Procedures

### 1. API Token Issues
1. Check `.env` file for correct token
2. Verify token is active in Helcim dashboard
3. Generate new token if compromised
4. Update all environments with new token

### 2. Payment Processing Issues
1. Check Helcim system status
2. Verify network connectivity
3. Review recent API changes
4. Fallback to manual processing if needed

### 3. Webhook Delivery Issues
1. Check webhook URL accessibility
2. Verify SSL certificate validity
3. Review firewall/security settings
4. Check webhook logs in Helcim dashboard

## Testing Error Conditions

### Unit Tests for Error Handling
```go
func TestHelcimErrorParsing(t *testing.T) {
    tests := []struct {
        name           string
        statusCode     int
        responseBody   string
        expectedError  string
    }{
        {
            name:         "Unauthorized error",
            statusCode:   401,
            responseBody: `{"errors": "Unauthorized"}`,
            expectedError: "Helcim API error 401: Unauthorized",
        },
        {
            name:         "Validation errors",
            statusCode:   400,
            responseBody: `{"errors": {"amount": "Invalid amount"}}`,
            expectedError: "Helcim API error 400:",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test error parsing logic
        })
    }
}
```

---

*This error handling reference should be used throughout the AVR Helcim integration to ensure robust error handling and recovery.*

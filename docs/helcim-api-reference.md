# Helcim API Reference Guide

This document serves as a comprehensive reference for integrating with the Helcim API for the AVR NPO donation system.

## Overview

Helcim is a payment processing company that provides APIs for handling credit card transactions, invoicing, customer management, and more. This guide focuses on the functionality needed for the AVR donation system.

## Authentication

### API Token Requirements
- All API requests require a valid `api-token` in the request header
- Tokens are generated through the Helcim platform in the API Access Configuration section
- Each token has specific permissions that control what endpoints can be accessed

### Headers Required
```http
api-token: your-api-token-here
Content-Type: application/json
Accept: application/json
```

### Testing Connection
Use the connection test endpoint to verify your API token:
```bash
curl -X GET "https://api.helcim.com/v2/connection-test" \
  -H "api-token: your-api-token"
```

**Success Response:**
```json
{
  "message": "Connected Successfully"
}
```

**Error Response:**
```json
{
  "errors": "Unauthorized"
}
```

### API Token Security Best Practices
- **Never expose tokens in client-side code** - only use on backend servers
- Store tokens in environment variables, not in source code
- Obscure all but the last 4 digits when sharing screenshots or logs
- Minimum token length is approximately 30 characters
- If compromised, immediately disable the token in the Helcim dashboard

## Base URLs

### Production
```
https://api.helcim.com/v2/
```

### Testing
- Use production URLs with test credit card numbers
- Test cards are provided in Helcim documentation

## Common HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | Success | Request completed successfully |
| 400 | Bad Request | Invalid request data or missing required fields |
| 401 | Unauthorized | Invalid or missing API token |
| 403 | Forbidden | API token lacks required permissions |
| 404 | Not Found | Endpoint or resource doesn't exist |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Helcim system error |

## Error Response Format

All Helcim API errors follow this format:
```json
{
  "errors": {
    "field_name": "Error message description",
    "another_field": "Another error message"
  }
}
```

Or for general errors:
```json
{
  "errors": "Error message"
}
```

## Rate Limiting

- Helcim implements rate limiting on API endpoints
- Specific limits are not documented but should be handled gracefully
- Implement exponential backoff for 429 responses
- Consider implementing client-side rate limiting for donation endpoints

## API Endpoints Used by AVR

### 1. HelcimPay Initialize (Currently Used)
**Purpose:** Create a checkout session for HelcimPay.js integration

**Endpoint:** `POST /helcim-pay/initialize`

**Request Body:**
```json
{
  "paymentType": "purchase",
  "amount": 100.00,
  "currency": "USD",
  "customer": {
    "firstName": "John",
    "lastName": "Doe", 
    "email": "john@example.com"
  },
  "companyName": "American Veterans Rebuilding"
}
```

**Response:**
```json
{
  "checkoutToken": "49702e8e38d5db9226b54f",
  "secretToken": "1d4b6437a8aabfe4b0ed93"
}
```

**Important Notes:**
- Tokens expire after 60 minutes
- Must be called from backend server (CORS restrictions on frontend)
- `secretToken` used for transaction validation
- `checkoutToken` used to render payment modal

### 2. Card Transaction API (Future Use)
**Purpose:** Retrieve transaction details after payment

**Endpoint:** `GET /card-transactions/{transactionId}`

**Response includes:**
- Transaction status (approved, declined, etc.)
- Amount and currency
- Customer information
- Payment method details
- Timestamps

## Next Phase: Webhooks Integration

### Webhook Configuration
1. Log into Helcim account
2. Navigate to All Tools → Integrations → Webhooks
3. Toggle "Webhooks ON"
4. Set Deliver URL (must be HTTPS, cannot contain "Helcim" in URL)
5. Enable "Notify events for Transactions"

### Webhook Event Format
```json
{
  "id": "25764674",
  "type": "cardTransaction"
}
```

### Webhook Headers
```http
webhook-signature: v1,CsvqmJB7JYdg74tlxbIdXe63H62QMOrMALNw51V/uYU=
webhook-timestamp: 1716412291
webhook-id: msg_2gq5VYqF4DlzM66mCpaXtsEBAkp
```

### Webhook Verification (Node.js Example)
```javascript
const crypto = require('crypto');

// Extract from webhook headers
const webhook_signature = "v1,CsvqmJB7JYdg74tlxbIdXe63H62QMOrMALNw51V/uYU=";
const webhook_timestamp = "1716412291";
const webhook_id = "msg_2gq5VYqF4DlzM66mCpaXtsEBAkp";
const body = JSON.stringify(request.body);

// Construct signed content
const signedContent = `${webhook_id}.${webhook_timestamp}.${body}`;

// Get verifier token from Helcim webhook settings
const verifierToken = process.env.HELCIM_WEBHOOK_VERIFIER_TOKEN;
const verifierTokenBytes = Buffer.from(verifierToken, "base64");

// Generate signature
const generated_signature = crypto
  .createHmac('sha256', verifierTokenBytes)
  .update(signedContent)
  .digest('base64');

// Verify signature
if (webhook_signature.includes(generated_signature)) {
  // Valid webhook - process the event
} else {
  // Invalid webhook - reject
}
```

### Webhook Retry Schedule
Failed webhooks are retried with this schedule:
- Immediately
- 5 seconds
- 5 minutes  
- 30 minutes
- 2 hours
- 5 hours
- 10 hours
- 10 hours (additional)

## Integration Patterns for Go

### Current Pattern (HelcimPay Initialize)
```go
type HelcimRequest struct {
    PaymentType string  `json:"paymentType"`
    Amount      float64 `json:"amount"`
    Currency    string  `json:"currency"`
    Customer    struct {
        FirstName string `json:"firstName"`
        LastName  string `json:"lastName"`
        Email     string `json:"email"`
    } `json:"customer"`
    CompanyName string `json:"companyName"`
}

type HelcimResponse struct {
    CheckoutToken string `json:"checkoutToken"`
    SecretToken   string `json:"secretToken"`
}
```

### Future Pattern (Webhook Handler)
```go
type WebhookEvent struct {
    ID   string `json:"id"`
    Type string `json:"type"`
}

func (w *WebhookHandler) VerifySignature(signature, timestamp, id string, body []byte) bool {
    // Implementation based on Node.js example above
}

func (w *WebhookHandler) HandleEvent(event WebhookEvent) error {
    switch event.Type {
    case "cardTransaction":
        return w.handleTransactionEvent(event.ID)
    default:
        return fmt.Errorf("unknown event type: %s", event.Type)
    }
}
```

## Common Integration Issues

### 1. CORS Errors
- **Problem:** Frontend making direct API calls to Helcim
- **Solution:** All Helcim API calls must be made from backend server

### 2. Token Expiration
- **Problem:** Checkout tokens expire after 60 minutes
- **Solution:** Generate new tokens for each payment session

### 3. Authentication Errors
- **Problem:** "Unauthorized" or "No access permission" errors
- **Solution:** Verify API token and permissions in Helcim dashboard

### 4. Missing Required Data
- **Problem:** 400 errors about missing fields
- **Solution:** Check error response for specific missing fields

### 5. PCI Compliance
- **Problem:** "Not allowed to send full card number" error
- **Solution:** Use HelcimPay.js for card tokenization, never handle raw card data

## Testing

### Test Credit Card Numbers
Use Helcim's provided test card numbers for development:
- Check Helcim developer documentation for current test card numbers
- Test both successful and declined transaction scenarios

### API Testing Commands
```bash
# Test connection
curl -X GET "https://api.helcim.com/v2/connection-test" \
  -H "api-token: ${HELCIM_PRIVATE_API_KEY}"

# Test HelcimPay initialize
curl -X POST "https://api.helcim.com/v2/helcim-pay/initialize" \
  -H "api-token: ${HELCIM_PRIVATE_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "paymentType": "purchase",
    "amount": 25.00,
    "currency": "USD",
    "companyName": "American Veterans Rebuilding"
  }'
```

## Environment Variables

Required environment variables for Helcim integration:
```bash
# Primary API token for payment processing
HELCIM_PRIVATE_API_KEY=your-api-token-here

# Future: Webhook verification token
HELCIM_WEBHOOK_VERIFIER_TOKEN=your-webhook-verifier-token

# Future: Webhook endpoint URL
HELCIM_WEBHOOK_URL=https://yourdomain.com/api/webhooks/helcim
```

## Resources

- [Helcim Developer Documentation](https://devdocs.helcim.com/)
- [API Reference](https://devdocs.helcim.com/reference)
- [Helcim Support](https://devdocs.helcim.com/docs/get-help)
- [PCI Compliance Guide](https://devdocs.helcim.com/docs/pci-compliance-scope)

---

*This document should be updated as new Helcim features are implemented in the AVR donation system.*

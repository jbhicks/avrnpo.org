# Helcim API Reference Guide

This document serves as a comprehensive reference for integrating with the Helcim API for the AVR NPO donation system.

## 🚨 CRITICAL: Official HelcimPay.js Integration Only

**⚠️ IMPORTANT:** This project now uses the **official HelcimPay.js integration** from Helcim. Previous implementations using custom modals were incorrect and have been replaced.

**Official Helcim Documentation:** https://devdocs.helcim.com/docs/overview-of-helcimpayjs

## Overview

Helcim is a payment processing company that provides APIs for handling credit card transactions, invoicing, customer management, and more. This guide focuses on the functionality needed for the AVR donation system using the **official HelcimPay.js library**.

## Correct Integration Architecture

### Backend (Go/Buffalo)
- Calls `/helcim-pay/initialize` API endpoint to get `checkoutToken` and `secretToken`
- Handles webhooks for payment confirmation
- Processes payment completion callbacks

### Frontend (JavaScript)
- Loads official HelcimPay.js: `https://secure.helcim.app/helcim-pay/services/start.js`
- Uses `appendHelcimPayIframe(checkoutToken)` to display payment modal
- Listens for postMessage events from the Helcim iframe
- Uses `removeHelcimPayIframe()` to clean up after payment

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

## Official HelcimPay.js Implementation

### 1. Frontend Integration

**Load the Official Library:**
```html
<script type="text/javascript" src="https://secure.helcim.app/helcim-pay/services/start.js"></script>
```

**Display Payment Modal:**
```javascript
// After getting checkoutToken from backend
appendHelcimPayIframe(checkoutToken);
```

**Listen for Payment Events:**
```javascript
window.addEventListener('message', (event) => {
  const helcimPayJsIdentifierKey = 'helcim-pay-js-' + checkoutToken;
  
  if (event.data.eventName === helcimPayJsIdentifierKey) {
    
    if (event.data.eventStatus === 'SUCCESS') {
      // Payment successful - eventMessage contains transaction data
      const transactionData = JSON.parse(event.data.eventMessage);
      handlePaymentSuccess(transactionData);
    }
    
    if (event.data.eventStatus === 'ABORTED') {
      // Payment failed
      handlePaymentError(event.data.eventMessage);
    }
    
    if (event.data.eventStatus === 'HIDE') {
      // Modal closed without payment
      handlePaymentCancelled();
    }
  }
});
```

**Clean Up After Payment:**
```javascript
// Remove the iframe and event listeners
removeHelcimPayIframe();
window.removeEventListener('message', eventHandler);
```

### 2. Backend API Integration

## API Endpoints Used by AVR
#### HelcimPay Initialize (Currently Used)

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
- `checkoutToken` used with `appendHelcimPayIframe(checkoutToken)` function

#### Transaction Response Format

When a payment is successful, the Helcim modal emits a postMessage event with this structure:

```javascript
{
  eventName: "helcim-pay-js-" + checkoutToken,
  eventStatus: "SUCCESS", // or "ABORTED" or "HIDE"
  eventMessage: "{\"status\":200,\"data\":{\"data\":{\"amount\":\"100.00\",\"approvalCode\":\"T5E5ST\",\"avsResponse\":\"X\",\"cardBatchId\":\"2578965\",\"cardHolderName\":\"Jane Doe\",\"cardNumber\":\"5454545454\",\"cardToken\":\"684a4a03400fadd1e7bdc9\",\"currency\":\"CAD\",\"customerCode\":\"CST1200\",\"dateCreated\":\"2022-01-05 12:30:45\",\"invoiceNumber\":\"INV000010\",\"status\":\"APPROVED\",\"transactionId\":\"17701631\",\"type\":\"purchase\"},\"hash\":\"dbcb570cca52c38d597941adbed03f01be78c43cba89048722925b2f168226a9\"}}"
}
```

**Note:** The `eventMessage` is a JSON.stringify'd string that needs to be parsed to access transaction details.

### 🚨 Previous Incorrect Implementation (FIXED)

**What Was Wrong:**
- Used a custom local modal instead of official HelcimPay.js
- Created `/js/helcim-pay.min.js` as a custom implementation
- Did not use the official Helcim iframe and postMessage system
- Manually styled and created payment forms (incorrect for PCI compliance)

**What We Fixed:**
- Replaced custom modal with official `https://secure.helcim.app/helcim-pay/services/start.js`
- Use `appendHelcimPayIframe(checkoutToken)` function provided by Helcim
- Listen for official postMessage events from Helcim's secure iframe
- Removed all custom payment form code - Helcim handles the secure payment collection

**Why This Matters:**
- **PCI Compliance:** Only Helcim's official iframe is PCI compliant
- **Security:** Custom payment forms can expose card data to our servers
- **Updates:** Official library gets security updates automatically
- **Features:** Access to all Helcim features (digital wallets, ACH, etc.)

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
**⚠️ IMPORTANT:** Official Helcim test cards - only work with test accounts!

| Card Type | Number | Expiry | CVV | Limit |
|-----------|--------|--------|-----|-------|
| **Visa** | `4124939999999990` | 01/28 | 100 | $100 |
| **Visa** | `4000000000000028` | 01/28 | 100 | $100 |
| **Mastercard** | `5413330089099130` | 01/28 | 100 | $100 |
| **Mastercard** | `5413330089020011` | 01/28 | 100 | $100 |
| **American Express** | `374245001751006` | 01/28 | 1000 | $1000 |

**Requirements:** 
- Must have Helcim developer test account (contact tier2support@helcim.com)
- Test cards will be declined on production accounts
- All test cards use expiry 01/28 and CVV 100 (except Amex: 1000)

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

## 📧 CRITICAL: Helcim Email Functionality (Updated June 24, 2025)

### ❌ HELCIM DOES NOT PROVIDE EMAIL SERVICES

**Key Discovery**: Helcim does **NOT** automatically send emails when payments are processed.

#### What Helcim Provides:
- ✅ **Payment Processing**: Credit card and ACH transaction processing
- ✅ **Webhooks**: Real-time payment status notifications to your server
- ✅ **Transaction Data**: Complete payment details via API responses
- ❌ **NO Email Services**: No automatic receipt sending or email functionality

#### What You Must Implement:
- **Custom Email Service**: Your application must handle all email communications
- **Receipt Generation**: Create and send donation receipts via your own SMTP service
- **Email Templates**: Design and maintain your own email templates
- **SMTP Configuration**: Set up email delivery through services like Gmail, SendGrid, etc.

### ✅ AVR NPO Email Implementation

The AVR donation system includes a complete email service implementation:

#### Email Service Features:
- **HTML & Text Receipts**: Professional donation receipts with AVR branding
- **501(c)(3) Compliance**: Tax-deductible information and EIN details
- **Automatic Sending**: Triggered by successful payment webhooks
- **SMTP Integration**: Works with Gmail, SendGrid, Mailgun, etc.
- **Error Handling**: Graceful fallback if email delivery fails

#### Configuration Required:
```bash
# SMTP settings for email delivery
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
FROM_EMAIL=donations@avrnpo.org
FROM_NAME=American Veterans Rebuilding
```

#### Email Flow:
```
Payment Success → Helcim Webhook → Your Server → Generate Receipt → Send Email
```

**Important**: Never rely on Helcim for email functionality. Always implement your own email service for donation receipts and donor communications.

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

## 🚨 CRITICAL: Helcim Recurring API Integration (Updated June 24, 2025)

**⚠️ IMPORTANT DISCOVERIES:** Through implementation and testing, we've learned critical details about Helcim's Recurring API that differ from initial assumptions.

### ✅ VERIFIED: Official Helcim Recurring API Structure

#### Payment Plans API (`POST /v2/payment-plans`)
- **Request Format**: Must use `paymentPlans` array wrapper
- **Required Fields**: `name`, `type`, `currency`, `recurringAmount`, `billingPeriod`
- **Response Format**: Returns array of created payment plans
- **ID Type**: Payment plan IDs are integers, not strings

```json
// CORRECT Payment Plan Request Structure
{
  "paymentPlans": [
    {
      "name": "Monthly Donation - $25.00",
      "description": "Monthly donation plan for $25.00",
      "type": "subscription",
      "currency": "USD", 
      "recurringAmount": 25.00,
      "billingPeriod": "monthly",
      "billingPeriodIncrements": 1,
      "dateBilling": "Sign-up",
      "termType": "forever",
      "paymentMethod": "card",
      "taxType": "no_tax",
      "status": "active"
    }
  ]
}
```

#### Subscriptions API (`POST /v2/subscriptions`)
- **Request Format**: Must use `subscriptions` array wrapper
- **Customer Linking**: Uses `customerId` from HelcimPay.js verify process
- **Payment Plan Linking**: Uses integer `paymentPlanId` from payment plan creation
- **Activation**: Uses `activationDate` field for immediate or scheduled activation

```json
// CORRECT Subscription Request Structure
{
  "subscriptions": [
    {
      "customerId": "customer_code_from_helcimpay",
      "paymentPlanId": 12345,
      "amount": 25.00,
      "paymentMethod": "card",
      "activationDate": "2025-06-24"
    }
  ]
}
```

### 🔄 RECOMMENDED IMPLEMENTATION FLOW

#### Step 1: HelcimPay.js Verify Mode
- Use `paymentType: "verify"` to collect payment details
- Creates customer in Helcim system with stored payment method
- Returns `customerCode` for subscription creation

#### Step 2: Payment Plan Creation
- Create payment plan using official Payment Plans API
- Store plan ID for subscription creation
- Plans can be reused for same donation amounts

#### Step 3: Subscription Creation
- Link customer to payment plan via Subscriptions API
- Set activation date for immediate or future billing
- Store subscription ID for management

### ❌ COMMON MISTAKES TO AVOID

1. **Wrong Request Structure**: Don't send individual objects, use array wrappers
2. **Incorrect Field Names**: Use `name` not `planName`, `customerId` not `customerCode`
3. **Wrong Data Types**: Payment plan IDs are integers, not strings
4. **Missing Customer Creation**: Must use HelcimPay.js verify mode first
5. **Incorrect Payment Method**: Use "card" not "cc" for credit cards

# Helcim Integration Guide

Complete technical guide for integrating with the Helcim payment processor for the AVR NPO donation system.

## 🚨 CRITICAL: Official Integration Only

**⚠️ IMPORTANT:** This project uses the **official HelcimPay.js integration** from Helcim. Any custom modal implementations are incorrect and should not be used.

**Official Documentation:** https://devdocs.helcim.com/docs/overview-of-helcimpayjs

## 🏗️ Integration Architecture

### Two-Tier Payment Processing

**Frontend (HelcimPay.js):**
- Secure, PCI-compliant payment data collection
- Tokenizes payment methods without storing card data
- Provides iframe-based payment forms
- Handles all sensitive payment information

**Backend (Go/Buffalo + Helcim APIs):**
- Initializes payment sessions via `/helcim-pay/initialize`
- Processes completions via Payment and Recurring APIs
- Handles webhooks for real-time status updates
- Manages donation records and user accounts

### Payment Flow Types

**One-Time Donations:**
1. HelcimPay.js collects payment data → card token
2. Backend calls Payment API `purchase` with token
3. Immediate processing and confirmation

**Recurring Donations:**
1. HelcimPay.js collects payment data → card token + customer ID
2. Backend creates subscription via Recurring API
3. Automatic monthly processing by Helcim

## 🔐 Authentication & Security

### API Token Management
- All API requests require `api-token` header
- **NEVER expose tokens in client-side code** 
- Store in environment variables only
- Minimum 30 character length
- Immediate revocation if compromised

### Environment Variables
```bash
# Required Helcim credentials
HELCIM_API_TOKEN=your-api-token-here
HELCIM_MERCHANT_ID=your-merchant-id-here

# Webhook verification
HELCIM_WEBHOOK_VERIFIER_TOKEN=your-webhook-verifier-here

# Application URLs
HELCIM_SUCCESS_URL=https://yourdomain.com/donation/success
HELCIM_ERROR_URL=https://yourdomain.com/donation/error
```

### Headers Required
```http
api-token: your-api-token-here
Content-Type: application/json
Accept: application/json
```

## 🔌 API Endpoints

### Base URLs
- **Production:** `https://api.helcim.com/v2/`
- **Testing:** Use production URLs with test card numbers

### Key Endpoints

#### Payment Initialization
```
POST /helcim-pay/initialize
```

**Request:**
```json
{
  "paymentType": "purchase",
  "amount": 100.00,
  "currency": "CAD",
  "customerCode": "unique-customer-id",
  "invoiceNumber": "inv-001",
  "successUrl": "https://yourdomain.com/donation/success",
  "errorUrl": "https://yourdomain.com/donation/error"
}
```

**Response:**
```json
{
  "checkoutToken": "token-for-frontend",
  "secretToken": "token-for-backend-verification"
}
```

#### Payment Processing (Direct API)
```
POST /payment
```

**For One-Time Payments:**
```json
{
  "paymentType": "purchase",
  "amount": 100.00,
  "currency": "CAD",
  "customerCode": "customer-123",
  "cardToken": "token-from-helcimpay-js"
}
```

#### Recurring Payment Plans
```
POST /payment-plans
```

**Create Payment Plan:**
```json
{
  "planName": "Monthly Donation - $50",
  "amount": 50.00,
  "frequency": "monthly",
  "currency": "CAD"
}
```

#### Customer Management
```
POST /customers
```

**Create Customer:**
```json
{
  "contactName": "John Doe",
  "email": "john@example.com",
  "billingAddress": {
    "name": "John Doe",
    "street1": "123 Main St",
    "city": "Toronto",
    "province": "ON", 
    "country": "CA",
    "postalCode": "M5V 3A3"
  }
}
```

#### Subscription Management
```
POST /recurring
```

**Create Subscription:**
```json
{
  "customerId": "customer-id-from-creation",
  "paymentPlanId": "plan-id-from-creation", 
  "amount": 50.00,
  "paymentMethod": "cc",
  "cardToken": "token-from-helcimpay-js"
}
```

**Cancel Subscription:**
```
POST /recurring/{subscriptionId}/cancel
```

**Get Subscription Details:**
```
GET /recurring/{subscriptionId}
```

## 📨 Webhook Implementation

### Webhook Events
- `payment.success` - Payment processed successfully
- `payment.declined` - Payment was declined  
- `payment.refunded` - Payment was refunded
- `payment.cancelled` - Payment was cancelled

### Webhook Handler (Go)
```go
type HelcimWebhookEvent struct {
    ID     string                 `json:"id"`
    Type   string                 `json:"type"`
    Object map[string]interface{} `json:"object"`
}

func (app *App) HelcimWebhookHandler(c buffalo.Context) error {
    // Verify webhook signature
    signature := c.Request().Header.Get("X-Helcim-Signature")
    if !verifyWebhookSignature(c.Request().Body, signature) {
        return c.Error(http.StatusUnauthorized, errors.New("invalid signature"))
    }
    
    var event HelcimWebhookEvent
    if err := c.Bind(&event); err != nil {
        return c.Error(http.StatusBadRequest, err)
    }
    
    // Process event based on type
    switch event.Type {
    case "payment.success":
        return app.handlePaymentSuccess(c, event)
    case "payment.declined":
        return app.handlePaymentDeclined(c, event)
    // ... other event types
    }
    
    return c.Render(http.StatusOK, r.JSON(map[string]string{
        "status": "received",
    }))
}
```

## 🧪 Testing

### Test Credit Cards
```
Visa: 4111 1111 1111 1111
Mastercard: 5555 5555 5555 4444
Amex: 3714 496353 98431
```

**Test CVV:** Any 3-digit number  
**Test Expiry:** Any future date

### Testing Scenarios
1. **Successful Payment** - Use valid test card with amount > $1.00
2. **Declined Payment** - Use amount exactly $0.05
3. **Processing Error** - Use amount exactly $0.01
4. **Webhook Testing** - Set up ngrok for local development

### Connection Test
```bash
curl -X GET "https://api.helcim.com/v2/connection-test" \
  -H "api-token: your-api-token"
```

**Expected Success Response:**
```json
{
  "message": "Connected Successfully"
}
```

## 🚨 Error Handling

### Common HTTP Status Codes
| Code | Meaning | Description |
|------|---------|-------------|
| 200 | Success | Request completed successfully |
| 400 | Bad Request | Invalid request data or missing fields |
| 401 | Unauthorized | Invalid or missing API token |
| 403 | Forbidden | Token lacks required permissions |
| 404 | Not Found | Endpoint or resource doesn't exist |
| 429 | Too Many Requests | Rate limit exceeded |

### Error Response Format
```json
{
  "errors": {
    "amount": "Amount must be greater than 0",
    "currency": "Currency is required"
  }
}
```

### Error Handling Best Practices
1. **Log all API responses** for debugging
2. **Graceful degradation** - don't break donation flow
3. **User-friendly messages** - translate technical errors
4. **Retry logic** - for transient network issues
5. **Monitoring** - alert on repeated failures

## 🔧 Go Implementation Examples

### Helcim Service Client
```go
type HelcimService struct {
    APIToken   string
    BaseURL    string
    HTTPClient *http.Client
}

func NewHelcimService(apiToken string) *HelcimService {
    return &HelcimService{
        APIToken:   apiToken,
        BaseURL:    "https://api.helcim.com/v2",
        HTTPClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (h *HelcimService) makeRequest(method, endpoint string, payload interface{}) (*http.Response, error) {
    var body io.Reader
    if payload != nil {
        jsonPayload, err := json.Marshal(payload)
        if err != nil {
            return nil, err
        }
        body = bytes.NewReader(jsonPayload)
    }
    
    req, err := http.NewRequest(method, h.BaseURL+endpoint, body)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("api-token", h.APIToken)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    
    return h.HTTPClient.Do(req)
}
```

### Payment Processing
```go
func (h *HelcimService) ProcessPayment(amount float64, cardToken, customerCode string) (*PaymentResponse, error) {
    payload := PaymentRequest{
        PaymentType:  "purchase",
        Amount:       amount,
        Currency:     "CAD", 
        CardToken:    cardToken,
        CustomerCode: customerCode,
    }
    
    resp, err := h.makeRequest("POST", "/payment", payload)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("payment failed with status: %d", resp.StatusCode)
    }
    
    var result PaymentResponse
    return &result, json.NewDecoder(resp.Body).Decode(&result)
}
```

## 📋 Implementation Checklist

### Initial Setup
- [ ] Obtain Helcim API credentials
- [ ] Configure environment variables
- [ ] Test API connection
- [ ] Set up webhook endpoint

### Frontend Integration  
- [ ] Load official HelcimPay.js library
- [ ] Implement payment form with HelcimPay.js
- [ ] Handle success/error callbacks
- [ ] Test payment flow end-to-end

### Backend Integration
- [ ] Implement payment initialization endpoint  
- [ ] Create webhook handler
- [ ] Add payment completion processing
- [ ] Implement error handling and logging

### Production Readiness
- [ ] SSL/HTTPS configuration
- [ ] Security audit of API token usage
- [ ] Webhook signature verification
- [ ] Monitoring and alerting setup
- [ ] Payment reconciliation procedures

For specific implementation details, see the related guides:
- [Donation Flow](./donation-flow.md) - Frontend form implementation
- [Recurring Payments](./recurring-payments-final.md) - **PRODUCTION READY** subscription system  
- [Webhooks](./webhooks.md) - Event processing
- [Testing](./testing.md) - Testing procedures and test cards

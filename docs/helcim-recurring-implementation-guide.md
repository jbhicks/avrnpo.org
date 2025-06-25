# Helcim Recurring Donations Implementation Guide

*Updated: June 24, 2025*

This document captures the verified implementation approach for Helcim recurring donations based on actual testing and integration with the official Helcim API.

## 🎯 VERIFIED IMPLEMENTATION APPROACH

### Official Helcim Documentation Sources
- **Recurring API**: https://devdocs.helcim.com/docs/recurring-api
- **Payment Plans**: https://devdocs.helcim.com/docs/recurring-payment-plans
- **Subscriptions**: https://devdocs.helcim.com/docs/recurring-subscriptions
- **API Reference**: https://devdocs.helcim.com/reference/payment-plan-create

## 🔄 CORRECT INTEGRATION FLOW

### Step 1: Customer Creation via HelcimPay.js
**Helcim Recommendation**: Use HelcimPay.js with `paymentType: "verify"`

```javascript
// Frontend: Initialize HelcimPay.js in verify mode
const helcimRequest = {
    paymentType: "verify",
    amount: 0,  // $0 for verification
    currency: "USD",
    customerRequest: {
        contactName: "John Doe",
        email: "john@example.com",
        billingAddress: {
            name: "John Doe",
            street1: "123 Main St",
            city: "Calgary",
            province: "AB",
            country: "CA",
            postalCode: "T2P 1J9"
        }
    }
};
```

**Result**: Creates customer in Helcim system with stored payment method, returns `customerCode`.

### Step 2: Payment Plan Creation
**API Endpoint**: `POST /v2/payment-plans`

```json
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

**Response Structure**:
```json
[
  {
    "id": 12345,
    "name": "Monthly Donation - $25.00",
    "recurringAmount": 25.00,
    "billingPeriod": "monthly",
    "status": "active"
  }
]
```

### Step 3: Subscription Creation
**API Endpoint**: `POST /v2/subscriptions`

```json
{
  "subscriptions": [
    {
      "customerId": "customer_code_from_step_1",
      "paymentPlanId": 12345,
      "amount": 25.00,
      "paymentMethod": "card",
      "activationDate": "2025-06-24"
    }
  ]
}
```

**Response Structure**:
```json
[
  {
    "id": 67890,
    "customerId": "customer_code",
    "paymentPlanId": 12345,
    "amount": 25.00,
    "status": "active",
    "nextBillingDate": "2025-07-24T00:00:00Z"
  }
]
```

## ⚙️ IMPLEMENTATION DETAILS

### Payment Plan Configuration
- **Type**: Use `"subscription"` for immediate billing on sign-up
- **Billing Period**: `"monthly"` with `billingPeriodIncrements: 1`
- **Term Type**: `"forever"` for indefinite recurring billing
- **Tax Type**: `"no_tax"` for tax-exempt donations (501c3)

### Subscription Configuration
- **Activation**: Use current date for immediate activation
- **Payment Method**: Use `"card"` for credit card payments
- **Amount**: Can override payment plan amount for custom donations

### Data Types
- **Payment Plan ID**: Integer (e.g., 12345)
- **Subscription ID**: Integer (e.g., 67890)
- **Customer ID**: String (e.g., "customer_abc123")
- **Amounts**: Float64 (e.g., 25.00)

## 🚨 CRITICAL REQUIREMENTS

### 1. Array-Based Requests
All creation endpoints expect array wrappers:
- Payment Plans: `{"paymentPlans": [...]}`
- Subscriptions: `{"subscriptions": [...]}`

### 2. Customer Must Exist First
- Cannot create subscription without existing customer
- Must use HelcimPay.js verify mode or Customer API
- Customer needs stored payment method

### 3. Payment Plan Must Exist
- Cannot create subscription without payment plan
- Can reuse payment plans for same amounts
- Plan defines billing behavior

### 4. Response Parsing
- All creation endpoints return arrays
- Must parse first element for single creations
- Check array length before accessing elements

## 🔧 GO IMPLEMENTATION STRUCTURES

### Payment Plan Structure
```go
type PaymentPlan struct {
    ID                      int     `json:"id"`
    Name                    string  `json:"name"`
    Description             string  `json:"description"`
    Type                    string  `json:"type"`
    Currency                string  `json:"currency"`
    RecurringAmount         float64 `json:"recurringAmount"`
    BillingPeriod           string  `json:"billingPeriod"`
    BillingPeriodIncrements int     `json:"billingPeriodIncrements"`
    DateBilling             string  `json:"dateBilling"`
    TermType                string  `json:"termType"`
    PaymentMethod           string  `json:"paymentMethod"`
    Status                  string  `json:"status"`
}
```

### Subscription Structures
```go
type SubscriptionRequest struct {
    CustomerID    string  `json:"customerId"`
    PaymentPlanID int     `json:"paymentPlanId"`
    Amount        float64 `json:"amount"`
    PaymentMethod string  `json:"paymentMethod"`
}

type SubscriptionResponse struct {
    ID              int       `json:"id"`
    CustomerID      string    `json:"customerId"`
    PaymentPlanID   int       `json:"paymentPlanId"`
    Amount          float64   `json:"amount"`
    Status          string    `json:"status"`
    ActivationDate  string    `json:"activationDate"`
    NextBillingDate time.Time `json:"nextBillingDate"`
    PaymentMethod   string    `json:"paymentMethod"`
}
```

## ❌ COMMON MISTAKES AVOIDED

1. **Wrong Field Names**: Using `planName` instead of `name`
2. **Incorrect Data Types**: Using strings for integer IDs
3. **Missing Array Wrappers**: Sending objects instead of arrays
4. **Wrong Payment Method**: Using `"cc"` instead of `"card"`
5. **No Customer Creation**: Attempting subscription without customer
6. **Response Parsing**: Not handling array-based responses correctly

## 📧 EMAIL INTEGRATION

**CRITICAL**: Helcim does NOT provide email services. You must implement:

1. **Custom Email Service**: SMTP-based email delivery
2. **Receipt Generation**: HTML/text donation receipts
3. **Webhook Handlers**: Process payment success events
4. **Automatic Sending**: Trigger emails on successful payments

See `/docs/helcim-api-reference.md` for detailed email implementation guidance.

## 🧪 TESTING APPROACH

### Unit Tests
- Mock Helcim API responses
- Test request structure generation
- Verify response parsing
- Test error handling

### Integration Tests
- Use Helcim test API credentials
- Test with official test card numbers
- Verify end-to-end flow
- Monitor webhook delivery

### Production Testing
- Start with small test donations
- Monitor logs for API errors
- Verify subscription creation
- Confirm billing behavior

## 📚 RESOURCES

- [Official Helcim Recurring API](https://devdocs.helcim.com/docs/recurring-api)
- [Payment Plans Documentation](https://devdocs.helcim.com/docs/recurring-payment-plans)
- [Subscriptions Documentation](https://devdocs.helcim.com/docs/recurring-subscriptions)
- [API Reference - Payment Plans](https://devdocs.helcim.com/reference/payment-plan-create)
- [API Reference - Subscriptions](https://devdocs.helcim.com/reference/subscription-create)

---

*This guide is based on actual implementation and testing with Helcim's official APIs as of June 2025.*

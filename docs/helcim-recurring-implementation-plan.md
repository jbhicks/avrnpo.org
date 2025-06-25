 Helcim Recurring Payment Implementation Plan

**Status:** PLANNING PHASE  
**Priority:** HIGH  
**Estimated Time:** 10-12 hours  

## 🎯 OBJECTIVE

Implement true recurring monthly donations using Helcim's Recurring API alongside the existing HelcimPay.js integration.

## 🚨 CURRENT PROBLEM

The donation system has UI for recurring donations but **only processes one-time payments**. Users selecting "Monthly recurring" are charged once, not set up for recurring billing.

## 💡 SOLUTION APPROACH

### Unified API Architecture (CLEANER APPROACH)

**Why this is better than the original plan:**
- Single payment collection method for both one-time and recurring
- Consistent user experience regardless of payment type  
- Unified backend API integration
- Less complex frontend logic

**Two-Step Process (REQUIRED by Helcim's security model):**

1. **Payment Data Collection** (HelcimPay.js with `verify` mode)
   - Use `paymentType: "verify"` and `amount: 0` for ALL donations
   - Collect and vault payment method without charging (PCI compliant)
   - Create customer record in Helcim system
   - Return card token for both one-time and recurring

2. **Payment Processing** (Direct API calls)
   - **One-time donations**: Use Payment API `purchase` with card token
   - **Recurring donations**: Use Recurring API to create subscription with card token
   - Both use the same vaulted payment method

**Key insight:** Instead of mixing HelcimPay.js `purchase` and `verify` modes, use `verify` for everything and handle the actual charging via direct API calls.

## 🔧 TECHNICAL IMPLEMENTATION

### Key API Findings

**Why we need HelcimPay.js (PCI Compliance):**
- Payment API blocks full card numbers by default (requires special approval)
- HelcimPay.js is the secure, PCI-compliant way to collect payment data
- All payment methods (one-time and recurring) need to be tokenized first

**Unified Architecture Benefits:**
- Single payment collection flow using `verify` mode for all donations
- Consistent user experience regardless of payment type
- Backend handles charging via appropriate API (Payment API or Recurring API)
- Cleaner code with less conditional logic

**API Integration Pattern:**
- HelcimPay.js `verify` mode → Card token + Customer ID
- One-time: Payment API `purchase` with card token  
- Recurring: Recurring API subscription with card token

### Backend Changes

#### 1. New Go Structures
```go
// Helcim Recurring API structures
type PaymentPlan struct {
    ID          string  `json:"id"`
    Name        string  `json:"planName"`
    Amount      float64 `json:"amount"`
    Frequency   string  `json:"frequency"` // "monthly"
    Currency    string  `json:"currency"`
}

type CustomerRequest struct {
    ContactName     string `json:"contactName"`
    Email          string `json:"email"`
    BillingAddress struct {
        Name       string `json:"name"`
        Street1    string `json:"street1"`
        City       string `json:"city"`
        Province   string `json:"province"`
        Country    string `json:"country"`
        PostalCode string `json:"postalCode"`
    } `json:"billingAddress"`
}

type SubscriptionRequest struct {
    CustomerID    string  `json:"customerId"`
    PaymentPlanID string  `json:"paymentPlanId"`
    Amount        float64 `json:"amount"`
    PaymentMethod string  `json:"paymentMethod"` // "cc" for credit card
}

type SubscriptionResponse struct {
    ID              string    `json:"id"`
    CustomerID      string    `json:"customerId"`
    PaymentPlanID   string    `json:"paymentPlanId"`
    Amount          float64   `json:"amount"`
    Status          string    `json:"status"`
    NextBillingDate time.Time `json:"nextBillingDate"`
}
```

#### 2. Unified Donation Handler
```go
func DonationInitializeHandler(c buffalo.Context) error {
    // Parse donation request (existing code)
    
    // UNIFIED APPROACH: Always use verify mode for payment collection
    helcimReq := HelcimPayRequest{
        PaymentType: "verify", // Always verify first, charge later via API
        Amount:      0,        // Verify mode requires $0
        Currency:    "USD",
        CustomerRequest: &CustomerRequest{
            ContactName: req.DonorName,
            Email:      req.DonorEmail,
            BillingAddress: BillingAddress{
                Name:       req.DonorName,
                Street1:    req.AddressLine1,
                City:       req.City,
                Province:   req.State,
                Country:    "US",
                PostalCode: req.Zip,
            },
        },
    }
    
    // Store donation details for later processing
    donation := &Donation{
        DonorName:    req.DonorName,
        DonorEmail:   req.DonorEmail,
        Amount:       req.Amount,
        DonationType: req.DonationType, // "one-time" or "monthly"
        Status:       "pending",
        // ...other fields
    }
    
    // Save to database
    tx := c.Value("tx").(*pop.Connection)
    if err := tx.Create(donation); err != nil {
        return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
            "error": "Failed to create donation record",
        }))
    }
    
    // Call Helcim API and return response with donation ID
    response, err := callHelcimAPI(helcimReq)
    if err != nil {
        return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
            "error": "Payment initialization failed",
        }))
    }
    
    response["donationId"] = donation.ID
    return c.Render(http.StatusOK, r.JSON(response))
}
```

#### 3. Payment Processing Handler
```go
func ProcessPaymentHandler(c buffalo.Context) error {
    var req struct {
        CustomerCode string  `json:"customerCode"`
        CardToken    string  `json:"cardToken"`
        DonationID   string  `json:"donationId"`
        Amount       float64 `json:"amount"`
    }
    
    if err := c.Bind(&req); err != nil {
        return c.Render(http.StatusBadRequest, r.JSON(map[string]string{
            "error": "Invalid request data",
        }))
    }
    
    // Get donation record
    tx := c.Value("tx").(*pop.Connection)
    donation := &Donation{}
    if err := tx.Find(donation, req.DonationID); err != nil {
        return c.Render(http.StatusNotFound, r.JSON(map[string]string{
            "error": "Donation not found",
        }))
    }
    
    if donation.DonationType == "monthly" {
        // RECURRING DONATION: Create subscription
        return c.handleRecurringPayment(req, donation)
    } else {
        // ONE-TIME DONATION: Process immediate payment
        return c.handleOneTimePayment(req, donation)
    }
}

func (c ActionContext) handleOneTimePayment(req PaymentRequest, donation *Donation) error {
    // Use Payment API to charge the card token
    paymentReq := PaymentAPIRequest{
        Amount:       donation.Amount,
        Currency:     "USD",
        CustomerCode: req.CustomerCode,
        CardData: CardData{
            CardToken: req.CardToken,
        },
    }
    
    transaction, err := processPaymentAPIPurchase(paymentReq)
    if err != nil {
        return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
            "error": "Payment processing failed",
        }))
    }
    
    // Update donation record
    donation.TransactionID = &transaction.ID
    donation.CustomerID = &req.CustomerCode
    donation.Status = "completed"
    
    tx := c.Value("tx").(*pop.Connection)
    if err := tx.Update(donation); err != nil {
        return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
            "error": "Failed to update donation",
        }))
    }
    
    return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
        "success":       true,
        "transactionId": transaction.ID,
        "type":          "one-time",
    }))
}

func (c ActionContext) handleRecurringPayment(req PaymentRequest, donation *Donation) error {
    // Create or get payment plan
    paymentPlanID, err := getOrCreateMonthlyDonationPlan(donation.Amount)
    if err != nil {
        return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
            "error": "Failed to setup payment plan",
        }))
    }
    
    // Create subscription using Recurring API
    subscription, err := createHelcimSubscription(SubscriptionRequest{
        CustomerID:    req.CustomerCode,
        PaymentPlanID: paymentPlanID,
        Amount:        donation.Amount,
        PaymentMethod: "cc",
    })
    if err != nil {
        return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
            "error": "Failed to create subscription",
        }))
    }
    
    // Update donation record
    donation.SubscriptionID = &subscription.ID
    donation.CustomerID = &req.CustomerCode  
    donation.PaymentPlanID = &paymentPlanID
    donation.Status = "active"
    
    tx := c.Value("tx").(*pop.Connection)
    if err := tx.Update(donation); err != nil {
        return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
            "error": "Failed to update donation",
        }))
    }
    
    return c.Render(http.StatusOK, r.JSON(map[string]interface{}{
        "success":        true,
        "subscriptionId": subscription.ID,
        "nextBilling":    subscription.NextBillingDate,
        "type":           "recurring",
    }))
}
```

### Frontend Changes

#### 1. Unified donation.js
```javascript
handlePaymentSuccess(eventMessage, donationId) {
    console.log('Payment verification completed successfully');
    
    // Parse Helcim response
    let transactionData;
    try {
        transactionData = typeof eventMessage === 'string' ? JSON.parse(eventMessage) : eventMessage;
    } catch (parseError) {
        console.error('Error parsing transaction response:', parseError);
        this.showError('Payment verification failed. Please try again.');
        return;
    }
    
    // Clean up modal
    this.cleanup();
    
    // UNIFIED APPROACH: Process payment regardless of type
    this.processPayment(transactionData, donationId);
}

async processPayment(transactionData, donationId) {
    try {
        const response = await fetch('/donations/process-payment', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                customerCode: transactionData.data?.data?.customerCode,
                cardToken: transactionData.data?.data?.cardToken,
                donationId: donationId,
                amount: this.currentAmount
            })
        });
        
        const result = await response.json();
        
        if (result.success) {
            if (result.type === 'recurring') {
                // Redirect to recurring success page
                window.location.href = `/donate/success/recurring?subscriptionId=${result.subscriptionId}`;
            } else {
                // Redirect to one-time success page  
                window.location.href = `/donate/success?transactionId=${result.transactionId}`;
            }
        } else {
            throw new Error(result.error || 'Payment processing failed');
        }
    } catch (error) {
        console.error('Payment processing error:', error);
        this.showError('Payment processing failed. Please contact support.');
    }
}

showError(message) {
    // Display user-friendly error message
    alert(message); // Replace with better UI
}
```

### Database Schema Updates

#### 1. Migration File: `add_recurring_fields_to_donations.up.fizz`
```sql
add_column("donations", "subscription_id", "string", {"null": true})
add_column("donations", "customer_id", "string", {"null": true})
add_column("donations", "payment_plan_id", "string", {"null": true})
add_index("donations", ["subscription_id"], {})
add_index("donations", ["customer_id"], {})
```

#### 2. Updated Donation Model
```go
type Donation struct {
    // Existing fields...
    SubscriptionID  *string `json:"subscription_id,omitempty" db:"subscription_id"`
    CustomerID      *string `json:"customer_id,omitempty" db:"customer_id"`
    PaymentPlanID   *string `json:"payment_plan_id,omitempty" db:"payment_plan_id"`
}
```

## 🚀 IMPLEMENTATION PHASES

### Phase 1: Backend Foundation (3-4 hours)
- [ ] Add Helcim Recurring API client
- [ ] Create payment plan management functions
- [ ] Add subscription creation handlers
- [ ] Update database schema with migration

### Phase 2: Frontend Integration (2-3 hours)
- [ ] Modify donation.js for two-step flow
- [ ] Add subscription creation after payment verification
- [ ] Update success pages for recurring vs one-time
- [ ] Add subscription management links

### Phase 3: Testing & Validation (3-4 hours)
- [ ] Test one-time donations (ensure no regression)
- [ ] Test recurring donation setup flow
- [ ] Test subscription cancellation
- [ ] Test failed payment handling
- [ ] Verify webhook events for subscriptions

### Phase 4: User Experience (2-3 hours)
- [ ] Create subscription management page
- [ ] Add email templates for recurring donations
- [ ] Add cancellation/modification flows
- [ ] Documentation and user guides

## 🔒 SECURITY CONSIDERATIONS

1. **PCI Compliance Maintained**
   - HelcimPay.js still handles all sensitive payment data
   - No card details stored on our servers

2. **API Security**
   - All Helcim API calls from backend only
   - Proper token validation and error handling

3. **Subscription Management**
   - User authentication required for subscription changes
   - Audit trail for all subscription modifications

## 📊 SUCCESS METRICS

1. **Functional Metrics**
   - [ ] Monthly recurring donations create actual subscriptions
   - [ ] Automatic monthly billing occurs successfully
   - [ ] Subscription cancellation works properly
   - [ ] Webhook events processed correctly

2. **Business Metrics**
   - Increased monthly recurring donor percentage
   - Higher total donation value from recurring donors
   - Reduced manual donation processing overhead

## 🚨 ROLLBACK PLAN

If issues arise:
1. **Immediate**: Disable recurring option in UI (comment out radio button)
2. **Short-term**: Revert to one-time only donations until fixes complete
3. **Data Protection**: All existing donation data remains intact

## 📝 TESTING CHECKLIST

### One-Time Donations (Regression Testing)
- [ ] Amount selection works
- [ ] Payment processing succeeds
- [ ] Success page displays correctly
- [ ] Email receipts sent
- [ ] Database records created properly

### Recurring Donations (New Feature)
- [ ] Payment method verification succeeds
- [ ] Customer created in Helcim
- [ ] Subscription created successfully
- [ ] Next billing date calculated correctly
- [ ] Subscription appears in Helcim dashboard
- [ ] Monthly billing occurs automatically
- [ ] Failed payment handling works
- [ ] Subscription cancellation works

### Edge Cases
- [ ] Invalid payment methods
- [ ] API timeouts/failures
- [ ] Partial completion scenarios
- [ ] Duplicate subscriptions prevention

---

**Next Steps:** Begin Phase 1 implementation with payment plan setup and backend API integration.

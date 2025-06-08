# AVR Donation System Improvement - Project Tracking

## Current Status: Phase 1 - Security & API Improvements

### Task 1.1: Convert GET `/api/checkout_token` to POST endpoint
**Status**: 🔄 IN PROGRESS  
**Priority**: HIGH  
**Assigned**: Next AI Assistant Session

#### Current Implementation Analysis
- **Location**: `main.go` line ~303-494
- **Current Method**: GET with query parameters
- **Security Issue**: Sensitive donation data passed in URL
- **Frontend**: `templates/donate.html` line ~167-185

#### Required Changes

**Backend Changes (`main.go`)**:
```go
// CURRENT (line ~303):
r.GET("/api/checkout_token", func(c *gin.Context) {
    // Gets data from c.Query("amount"), c.Query("firstName"), etc.
    
// TARGET:
r.POST("/api/checkout_token", func(c *gin.Context) {
    var request struct {
        Amount     float64 `json:"amount" binding:"required"`
        FirstName  string  `json:"firstName"`
        LastName   string  `json:"lastName"`
        Email      string  `json:"email"`
        Purpose    string  `json:"purpose"`
        Referral   string  `json:"referral"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
        return
    }
```

**Frontend Changes (`templates/donate.html`)**:
```javascript
// CURRENT (line ~167):
const apiUrl = new URL('/api/checkout_token', window.location.origin);
apiUrl.searchParams.append('amount', donationAmount);
// ... more searchParams.append calls
const response = await fetch(apiUrl);

// TARGET:
const response = await fetch('/api/checkout_token', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        amount: parseFloat(donationAmount),
        firstName: firstName,
        lastName: lastName,
        email: email,
        purpose: purpose,
        referral: referralText
    })
});
```

#### Files to Modify
1. `main.go` - Lines ~303-494 (checkout_token endpoint)
2. `templates/donate.html` - Lines ~167-200 (initiateHelcimCheckout function)

#### Testing Requirements
- [ ] Test successful donation flow
- [ ] Test validation errors
- [ ] Test API error handling
- [ ] Verify no data leakage in logs
- [ ] Test frontend error display

#### Definition of Done
- [ ] API endpoint uses POST method
- [ ] Request validation implemented
- [ ] Frontend sends JSON payload
- [ ] All error cases handled
- [ ] Testing completed successfully
- [ ] No sensitive data in URL/logs

---

### Task 1.2: Add request validation and sanitization
**Status**: 📋 PLANNED  
**Priority**: HIGH  
**Depends On**: Task 1.1

#### Scope
- Input validation for all donation fields
- Rate limiting for donation endpoints
- Sanitization of user inputs
- Enhanced error messages

---

## Phase 2: Webhook Integration (NEXT)

### Task 2.1: Implement Helcim webhook handler
**Status**: 📋 PLANNED  
**Priority**: MEDIUM

#### Requirements
- Create `/api/webhooks/helcim` endpoint
- Verify webhook signatures
- Handle payment status updates
- Log all webhook events

#### Helcim Webhook Events to Handle
- `payment.success`
- `payment.failed`
- `payment.refunded`
- `payment.cancelled`

---

## Phase 3-5: Future Phases
See README.md for complete phase breakdown.

---

## Development Notes

### Environment Setup
- **Port**: 3001 (default)
- **Database**: None (currently using email notifications only)
- **Payment Processor**: Helcim Pay
- **Email**: Internal SMTP server on port 1025

### Current Limitations
1. No persistent storage of donations
2. No webhook verification
3. No automated receipts
4. No admin dashboard
5. Security improvements needed

### Quick Start Commands
```bash
# Start development server
go run main.go

# Test current donation flow
curl "http://localhost:3001/api/checkout_token?amount=25&firstName=Test&lastName=User&email=test@example.com&purpose=general&referral=test"
```

### Important Files
- `main.go`: Main application (888 lines)
- `templates/donate.html`: Donation form (314 lines)
- `.env`: Environment configuration
- `README.md`: Project overview and phase tracking

---

## For Future AI Assistants

### Getting Started Checklist
1. [ ] Read README.md for current phase status
2. [ ] Check this file for detailed task breakdown
3. [ ] Review current implementation in main.go line ~303
4. [ ] Understand Helcim integration in donate.html
5. [ ] Set up development environment (port 3001)
6. [ ] Run current code to understand flow

### After Completing Work - MANDATORY UPDATES
1. [ ] **Update task status** in this file (📋 PLANNED → 🔄 IN PROGRESS → ✅ COMPLETED)
2. [ ] **Check off completed items** in README.md phase checklists
3. [ ] **Update current priority** in README.md to point to next task
4. [ ] **Document any issues encountered** or changes in approach
5. [ ] **Update .github/copilot-instructions.md** if new patterns are discovered
6. [ ] **Add any new testing requirements** or edge cases found

### Next Session Priority
**START HERE**: Task 1.1 - Convert GET to POST for `/api/checkout_token`

This is a critical security improvement that must be completed before moving to Phase 2.

### Documentation Maintenance
**REMEMBER**: The documentation system only works if it's kept up to date. Always update progress tracking when completing tasks!

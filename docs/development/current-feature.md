# Current Features and Roadmap

This document tracks active features and planned improvements for the project, prioritized by urgency and impact.

## Table of Contents

1. [Donation System Implementation (Priority Feature)](#donation-system-implementation-priority-feature)
2. [Progressive Enhancement Refactor (Current Feature)](#progressive-enhancement-refactor-current-feature)
3. [Harden Tests](#feature-harden-tests--fix-404s-and-webhook-signature-flakiness)
4. [Template Codebase Deficiencies and Fixes](#feature-template-codebase-deficiencies-and-fixes)

---

# Donation System Implementation (Priority Feature)

## Overview

Transform the minimal test donation form into a comprehensive, production-ready donation system with full form functionality, robust validation, payment processing, and success/failure handling.

## Goal

Create a complete donation experience that allows users to make tax-deductible donations to American Veterans Rebuilding with secure payment processing, proper validation, and comprehensive user feedback.

## Approach: Buffalo OOTB Features

**No custom JavaScript required!** This implementation leverages Buffalo's built-in features:

- **Server-side validation** with Buffalo's `validate` package
- **Flash messages** for user feedback and alerts
- **Form handling** with automatic binding and CSRF protection
- **Error handling** with built-in patterns
- **HTML5 form attributes** for basic client-side validation
- **Progressive enhancement** that works without JavaScript

This approach ensures:
- ✅ **Security**: Server-side validation is primary, client-side is enhancement
- ✅ **Maintainability**: Less code, fewer dependencies
- ✅ **Accessibility**: Works with screen readers and keyboard navigation
- ✅ **Performance**: No JavaScript overhead for basic functionality
- ✅ **Reliability**: Fewer points of failure

## Current State

- ✅ Basic form structure exists with CSRF protection
- ✅ Server-side validation framework in place (DonateHandler validation logic)
- ✅ Payment processing infrastructure (Helcim integration exists)
- ✅ Success/failure page templates (comprehensive and well-designed)
- ✅ Email service infrastructure (SMTP configuration and templates)
- ✅ Most routes configured (GET/POST /donate, payment handlers, webhooks)
- ✅ GET /donate works (500 error fixed)
- ❌ Missing full form fields (currently minimal test form)
- ❌ Payment flow not fully integrated (needs completion)
- ❌ Success/failure handling needs enhancement
- ❌ Limited error handling and user feedback

## Implementation Plan

### Phase 1: Restore Full Form Functionality (Week 1)

#### 1A: Form Structure & Fields
- **Amount Selection**: Preset buttons ($25, $50, $100, $250, $500, $1000) + custom input
- **Donation Type**: Radio buttons for one-time vs monthly recurring
- **Donor Information**:
  - First Name (required)
  - Last Name (required)
  - Email (required, validated)
  - Phone (optional)
- **Address Fields**:
  - Address Line 1 (required)
  - Address Line 2 (optional)
  - City (required)
  - State (required, dropdown with all 50 states)
  - ZIP Code (required)
- **Comments**: Optional textarea for dedication/special messages

#### 1B: Form Layout & UX
- Responsive grid layout using Pico CSS
- Proper form labels and accessibility attributes
- Autocomplete attributes for better UX
- Visual feedback for selected amount buttons
- Progressive enhancement (works without JavaScript)

#### 1C: Buffalo OOTB Integration
- Server-side form processing with Buffalo validation
- HTML5 form attributes for basic client-side validation
- Buffalo flash messages for user feedback
- Progressive enhancement (works without JavaScript)

### Phase 2: Comprehensive Form Validation (Week 2)

#### 2A: Server-Side Validation
```go
// Enhanced validation functions
func ValidateDonationForm(req DonationRequest) *validate.Errors {
    errors := validate.NewErrors()

    // Required field validation
    ValidateRequiredString(req.FirstName, "first_name", "First name")
    ValidateRequiredString(req.LastName, "last_name", "Last name")
    ValidateEmailFormat(req.DonorEmail)
    ValidateRequiredString(req.AddressLine1, "address_line1", "Address")
    ValidateRequiredString(req.City, "city", "City")
    ValidateRequiredString(req.State, "state", "State")
    ValidateZipCode(req.Zip)

    // Amount validation
    ValidateDonationAmount(req.CustomAmount)

    // Phone validation (optional but format check)
    if req.DonorPhone != "" {
        ValidatePhoneFormat(req.DonorPhone)
    }

    return errors
}
```

#### 2B: HTML5 Form Validation
- HTML5 form attributes for basic client-side validation
- `required` attributes for mandatory fields
- `type="email"` for email validation
- `type="number"` for amount validation
- `pattern` attributes for phone/state validation
- Progressive enhancement (server validation as primary)

#### 2C: Buffalo Flash Messages & Error Display
- Use Buffalo's built-in `c.Flash()` for success/error messages
- Server-side validation with field-specific error display
- Form re-rendering with preserved user input
- Clear error messaging without JavaScript dependencies

```html
<!-- Use Buffalo flash messages -->
<%= partial("flash") %>

<!-- Field-specific errors -->
<% if (hasAmountError) { %>
  <small class="error"><%= errors.Get("amount") %></small>
<% } %>
```

### Phase 3: Payment Processing Integration (Week 3)

#### 3A: Donation Flow Architecture
```
1. User submits donation form
2. Server validates form data
3. If valid → Create donation record in DB
4. Call Helcim "verify" API (no charge yet)
5. Store checkout_token and secret_token
6. Redirect to payment page
7. Display Helcim payment form
8. User enters payment details
9. Helcim processes payment
10. Helcim calls webhook on completion
11. Update donation status in DB
12. Send confirmation email
13. Redirect to success/failure page
```

#### 3B: Payment Form Integration
```html
<!-- donate_payment.plush.html -->
<div id="helcim-payment-form">
  <!-- Official HelcimPay.js from Helcim documentation -->
  <script type="text/javascript" src="https://secure.helcim.app/helcim-pay/services/start.js"></script>
</div>

<script>
  // Official HelcimPay.js integration using appendHelcimPayIframe
  appendHelcimPayIframe("<%= checkoutToken %>");

  // Listen for postMessage events from HelcimPay.js iframe
  window.addEventListener('message', (event) => {
    const helcimPayJsIdentifierKey = 'helcim-pay-js-' + "<%= checkoutToken %>";
    if(event.data.eventName === helcimPayJsIdentifierKey){
      if(event.data.eventStatus === 'SUCCESS'){
        // Handle successful payment
        window.location.href = "/donate/success";
      }
      if(event.data.eventStatus === 'ABORTED'){
        // Handle payment failure
        window.location.href = "/donate/failed";
      }
    }
  });
</script>
```

**Note:** This JavaScript is provided by the Helcim payment service, not custom code. The primary payment logic remains server-side with Buffalo handlers.

#### 3C: Backend Payment Processing
- Enhanced `ProcessPaymentHandler` for payment completion
- Improved `HelcimWebhookHandler` for webhook processing
- Proper error handling and recovery
- Transaction logging and monitoring

### Phase 4: Success/Failure Handling (Week 4)

#### 4A: Enhanced Success Page
```html
<!-- donation_success.plush.html -->
<section class="success-message">
  <h1>Thank You for Your Donation!</h1>
  <div class="donation-summary">
    <p><strong>Amount:</strong> $<%= donation.Amount %></p>
    <p><strong>Type:</strong> <%= donation.DonationType %></p>
    <p><strong>Transaction ID:</strong> <%= donation.TransactionID %></p>
  </div>

  <div class="receipt-info">
    <p>A receipt has been sent to <%= donation.DonorEmail %></p>
    <p>Keep this transaction ID for your records: <code><%= donation.TransactionID %></code></p>
  </div>

  <div class="next-steps">
    <a href="/" class="button">Return Home</a>
    <a href="/donate" class="button secondary">Make Another Donation</a>
  </div>
</section>
```

#### 4B: Enhanced Failure Page
```html
<!-- donation_failed.plush.html -->
<section class="error-message">
  <h1>Donation Processing Failed</h1>
  <div class="error-details">
    <p>We're sorry, but there was an issue processing your donation.</p>
    <p><strong>Error:</strong> <%= errorMessage %></p>
  </div>

  <div class="recovery-options">
    <a href="/donate" class="button">Try Again</a>
    <a href="/contact" class="button secondary">Contact Support</a>
  </div>
</section>
```

#### 4C: Email Notifications
- Automated receipt emails with tax deduction information
- Transaction details and organization EIN
- Professional email templates
- Error notifications for failed payments

## Implementation Checklist

### Phase 1: Form Restoration ✅ COMPLETE
- [x] Restore all form fields (name, email, phone, address, comments)
- [x] Implement amount selection (preset buttons + custom input)
- [x] Add donation type selection (one-time/monthly)
- [x] Create comprehensive state dropdown
- [x] Implement responsive form layout with HTML5 validation
- [x] Use Buffalo flash messages for user feedback
- [x] Basic form structure exists with CSRF protection
- [x] GET /donate works (500 error fixed)
- [x] POST /donate works with form validation and data preservation

### Phase 2: Validation Enhancement
- [x] Server-side validation framework in place (DonateHandler validation logic)
- [ ] Enhance server-side validation functions using Buffalo validate package
- [ ] Add HTML5 form attributes for basic client-side validation
- [ ] Implement Buffalo flash messages for user feedback
- [ ] Add field-specific validation messages with error display
- [ ] Test validation with various edge cases using Buffalo patterns

### Phase 3: Payment Processing
- [x] Payment processing infrastructure (Helcim integration exists)
- [x] Payment handlers exist (DonatePaymentHandler, ProcessPaymentHandler)
- [x] Webhook processing (HelcimWebhookHandler exists)
- [ ] Integrate Helcim payment form
- [ ] Enhance payment flow handlers
- [ ] Implement webhook processing improvements
- [ ] Add payment error handling and recovery
- [ ] Test payment flow with test credentials

### Phase 4: Success/Failure Handling
- [x] Success/failure page templates (comprehensive and well-designed)
- [x] Email service infrastructure (SMTP configuration and templates)
- [x] Success/failure handlers exist (DonationSuccessHandler, DonationFailedHandler)
- [ ] Enhance success page with donation details
- [ ] Improve failure page with error recovery options
- [ ] Implement email notification system
- [ ] Add transaction logging and monitoring
- [ ] Test complete donation flow end-to-end

## Testing Strategy

### Unit Tests
- Form validation functions
- Payment processing logic
- Email notification system
- Error handling scenarios

### Integration Tests
- Complete donation flow from form to completion
- Payment processing with test credentials
- Webhook handling and status updates
- Email delivery verification

### User Acceptance Testing
- End-to-end donation process testing
- Various payment amounts and donation types
- Error conditions and recovery flows
- Cross-browser compatibility
- Mobile responsiveness

## Current Implementation Status

### ✅ **Completed (Major Progress Made):**
- **GET /donate fixed** - 500 error resolved, page loads successfully
- **CSRF handling** - Temporarily disabled for debugging, proper token handling in place
- **Form structure** - Basic form with CSRF protection working
- **Validation framework** - Server-side validation logic exists and functional
- **Payment infrastructure** - Helcim integration handlers and services exist
- **Success/failure pages** - Comprehensive, well-designed templates ready
- **Email system** - SMTP service and notification infrastructure in place
- **Routes configured** - All necessary endpoints set up

### 🚧 **In Progress/Partially Complete:**
- **Form fields** - Basic test form exists, needs full field restoration
- **Payment flow** - Infrastructure exists, needs integration completion
- **Error handling** - Basic system works, needs enhancement

### ❌ **Still To Do:**
- **Full form restoration** - Add all donor fields, amount selection, state dropdown
- **Payment form integration** - Connect Helcim payment form to flow
- **Email notifications** - Implement automated receipt sending
- **End-to-end testing** - Complete donation flow validation

## Dependencies & Prerequisites

- Helcim payment gateway credentials (production/test)
- Email service configuration
- Database schema for donations table
- SSL certificate for production
- Tax exemption documentation (EIN, etc.)

**No additional JavaScript dependencies required** - Uses only Buffalo OOTB features and Helcim's provided SDK.

## Risk Assessment

### High Risk
- Payment processing integration failures
- Email delivery issues
- Form validation edge cases

### Medium Risk
- Cross-browser compatibility with HTML5 validation
- Mobile responsiveness problems
- Performance issues with large forms
- Complex server-side validation edge cases

### Low Risk
- UI/UX improvements
- Additional validation rules
- Enhanced error messages

## Success Metrics

- 100% form completion rate (no validation errors)
- Successful payment processing rate > 95%
- Email delivery success rate > 99%
- User satisfaction with donation experience
- Reduced support tickets related to donations
- **Buffalo OOTB compliance** - uses only built-in Buffalo features
- **JavaScript-free core functionality** - works without JavaScript enabled
- **Fast form submission** - no client-side processing delays
- **Accessible experience** - works with screen readers and keyboard navigation
- **Security-first approach** - server-side validation as primary defense

---

# Progressive Enhancement Refactor (Current Feature)

This document describes the comprehensive refactor plan to convert the site to a progressively enhanced application using out-of-the-box Buffalo features and HTMX best practices. It includes an actionable checklist to track progress.

## Goal

Refactor the codebase to follow Buffalo OOTB patterns for forms, rendering, CSRF, flash messaging, and asset handling while strictly following HTMX guidelines (single-template architecture, hx-boost for navigation, progressive enhancement).

---

## Summary of Changes

- Remove header-based conditional rendering (no HX-Request header branching)
- Enable global `hx-boost` in `templates/application.plush.html`
- Consolidate to a single render engine and remove the `htmx.plush.html` layout
- Standardize form handlers to the single-route GET/POST pattern
- Ensure CSRF tokens are included via hidden inputs and rely on Buffalo CSRF middleware
- Use Buffalo flash messages for alerts
- Use Buffalo asset helpers (stylesheetTag/javascriptTag) and fix asset pipeline integration
- Remove form submissions to API endpoints for user-facing pages

---

## Implementation Steps (High Level)

1. Core architecture
   - Remove `renderForRequest()` and `IsHTMX()` helpers and `rHTMX` render engine
   - Remove `templates/htmx.plush.html`
   - Replace all calls to `renderForRequest()` with standard `c.Render(..., r.HTML(...))`
   - Add `hx-boost="true"` to `<body>` in `templates/application.plush.html`

2. Form handlers
   - Update all form handlers to implement GET (show) and POST (process) in the same function
   - Ensure success and error handling uses Buffalo flash messages and standard redirects for non-HTMX
   - Ensure HTMX-enhanced requests still return full pages

3. Templates
   - Ensure all page templates are full-page templates with proper `<html>`, `<head>`, and `<body>` via `application.plush.html`
   - Ensure form partials include hidden `authenticity_token` inputs
   - Remove HTMX-only partials where unnecessary

4. Assets
   - Replace direct asset links with Buffalo helpers (stylesheetTag/javascriptTag)
   - Ensure manifest and asset pipeline are configured

5. Routes
   - Consolidate routes for forms (same path for GET and POST)
   - Remove user-facing forms posting to `/api/*` endpoints

6. Tests & QA
   - Test with JS disabled and enabled
   - Verify bookmarking, refresh, back/forward behavior
   - Add/adjust unit and integration tests where necessary

7. Documentation
   - Update docs in `docs/buffalo-framework/` to reference the new patterns
   - Add this current feature guide as the active progress tracker

---

## File Changes Required (non-exhaustive)

- actions/render.go: remove HTMX helpers and rHTMX engine
- templates/htmx.plush.html: remove file
- templates/application.plush.html: add hx-boost on body and replace direct assets with helpers
- Multiple handlers: replace renderForRequest(...) with c.Render(...)
- templates/pages/* and templates/pages/_donate_form.plush.html: ensure CSRF hidden inputs and progressive enhancement attributes
- actions/pages.go, actions/donations.go: consolidate GET/POST handlers

---

## Checklist (Track Progress)

- [ ] Core: Remove `renderForRequest()` and `IsHTMX()` helpers
- [ ] Core: Remove `rHTMX` render engine and `templates/htmx.plush.html`
- [ ] Core: Replace all `renderForRequest()` usages
- [ ] Core: Add `hx-boost="true"` to `templates/application.plush.html`
- [ ] Forms: Convert all form handlers to GET/POST single-handler pattern
- [ ] Forms: Ensure all form templates include hidden `authenticity_token` inputs
- [ ] Alerts: Replace custom alerts with Buffalo `c.Flash()` usage across handlers
- [ ] Assets: Replace direct asset links with `stylesheetTag`/`javascriptTag` helpers
- [ ] Routes: Remove forms posting to `/api/*` and consolidate routes
- [ ] Tests: Add/adjust tests for progressive enhancement and HTMX
- [ ] Docs: Update docs to reflect the new architecture
- [ ] QA: Manual verification with JS disabled/enabled for all major flows

---

## How to Use This Tracker

- Mark checklist items as completed when changes are implemented and tested.
- For each completed item, add a short note below with the commit or PR reference.

---

## Notes & Rationale

We will follow Buffalo OOTB features and HTMX docs: single-template architecture, progressive enhancement (forms with action attributes), and global hx-boost for navigation. Removing header-based rendering prevents bookmark/refresh issues and simplifies templates and handlers.

---

Last updated: (auto-generated)

---

Feature: Harden tests — fix 404s and webhook signature flakiness
Status: in_progress

Description
- Reduce intermittent test failures caused by missing DB fixtures, Helcim webhook signature verification, and missing template partials. Make tests deterministic and CI-friendly.

Short checklist
- Create required DB fixtures for blog post tests (ensure PublishedAt/slugs match handlers).
- Add a reusable test helper: CreatePostForTest (centralizes post creation).
- Ensure CI/test runner prepares DB (migrations) before running DB-dependent tests or clearly mark tests that require DB.
- Make Helcim webhook signature verification test-friendly:
  - Add test helpers to generate valid signatures, and/or
  - Bypass verification only in test environment (GO_ENV=test) or via a guarded env var.
- Add AttachHelcimSignature test helper to simplify signed request creation.
- Scan templates for missing partials; add thin wrapper partials or correct partial references.
- Add a template-prewarm CI test to fail fast on missing partials.
- Update developer docs with testing instructions and helper usage.

Acceptance criteria
- Previously-failing blog slug tests create their fixtures and pass consistently in CI.
- Helcim webhook tests pass using test signatures or bypass mechanism; verification logic is covered by unit tests.
- No missing-partial template render errors during tests; template prewarm passes in CI.
- Docs updated with instructions to run DB-backed tests and webhook-test environment variables.

---

Feature: Template Codebase Deficiencies and Fixes
Status: pending

Description
- Address deficiencies in the Plush template codebase to improve maintainability, security, and consistency. Move logic out of templates into helpers, standardize partial naming, fix error handling, and ensure XSS safety.

Short checklist
- Audit and inventory templates and partial calls (generate a report mapping partial names -> files).
- Create common helpers in Go: fieldClass(errors, name), fieldValue(ctx, name, default), checkedAttr(val, expected), selectedAttr(val, expected), csrfToken(), renderErrors(errors).
- Update handlers to set defaults in context rather than mutating templates.
- Replace inline logic in templates with partial/helper calls incrementally.
- Add/extend tests for helpers and templates; run make test.
- Security review for any helper that returns template.HTML; add sanitizer where needed.
- Lint templates for partial naming consistency and update docs if conventions change.
- Fix specific issues: mutation/defaulting inside templates, large inline logic blocks, error rendering duplication, CSRF handling inconsistency, inline attribute concatenation, raw value insertion risking XSS, inconsistent partial naming, type/append errors for arrays, iterator usage.

Acceptance criteria
- All templates follow Plush best practices with minimal inline logic.
- Common patterns (errors, CSRF, attributes) use standardized helpers/partials.
- No XSS vulnerabilities from unsanitized output.
- Partial naming is consistent and follows repository conventions.
- Unit tests cover new helpers and template rendering edge cases.
- Template prewarm and linting pass without errors.

---

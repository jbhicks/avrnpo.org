# HelcimPay.js Integration Validation Checklist

This checklist ensures the HelcimPay.js integration is working correctly after standardization to the official URL and API patterns.

## Frontend Validation

### Script Loading
- [ ] **Canonical URL loads successfully**: `https://secure.helcim.app/helcim-pay/services/start.js` returns 200 in production/dev
- [ ] **appendHelcimPayIframe function available**: `window.appendHelcimPayIframe` exists after script loads
- [ ] **No console errors**: Script loads without JavaScript errors in browser console
- [ ] **CSP compatibility**: Script loads without Content Security Policy violations

### Modal Rendering
- [ ] **Modal displays correctly**: `appendHelcimPayIframe(checkoutToken)` shows Helcim payment modal
- [ ] **Modal styling intact**: Official Helcim styling renders properly
- [ ] **Responsive design**: Modal works on mobile and desktop devices
- [ ] **Exit functionality**: Modal can be closed by user (allowExit=true by default)

### Event Handling
- [ ] **PostMessage events received**: Browser receives events from Helcim iframe
- [ ] **SUCCESS event processed**: Successful payments trigger `processPaymentSuccess()`
- [ ] **ABORTED event processed**: Failed payments trigger `processPaymentError()`
- [ ] **HIDE event processed**: Modal closure redirects appropriately

## Backend Validation

### API Integration
- [ ] **Initialize endpoint succeeds**: `POST /v2/helcim-pay/initialize` returns checkoutToken and secretToken
- [ ] **Tokens stored correctly**: checkoutToken and secretToken saved to donation record
- [ ] **Token expiration handled**: 60-minute token expiry doesn't break user flow
- [ ] **Error handling robust**: API failures don't crash application

### Payment Processing
- [ ] **Transaction creation succeeds**: Payment API processes card tokens correctly
- [ ] **Webhook events processed**: Real-time status updates work
- [ ] **Database updates correct**: Donation status changes from pending → completed
- [ ] **Email receipts sent**: Successful payments trigger receipt emails

## Documentation Validation

### Code Comments
- [ ] **API endpoint references**: Code comments link to correct Helcim API docs
- [ ] **Token usage documented**: Comments explain checkoutToken vs secretToken
- [ ] **Function purposes clear**: All Helcim-related functions have descriptive comments

### External Documentation
- [ ] **No outdated URLs**: All references to old Helcim script URLs removed
- [ ] **Canonical URL documented**: Official script URL documented in all relevant places
- [ ] **Integration examples updated**: Code examples use correct `appendHelcimPayIframe()` pattern
- [ ] **Testing instructions current**: Test docs reference correct URLs and patterns

## Testing Validation

### Unit Tests
- [ ] **Template rendering tests**: donate_payment.plush.html includes canonical script URL
- [ ] **Mock token tests**: Test environment generates valid mock tokens
- [ ] **Event handling tests**: PostMessage event processing works correctly

### Integration Tests
- [ ] **End-to-end flow**: Complete donation process works from form to completion
- [ ] **Error scenarios**: Payment failures handled gracefully
- [ ] **Webhook processing**: Real-time status updates work in test environment

## Production Readiness

### Security
- [ ] **No token exposure**: Sensitive tokens never logged or exposed to client
- [ ] **HTTPS required**: All Helcim endpoints use secure connections
- [ ] **CSP headers**: Content Security Policy allows Helcim domains

### Performance
- [ ] **Script load time**: Official CDN provides fast loading
- [ ] **Modal render time**: Payment modal appears quickly
- [ ] **Error recovery**: Failed loads don't break donation flow

### Monitoring
- [ ] **Error logging**: Helcim API failures logged appropriately
- [ ] **Success metrics**: Payment success rates tracked
- [ ] **User experience**: Modal load times and completion rates monitored

## Migration Validation

### Backward Compatibility
- [ ] **No breaking changes**: Existing donation flow still works
- [ ] **Session handling intact**: Token storage in session works correctly
- [ ] **Database schema unchanged**: No migration required for existing data

### Cleanup Verification
- [ ] **Old URLs removed**: No references to deprecated Helcim script URLs
- [ ] **Fallback logic removed**: No redundant URL fallback mechanisms
- [ ] **Code consistency**: All Helcim integration uses same pattern

## Emergency Procedures

### Rollback Plan
- [ ] **Quick revert available**: Changes can be rolled back if issues arise
- [ ] **Alternative payment method**: Manual payment processing available if needed
- [ ] **User communication**: Clear messaging for any temporary issues

### Support Resources
- [ ] **Helcim support contact**: Technical support information available
- [ ] **Internal documentation**: Team knows how to troubleshoot Helcim issues
- [ ] **Test environment**: Isolated testing environment for future changes

---

## Validation Commands

### Frontend Testing
```bash
# Check script loading
curl -I https://secure.helcim.app/helcim-pay/services/start.js

# Test modal rendering (manual)
# 1. Navigate to /donate
# 2. Complete form
# 3. Verify modal appears
# 4. Check browser console for errors
```

### Backend Testing
```bash
# Test API connectivity
curl -X POST https://api.helcim.com/v2/helcim-pay/initialize \
  -H "api-token: YOUR_TEST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"paymentType": "purchase", "amount": 10.00, "currency": "USD"}'

# Check application logs
tail -f logs/application.log | grep -i helcim
```

### Documentation Audit
```bash
# Find any remaining old URLs
grep -r "gateway.helcim.com" docs/
grep -r "helcim-pay.min.js" docs/
grep -r "secure.helcim.app/helcim-pay/services/start.js" docs/
```

---

*Last updated: After HelcimPay.js standardization*
# Donation System Testing Guide

This guide provides comprehensive instructions for testing the AVR NPO donation system in development mode.

## Overview

The donation system uses real Helcim payment processing but operates in test mode for development, ensuring:
- ✅ **Safe Testing**: No real money is charged
- ✅ **Real Integration**: Uses actual Helcim API and workflows  
- ✅ **Complete Flow**: Tests entire donation process end-to-end
- ✅ **Email Receipts**: Optional real email sending for testing

## Quick Start Testing

### 1. Basic Test (5 minutes)
1. Navigate to: `http://localhost:3000/donate`
2. Enter any amount and donor information
3. Click "Donate Now"
4. Use **official Helcim test card**: **4124 9399 9999 9990**, CVV: **100**, Expiry: **01/28**
5. Verify success page appears
6. Check admin dashboard: `http://localhost:3000/admin/donations`

### 2. Email Testing
**To receive test emails**:
1. Configure email settings in `.env`:
   ```bash
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your_email@gmail.com
   SMTP_PASSWORD=your_app_password
   FROM_EMAIL=donations@avrnpo.org
   FROM_NAME=American Veterans Rebuilding
   ```
2. Use your real email address as the donor email
3. Complete a test donation
4. Check your inbox for receipt email

**Without email configuration**:
- System works normally but no emails are sent
- Buffalo logs will show: `email service not configured`

## Official Helcim Test Card Numbers

**⚠️ IMPORTANT:** These are official Helcim test cards - only work with test accounts!

### Successful Payments
- **Visa**: `4124939999999990` or `4000000000000028`
- **Mastercard**: `5413330089099130` or `5413330089020011`
- **American Express**: `374245001751006`
- **Discover**: `6011973700000005`
- **CVV**: 100 (or 1000 for Amex)
- **Expiry**: 01/28 (all test cards)

### Test Account Required
- Contact `tier2support@helcim.com` to request a Helcim test account
- Test cards will be **declined** on production accounts
- Real cards should **never** be used for testing

### Declined Payments (for testing error handling)
- Use any test card with incorrect CVV, expired date, or insufficient amount
- Test cards may have built-in decline scenarios in test environment

## Testing Scenarios

### Scenario 1: Successful One-Time Donation
1. **Amount**: $25 (preset button)
2. **Donor Info**: John Doe, john@example.com
3. **Type**: One-time
4. **Card**: `4124939999999990`, CVV: 100, Exp: 01/28
5. **Expected**: Success page, email receipt, "completed" status in admin

### Scenario 2: Failed Payment
1. **Amount**: $50
2. **Donor Info**: Jane Smith, jane@example.com  
3. **Type**: One-time
4. **Card**: Use test card with invalid CVV or expired date
5. **Expected**: Error message, stays on form, "failed" status in admin

### Scenario 3: Custom Amount
1. **Amount**: Custom $73.50
2. **Donor Info**: Complete address information
3. **Type**: Recurring (monthly)
4. **Card**: 5555 5555 5555 4444
5. **Expected**: Success with recurring notation

### Scenario 4: Webhook Processing
1. Complete any successful payment
2. Monitor Buffalo logs for webhook events:
   ```
   INFO: Received Helcim webhook: type=payment_success
   INFO: Payment completed for donation ID: abc-123
   ```
3. Verify donation status changes from "pending" to "completed"

## Admin Dashboard Testing

**URL**: `http://localhost:3000/admin/donations`
**Requirement**: Admin user account

### Features to Test:
- **Donation List**: Paginated list of all donations
- **Status Filter**: Filter by pending/completed/failed
- **Search**: Search by donor name or email
- **Statistics**: Real-time donation totals and counts
- **Individual View**: Click donation ID for detailed view

### Creating Admin User (if needed):
```sql
UPDATE users SET role = 'admin' WHERE email = 'your_email@example.com';
```

## Database Verification

### Direct Database Queries
```sql
-- View recent donations
SELECT id, amount, donor_name, donor_email, status, donation_type, created_at 
FROM donations 
ORDER BY created_at DESC 
LIMIT 10;

-- Check donation statistics
SELECT 
    status,
    COUNT(*) as count,
    SUM(amount) as total_amount,
    AVG(amount) as avg_amount
FROM donations 
GROUP BY status;

-- View donations from last 24 hours
SELECT * FROM donations 
WHERE created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;
```

## API Testing

### Test Donation Initialization
```bash
curl -X POST http://localhost:3000/api/donations/initialize \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "25.00",
    "donation_type": "one-time",
    "donor_name": "Test Donor",
    "donor_email": "test@example.com",
    "donor_phone": "555-123-4567"
  }'
```

**Expected Response**:
```json
{
  "checkoutToken": "chkt_...",
  "donationId": "uuid-here"
}
```

### Test Webhook Endpoint
```bash
curl -X POST http://localhost:3000/api/webhooks/helcim \
  -H "Content-Type: application/json" \
  -H "X-Helcim-Signature: sha256=test" \
  -d '{
    "id": "event_123",
    "type": "payment_success",
    "data": {
      "transactionId": "txn_test_123",
      "amount": 25.00,
      "status": "completed"
    }
  }'
```

## Troubleshooting

### Common Issues

#### "Payment modal doesn't open"
- **Check**: Browser console for JavaScript errors
- **Check**: HelcimPay.js is loaded (`/js/helcim-pay.min.js`)
- **Solution**: Ensure Buffalo is serving static files

#### "Invalid signature" webhook errors  
- **Check**: `HELCIM_WEBHOOK_VERIFIER_TOKEN` in `.env`
- **Solution**: Set token or ensure `GO_ENV=development` (bypasses signature check)

#### "No emails received"
- **Check**: SMTP configuration in `.env`
- **Check**: Spam folder
- **Check**: Buffalo logs for email errors
- **Solution**: Configure valid SMTP credentials

#### "Donation not found" in webhook processing
- **Check**: Donation was created before webhook
- **Check**: Transaction ID matches between systems
- **Solution**: Verify donation initialization worked

#### "404 on admin pages"
- **Check**: User has admin role
- **Check**: Logged in as admin user
- **Solution**: Update user role in database

### Log Analysis

**Successful donation logs should show**:
```
INFO: Starting donation initialization
INFO: Created donation record: uuid-abc-123
INFO: Helcim checkout token created successfully
INFO: Received Helcim webhook: type=payment_success
INFO: Payment completed for donation ID: abc-123, amount: $25.00
```

**Failed donation logs might show**:
```
ERROR: Helcim API error: Card declined
INFO: Received Helcim webhook: type=payment_declined
INFO: Payment declined for donation ID: abc-123
```

## Production Readiness Checklist

Before deploying to production, verify:
- [ ] All test scenarios pass
- [ ] Email receipts working correctly
- [ ] Admin dashboard accessible and functional
- [ ] Webhook signature validation working
- [ ] Database properly configured
- [ ] HTTPS enabled (required for production)
- [ ] Real Helcim merchant account configured
- [ ] `HELCIM_TEST_MODE=false` in production

## Support

For testing issues:
1. Check Buffalo logs: `tail -f buffalo.log`
2. Verify environment variables in `.env`
3. Test individual components (API, webhooks, admin) separately
4. Consult Helcim documentation for payment processing issues

The donation system testing is designed to be comprehensive yet simple - most functionality can be verified with a single test donation using the provided test cards.

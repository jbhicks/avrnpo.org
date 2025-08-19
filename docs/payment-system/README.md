# Payment System Overview

The AVR NPO donation system uses Helcim for secure payment processing, supporting both one-time and recurring monthly donations.

## 📋 Quick Navigation

- [Helcim Integration](./helcim-integration.md) - Complete integration guide
- [Donation Flow](./donation-flow.md) - User experience and form handling  
- [Recurring Payments](./recurring-payments.md) - Subscription management
- [Webhooks](./webhooks.md) - Event processing and notifications
- [Testing](./testing.md) - Payment testing procedures and test cards

## 🎯 System Architecture

### Frontend Payment Collection
- **HelcimPay.js** - Official Helcim library for PCI-compliant card collection
- **Progressive Enhancement** - Works with and without JavaScript
- **Unified UX** - Same form for one-time and recurring donations

### Backend Processing
- **Buffalo/Go Handlers** - Payment initialization and completion
- **Helcim APIs** - Payment and Recurring APIs for transaction processing
- **Database Tracking** - Full donation history and subscription management
- **Email Receipts** - Automatic confirmation emails

### Security & Compliance
- **PCI Compliance** - No card data stored locally, all handled by Helcim
- **Environment Variables** - API credentials never in source code
- **HTTPS Only** - All payment communications encrypted
- **Session-Based Auth** - User account linking for subscription management

## ✅ Current Implementation Status

### ✅ Completed Features
- **One-time donations** - Full implementation with Helcim Payment API
- **Recurring donations** - Monthly subscriptions via Helcim Recurring API
- **User account linking** - Donations tied to user accounts when logged in
- **Subscription management** - View, cancel, and update subscriptions
- **Receipt system** - Email confirmations for all donations
- **Database tracking** - Complete audit trail of all transactions
- **Webhook processing** - Real-time payment status updates

### 🎯 Phase 2 Complete
The donation system has reached **Phase 2** completion with full recurring payment support:
- Recurring payment backend implementation ✅
- Subscription management UI ✅  
- User account integration ✅
- Payment flow unification ✅
- Comprehensive testing ✅

## 🔄 Payment Flow Summary

### One-Time Donations
1. User fills donation form
2. HelcimPay.js collects payment data securely
3. Backend processes via Helcim Payment API
4. Success confirmation and email receipt
5. Database records transaction

### Recurring Donations  
1. User selects "Monthly recurring" option
2. HelcimPay.js collects payment data securely
3. Backend creates subscription via Helcim Recurring API
4. Success confirmation with subscription details
5. Database records initial donation and subscription
6. User can manage subscription in account area

### Subscription Management
1. User logs in to account
2. Views active subscriptions with live status
3. Can cancel or modify subscriptions
4. Receives confirmation of changes

## 📊 Key Metrics Tracked

- **Donation amounts** - One-time and recurring totals
- **Conversion rates** - Form completion and payment success
- **Subscription lifecycle** - Active, cancelled, failed payments
- **User engagement** - Account creation and login patterns
- **Payment methods** - Card type and geographic distribution

## 🔗 Integration Points

### External Services
- **Helcim Payment Processor** - All financial transactions
- **Email Service** - Receipt delivery (SMTP configuration required)
- **Database** - PostgreSQL for all donation and user data

### Internal Systems  
- **User Authentication** - Buffalo session management
- **Admin Interface** - Donation and subscription oversight
- **Reporting** - Financial and donor analytics
- **Blog System** - Donation campaign integration

## 📝 Documentation Standards

All payment system documentation follows these principles:
- **Security first** - Never expose real credentials in examples
- **Code examples** - Working snippets with placeholder values
- **Error handling** - Comprehensive failure scenarios
- **Testing guidance** - How to verify implementations
- **Troubleshooting** - Common issues and solutions

For detailed implementation guidance, see the specific topic guides linked above.

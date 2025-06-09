# AVR Helcim Documentation Index

## 🚨 [SECURITY GUIDELINES - READ FIRST](./SECURITY-GUIDELINES.md)
**CRITICAL: Security rules for handling sensitive data - MUST READ before any development**

This directory contains comprehensive documentation for integrating with the Helcim payment processing API in the AVR NPO donation system.

## Documentation Files

### 📖 [Helcim API Reference](./helcim-api-reference.md)
**Complete reference guide for Helcim API integration**
- Authentication and API tokens
- Endpoint documentation
- Request/response formats
- Security best practices
- Environment configuration
- Testing procedures

### 🔗 [Helcim Webhooks Implementation Guide](./helcim-webhooks-guide.md)
**Step-by-step guide for Phase 2 webhook implementation**
- Webhook configuration in Helcim dashboard
- Go implementation patterns
- Signature verification
- Event processing logic
- Security considerations
- Testing and monitoring

### ⚠️ [Helcim Error Handling Reference](./helcim-error-handling.md)
**Comprehensive error handling patterns and recovery procedures**
- Common HTTP status codes
- Error response formats
- Go error handling patterns
- Retry logic and backoff strategies
- Frontend error display
- Monitoring and alerting

## Quick Reference

### Current Implementation (Phase 1)
- ✅ HelcimPay.js integration for donation processing
- ✅ POST endpoint for checkout token generation
- ✅ Input validation and sanitization
- ✅ Rate limiting protection

### Next Phase (Phase 2)
- 🔄 Webhook integration for real-time payment notifications
- 🔄 Signature verification for webhook security
- 🔄 Event processing for different payment statuses

### Future Phases
- 📋 Database integration for donation storage (Phase 3)
- 📋 Receipt and email system (Phase 4)
- 📋 Admin dashboard for donation management (Phase 5)

## Integration Checklist

### Environment Setup
- [ ] `HELCIM_PRIVATE_API_KEY` configured
- [ ] API token permissions verified in Helcim dashboard
- [ ] Connection test successful

### Current Features
- [ ] Donation form validation working
- [ ] HelcimPay.js checkout integration functional
- [ ] Rate limiting protecting endpoints
- [ ] Error handling providing user feedback

### Phase 2 Requirements
- [ ] Webhook URL configured in Helcim
- [ ] `HELCIM_WEBHOOK_VERIFIER_TOKEN` added to environment
- [ ] Webhook endpoint implemented with signature verification
- [ ] Event processing logic for payment statuses
- [ ] Webhook testing completed

## File Locations in Project

```
avrnpo.org/
├── docs/                           # This documentation directory
│   ├── README.md                   # This index file
│   ├── helcim-api-reference.md     # Complete API reference
│   ├── helcim-webhooks-guide.md    # Webhook implementation guide
│   └── helcim-error-handling.md    # Error handling patterns
├── main.go                         # Main application with Helcim integration
├── templates/donate.html           # Donation form with HelcimPay.js
├── .env                           # Environment variables (not in git)
└── README.md                      # Project overview and phase tracking
```

## Code Examples Location

All code examples in the documentation are designed to integrate with the current `main.go` structure:

- **API Reference**: Contains working Go code snippets for Helcim API calls
- **Webhooks Guide**: Provides complete webhook implementation for Phase 2
- **Error Handling**: Shows patterns for robust error handling throughout the system

## Development Workflow

1. **Read the relevant documentation** before implementing features
2. **Follow the Go patterns** shown in the documentation
3. **Test thoroughly** using the provided testing procedures
4. **Update documentation** if new patterns or issues are discovered
5. **Update project tracking** in README.md and PROJECT_TRACKING.md

## Security Notes

🔒 **Important**: All Helcim API calls must be made from the backend server, never from frontend JavaScript, to maintain PCI compliance and API security.

🔐 **Critical**: Always verify webhook signatures and validate timestamps to prevent security vulnerabilities.

💳 **Required**: Never handle raw credit card data - always use HelcimPay.js for PCI-compliant tokenization.

## Support and Resources

- **Helcim Documentation**: https://devdocs.helcim.com/
- **Helcim API Reference**: https://devdocs.helcim.com/reference
- **Helcim Support**: https://devdocs.helcim.com/docs/get-help
- **Project Issues**: Document in PROJECT_TRACKING.md

---

*This documentation is maintained as part of the AVR NPO donation system and should be updated as new features are implemented.*

# AVR NPO Website

American Veterans Rebuilding (AVR) official website - A Go-based web application with donation processing via Helcim API.

## Tech Stack
- **Backend**: Go with Gin framework
- **Frontend**: HTML templates with HTMX for interactivity
- **Styling**: Tailwind CSS + DaisyUI components
- **Payment Processing**: Helcim Pay API
- **Environment**: WSL Arch Linux

## Current Status
The website is functional with basic donation processing, but requires improvements for production-ready nonprofit operations.

## 🚧 DONATION SYSTEM IMPROVEMENT PLAN

### Phase 1: Security & API Improvements ✅ PLANNED
- [x] **Task 1.1**: Convert GET `/api/checkout_token` to POST endpoint
  - ✅ COMPLETED: Uses POST with JSON body for sensitive data
  - ✅ COMPLETED: Added proper request validation
  - Files modified: `main.go`, `templates/donate.html`
  
- [ ] **Task 1.2**: Add request validation and sanitization
  - Implement proper input validation
  - Add rate limiting for donation endpoints
  - Sanitize all user inputs

### Phase 2: Webhook Integration ⏳ NEXT UP
- [ ] **Task 2.1**: Implement Helcim webhook handler
  - Create `/api/webhooks/helcim` endpoint
  - Verify webhook signatures
  - Handle payment status updates (success, failure, refund)
  
- [ ] **Task 2.2**: Add webhook security
  - Implement HMAC signature verification
  - Add webhook authentication
  - Log all webhook events for debugging

### Phase 3: Database Integration 📋 PLANNED
- [ ] **Task 3.1**: Choose and setup database
  - Options: SQLite (simple) or PostgreSQL (production)
  - Create database schema for donations
  - Add database connection management
  
- [ ] **Task 3.2**: Create donation models and repositories
  - Donation struct with all required fields
  - CRUD operations for donations
  - Database migration system

### Phase 4: Receipt & Email System 📧 PLANNED
- [ ] **Task 4.1**: Integrate with existing email system
  - Use current SMTP setup for receipts
  - Create receipt email templates
  - Send automated thank you emails
  
- [ ] **Task 4.2**: Tax receipt generation
  - Generate PDF receipts for tax purposes
  - Include 501(c)(3) information
  - Store receipt records

### Phase 5: Admin Dashboard 📊 PLANNED
- [ ] **Task 5.1**: Create admin interface
  - View donation history
  - Export donation data
  - Manage refunds and disputes
  
- [ ] **Task 5.2**: Reporting system
  - Monthly/yearly donation reports
  - Donor analytics
  - Export for accounting software

## Current Implementation Details

### Donation Flow
1. User fills form on `/donate` page
2. Frontend calls `/api/checkout_token` (currently GET - needs to be POST)
3. Backend calls Helcim API to initialize payment
4. Helcim returns checkout token
5. Frontend opens Helcim hosted checkout
6. Payment processed by Helcim
7. Success/failure handled via JavaScript message events

### Environment Variables
```
CSRF_SECRET=<secret>
GIN_MODE=release
HELCIM_PRIVATE_API_KEY=<api_key>
PORT=3001 (default if not set)
```

### Key Files
- `main.go` - Main application with routes and Helcim integration
- `templates/donate.html` - Donation form and payment flow
- `.env` - Environment configuration
- `static/` - CSS, JS, and asset files

## Development Commands

```bash
# Run the application
go run main.go

# Build for production
go build -o tmp/main main.go

# Test local API
./test-api-local.sh

# Test production API
./test-api-prod.sh
```

## Next Steps for AI Assistant

**PRIORITY**: Always check this README first to understand current progress and what task to work on next.

1. **Current Phase**: Phase 1 (Security & API Improvements)
2. **Next Task**: Task 1.1 - Convert GET to POST endpoint
3. **Focus Area**: Improve donation API security before adding new features

### CRITICAL: Documentation Update Requirements
After completing any task or making significant changes:
- [ ] **Update checkboxes in this README.md** to reflect completed work
- [ ] **Update PROJECT_TRACKING.md** with task status changes (📋 PLANNED → 🔄 IN PROGRESS → ✅ COMPLETED)
- [ ] **Update .github/copilot-instructions.md** if new patterns or issues are discovered
- [ ] **Document any new requirements** or changes in approach
- [ ] **Move to next task** in sequence and update current priority section

## Notes
- Project uses WSL Arch Linux environment
- VS Code configured for Arch WSL terminal
- Follows clean Go practices with minimal comments
- Uses DaisyUI components for consistent styling
- Security-focused development required for nonprofit operations
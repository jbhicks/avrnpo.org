# AVR NPO Documentation

Comprehensive documentation for the American Veterans Rebuilding (AVR) donation system built with Buffalo framework and Helcim payment processing.

## 🚀 Quick Start

New to the project? Start here:

1. **[Development Setup](#-getting-started)** - Environment setup and first run
2. **[Payment System Overview](./payment-system/README.md)** - Core donation functionality  
3. **[Buffalo Framework Guide](./buffalo-framework/README.md)** - Daily development workflow

## 📚 Documentation Structure

### 🎯 Getting Started
Essential guides for new developers and daily development workflows.

- **[Quick Start](./getting-started/quick-start.md)** - Environment setup and first run
- **[Development Workflow](./getting-started/development-workflow.md)** - Daily development commands
- **[Testing Guide](./getting-started/testing-guide.md)** - How to run tests properly

### 💳 Payment System  
Complete donation and subscription management system documentation.

- **[Payment Overview](./payment-system/README.md)** - System architecture and status
- **[Helcim Integration](./payment-system/helcim-integration.md)** - Complete API integration guide
- **[Donation Flow](./payment-system/donation-flow.md)** - User experience and form handling
- **[Recurring Payments](./payment-system/recurring-payments.md)** - Subscription management
- **[Webhooks](./payment-system/webhooks.md)** - Event processing and notifications
- **[Testing](./payment-system/testing.md)** - Payment testing procedures

### 🦬 Buffalo Framework
Buffalo web framework development guides and best practices.

- **[Buffalo Overview](./buffalo-framework/README.md)** - Framework guide and critical rules  
- **[Templates](./buffalo-framework/templates.md)** - Plush templating and partial naming
- **[Routing & HTMX](./buffalo-framework/routing-htmx.md)** - Route configuration and HTMX
- **[Authentication](./buffalo-framework/authentication.md)** - Auth patterns and testing
- **[Database](./buffalo-framework/database.md)** - Migrations and database operations
- **[Troubleshooting](./buffalo-framework/troubleshooting.md)** - Common issues and solutions

### 🎨 Frontend Development
Styling, interactions, and user interface implementation.

- **[Pico CSS Guide](./frontend/pico-css.md)** - Styling with Pico CSS variables
- **[HTMX Patterns](./frontend/htmx-patterns.md)** - HTMX best practices and progressive enhancement
- **[Assets](./frontend/assets.md)** - Asset pipeline and management

### 🚀 Deployment & Production
Production deployment, security, and monitoring.

- **[Production Checklist](./deployment/production-checklist.md)** - Go-live requirements
- **[Security Guidelines](./deployment/security.md)** - Security best practices
- **[Monitoring](./deployment/monitoring.md)** - Logging and monitoring setup

### 📖 Reference
API documentation, schemas, and technical references.

- **[API Endpoints](./reference/api-endpoints.md)** - Complete API reference
- **[Database Schema](./reference/database-schema.md)** - Current database structure
- **[Dependencies](./reference/dependencies.md)** - Dependency management rules  
- **[Changelog](./reference/changelog.md)** - Major changes and updates

## 🔍 Finding Information

### By Developer Role

**🆕 New Developer:**
1. [Quick Start](./getting-started/quick-start.md) - Get running quickly
2. [Buffalo Overview](./buffalo-framework/README.md) - Learn the framework
3. [Payment Overview](./payment-system/README.md) - Understand core functionality

**💻 Daily Development:**
1. [Development Workflow](./getting-started/development-workflow.md) - Common commands
2. [Testing Guide](./getting-started/testing-guide.md) - How to test changes
3. [Troubleshooting](./buffalo-framework/troubleshooting.md) - Fix common issues

**💳 Payment Features:**
1. [Helcim Integration](./payment-system/helcim-integration.md) - Complete API guide
2. [Donation Flow](./payment-system/donation-flow.md) - Frontend implementation
3. [Recurring Payments](./payment-system/recurring-payments.md) - Subscription system

**🎨 Frontend Work:**
1. [Pico CSS Guide](./frontend/pico-css.md) - Styling and theming
2. [HTMX Patterns](./frontend/htmx-patterns.md) - Progressive enhancement
3. [Templates](./buffalo-framework/templates.md) - Template development

### By Problem Type

**🐛 Something Broken:**
- [Troubleshooting](./buffalo-framework/troubleshooting.md) - Common Buffalo issues
- [Testing Guide](./getting-started/testing-guide.md) - How to verify fixes
- [Security Guidelines](./deployment/security.md) - Security concerns

**🚀 Adding Features:**
- [Buffalo Overview](./buffalo-framework/README.md) - Framework patterns
- [API Endpoints](./reference/api-endpoints.md) - Existing API structure
- [Database Schema](./reference/database-schema.md) - Current database design

**🎯 Payment Issues:**
- [Payment Testing](./payment-system/testing.md) - Test procedures and cards
- [Helcim Integration](./payment-system/helcim-integration.md) - API troubleshooting  
- [Webhooks](./payment-system/webhooks.md) - Event processing issues

## 🎯 Current Project Status

### ✅ Completed Features (Phase 2)
- **One-time donations** - Full Helcim Payment API integration
- **Recurring donations** - Monthly subscriptions via Helcim Recurring API  
- **User account linking** - Donations tied to user accounts when logged in
- **Subscription management** - View, cancel, and update subscriptions
- **Receipt system** - Email confirmations for all donations
- **Webhook processing** - Real-time payment status updates

### 🔄 Current Focus
- **Documentation organization** - Consolidating and improving developer experience
- **Testing procedures** - Ensuring robust payment system operation
- **User experience refinement** - Optimizing donation and subscription flows

### 🎯 Future Enhancements
- **Enhanced reporting** - Donation analytics and donor insights
- **Campaign integration** - Blog-driven donation campaigns  
- **Admin interface** - Donation and subscription oversight tools

## 📋 Quick Reference

### Essential Commands
```bash
# Start development environment
make dev

# Run comprehensive tests  
make test

# Run quick tests (assumes database running)
make test-fast

# Database migrations
soda migrate up

# Check Buffalo status
ps aux | grep buffalo
lsof -i :3000
```

### Key URLs (Development)
- **Application:** http://127.0.0.1:3000
- **Donation Page:** http://127.0.0.1:3000/donation  
- **User Account:** http://127.0.0.1:3000/account
- **Admin Panel:** http://127.0.0.1:3000/admin

### Environment Files
- **`.env`** - Development environment variables
- **`database.yml`** - Database configuration
- **`config/buffalo-app.toml`** - Buffalo application settings

## 🆘 Getting Help

1. **Check troubleshooting guides** in relevant topic areas
2. **Search this documentation** for specific error messages or concepts
3. **Verify test procedures** to ensure changes work correctly
4. **Review error logs** in Buffalo console output or log files

## 📝 Contributing to Documentation

When updating documentation:
- **Place in appropriate topic directory** based on functional area
- **Update relevant README.md files** to maintain navigation
- **Use consistent formatting** and include practical examples
- **Test all code examples** to ensure they work correctly
- **Follow security guidelines** - never expose real credentials

---

**Documentation Organization:** This structure replaces the previous 39-file sprawl with a logical hierarchy organized around developer workflows and functional areas. Each topic directory contains a README.md that provides navigation within that area, while this main README provides navigation across all areas.

## 🗂️ Legacy Documentation

The previous documentation files have been reorganized. If you're looking for specific content from the old structure, check the reorganization plan:

- **[Documentation Reorganization Plan](./DOCUMENTATION_REORGANIZATION_PLAN.md)** - Complete migration strategy and file mapping

For immediate access to legacy content while the migration is in progress, the old files remain available in the root docs directory.

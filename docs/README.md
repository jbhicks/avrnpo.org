# AVR NPO Documentation

Comprehensive documentation for the American Veterans Rebuilding (AVR) donation system built with PocketBase and Helcim payment processing.

## 🚀 Quick Start

New to the project? Start here:

1. **[Development Setup](./getting-started/quick-start.md)** - Environment setup and first run
2. **[Payment System Overview](./payment-system/README.md)** - Core donation functionality  
3. **[PocketBase Development Guide](./DEVELOPMENT_GUIDE.md)** - Daily development workflow

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
- **[Recurring Payments](./payment-system/recurring-payments-final.md)** - Subscription management
- **[Webhooks](./payment-system/webhooks.md)** - Event processing and notifications
- **[Testing](./payment-system/testing.md)** - Payment testing procedures

### 🎨 Frontend Development
Styling, interactions, and user interface implementation.

- **[Pico CSS Guide](./frontend/pico-css.md)** - Styling with Pico CSS variables
- **[HTMX Patterns](./frontend/htmx-patterns.md)** - HTMX best practices and progressive enhancement
- **[Templ Templates](./frontend/templ-guide.md)** - Type-safe Go templates

### 🚀 Deployment & Production
Production deployment, security, and monitoring.

- **[Production Checklist](./deployment/production-checklist.md)** - Go-live requirements
- **[Security Guidelines](./deployment/security.md)** - Security best practices
- **[Deployment History](./changelog/deployment-history.md)** - Coolify deployment notes

### 🔧 Development & Setup
Development tools, configuration, and contributor resources.

- **[Development Guide](./DEVELOPMENT_GUIDE.md)** - Current development workflow
- **[Current Feature](./development/current-feature.md)** - Active feature work (Phase 3: Templ UI)
- **[Refactoring Status](../REFACTORING_STATUS.md)** - PocketBase migration status
- **[Setup & Configuration](./setup/README.md)** - System setup and environment configuration

### 📖 Reference
API documentation, schemas, and technical references.

- **[Dependencies](./reference/dependencies.md)** - Dependency management rules
- **[Subscription API Reference](./reference/subscription-api-reference.md)** - PocketBase collections schema

## 🔍 Finding Information

### By Developer Role

**🆕 New Developer:**
1. [Quick Start](./getting-started/quick-start.md) - Get running quickly
2. [Development Guide](./DEVELOPMENT_GUIDE.md) - Learn the stack
3. [Payment Overview](./payment-system/README.md) - Understand core functionality

**💻 Daily Development:**
1. [Development Workflow](./getting-started/development-workflow.md) - Common commands
2. [Testing Guide](./getting-started/testing-guide.md) - How to test changes
3. [Refactoring Status](../REFACTORING_STATUS.md) - Current project state

**🤖 AI Contributors:**
1. [Current Feature](./development/current-feature.md) - What we're currently working on (Phase 3)
2. [Development Guide](./DEVELOPMENT_GUIDE.md) - Development tools and patterns
3. [Refactoring Status](../REFACTORING_STATUS.md) - Migration progress

**💳 Payment Features:**
1. [Helcim Integration](./payment-system/helcim-integration.md) - Complete API guide
2. [Donation Flow](./payment-system/donation-flow.md) - Frontend implementation
3. [Recurring Payments](./payment-system/recurring-payments-final.md) - Subscription system

**🎨 Frontend Work:**
1. [Pico CSS Guide](./frontend/pico-css.md) - Styling and theming
2. [HTMX Patterns](./frontend/htmx-patterns.md) - Progressive enhancement
3. [Templ Templates](./frontend/templ-guide.md) - Template development

**⚙️ Setup & Configuration:**
1. [Setup Guide](./setup/README.md) - System configuration and admin setup
2. [Receipt Setup](./setup/receipt-setup.md) - Payment confirmation configuration
3. [Environment Setup](./getting-started/setup-checklist.md) - Development environment

## 🎯 Current Project Status

### ✅ Phase 2 Complete: PocketBase Migration
- **PocketBase initialized** - Single binary, SQLite database
- **Collections created** - `posts`, `donations`, `contact_submissions`, `users`
- **Migrations working** - Go-based migration system
- **Buffalo archived** - All old code preserved in `archive/buffalo/`
- **Services preserved** - Helcim and Email services ready for integration

### 🔄 Current Focus: Phase 3
- **Template strategy** - Implementing Templ for type-safe Go templates
- **UI implementation** - Building pages (home, blog, donate, contact)
- **Service integration** - Connecting Helcim and Email to PocketBase
- **Route handlers** - Updating handlers to use PocketBase SDK

### 🎯 Future Phases
- **Phase 4:** Data migration from Buffalo/PostgreSQL (if needed)
- **Phase 5:** Testing and deployment to production

## 📋 Quick Reference

### Essential Commands
```bash
# Build and run PocketBase
go build -o avrnpo ./main.go
./avrnpo serve --dev

# Apply migrations
./avrnpo migrate up

# Create superuser
./avrnpo superuser create admin@avrnpo.org password

# Generate Templ templates
templ generate

# Watch Templ files
templ generate --watch
```

### Key URLs (Development)
- **Application:** http://127.0.0.1:8090
- **PocketBase Admin:** http://127.0.0.1:8090/_/
- **API Base:** http://127.0.0.1:8090/api/

### Environment Files
- **`.env`** - Environment variables
- **`pb_data/`** - PocketBase SQLite database and files
- **`migrations/`** - Go-based database migrations

## 🆘 Getting Help

1. **Check current status** - [REFACTORING_STATUS.md](../REFACTORING_STATUS.md)
2. **Search this documentation** for specific topics
3. **Review PocketBase docs** - https://pocketbase.io/docs/
4. **Check archived Buffalo docs** - [ARCHIVE.md](./ARCHIVE.md)

## 📝 Documentation Organization

This documentation is organized by functional area:

- **`getting-started/`** - Setup and initial workflows
- **`payment-system/`** - Helcim integration and donation system
- **`frontend/`** - UI, styling, and templates
- **`development/`** - Active feature work and refactoring
- **`deployment/`** - Production deployment guides
- **`setup/`** - System configuration
- **`reference/`** - API and technical references
- **`changelog/`** - Project history
- **`buffalo-framework/`** - **ARCHIVED** - Buffalo-era documentation

## 🗂️ Archived Documentation

Buffalo framework documentation has been archived. See:

- **[ARCHIVE.md](./ARCHIVE.md)** - Links to all Buffalo-era documentation

The project migrated from Buffalo to PocketBase in October 2025. All Buffalo, Plush, Pop, and Fizz references are historical only.

## 🔄 Current Stack

- **Backend:** PocketBase (Go)
- **Database:** SQLite (embedded)
- **Templates:** Templ (type-safe Go templates)
- **Frontend:** HTMX + Pico CSS
- **Payments:** Helcim API
- **Email:** PocketBase mailer
- **Deployment:** Coolify (Docker)

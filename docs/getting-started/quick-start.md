# Quick Start Guide

Get the AVR NPO donation system running locally in under 5 minutes.

## 🎯 Prerequisites

- Go 1.23+ installed
- Templ CLI installed (`go install github.com/a-h/templ/cmd/templ@latest`)
- Git installed

## ⚡ Quick Setup

### 1. Clone and Setup
```bash
git clone https://github.com/jbhicks/avrnpo.org.git
cd avrnpo.org
```

### 2. Environment Configuration
```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your settings
# Required: Set PB_ADMIN_EMAIL and PB_ADMIN_PASSWORD
```

### 3. Start Development
```bash
make dev
```

**That's it!** The application will be running at http://127.0.0.1:8090

### 4. Verify Setup
- Visit http://127.0.0.1:8090 - you should see the AVR homepage
- Visit http://127.0.0.1:8090/donate - you should see the donation form
- Visit http://127.0.0.1:8090/_/ - PocketBase admin UI (login with PB_ADMIN_EMAIL/PASSWORD)

## 🧪 Test Everything Works

```bash
# Run unit tests
go test ./...

# Run E2E tests
E2E_TESTS=1 go test -v -run E2E
```

## 🎯 What Just Happened?

1. **Database Initialized**: SQLite database created at `pb_data/data.db`
2. **Admin User Created**: Auto-created from .env on first run
3. **Migrations Applied**: Collections and schema created automatically
4. **Server Started**: PocketBase + custom handlers running with hot reload
5. **Assets Served**: Static files from `public/assets/`

## 📋 Essential Commands

```bash
# Start development server with hot reload
make dev

# Run unit tests
go test ./...

# Run E2E tests  
E2E_TESTS=1 go test -v -run E2E

# Regenerate Templ templates after editing
templ generate

# Build production binary
go build
```

## 🎯 Development URLs

- **Application**: http://127.0.0.1:8090
- **Donation Page**: http://127.0.0.1:8090/donate
- **PocketBase Admin**: http://127.0.0.1:8090/_/
- **API Base**: http://127.0.0.1:8090/api/

## 🔧 If Something Goes Wrong

1. **Port already in use**: Check if port 8090 is being used by another process
2. **Database errors**: Delete `pb_data/` folder and restart to recreate
3. **Template errors**: Run `templ generate` to regenerate template code
4. **Admin login fails**: Check PB_ADMIN_EMAIL/PASSWORD in .env

## 📚 Next Steps

- **[Development Workflow](./development-workflow.md)** - Daily development patterns
- **[Testing Guide](./testing-guide.md)** - How to test properly
- **[Development Guide](../DEVELOPMENT_GUIDE.md)** - Complete development reference
- **[Payment System Overview](../payment-system/README.md)** - Donation functionality

## 🆘 Need Help?

- Check **[Development Guide](../DEVELOPMENT_GUIDE.md)** for comprehensive documentation
- Review **[Testing Guide](./testing-guide.md)** for testing patterns
- See PocketBase docs at https://pocketbase.io/docs/

# Development Setup Checklist

Quick checklist to get the AVR NPO project running locally.

## ✅ Initial Setup

### Prerequisites
- [ ] Go 1.23+ installed (`go version`)
- [ ] Templ CLI installed (`templ version`)
- [ ] Git installed (`git --version`)

### Project Setup
- [ ] Clone repository: `git clone https://github.com/jbhicks/avrnpo.org.git`
- [ ] Navigate to directory: `cd avrnpo.org`
- [ ] Copy environment file: `cp .env.example .env`
- [ ] Edit `.env` with your settings (minimum: PB_ADMIN_EMAIL, PB_ADMIN_PASSWORD)

### First Run
- [ ] Start development server: `make dev`
- [ ] Verify app loads at http://127.0.0.1:8090
- [ ] Check PocketBase admin at http://127.0.0.1:8090/_/
- [ ] Login with credentials from .env

## ✅ Verify Core Functionality

### Pages Load
- [ ] Homepage: http://127.0.0.1:8090/
- [ ] Donate page: http://127.0.0.1:8090/donate
- [ ] Blog page: http://127.0.0.1:8090/blog
- [ ] About page: http://127.0.0.1:8090/about
- [ ] Contact page: http://127.0.0.1:8090/contact

### PocketBase Admin
- [ ] Collections visible (users, posts, donations)
- [ ] Admin user exists
- [ ] Can create test blog post
- [ ] Can view logs

### Testing
- [ ] Run unit tests: `go test ./...`
- [ ] All tests pass
- [ ] (Optional) Run E2E tests: `E2E_TESTS=1 go test -v -run E2E`

## ✅ Development Workflow

### Daily Development
- [ ] Start server: `make dev`
- [ ] Edit files (server auto-reloads)
- [ ] Check terminal for errors
- [ ] Run tests in separate terminal

### Template Development
- [ ] Edit `.templ` files in `templates/`
- [ ] Run `templ generate` (or let Air do it)
- [ ] Check browser for changes
- [ ] Format templates: `templ fmt templates/`

### Database Changes
- [ ] Create migration file in `pb_migrations/`
- [ ] Follow naming: `TIMESTAMP_description.go`
- [ ] Restart server to apply migration
- [ ] Verify changes in admin UI

## ✅ Optional Configuration

### Payment System (Helcim)
- [ ] Get Helcim API credentials
- [ ] Add to `.env`: HELCIM_API_TOKEN, HELCIM_TERMINAL_ID
- [ ] Test donation flow at /donate

### Email System (Resend)
- [ ] Get Resend API key
- [ ] Add to `.env`: RESEND_API_KEY
- [ ] Test contact form

### Production Build
- [ ] Build binary: `go build`
- [ ] Run production: `./avrnpo serve`
- [ ] Verify production mode works

## 🚨 Common Issues

### Port Already in Use
```bash
# Check what's using port 8090
lsof -ti:8090

# Kill the process
pkill -f "avrnpo serve"
```

### Database Reset Needed
```bash
# WARNING: Deletes all data
rm -rf pb_data/
make dev  # Recreates database
```

### Template Compilation Errors
```bash
# Manually regenerate templates
templ generate

# Check for syntax errors
templ fmt templates/
```

### Tests Failing
```bash
# Clean test cache
go clean -testcache

# Run tests with verbose output
go test -v ./...
```

## 📚 Next Steps

Once setup is complete:
- [ ] Read [Development Guide](../DEVELOPMENT_GUIDE.md)
- [ ] Review [Development Workflow](./development-workflow.md)
- [ ] Check [Testing Guide](./testing-guide.md)
- [ ] Explore [Payment System docs](../payment-system/README.md)

## ✅ Setup Complete!

If all checkboxes above are complete, you're ready to start developing!

**Key Commands:**
- `make dev` - Start development server
- `go test ./...` - Run tests
- `templ generate` - Regenerate templates
- `go build` - Build production binary

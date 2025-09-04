# Quick Start Guide

Get the AVR NPO donation system running locally in under 10 minutes.

## 🎯 Prerequisites

- Go 1.24+ installed
- Podman or Docker installed  
- Git installed

## ⚡ Quick Setup

### 1. Clone and Setup
```bash
git clone https://github.com/jbhicks/avrnpo.org.git
cd avrnpo.org
```

### 2. Start Development Environment
```bash
# This command starts everything: PostgreSQL + Buffalo
make dev
```

**That's it!** The application will be running at http://127.0.0.1:3001

### 3. Verify Setup
- Visit http://127.0.0.1:3001 - you should see the AVR homepage
- Visit http://127.0.0.1:3001/donation - you should see the donation form
- Buffalo console will show logs in your terminal

## 🧪 Test Everything Works

```bash
# In a new terminal (keep Buffalo running)
make test-fast
```

## 🎯 What Just Happened?

1. **Database Started**: PostgreSQL container launched via Podman
2. **Migrations Applied**: Database schema created automatically  
3. **Buffalo Started**: Web server running with hot reload
4. **Assets Served**: CSS/JS files available at `/assets/`

## 📋 Essential Commands

```bash
# Start development (run once)
make dev

# Run tests (Buffalo keeps running)
make test-fast

# Check what's running
ps aux | grep buffalo
lsof -i :3000

# Database operations
soda migrate up
soda reset  # Reset database if needed
```

## 🚨 Important Rules

- **Don't kill Buffalo** - It auto-reloads on all file changes
- **Use `make test-fast`** - Never use `go test` directly
- **Use `soda` commands** - Not `buffalo pop` (removed in v0.18.14+)

## 🔧 If Something Goes Wrong

1. **Buffalo won't start**: Check if port 3000 is already in use
2. **Database errors**: Run `soda reset` to recreate database
3. **Template errors**: Check partial naming (use `partial("dir/file")` not `partial("dir/_file.plush.html")`)
4. **Test failures**: Ensure PostgreSQL is running before tests

## 📚 Next Steps

- **[Development Workflow](./development-workflow.md)** - Daily development commands
- **[Testing Guide](./testing-guide.md)** - How to test properly
- **[Buffalo Framework Guide](../buffalo-framework/README.md)** - Framework patterns
- **[Payment System Overview](../payment-system/README.md)** - Donation functionality

## 🎯 Development URLs

- **Application**: http://127.0.0.1:3000
- **Donation Page**: http://127.0.0.1:3000/donation
- **Admin Panel**: http://127.0.0.1:3000/admin
- **User Account**: http://127.0.0.1:3000/account

## 🆘 Need Help?

- Check **[Troubleshooting](../buffalo-framework/troubleshooting.md)** for common issues
- Review **[Development Workflow](./development-workflow.md)** for detailed commands
- See **[Testing Guide](./testing-guide.md)** for testing problems

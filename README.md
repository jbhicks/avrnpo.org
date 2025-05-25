# My Go SaaS Template

A Buffalo-based SaaS application template with containerized PostgreSQL database and complete authentication system.

## ✅ Setup Checklist

- [x] **Buffalo application generated** - Basic Buffalo app structure created
- [x] **PostgreSQL containerized** - Docker Compose setup with PostgreSQL 15
- [x] **Database configuration** - All databases created (development, test, production)
- [x] **Database migrations** - Schema up to date and working
- [x] **Application running** - Buffalo dev server successfully connecting to database
- [x] **Authentication system** - Complete user registration, login, logout with session management
- [x] **User dashboard** - Protected dashboard with user dropdown menu
- [x] **Template system** - Plush templates with proper syntax and Alpine.js integration
- [ ] User profile management
- [ ] Billing/subscription features
- [ ] Email services
- [ ] Production deployment

## 🚀 Quick Start

### Prerequisites
- Go 1.19+
- Docker and Docker Compose
- Buffalo CLI

### Development Mode (Recommended)

The easiest way to get started is using the Makefile shortcuts:

```console
# First time setup (creates database and runs migrations)
make setup

# Start development mode (database + Buffalo dev server)
make dev
```

Visit [http://127.0.0.1:3000](http://127.0.0.1:3000) to see your application.

### Manual Setup

If you prefer to run commands manually:

```console
# Start PostgreSQL container
docker-compose up -d

# Run database migrations (if needed)
buffalo pop migrate

# Start the development server
buffalo dev
```

### Available Make Commands

- **`make dev`** - Start database and Buffalo development server
- **`make setup`** - Initial setup (database + migrations)
- **`make db-up`** - Start only the database
- **`make test`** - Run tests with database
- **`make clean`** - Stop all services and clean up

## 🔐 Authentication Features

### User Registration & Login
- **Registration**: `/users/new` - Create new user accounts
- **Login**: `/auth/new` - Sign in with email/password
- **Dashboard**: `/dashboard` - Protected area for authenticated users
- **Logout**: Available via user dropdown menu

### User Interface
- **Landing Page**: Marketing page with conditional CTAs based on auth status
- **User Dropdown**: Professional dropdown menu in dashboard top-right with:
  - User avatar (initials with gradient background)
  - User name display
  - Profile Settings (placeholder)
  - Account Settings (placeholder)  
  - Sign Out functionality

### Authentication Flow
1. Unauthenticated users see the landing page with sign-up/sign-in options
2. After successful login, users are redirected to `/dashboard`
3. Dashboard features protected content and user dropdown menu
4. Logout clears session and returns to landing page

## 📊 Architecture

- **Backend**: Buffalo (Go web framework)
- **Database**: PostgreSQL 15 (containerized)
- **Frontend**: Plush templates with Tailwind CSS and Alpine.js
- **Authentication**: Session-based with bcrypt password hashing
- **Background Jobs**: Buffalo workers

## 🛠️ Development

### Template Development

This project uses Buffalo's Plush templating engine. Key documentation:

- **Template Syntax**: See `/docs/buffalo-template-syntax.md` for Plush syntax reference
- **String Operations**: Use built-in helpers like `capitalize()`, avoid Go-style syntax
- **Conditionals**: Use `<%= if (condition) { %>` format
- **Interactive Elements**: Alpine.js is included for dropdowns and dynamic behavior

### Database Management

The application includes a PostgreSQL container configured via `docker-compose.yml`:

- **Development DB**: `my_go_saas_template_development`
- **Test DB**: `my_go_saas_template_test`
- **Production DB**: `my_go_saas_template_production`

### Common Commands

```console
# Development shortcuts (recommended)
make dev                       # Start everything for development
make setup                     # First-time setup
make test                      # Run tests
make clean                     # Stop and cleanup

# Manual commands
# Database operations
buffalo pop create -a          # Create databases
buffalo pop migrate            # Run migrations
buffalo pop generate migration # Create new migration

# Development
buffalo dev                    # Start dev server with hot reload
buffalo build                  # Build production binary
buffalo test                   # Run tests

# Docker
docker-compose up -d           # Start database
docker-compose down            # Stop database

# User management (examples)
# Register: POST /users with {email, password, first_name, last_name}
# Login: POST /auth with {email, password}
# Logout: DELETE /auth
```

## 🤖 Bot Instructions

When working with this Buffalo SaaS template:

### Template Development
1. **Always use proper Plush syntax** - Refer to `/docs/buffalo-template-syntax.md`
2. **Avoid Go-style operations** - Use Plush helpers instead of `[0:1]`, `strings.Split()`, etc.
3. **Test template changes** - Template syntax errors cause 500 errors
4. **Use Alpine.js** - Already included for interactive components

### Authentication Features
1. **User dropdown** - Located in dashboard template, uses Alpine.js
2. **Protected routes** - Use `Authorize` middleware for protected areas
3. **Session management** - User ID stored in `current_user_id` session key
4. **Post-login redirect** - Currently goes to `/dashboard`, customize in `AuthCreate`

### Common Patterns
- **User avatar**: Uses `capitalize()` helper for first letter
- **Conditional content**: Check `current_user` existence for auth-specific content
- **Form handling**: Use Buffalo's `linkTo()` helper with proper HTTP methods
- **Error handling**: Template errors appear in Buffalo logs

### Troubleshooting
- **500 errors**: Usually template syntax issues, check Buffalo logs
- **Auth issues**: Verify session middleware and user context
- **Dropdown not working**: Ensure Alpine.js is loaded and proper `x-data` syntax

## 🐛 Troubleshooting Log

### Issue 1: Database Connection Refused (Resolved ✅)
**Date**: May 24, 2025  
**Problem**: Buffalo app couldn't connect to PostgreSQL - "dial tcp 127.0.0.1:5432: connect: connection refused"  
**Root Cause**: PostgreSQL was not running  
**Solution**: 
1. Created `docker-compose.yml` with PostgreSQL 15 container
2. Added `init.sql` to create test and production databases
3. Started container with `docker-compose up -d`
4. Restarted Buffalo dev server to establish connection

**Files Modified**:
- Added `docker-compose.yml`
- Added `init.sql`

### Issue 2: Port Conflict During Restart (Resolved ✅)
**Date**: May 24, 2025  
**Problem**: "address already in use" when restarting Buffalo dev server  
**Solution**: Properly stopped old processes with `pkill -f "my-go-saas-template-build"`

### Issue 3: Template Syntax Errors in User Dropdown (Resolved ✅)
**Date**: May 24, 2025  
**Problem**: HTTP 500 errors when accessing dashboard with user dropdown  
**Root Cause**: Used Go-style string slicing syntax `[0:1]` which Plush doesn't support  
**Solution**: 
1. Updated template to use `capitalize()` helper instead of string slicing
2. Simplified conditional logic to avoid complex string operations
3. Added proper Plush template syntax documentation

**Files Modified**:
- Fixed `templates/home/dashboard.plush.html`
- Added `docs/buffalo-template-syntax.md`
- Added Alpine.js to `templates/application.plush.html`

## 📁 Project Structure

```
.
├── actions/          # Controllers and HTTP handlers
│   ├── auth.go      # Authentication (login/logout)
│   ├── users.go     # User registration
│   └── home.go      # Dashboard and landing pages
├── cmd/app/         # Application entry point
├── config/          # Buffalo configuration
├── docs/            # Documentation
│   └── buffalo-template-syntax.md  # Plush template guide
├── migrations/      # Database migrations
├── models/          # Database models
│   └── user.go      # User model with bcrypt
├── templates/       # HTML templates
│   ├── home/
│   │   ├── index.plush.html      # Landing page
│   │   └── dashboard.plush.html  # Protected dashboard
│   ├── auth/        # Authentication templates
│   └── users/       # User registration templates
├── public/          # Static assets
├── docker-compose.yml  # PostgreSQL container setup
├── database.yml     # Database configuration
└── README.md        # This file
```

## 🔗 Resources

- [Buffalo Documentation](http://gobuffalo.io)
- [Plush Template Documentation](https://github.com/gobuffalo/plush)
- [Alpine.js Documentation](https://alpinejs.dev/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Tailwind CSS Documentation](https://tailwindcss.com/)

---

**Status**: ✅ Authentication system complete - Ready for advanced SaaS features  
**Last Updated**: May 24, 2025

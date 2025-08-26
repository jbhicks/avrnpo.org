# Buffalo Framework Development Guide

Complete guide for developing with the Buffalo web framework for the AVR NPO donation system.

## 📋 Quick Navigation

- [Templates](./templates.md) - Plush templating syntax and patterns
- [Forms & HTMX Patterns](./forms-and-htmx-patterns.md) - **CRITICAL** form handling and URL management
- [Routing & HTMX](./routing-htmx.md) - Route configuration and HTMX integration  
- [Authentication](./authentication.md) - Auth patterns and testing strategies
- [Database](./database.md) - Migrations and database operations
- [Troubleshooting](./troubleshooting.md) - Common issues and solutions

## 🎯 Buffalo Project Structure

```
avrnpo.org/
├── actions/           # Route handlers and business logic
├── models/           # Database models and validation  
├── templates/        # Plush HTML templates
├── public/          # Static assets (CSS, JS, images)
├── migrations/      # Database schema changes (.fizz files)
├── grifts/          # Background tasks and commands
├── config/          # Buffalo configuration
└── cmd/app/         # Application entry point
```

## 🚨 CRITICAL Buffalo Development Rules

### 🚨 NEVER SUBMIT FORMS TO API ENDPOINTS
**CRITICAL WARNING:** Form submission to API endpoints causes URL issues that break user experience.

**❌ FORBIDDEN:** `<form action="/api/anything">`  
**✅ REQUIRED:** `<form action="/route/submit" hx-post="/route/submit" hx-target="body" hx-swap="outerHTML" hx-push-url="true">`

**Why API endpoints break forms:**
- Form submits to `POST /api/endpoint`
- Handler redirects to `/success`  
- Browser URL shows `/api/endpoint` (NOT `/success`)
- Refreshing page tries `GET /api/endpoint` → 404 ERROR

**📖 MANDATORY READING:** [Forms & HTMX Patterns](./forms-and-htmx-patterns.md)

### NEVER KILL BUFFALO UNLESS ABSOLUTELY NECESSARY
Buffalo has intelligent auto-reload that handles:
- **Go code changes** → Automatic recompilation and restart
- **Template changes** → Automatic template reload  
- **Static assets** → Automatic asset pipeline refresh
- **CSS/JS changes** → Hot reload without server restart

**✅ Trust Buffalo's auto-reload - it's designed to stay running throughout development**

### Buffalo Auto-Reload Works For:
- Editing Go files in `/actions/`, `/models/`, `/grifts/`
- Editing templates in `/templates/` (`.plush.html` files)
- Editing CSS in `/public/assets/css/`
- Editing JavaScript in `/public/assets/js/`
- Running database migrations (`soda migrate up`)
- Configuration changes

### 🚨 ONLY Restart Buffalo When:
- **Compilation errors prevent auto-reload** (syntax errors in Go code)
- **User explicitly requests restart**
- **Adding new routes or middleware** that requires full restart
- **Environment variable changes** (rare in development)
- **Debugging why auto-reload isn't working**

## 🗄️ Database Operations

### 🚨 CRITICAL: Use `soda` Commands, NOT `buffalo pop`

Buffalo v0.18.14+ does not include the `pop` plugin. Use these commands:

```bash
# Run pending migrations
soda migrate up

# Reset database (drop, create, migrate)  
soda reset

# Reset test database
GO_ENV=test soda reset

# Create all databases
soda create -a

# Create new migration
soda generate migration create_posts
```

### Migration Best Practices
- **Use `.fizz` files** instead of `.sql` for cross-database compatibility
- **Never manually edit** auto-generated `schema.sql` files
- **Delete problematic schema.sql files** if they cause version conflicts
- **Atomic migrations** - each migration should be reversible

```bash
# Remove auto-generated SQL files that can cause issues
rm -f db/schema.sql migrations/schema.sql
```

## 🧪 Buffalo Testing System

### 🚨 CRITICAL: Proper Buffalo Testing Commands

**NEVER use `go test` directly** - Buffalo requires special setup.

**✅ Use these Makefile commands:**

```bash
# Comprehensive test suite (recommended)
make test

# Quick testing (assumes database running)  
make test-fast

# Automatic database management
make test-resilient
```

### Buffalo Testing Requirements
Buffalo tests need special setup that `go test` alone cannot provide:
- **PostgreSQL connection** - Test database must be running
- **Environment variables** - `GO_ENV=test` properly configured
- **Database migrations** - Test database with proper schema
- **Transaction isolation** - ActionSuite handles test data cleanup
- **Session management** - Buffalo test session handling

### Testing Best Practices
```go
// ✅ Correct ActionSuite pattern
func (as *ActionSuite) Test_DonationFlow() {
    res := as.HTML("/donation").Get()
    as.Equal(http.StatusOK, res.Code)
    as.Contains(res.Body.String(), "Donate to AVR")
}

// ✅ User creation with unique data
timestamp := time.Now().UnixNano()
user := &models.User{
    Email: fmt.Sprintf("test-%d@example.com", timestamp),
    // ... other fields
}

// ✅ Test both direct loads and HTMX navigation
res := as.HTML("/account/subscriptions").Get()
as.Equal(http.StatusOK, res.Code)
```

## 🎨 Template System

### 🚨 CRITICAL: Buffalo Partial Naming Convention

**This is the #1 source of recurring template errors:**

Buffalo automatically adds underscore prefix to partial filenames, causing double underscore issues if not handled correctly.

**✅ CORRECT Pattern:**
```html
<!-- Call partial WITHOUT underscore or extension -->
<%= partial("auth/new") %>
<!-- Buffalo looks for: templates/auth/_new.plush.html -->
```

**❌ WRONG Pattern:**
```html
<!-- DON'T include underscore - causes double underscore error -->
<%= partial("auth/_new.plush.html") %>
<!-- Buffalo looks for: templates/auth/__new.plush.html (FAILS) -->
```

### Template File Naming
- **Partial files:** `_filename.plush.html` (single underscore prefix)
- **Partial calls:** `partial("directory/filename")` (no underscore, no extension)
- **Layout files:** `application.plush.html`
- **Page templates:** `filename.plush.html`

## 🌐 Development Workflow

### Startup Commands
```bash
# Start everything (PostgreSQL + Buffalo)
make dev

# Check database status
podman-compose ps

# Check Buffalo status  
ps aux | grep buffalo
lsof -i :3000
```

### Daily Development Flow
1. **Start once:** `make dev` 
2. **Leave running:** Buffalo handles all reloading
3. **Edit files:** Make changes, Buffalo auto-reloads
4. **Refresh browser:** See changes immediately
5. **Run tests:** `make test-fast` (doesn't affect Buffalo)
6. **Only restart if:** Compilation errors or explicit need

### Environment Management
```bash
# Development environment
GO_ENV=development

# Test environment  
GO_ENV=test

# Production environment
GO_ENV=production
```

## 📦 Asset Pipeline

### Asset Structure
```
public/
├── assets/
│   ├── css/          # Stylesheets
│   ├── js/           # JavaScript files  
│   └── images/       # Static images
└── favicon.ico
```

### Buffalo Asset Helpers
```html
<!-- ✅ Correct - Buffalo adds /assets/ prefix -->
<%= stylesheetTag("app.css") %>
<%= javascriptTag("app.js") %>

<!-- ❌ Wrong - Don't include /assets/ manually -->
<%= stylesheetTag("/assets/app.css") %>
```

### Asset Pipeline Commands
```bash
# Development asset serving (automatic)
# Buffalo serves assets with hot reload

# Production asset compilation
buffalo build
```

## 🔧 Common Buffalo Patterns

### Route Definition
```go
// actions/app.go
app.GET("/", HomeHandler)
app.GET("/donation", DonationHandler)
app.POST("/donation/process", ProcessDonationHandler)

// With middleware
app.GET("/account", SetCurrentUser(Authorize(AccountHandler)))
```

### Handler Pattern
```go
func DonationHandler(c buffalo.Context) error {
    // Set template data
    c.Set("title", "Make a Donation")
    
    // Render template
    return c.Render(http.StatusOK, r.HTML("pages/donation.plush.html"))
}
```

### Model Validation
```go
// models/donation.go
var ValidateCreate = validators.ValidatorFunc(func(errors *validate.Errors) {
    v.MinLength("donor_name", 1, "Donor name is required")
    v.Email("donor_email", "Email must be valid")
    v.Range("amount", 1, 10000, "Amount must be between $1 and $10,000")
})
```

## 🔍 Debugging & Logs

### Log Levels
```go
// In handlers
c.Logger().Info("Processing donation", "amount", donation.Amount)
c.Logger().Error("Payment failed", "error", err)
c.Logger().Debug("Payment response", "response", resp)
```

### Log Files
```bash
# View Buffalo logs
tail -f buffalo.log

# View application logs  
tail -f logs/application.log
```

### Common Debug Techniques
1. **Add logging** to handlers for request tracing
2. **Check template paths** for partial errors
3. **Verify database connections** with connection tests
4. **Test API endpoints** with curl before frontend integration
5. **Use Buffalo console** for interactive debugging

## 🎯 Performance Optimization

### Template Caching
- **Development:** Templates reload on every request
- **Production:** Templates cached in memory
- **Partial caching:** Buffalo automatically optimizes partial rendering

### Database Optimization
```go
// Use Pop query optimization
users := []models.User{}
err := tx.Where("active = ?", true).All(&users)

// Avoid N+1 queries with eager loading
err := tx.Eager().All(&users)
```

### Asset Optimization
- **Development:** Individual file serving
- **Production:** Asset concatenation and minification  
- **CDN integration:** For static asset serving

This guide covers the essential Buffalo framework knowledge for effective development. For specific implementation details, see the individual topic guides linked above.

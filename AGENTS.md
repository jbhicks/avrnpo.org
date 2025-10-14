# avrnpo.org Development Guidelines

## Important: Server and Long-Running Process Management

### Running the Development Server

**The `make dev` command now automatically runs in the background with logging.**

- Starting server: `make dev` (logs to `/tmp/avrnpo-dev.log`)
- Stopping server: `make stop`
- Viewing logs: `tail -f /tmp/avrnpo-dev.log` or `tail /tmp/avrnpo-dev.log` (without `-f` to avoid blocking)

### General Pattern for Long-Running Processes

**ALWAYS** redirect output to log files and run in background to avoid blocking execution:

```bash
# Good: Background with logging
command > /tmp/process.log 2>&1 &
PROCESS_PID=$!
sleep 2
tail /tmp/process.log  # View logs without blocking

# Bad: Direct execution (blocks thread)
command

# Bad: tail -f (blocks thread)
tail -f /tmp/process.log
```

### When to Ask User vs Automate

**Ask the user** to start the server when:
- Making code changes that don't require immediate testing
- The user has their own terminal setup/workflow

**Use `make dev` automatically** when:
- You need to test functionality immediately
- Running E2E tests that require server
- Debugging live server behavior

**Always clean up:** Use `make stop` or `pkill -f "avrnpo serve"` when done testing.

## Active Technologies
- Go 1.23.0 + PocketBase v0.22+
- Templ v0.3.943 (template engine)
- HTMX (client-side interactions)
- Pico CSS (styling framework)

## Project Structure
```
pb_migrations/          # Database migrations (Go files)
templates/             # Templ templates (compiled to .go files)
public/               # Static assets (CSS, JS, images)
  assets/
    css/                # Pico CSS, custom theme variables
    js/                 # HTMX, theme toggler, editor
services/             # External service integrations (Helcim, email)
middleware/           # CSRF, rate limiting, validation
pb_data/              # PocketBase database and logs (SQLite)
main.go               # Application entry point
```

## Commands
```bash
make install      # Install development tools (Air, Templ)
make dev          # Start development server (background, logs to /tmp/avrnpo-dev.log)
make stop         # Stop development server
go test ./...     # Run unit tests
E2E_TESTS=1 go test -v -run E2E  # Run E2E tests
templ generate    # Regenerate template Go files
go build          # Build binary
```

## Template Design System

All Templ templates use either `Base` or `BasePage` for consistency:

**Structure:**
```go
templ MyPage(csrfToken string) {
    @Base("Page Title", csrfToken, myPageContent())
}

templ myPageContent() {
    // page content here
}
```

**Base Components:**
- `Base(title, csrfToken, content)` - Full page with navigation
- `BasePage(title, csrfToken, content)` - Simplified page (admin/login)
- Both include CSRF meta tag when csrfToken is provided

**Styling Guidelines:**
1. Use Pico CSS classes and semantic HTML
2. Forms: Always include `<input type="hidden" name="csrf_token" value={ csrfToken }/>`
3. HTMX forms: Use `hx-post`, `hx-put`, etc. with CSRF token
4. Cards: Use `<article>` with optional `<header>` and `<footer>`
5. Buttons: Use `role="button"` or `<button>`
6. Grids: Use `<div class="grid">` for responsive columns

## Content Management

**Blog Posts & Content:**
- Use PocketBase Admin UI at `/_/` for managing posts, users, content
- Templates are READ-ONLY for displaying content
- Do NOT create custom CRUD forms for content managed by PocketBase admin

## Security

**CSRF Protection:**
- All forms MUST include CSRF token: `<input type="hidden" name="csrf_token" value={ csrfToken }/>`
- All POST/PUT/DELETE handlers wrapped with `middleware.CSRFProtection()`
- CSRF meta tag included in all pages for HTMX: `<meta name="csrf-token" content={ csrfToken }>`
- Tokens retrieved in handlers: `csrfToken, _ := middleware.GetCSRFToken(e)`

**Rate Limiting:**
- Login endpoint has rate limiter: `loginRateLimiter.RequestEventMiddleware()`
- Contact form has rate limiter

## Login Debugging

The login handler at `main.go:handleLoginPost()` includes logging:
- `[LOGIN] Attempt for email: <email>`
- `[LOGIN] Failed to find user <email>: <error>`
- `[LOGIN] Invalid password for user: <email>`

Check server logs for these messages when troubleshooting authentication.

## Admin User Setup
- Admin user auto-created from `.env` on first run
- Email: `PB_ADMIN_EMAIL`, Password: `PB_ADMIN_PASSWORD`
- User created with `role = "admin"` in `users` collection

## PocketBase Development Patterns

### Database Access
```go
// Query records
records, err := app.Dao().FindRecordsByFilter(
    "posts",
    "published = true && status = 'published'",
    "-created",  // sort descending
    10,          // limit
    0,           // offset
)

// Get single record
record, err := app.Dao().FindFirstRecordByFilter(
    "users",
    "email = {:email}",
    dbx.Params{"email": email},
)

// Create record
collection, _ := app.Dao().FindCollectionByNameOrId("posts")
record := models.NewRecord(collection)
record.Set("title", "My Title")
record.Set("content", "Content here")
err := app.Dao().SaveRecord(record)

// Update record
record.Set("status", "published")
err := app.Dao().SaveRecord(record)

// Delete record
err := app.Dao().DeleteRecord(record)
```

### File Uploads
```go
// Handle file upload in POST handler
file, err := e.FormFile("image")
if err != nil {
    return err
}

// Save to record
record.Set("featured_image", file.Filename)
form := forms.NewRecordUpsert(app, record)
form.AddFiles("featured_image", file)
if err := form.Submit(); err != nil {
    return err
}
```

### Migrations
- Location: `pb_migrations/`
- Format: `TIMESTAMP_description.go`
- Auto-run on server start
- Use `migrate.Collection()` API for schema changes
- Example:
```go
migrate.Collection("posts").SetFields(
    &schema.TextField{Name: "title", Required: true},
    &schema.TextField{Name: "content"},
    &schema.FileField{Name: "featured_image", MaxSelect: 1},
)
```

### Route Registration
All routes registered in `main.go`:
```go
// Public routes
app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
    e.Router.GET("/", handleHome)
    e.Router.GET("/blog", handleBlogList)
    e.Router.GET("/blog/:slug", handleBlogPost)
    
    // Protected routes with middleware
    e.Router.POST("/contact", handleContact, 
        middleware.CSRFProtection(),
        contactRateLimiter.RequestEventMiddleware())
    
    return nil
})
```

### Authentication Checks
```go
// Get authenticated user
authRecord := e.Get(apis.ContextAuthRecordKey)
if authRecord == nil {
    return apis.NewUnauthorizedError("Unauthorized", nil)
}
user := authRecord.(*models.Record)

// Check admin role
if user.GetString("role") != "admin" {
    return apis.NewForbiddenError("Admin access required", nil)
}
```

### Response Patterns
```go
// HTML response with Templ
component := templates.MyPage(csrfToken)
return component.Render(context.Background(), e.Response())

// JSON response
return e.JSON(200, map[string]interface{}{
    "success": true,
    "data": data,
})

// Redirect
return e.Redirect(302, "/success-page")

// Error handling
return apis.NewBadRequestError("Invalid input", err)
return apis.NewNotFoundError("Post not found", nil)
```

### Environment Variables
Required in `.env`:
```bash
# PocketBase Admin
PB_ADMIN_EMAIL=admin@example.com
PB_ADMIN_PASSWORD=secure_password

# Email Service
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
FROM_EMAIL=noreply@avrnpo.org
FROM_NAME=American Veterans Rebuilding
EMAIL_ENABLED=true
CONTACT_EMAIL=info@avrnpo.org

# Helcim Payment
HELCIM_API_TOKEN=your_token
HELCIM_TEST_MODE=true
```

### Testing Patterns
```go
// Setup test app
app, cleanup := setupTestApp(t)
defer cleanup()

// Create test record
collection, _ := app.Dao().FindCollectionByNameOrId("posts")
record := models.NewRecord(collection)
record.Set("title", "Test Post")
app.Dao().SaveRecord(record)

// Make HTTP request
req := httptest.NewRequest("GET", "/blog", nil)
rec := httptest.NewRecorder()
// ... test response
```

### Common Pitfalls
1. **Don't query in templates** - Fetch all data in handlers
2. **Always sanitize user input** - Use `services.SanitizeHTML()` for content
3. **Check auth before data access** - Middleware or manual checks
4. **Use transactions for multi-step operations** - `app.Dao().RunInTransaction()`
5. **Clean up test data** - Use `defer cleanup()` in tests
6. **CSRF tokens required** - All forms need CSRF protection
7. **Files are special** - Use `forms.RecordUpsert` for file uploads

### Debugging
```bash
# View logs (use tail without -f to avoid blocking)
tail /tmp/avrnpo-dev.log

# Follow logs (only when you want to monitor continuously)
tail -f /tmp/avrnpo-dev.log

# Check database directly
sqlite3 pb_data/data.db "SELECT * FROM posts;"

# Test specific handler
go test -v -run TestHandleContact

# Run with verbose PocketBase logs
LOG_LEVEL=debug ./avrnpo serve --http=0.0.0.0:8090 > /tmp/avrnpo-debug.log 2>&1 &
```

### Tool Usage Guidelines
- **Avoid blocking commands**: Do not use `tail -f` or other commands that block execution. Use `tail` without the `-f` flag to view recent logs without blocking.
- **Background processes**: When starting servers or long-running processes, use background execution (`&`) and capture output to log files.

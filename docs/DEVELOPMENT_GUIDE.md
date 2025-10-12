# American Veterans Rebuilding (AVR NPO) Development Guide

Official website for American Veterans Rebuilding, a 501(c)(3) non-profit organization dedicated to helping combat veterans rebuild their lives through housing projects, skills training, and community support programs.

## About AVR NPO

American Veterans Rebuilding is formed by Combat Veterans of the wars in Afghanistan and Iraq. We are soldiers who have lived through hell on earth and found a way to continue to dedicate our lives to the military's core values of Loyalty, Duty, Respect, Selfless Service, Honor, Integrity and Personal Courage.

## Technology Stack

### Backend
- **Go 1.23.0** - Modern, compiled language
- **PocketBase v0.22+** - Embedded backend (SQLite + Admin UI)
- **Templ v0.3.943** - Type-safe Go templates

### Frontend
- **HTMX** - Dynamic interactions without heavy JavaScript
- **Pico CSS v2** - Semantic CSS framework
- **Progressive Enhancement** - Works without JavaScript

### Services
- **Helcim** - Payment processing (one-time + recurring)
- **SMTP** - Email delivery (custom Go implementation)

## Quick Start

### Prerequisites
- Go 1.23+
- Templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`

### Setup

```bash
# Clone repository
git clone <repository-url>
cd avrnpo.org

# Copy environment template
cp .env.example .env

# Edit .env with your settings
# Required: PB_ADMIN_EMAIL, PB_ADMIN_PASSWORD

# Build and run
make dev
```

### Access Points
- **Website**: http://127.0.0.1:8090
- **Admin UI**: http://127.0.0.1:8090/_/

## Development Workflow

### Daily Development

```bash
# Start development server (with hot reload via Air)
make dev

# Run unit tests
go test ./...

# Run E2E tests
E2E_TESTS=1 go test -v -run E2E

# Regenerate templates after changes
templ generate

# Build production binary
go build
```

### Project Structure

```
avrnpo.org/
├── main.go                # Application entry point + all route handlers
├── templates/             # Templ templates (.templ files)
│   ├── base.templ        # Main layout with navigation
│   ├── helpers.templ     # Reusable components
│   └── *_templ.go        # Generated Go files (don't edit)
├── services/             # External integrations
│   ├── email.go          # SMTP email service
│   ├── helcim.go         # Payment processing
│   └── content_sanitizer.go
├── middleware/           # Security & validation
│   ├── csrf.go           # CSRF protection
│   ├── ratelimit.go      # Rate limiting
│   └── validation.go     # Input validation
├── pb_migrations/        # Database schema migrations
├── pb_data/              # SQLite database + logs (gitignored)
└── public/               # Static assets
    └── assets/
        ├── css/          # Pico CSS + custom theme
        └── js/           # HTMX, editor, theme toggle
```

## Template Development

### Template Architecture

All templates use either `Base` or `BasePage` for consistency:

```go
// Public page with full navigation
templ BlogList(csrfToken string, posts []*models.Record) {
    @Base("Blog - AVR NPO", csrfToken, blogListContent(posts))
}

templ blogListContent(posts []*models.Record) {
    // page content here
}

// Simplified page (login/admin)
templ LoginPage(csrfToken string) {
    @BasePage("Login", csrfToken, loginForm(csrfToken))
}
```

### Base Components
- `Base(title, csrfToken, content)` - Full page with navigation
- `BasePage(title, csrfToken, content)` - Simplified (admin/login)
- Both include CSRF meta tag when csrfToken provided

### Templ Best Practices

1. **Compile templates**: Run `templ generate` after changes
2. **Don't edit *_templ.go files**: Generated automatically
3. **Type safety**: Templates are compiled, errors caught at build time
4. **Component composition**: Build larger pages from smaller components

## PocketBase Patterns

### Database Access

```go
// Query multiple records
posts, err := app.Dao().FindRecordsByFilter(
    "posts",
    "status = 'published' && published = true",
    "-created",  // sort
    10,          // limit
    0,           // offset
)

// Get single record
user, err := app.Dao().FindFirstRecordByFilter(
    "users",
    "email = {:email}",
    dbx.Params{"email": email},
)

// Create record
collection, _ := app.Dao().FindCollectionByNameOrId("posts")
record := models.NewRecord(collection)
record.Set("title", "My Post")
record.Set("content", "Content here")
app.Dao().SaveRecord(record)

// Update record
record.Set("status", "published")
app.Dao().SaveRecord(record)
```

### Route Registration

All routes in `main.go`:

```go
app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
    // Public routes
    e.Router.GET("/", handleHome)
    e.Router.GET("/blog", handleBlogList)
    
    // Protected routes with middleware
    e.Router.POST("/contact", handleContact,
        middleware.CSRFProtection(),
        contactRateLimiter.RequestEventMiddleware())
    
    return nil
})
```

### Authentication

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
component := templates.BlogPost(csrfToken, post)
return component.Render(context.Background(), e.Response())

// JSON response
return e.JSON(200, map[string]interface{}{
    "success": true,
    "data": data,
})

// Redirect
return e.Redirect(302, "/success")

// Errors
return apis.NewBadRequestError("Invalid input", err)
```

## Content Management

### Blog Posts

**Managed via PocketBase Admin UI** at `/_/`

- Use Admin UI for creating/editing posts
- Templates are READ-ONLY for display
- Do NOT create custom CRUD forms

### Collections

- **users** - User accounts with role field (user/admin)
- **posts** - Blog posts with markdown content + featured images
- **donations** - Donation records from Helcim
- **subscriptions** - Recurring payment subscriptions

## Security

### CSRF Protection

```templ
// In all forms
<input type="hidden" name="csrf_token" value={ csrfToken }/>

// In handlers
csrfToken, _ := middleware.GetCSRFToken(e)

// Protect routes
e.Router.POST("/contact", handleContact,
    middleware.CSRFProtection())
```

### Rate Limiting

```go
// Create limiter
loginRateLimiter := middleware.NewRateLimiter(5, 1*time.Minute)

// Apply to route
e.Router.POST("/login", handleLoginPost,
    loginRateLimiter.RequestEventMiddleware())
```

## Styling with Pico CSS

### Design System

Use semantic HTML + Pico CSS variables:

```css
/* Theme variables in public/assets/css/custom.css */
--pico-primary: #ffb627;        /* Army gold */
--pico-secondary: #4a5d23;      /* Army green */
--pico-background-color: #e8e9ea; /* Concrete gray */
```

### Common Patterns

```html
<!-- Cards -->
<article>
  <header><h3>Title</h3></header>
  <p>Content here</p>
  <footer><button>Action</button></footer>
</article>

<!-- Forms -->
<form method="post">
  <input type="hidden" name="csrf_token" value="..."/>
  <label>
    Email
    <input type="email" name="email" required/>
  </label>
  <button type="submit">Submit</button>
</form>

<!-- Grids -->
<div class="grid">
  <div>Column 1</div>
  <div>Column 2</div>
  <div>Column 3</div>
</div>
```

### Styling Guidelines

1. **Semantic HTML first** - Use proper elements
2. **Pico CSS variables** - Don't write custom CSS rules
3. **Component classes** - Use `.blog-post-item`, `.team-card` for custom needs
4. **Theme compatible** - Test dark/light modes

## Payment System

### Helcim Integration

```go
// Initialize payment (one-time)
token, err := helcimService.InitializePayment(amount, currency)

// Create subscription (recurring)
subID, err := helcimService.CreateSubscription(userEmail, amount, cardToken)

// Process callback
result := helcimService.ProcessCallback(callbackData)
```

### Donation Flow

1. User selects amount + type (one-time/recurring)
2. Frontend: Helcim.js collects card, returns token
3. Backend: Process via Helcim API
4. Save donation record to database
5. Send receipt email via `services/email.go`

## Email System

### Configuration

```bash
# In .env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
FROM_EMAIL=donations@avrnpo.org
FROM_NAME=American Veterans Rebuilding
EMAIL_ENABLED=true
CONTACT_EMAIL=info@avrnpo.org
```

### Sending Emails

```go
emailService := services.NewEmailService()

// Donation receipt
err := emailService.SendDonationReceipt(
    donation,
    transactionID,
    subscriptionID,
)

// Contact form
err := emailService.SendContactNotification(
    name,
    email,
    message,
)
```

## Testing

### Unit Tests

```bash
# All tests
go test ./...

# Specific package
go test ./services

# Verbose
go test -v ./middleware
```

### E2E Tests

```bash
# Run E2E tests
E2E_TESTS=1 go test -v -run E2E

# Test includes:
# - Homepage rendering
# - Blog list/detail
# - Contact form submission
# - Login flow
# - Admin operations
```

### Test Patterns

```go
func TestHandler(t *testing.T) {
    app, cleanup := setupTestApp(t)
    defer cleanup()
    
    // Create test data
    collection, _ := app.Dao().FindCollectionByNameOrId("posts")
    record := models.NewRecord(collection)
    record.Set("title", "Test")
    app.Dao().SaveRecord(record)
    
    // Make request
    req := httptest.NewRequest("GET", "/blog", nil)
    rec := httptest.NewRecorder()
    
    // Test response
    // ...
}
```

## Deployment

See [Coolify Deployment Guide](./deployment/coolify-pocketbase-migration.md) for complete instructions.

### Environment Variables

```bash
# PocketBase Admin
PB_ADMIN_EMAIL=admin@avrnpo.org
PB_ADMIN_PASSWORD=secure_password

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=...
SMTP_PASSWORD=...
FROM_EMAIL=noreply@avrnpo.org
EMAIL_ENABLED=true

# Helcim
HELCIM_API_TOKEN=your_token
HELCIM_TEST_MODE=false

# Server
PORT=8090
```

### Production Build

```bash
# Build binary
go build -o avrnpo

# Run
./avrnpo serve --http=0.0.0.0:8090
```

## Troubleshooting

### Common Issues

**Templates not updating**
```bash
templ generate
make dev
```

**Database migrations not running**
- Migrations auto-run on server start
- Check `pb_data/logs/` for errors

**CSRF token errors**
- Ensure form includes `<input type="hidden" name="csrf_token".../>`
- Check middleware is applied to route

**Email not sending**
- Verify `EMAIL_ENABLED=true`
- Check SMTP credentials
- Review server logs

### Debug Logging

```bash
# Enable verbose logging
LOG_LEVEL=debug ./avrnpo serve

# View logs
tail -f pb_data/logs/data.log

# Check database
sqlite3 pb_data/data.db "SELECT * FROM users;"
```

## Resources

- **PocketBase Docs**: https://pocketbase.io/docs/
- **Templ Guide**: https://templ.guide/
- **Pico CSS**: https://picocss.com/
- **HTMX**: https://htmx.org/
- **Project Docs**: `./docs/`

## Contributing

1. Review [AGENTS.md](../AGENTS.md) for patterns
2. Run tests before committing: `go test ./...`
3. Update Templ templates: `templ generate`
4. Follow security guidelines
5. Test in light/dark modes

---

*Supporting combat veterans in rebuilding their lives and strengthening communities.*

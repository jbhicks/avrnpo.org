# avrnpo.org Development Guidelines

## Important: DO NOT Run the Server Directly

**NEVER** run `./avrnpo serve` or `make dev` directly in your shell as it will block your thread of execution and prevent you from doing anything else.

### If You Must Start the Server

If you absolutely need to start the server for testing:

1. **Use background execution with logging:**
   ```bash
   ./avrnpo serve --http=0.0.0.0:8090 > /tmp/avrnpo.log 2>&1 &
   SERVER_PID=$!
   sleep 2
   tail -f /tmp/avrnpo.log
   ```

2. **Or use `make dev` in background:**
   ```bash
   make dev > /tmp/avrnpo.log 2>&1 &
   SERVER_PID=$!
   sleep 2
   tail /tmp/avrnpo.log
   ```

3. **Always remember to kill the process when done:**
   ```bash
   kill $SERVER_PID
   # or
   pkill -f "avrnpo serve"
   ```

### Preferred Approach

**ASK THE USER** to start the server themselves. The user can run `make dev` in their own terminal and you can focus on code changes and testing.

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
make dev          # Start development server
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

# AGENTS.md - Development Guide for Agentic Coding Agents

This document provides a single, authoritative guide for AI agents working on this project. All agents must adhere to these rules.

## 🚨 CRITICAL SECURITY RULES 🚨

**NEVER EXPOSE SENSITIVE DATA IN DOCUMENTATION OR CODE**

- **NEVER** include real API keys, tokens, passwords, or secrets in ANY documentation files (.md, .txt, etc.)
- **ALWAYS** use placeholder values like `your_api_key_here` or `REPLACE_WITH_ACTUAL_KEY`.
- **ALWAYS** use environment variables for sensitive configuration.
- **IMMEDIATELY** flag and remove any exposed credentials found in files.

## 🚨 CRITICAL: PROCESS MANAGEMENT RULES 🚨

**NEVER START OR KILL LONG-RUNNING PROCESSES WITHOUT EXPLICIT USER REQUEST**

- **NEVER run `make dev`, `buffalo dev`, `./bin/app`** or any other operation that will block or run a long-lived process.
- **NEVER use `isBackground=true`** unless the user explicitly requests it.
- **NEVER kill processes** with `kill -9`, `pkill buffalo`, `kill $(lsof -t -i:3000)` unless user explicitly asks.
- **ASSUME Buffalo is already running** - it auto-reloads on ALL file changes (Go, templates, assets).
- **Agents should only run** build, test, or migration commands that complete and return control.

## 🚨 CRITICAL: HTMX AND FORM HANDLING 🚨

**Our application uses `hx-boost` for all navigation and form submissions. This is the primary and required pattern.**

- **ALWAYS** return full HTML pages from handlers. `hx-boost` will automatically extract the `<body>` content.
- **NEVER** check for the `HX-Request` header to serve different content (e.g., partial vs. full page). This is an anti-pattern in our architecture.
- **ALL** forms should have a `method` and `action` attribute for progressive enhancement. `hx-boost` will enhance them automatically.
- For specific components that need to be updated without a full page reload (e.g., a search results box), you may use explicit `hx-get` or `hx-post` with an `hx-target` that is not `body`. This is the exception, not the rule.

**✅ Correct Handler (for `hx-boost`):**
```go
func MyPageHandler(c buffalo.Context) error {
    // ... logic to fetch data ...
    // Always render the full page.
    return c.Render(http.StatusOK, r.HTML("pages/my_page.plush.html"))
}
```

## Code Style and Conventions

### Go Code Style
- Package imports: stdlib, third-party, local packages (separated by blank lines).
- Use `uuid.UUID` for IDs, `time.Time` for timestamps.
- Error handling: wrap with `github.com/pkg/errors.WithStack(err)`.
- Validation: use Buffalo's `validate.Errors` pattern.

### Template Conventions (Critical)
- **Partial Naming**: Call `partial("directory/filename")` - NO underscore, NO extension. The file itself should be named `_filename.plush.html`.
- **Plush Syntax**: Use `<%= %>` for output and `<% %>` for execution. See the [Plush documentation](https://github.com/gobuffalo/plush) for more details.

### Database & Migrations
- Use `.fizz` migrations only (NOT .sql files).
- Delete auto-generated `schema.sql` files immediately.
- Use `soda` commands, not `buffalo pop` (v0.18.14+ compatibility).

### Styling with Pico.css
- Use CSS variables (`--pico-primary`, `--pico-background-color`) for customization.
- Use semantic HTML first: `<nav>`, `<article>`, `<section>`, `role="button"`.

## Build and Test Commands

- `make build`: Run build to validate templates and catch issues after template changes.
- `make test`: Run all tests with database setup (recommended).
- `make test-fast`: Run tests without database setup (when you know the DB is running and migrated).
- `soda migrate up`: Run database migrations.

## Test-only signature bypass
- Any bypass for signature verification must be strictly limited to the test environment (`GO_ENV=test`).
- See existing tests for examples of how to use the `AttachHelcimSignature` helper or the bypass.
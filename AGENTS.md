# AGENTS.md - Development Guide for Agentic Coding Agents

## Build/Test Commands
- NEVER run `make dev`, `buffalo dev`, `./bin/app`, or any other operation that will block or run a long-lived process. These commands are for manual use only and should not be executed by agents. Agents should only run build, test, or migration commands that complete and return control. Starting the server (development or production) must always be done manually by a developer.
- Agents **should** test direct asset links (e.g., `/css/custom.css`, `/js/application.js`, `/images/logo.avif`, `/favicon.svg`) using HTTP requests to verify asset serving, as long as the server is already running.
- `make test` - Run all tests with database setup (recommended)
- `make test-fast` - Run tests without database setup
- `soda migrate up` - Run database migrations (NOT `buffalo pop migrate`)
- `make db-reset` - Reset database (drop, create, migrate)

## Code Style Guidelines

### Go Code Style
- Package imports: stdlib, third-party, local packages (separated by blank lines)
- Use `uuid.UUID` for IDs, `time.Time` for timestamps
- Error handling: wrap with `github.com/pkg/errors.WithStack(err)`
- Validation: use Buffalo's `validate.Errors` pattern
- Database: use Pop ORM with `*pop.Connection` parameter

### Template Conventions (Critical)
- **Partial naming**: Call `partial("directory/filename")` - NO underscore, NO extension
- **Partial files**: Named `_filename.plush.html` with single underscore prefix
- **HTMX patterns**: Use `hx-boost="true"` for navigation, return full pages
- **Template structure**: Include complete HTML (nav, main, footer) for HTMX compatibility
- **Form Layout with Pico CSS**: 
  - NEVER put form labels/inputs/errors directly in grid - this breaks layout completely
  - ALWAYS wrap related form fields in div containers within grids: `<div class="grid"><div><label><input><error></div></div>`
  - Use selective grids: group related fields (First/Last Name, City/State/ZIP) but keep others full-width
  - Template error handling: Use `<%= err %>` not `<%= err.Message %>` in error loops
- **Form Validation**:
  - ALWAYS use Buffalo's server-side validation instead of client-side alerts
  - Forms must have proper `action` attribute pointing to correct handler endpoint
  - Use `novalidate` attribute to disable browser validation and rely on server validation
  - Handler functions should detect API vs form requests and respond appropriately (JSON vs template rendering)

### Database & Migrations
- Use `.fizz` migrations only (NOT .sql files)
- Delete auto-generated `schema.sql` files immediately
- Use `soda` commands, not `buffalo pop` (v0.18.14+ compatibility)

### Styling with Pico.css
- Use CSS variables (`--pico-primary`, `--pico-background-color`) for customization
- Semantic HTML first: `<nav>`, `<article>`, `<section>`, `role="button"`
- Check `/docs/pico-css-variables.md` before any styling changes

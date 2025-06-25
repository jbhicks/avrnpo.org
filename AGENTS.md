# AGENTS.md - Development Guide for Agentic Coding Agents

## Build/Test Commands
- `make dev` - Start database and Buffalo development server
- `make test` - Run all tests with database setup (recommended)
- `make test-fast` - Run tests without database setup
- `buffalo test` - Run Buffalo test suite (NEVER use `go test` directly)
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

### Database & Migrations
- Use `.fizz` migrations only (NOT .sql files)
- Delete auto-generated `schema.sql` files immediately
- Use `soda` commands, not `buffalo pop` (v0.18.14+ compatibility)

### Styling with Pico.css
- Use CSS variables (`--pico-primary`, `--pico-background-color`) for customization
- Semantic HTML first: `<nav>`, `<article>`, `<section>`, `role="button"`
- Check `/docs/pico-css-variables.md` before any styling changes
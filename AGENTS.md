# AGENTS.md - Development Guide for Agentic Coding Agents

## 🚨 CRITICAL: PROCESS MANAGEMENT RULES 🚨

**NEVER START OR KILL LONG-RUNNING PROCESSES WITHOUT EXPLICIT USER REQUEST**

- **NEVER run `make dev`, `buffalo dev`, `./bin/app`** or any other operation that will block or run a long-lived process
- **NEVER use `isBackground=true`** unless the user explicitly requests it
- **NEVER kill processes** with `kill -9`, `pkill buffalo`, `kill $(lsof -t -i:3000)` unless user explicitly asks
- **ASSUME Buffalo is already running** - it auto-reloads on ALL file changes (Go, templates, assets)
- **These commands are for manual use only** and should not be executed by agents
- **Agents should only run** build, test, or migration commands that complete and return control
- **Starting/stopping servers** (development or production) must always be done manually by a developer

**🚨 FORBIDDEN COMMANDS (unless user explicitly requests):**
- `make dev` (starts long-running server)
- `buffalo dev` (starts long-running server)  
- `npm start` (starts long-running server)
- Any command that starts a background server or daemon process
- Any command with `isBackground=true` unless user specifically asks for it

## Build/Test Commands
- Agents **should** test direct asset links (e.g., `/css/custom.css`, `/js/application.js`, `/images/logo.avif`, `/favicon.svg`) using HTTP requests to verify asset serving, as long as the server is already running.
- `make build` - Run build to validate templates and catch issues after template changes
- `make test` - Run all tests with database setup (recommended)
- `make test-fast` - Run tests without database setup
- `soda migrate up` - Run database migrations (NOT `buffalo pop migrate`)
- `make db-reset` - Reset database (drop, create, migrate)


## Test-only signature bypass: policy and safe-guards

Rationale
- Webhook signature verification uses secrets and precise hashing timing; this makes integration tests brittle if they rely on production signing flows. To keep integration tests focused on handler behavior, tests must have a deterministic way to provide valid signatures or a safe bypass.

Policy
- Any bypass for signature verification must be strictly limited to the test environment.
  - Detect test environment via GO_ENV == "test" or a dedicated env var such as HELCIM_TEST_BYPASS.
  - Never allow bypass to be set in non-test environments.
- Prefer generating valid signatures in tests using the real signing algorithm and a test-only secret:
  - Use AttachHelcimSignature helper which uses HELCIM_SECRET from environment (set in CI/test runs).
- Only use the bypass for high-level handler tests where signature verification is orthogonal to the behavior being tested.
- Always include unit tests that exercise the verification logic (positive and negative cases) so bypassing does not hide verification regressions.

Implementation guidelines
- In signature verification code:
  - Add a clear, minimal guard:
    - if os.Getenv("GO_ENV") == "test" && os.Getenv("HEL_CIM_TEST_BYPASS") == "true" { return true }
  - Keep this guard short and obvious; add a comment linking to AGENTS.md and the test helper.
- CI setup:
  - By default, prefer using HELCIM_SECRET in CI and generate valid signatures for webhook tests.
  - If bypass is used, document why and for which tests in the commit message and feature tracking docs.

Audit and review
- Any PR that introduces or modifies a test bypass must:
  - Include rationale in the PR description.
  - Add unit tests for the verification function to demonstrate coverage.
  - Add a line in CURRENT_FEATURE.md noting the bypass and its acceptance criteria.


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

### Plush template syntax (agent guidance)

- Output vs code:
  - `<%= expression %>` — evaluate expression and insert the result into the template.
  - `<% statement %>` — execute code without inserting output.
  - `<%# comment %>` — template comment, not rendered.
- Variables and context:
  - Declare with `let`: `<% let x = 1 %>`.
  - Set values from Go via `ctx.Set("name", value)` and access in templates with `<%= name %>`.
  - Maps and arrays map to `map[string]interface{}` and `[]interface{}` in Go.
- Control flow and output:
  - Use `if/else` and `for` constructs. When controlling HTML output, wrap the flow with `<%= if (...) { %> ... <% } %>` so the HTML is emitted correctly.
  - For loops: `<%= for (key, val) in expr { %> ... <% } %>`; iterators must implement `Next()` in Go.
- Helpers and blocks:
  - Register helpers in Go with `ctx.Set("name", fn)` and call them in templates.
  - Block helpers accept a `HelperContext` to capture a template block; use `help.Block()` inside the helper to render it.
- Partials:
  - Call `partial("dir/name")` to render `templates/dir/_name.plush.html`.
  - Do NOT include underscore or extension in `partial()` call.
- Safety and best practices:
  - Avoid putting business logic in templates; prefer helpers and Go.
  - Return sanitized content. When returning `template.HTML`, ensure content is safe to avoid XSS.
  - Use `<%= %>` intentionally — missing it silently suppresses output.


### Database & Migrations
- Use `.fizz` migrations only (NOT .sql files)
- Delete auto-generated `schema.sql` files immediately
- Use `soda` commands, not `buffalo pop` (v0.18.14+ compatibility)

### Styling with Pico.css
- Use CSS variables (`--pico-primary`, `--pico-background-color`) for customization
- Semantic HTML first: `<nav>`, `<article>`, `<section>`, `role="button"`
- Check `/docs/pico-css-variables.md` before any styling changes

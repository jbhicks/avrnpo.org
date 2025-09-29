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

## 🚨 CRITICAL: BUFFALO + PLUSH TEMPLATES 🚨

**Our application uses standard Buffalo patterns with Plush templates and progressive JavaScript enhancement.**

- **ALWAYS** return full HTML pages from handlers using Buffalo's standard render patterns.
- **NEVER** check for AJAX/HTMX headers to serve different content. Use single-template architecture.
- **ALL** forms should use standard HTML `method` and `action` attributes.
- **PROGRESSIVE ENHANCEMENT**: Add JavaScript functionality that enhances the basic HTML experience.
- **ONE TEMPLATE PER VIEW**: Avoid duplicate partial/full template patterns.

**✅ Correct Handler Pattern:**
```go
func MyPageHandler(c buffalo.Context) error {
    // ... logic to fetch data ...
    // Always render the full page with standard Buffalo patterns
    return c.Render(http.StatusOK, r.HTML("pages/my_page.plush.html"))
}
```

**✅ Correct Form Pattern:**
```html
<form method="POST" action="/submit">
    <%= csrf() %>
    <!-- form fields -->
    <button type="submit">Submit</button>
</form>
```

## Code Style and Conventions

### Go Code Style
- Package imports: stdlib, third-party, local packages (separated by blank lines).
- Use `uuid.UUID` for IDs, `time.Time` for timestamps.
- Error handling: wrap with `github.com/pkg/errors.WithStack(err)`.
- Validation: use Buffalo's `validate.Errors` pattern.

### Template Conventions (Critical)
- **Single Template Per View**: Use one template per route. Avoid duplicate partial/full template patterns.
- **Partial Naming**: Call `partial("directory/filename")` - NO underscore, NO extension. The file itself should be named `_filename.plush.html`.
- **Plush Syntax**: Use `<%= %>` for output and `<% %>` for execution. See the [Plush documentation](https://github.com/gobuffalo/plush) for more details.
- **Progressive Enhancement**: Build templates that work without JavaScript, then enhance with JS.

### JavaScript Enhancement Patterns
- **Vanilla JavaScript Preferred**: Use modern JavaScript features (fetch, DOM APIs) over libraries.
- **Progressive Enhancement**: Ensure base functionality works without JavaScript.
- **Form Enhancement**: Use `fetch()` API to enhance form submissions while maintaining fallback behavior.
- **Event Delegation**: Use event delegation for dynamically added content.

**✅ Example Progressive Form Enhancement:**
```javascript
// Enhance forms without breaking basic functionality
document.addEventListener('submit', async (e) => {
    if (e.target.dataset.enhance === 'true') {
        e.preventDefault();
        const response = await fetch(e.target.action, {
            method: e.target.method,
            body: new FormData(e.target)
        });
        // Handle response and update UI
    }
});
```

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

## 🚨 CRITICAL: DEPLOYMENT WORKFLOW 🚨

**ALWAYS CONFIRM WITH USER BEFORE PUSHING TO PRODUCTION**

- **NEVER** push commits directly to production without explicit user confirmation
- **ALWAYS** ask the user to confirm locally before pushing changes
- **ALWAYS** provide a summary of what will be deployed and ask for approval
- **ONLY** push after receiving explicit confirmation from the user

## 📋 WORK TRACKING & DOCUMENTATION WORKFLOW

**ALWAYS USE DOCUMENTATION TO TRACK WORK IN PROGRESS AND COMPLETION**

### Current Feature Tracking
- **UPDATE** `docs/development/current-feature.md` when starting work on major features
- **DOCUMENT** progress, blockers, and next steps in real-time
- **MARK COMPLETED** tasks with ✅ checkboxes in the document
- **REFERENCE** specific file locations (e.g., `actions/donations.go:45`) for context

### Development Planning Documentation
- **ROADMAP PLANNING**: Use `docs/development/donation-system-roadmap.md` for long-term planning
- **REFACTORING WORK**: Document architectural changes in `docs/development/refactoring-plan.md`
- **BUG TRACKING**: Add debugging notes to `docs/development/css-debug.md` or relevant files

### Implementation History
- **DOCUMENT MAJOR CHANGES**: Add implementation summaries to `docs/changelog/` when completing significant features
- **MIGRATION NOTES**: Include upgrade paths and breaking changes
- **DECISION RATIONALE**: Explain why certain approaches were chosen

### Work Session Pattern
1. **START**: Check `docs/development/current-feature.md` for current priorities
2. **PLAN**: Update the document with your intended approach
3. **IMPLEMENT**: Make code changes, referencing documentation
4. **DOCUMENT**: Update progress and mark completed items
5. **COMPLETE**: Move to changelog if it's a major feature completion

### Status Indicators in Documentation
Use consistent status markers:
- ✅ **Completed** - Feature/task finished and tested
- 🔄 **In Progress** - Currently working on this
- ⚠️ **Blocked** - Needs user input or external dependency
- 📋 **Planned** - Identified for future work

### Example Current Feature Entry
```markdown
## 🔄 Payment Form Validation (In Progress)
- ✅ Add server-side validation rules in `actions/donations.go:85`
- 🔄 Implement client-side progressive enhancement 
- 📋 Add comprehensive error messaging
- ⚠️ Waiting for Helcim test credentials from user
```

### Documentation-First Development
- **BEFORE CODING**: Document the plan in appropriate files
- **DURING CODING**: Update progress and reference specific code locations  
- **AFTER CODING**: Mark completion and document lessons learned
- **NEVER** leave documentation outdated - it's the project's memory

## Test-only signature bypass
- Any bypass for signature verification must be strictly limited to the test environment (`GO_ENV=test`).
- See existing tests for examples of how to use the `AttachHelcimSignature` helper or the bypass.
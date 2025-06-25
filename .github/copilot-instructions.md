# Copilot Instructions

## 🚨 CRITICAL PROCESS MANAGEMENT RULES 🚨

**NEVER KILL THE BUFFALO DEVELOPMENT SERVER PROCESS**

- **Buffalo automatically reloads** on ALL file changes (Go code, templates, assets)
- **DO NOT run `kill -9`, `pkill buffalo`, `kill $(lsof -t -i:3000)` or similar commands**
- **DO NOT restart Buffalo** unless there are compilation errors or the user explicitly asks
- **Assume Buffalo is running and working** - it should stay running throughout development
- **Let Buffalo handle recompilation** - it's designed to auto-reload everything
- **Only check processes** - don't kill them: `ps aux | grep buffalo` or `lsof -i :3000`

**When Buffalo is running properly:**
- ✅ Go code changes trigger automatic recompilation
- ✅ Template changes reload immediately 
- ✅ Static asset changes update automatically
- ✅ Database migration commands work while Buffalo runs
- ✅ Just refresh the browser to see changes

## 🎯 CURRENT DEVELOPMENT FOCUS 🎯

**American Veterans Rebuilding (AVR NPO) Donation System Improvements**

This project is currently focused on improving the donation flow for the AVR NPO website. Key areas of development:

- **Donation Page Enhancement** - Improving user experience and conversion rates
- **Payment Processing Integration** - Secure payment gateway implementation
- **Donor Management** - Backend systems for donation tracking and receipts
- **Compliance & Security** - 501(c)(3) requirements and PCI compliance considerations

**Reference Documentation:**
- Check `/docs/donation-system-roadmap.md` for detailed implementation plans
- Follow AVR-specific content guidelines when working with donation-related features
- Maintain the organization's mission focus in all donation flow improvements

**Important Context:**
- This is a real non-profit serving combat veterans
- Donation functionality directly impacts the organization's ability to help veterans
- All changes should prioritize security, usability, and donor trust

## 🚨 CRITICAL: Buffalo Template Partial Naming Convention 🚨

**⚠️ THIS IS THE #1 SOURCE OF RECURRING TEMPLATE ERRORS - READ CAREFULLY**

Buffalo automatically adds an underscore prefix to partial filenames. This causes double underscore issues if not handled correctly:

**🚨 CRITICAL RULE: Never include underscores or extensions in partial() calls 🚨**

**✅ CORRECT Pattern:**
```html
<!-- Call partial WITHOUT underscore or extension -->
<%= partial("auth/new") %>
<!-- Buffalo automatically looks for: templates/auth/_new.plush.html -->
```

**❌ WRONG Pattern (causes double underscore error):**
```html
<!-- DON'T include underscore - causes Buffalo to look for __new.plush.html -->
<%= partial("auth/_new.plush.html") %>
```

**How Buffalo Partial Resolution Works:**
1. You call: `partial("directory/filename")`
2. Buffalo automatically looks for: `templates/directory/_filename.plush.html`
3. If you include an underscore, Buffalo looks for: `templates/directory/__filename.plush.html` (FAILS)

**File Naming Convention:**
- Partial files: `_filename.plush.html` (single underscore prefix)
- Partial calls: `partial("directory/filename")` (no underscore, no extension)

**🚨 NEVER DO THESE:**
- `partial("auth/_new.plush.html")` → looks for `__new.plush.html` (double underscore)
- `partial("auth/_new")` → looks for `__new.plush.html` (double underscore)  
- `partial("auth/new.plush.html")` → looks for `_new.plush.html.plush.html` (wrong extension)

**✅ ALWAYS DO THIS:**
- `partial("auth/new")` → correctly finds `_new.plush.html`
- `partial("pages/contact")` → correctly finds `_contact.plush.html`
- `partial("admin/nav")` → correctly finds `_nav.plush.html`

**When You See "could not find template" Errors:**
1. Check that partial file exists with single underscore: `_filename.plush.html`
2. Check that partial call has NO underscore: `partial("directory/filename")`
3. Check that partial call has NO extension: `partial("directory/filename")`

This rule prevents the recurring double underscore template errors that keep appearing in tests and development.

### Pico.css Styling Guidelines

**CRITICAL: Always use Pico.css variables instead of custom CSS**

- **For ALL styling changes**: Consult `/docs/pico-css-variables.md` and `/docs/pico-implementation-guide.md` FIRST
- **Use CSS variables**: Modify `--pico-primary`, `--pico-background-color`, etc. instead of writing custom CSS
- **Follow Pico patterns**: Use semantic HTML with Pico's built-in classes and roles
- **Never override Pico directly**: Always work within Pico's variable system for customization
- **Check `/docs/` first**: All styling requests should reference the Pico documentation in `/docs/`

## Buffalo Development Environment Guidelines

### Buffalo Development Server
- **Buffalo runs on port 3000** and automatically reloads on file changes
- **🚨 NEVER KILL THE BUFFALO PROCESS 🚨** when testing changes - it has hot reload built-in
- **🚨 DO NOT RUN `kill -9`, `pkill buffalo`, or similar commands 🚨** - Buffalo should stay running
- **🚨 DO NOT RESTART Buffalo unless explicitly asked by the user 🚨**
- **Check for existing Buffalo instances** before starting a new one:
  - Use `ps aux | grep buffalo` or `lsof -i :3000` to check for running instances
  - Buffalo dev server should be left running in a background terminal
  - Changes to Go files, templates, and assets will auto-reload automatically
- **Buffalo automatically handles**:
  - Go code changes (recompiles and restarts the process)
  - Template changes (reloads templates)
  - Static asset changes (updates assets)
  - Database schema changes (when migrations are run)

### Development Workflow
1. **Start once**: Use `make dev` to start PostgreSQL + Buffalo
2. **Keep running**: Leave Buffalo running in the background - DO NOT STOP IT
3. **Make changes**: Edit files and let Buffalo auto-reload - NO MANUAL RESTARTS NEEDED
4. **Test changes**: Refresh browser or use the running instance
5. **Only restart if**: There are compilation errors that prevent auto-reload, or you need to reset the database, OR the user explicitly asks you to restart
6. **🚨 CRITICAL**: Assume Buffalo is running and working unless proven otherwise

### Database Management
- **PostgreSQL**: Runs in a Podman container on port 5432
- **Use `podman-compose ps`** to check container status
- **Database persists** between restarts via Docker volumes
- **Migrations**: Use `soda migrate up` for running migrations (NOT `buffalo pop migrate`)

**🚨 CRITICAL: Use `soda` for database operations, NOT `buffalo pop` 🚨**

Buffalo v0.18.14+ does not include the `pop` plugin. Use these commands:
- `soda migrate up` - Run pending migrations
- `soda reset` - Reset database (drop, create, migrate)
- `GO_ENV=test soda reset` - Reset test database
- `soda create -a` - Create all databases
- `soda generate migration create_posts` - Create new migration

**Legacy Documentation Warning**: Older Buffalo docs reference `buffalo pop` commands, but these don't work in v0.18.14+.

**🚨 CRITICAL: AVOID .SQL FILES IN BUFFALO DEVELOPMENT 🚨**

**✅ Use .fizz migrations instead of .sql files:**
- **Use `.fizz` files** for all database schema changes (e.g., `20250608120000_create_donations.up.fizz`)
- **Cross-database compatible** - Works with PostgreSQL, MySQL, SQLite
- **Automatic rollbacks** - `.up.fizz` and `.down.fizz` files provide safe migrations
- **Buffalo ecosystem integration** - Works seamlessly with `soda` commands
- **Version controlled** - Each migration is timestamped and tracked

**❌ NEVER manually edit these auto-generated files:**
- **`db/schema.sql`** - Auto-generated by `pg_dump`, should not be manually edited
- **`migrations/schema.sql`** - Also auto-generated, delete if it causes issues
- These files can contain database-specific settings that break in different environments

**🚨 CRITICAL: Always delete auto-generated schema.sql files immediately:**
```bash
# Remove problematic auto-generated SQL files
rm -f db/schema.sql migrations/schema.sql
```

**🚨 Common Issue**: If you see errors like "unrecognized configuration parameter", check for problematic settings in schema.sql files like `SET transaction_timeout = 0;` which don't work in all PostgreSQL versions.

**When to use .sql files (rare):**
- Complex stored procedures (uncommon in Buffalo apps)
- One-time data imports/exports
- Database maintenance scripts outside Buffalo

**For everything else, use Buffalo's .fizz migrations and Pop/Soda ORM.**

### Testing Changes
- **Templates**: Auto-reload on save, just refresh the browser - NO RESTART NEEDED
- **Go code**: Auto-compiles and restarts Buffalo server automatically
- **Static assets**: Auto-reload via Buffalo's asset pipeline
- **Database changes**: Require migration runs but Buffalo stays running
- **🚨 IMPORTANT**: Let Buffalo handle all reloading - manual intervention not needed

### Buffalo Testing Guidelines

**🚨 CRITICAL: ALWAYS USE BUFFALO TESTING COMMANDS 🚨**

- **NEVER use `go test` directly** - Buffalo has its own testing workflow
- **ALWAYS use `buffalo test`** to run tests in Buffalo applications
- **Read `/docs/` folder** for Buffalo-specific testing patterns and best practices
- **Follow Buffalo suite patterns** as documented in `/docs/buffalo/auth-and-testing-patterns.md`

**Proper Buffalo Testing Commands:**
- `buffalo test` - Run all tests directly
- `make test` - Run comprehensive test suite with database setup (recommended)
- `make test-fast` - Run tests quickly (assumes database is already running)
- `buffalo test --timeout=60s` - Run tests with timeout
- `buffalo test -v` - Run tests with verbose output

**Buffalo Testing Best Practices:**
- Always consult `/docs/buffalo/auth-and-testing-patterns.md` for authentication testing patterns
- Use Buffalo's ActionSuite for HTTP endpoint testing
- Follow Buffalo's middleware testing patterns
- Use Buffalo's database transaction handling for test isolation
- Leverage Buffalo's built-in test helpers and fixtures

**Documentation Requirements:**
- Always check `/docs/` folder before implementing tests
- Follow patterns documented in Buffalo testing guides
- Reference `/docs/buffalo/development-workflow.md` for testing workflow

### Dependency and Technology Guidelines

**🚨 CRITICAL: STRICT DEPENDENCY REQUIREMENTS 🚨**

**ALWAYS follow the dependency guidelines in `/docs/dependency-guidelines.md` before adding ANY new dependencies:**

- **Go-Only**: Only use Go modules and libraries - NO Node.js, Python, PHP, Ruby, or other language dependencies
- **Open Source Only**: NEVER use commercial, SaaS, or corporate solutions (e.g., Strapi, Contentful, WordPress)  
- **No External Services**: NEVER integrate commercial APIs or third-party services requiring paid plans
- **Buffalo Ecosystem**: Prefer Buffalo-compatible modules and official Buffalo plugins
- **Database-First**: Use Buffalo's built-in Pop/Soda ORM instead of external CMSs or headless solutions
- **Self-Contained**: All functionality must be implemented within the Go application

**Before adding any dependency:**
1. **Read `/docs/dependency-guidelines.md`** - Check all requirements and restrictions
2. **Verify it's Go-only** - No JavaScript/Node.js, Python, or other language requirements
3. **Confirm open source** - Check license and ensure no commercial restrictions
4. **Test compatibility** - Ensure it works with Buffalo and our current stack
5. **Document the choice** - Add rationale to appropriate documentation

**Forbidden Technologies:**
- Content Management Systems (Strapi, WordPress, Drupal, Contentful, etc.)
- Node.js/JavaScript backends or build tools (except for frontend assets)
- Python/Django applications or services
- PHP applications or frameworks
- SaaS APIs requiring paid subscriptions
- Docker images that aren't pure Go applications

**For CMS-like functionality**: Use Buffalo's built-in database operations with Pop/Soda ORM instead of external CMS solutions.

### Common Commands
- `make dev` - Start everything (use once)
- `make test` - Run comprehensive test suite (recommended command)
- `make test-fast` - Run tests quickly (assumes database running)
- `buffalo test` - Run all tests directly (NEVER use `go test` directly)
- `buffalo test -v` - Run tests with verbose output
- `podman-compose ps` - Check database status
- `soda migrate up` - Run new database migrations
- `ps aux | grep buffalo` - Check for running Buffalo instances
- `lsof -i :3000` - See what's using port 3000

### Troubleshooting
- **Port 3000 in use**: Check if Buffalo is already running before starting new instance
- **Database connection issues**: Check `podman-compose ps` and container logs
- **Template errors**: Check Buffalo console output for Plush syntax errors
- **Hot reload not working**: Restart Buffalo only if auto-reload stops working

### HTMX Development Guidelines

**🚨 CRITICAL: ALWAYS follow HTMX best practices documented in `/docs/htmx-best-practices.md` 🚨**

**Navigation and Progressive Enhancement:**
- **Use `hx-boost="true"`** for navigation links instead of explicit `hx-get`/`hx-target` attributes
- **Return full pages** from handlers - let HTMX extract content automatically
- **Avoid `HX-Request` header checks** in handlers - serve the same full page for both direct and HTMX requests
- **Progressive enhancement first** - ensure all functionality works without JavaScript

**Required Reading Before Any HTMX Work:**
- **ALWAYS read `/docs/htmx-best-practices.md`** before implementing HTMX features
- **Follow official HTMX patterns** documented in the best practices guide
- **Use `hx-boost` for navigation** rather than manual HTMX attributes on every link

**Development and Testing:**
- **Content loaded via HTMX**: Changes to templates auto-reload with Buffalo
- **JavaScript changes**: May require browser hard refresh (Ctrl+F5)
- **Test both scenarios**: Direct page load vs HTMX navigation must both work correctly
- **Progressive enhancement**: All features must work without JavaScript enabled

**Template Structure Guidelines:**
- **Full page templates**: Include complete HTML structure (nav, main, footer)
- **HTMX boost navigation**: Uses full pages and extracts content automatically
- **Avoid nested main elements**: Let HTMX handle content swapping correctly
- **Semantic HTML**: Use proper HTML5 elements for better HTMX extraction

### Browser Testing Guidelines

**🚨 CRITICAL: DO NOT TEST PROTECTED PAGES IN BROWSER WITHOUT LOGIN 🚨**

- **NEVER use `open_simple_browser` for protected routes** like `/account`, `/profile`, `/dashboard`, `/admin/*`
- **Protected pages require authentication** - opening them just shows the login page, not the actual functionality
- **Use tests instead**: Run `buffalo test` to verify protected page functionality
- **For public pages only**: Use browser for `/`, `/blog`, `/auth/new`, `/users/new` (login/signup)
- **Testing protected functionality**: 
  - Use Buffalo tests with authenticated users
  - Test HTMX behavior through automated tests
  - Verify template rendering through test assertions

**🚨 CRITICAL: ALWAYS VERIFY PAGES WORK PROPERLY 🚨**

- **NEVER assume a page works** just because the browser opens or a tool call succeeds
- **Always check for Buffalo 500 error pages** - they look like regular pages but contain error details
- **Immediately check Buffalo logs** when any page shows unexpected behavior or errors
- **Verify HTTP status codes** with `curl -s -I` before assuming success
- **Look at actual page content** for error messages, don't just check if browser opens
- **Check server logs immediately** when encountering issues: `tail -20 buffalo.log` or similar
- **Validate template partial references** - ensure underscore-prefixed partials exist and are referenced correctly
- **Test both direct loads and HTMX navigation** for each page to ensure both work properly

**Protected Routes to AVOID in Browser:**
- `/account` - Account settings (requires login)
- `/profile` - Profile settings (requires login) 
- `/dashboard` - User dashboard (requires login)
- `/admin/*` - Admin pages (requires admin role)
- Any HTMX endpoint for authenticated content

**Safe Public Routes for Browser Testing:**
- `/` - Home/landing page
- `/blog` - Public blog listing
- `/auth/new` - Login page
- `/users/new` - Registration page
- `/blog/[slug]` - Individual blog posts (if public)

## Project-Specific Notes

- This is a Go Buffalo SaaS template project
- Templates use Plush templating engine (.plush.html files)
- Styling is handled through Pico.css - a semantic CSS framework with automatic theming
- Custom styles can be added with CSS variables to maintain Pico.css design consistency
- Dark/light mode switching is built-in with localStorage persistence

## Implementation Guidelines

**CRITICAL: For ALL styling changes, always check `/docs/` folder FIRST**

**IMPORTANT**: Always refer to the documentation in `/docs/` folder for implementation strategies:

- **Pico.css CSS Variables**: Read `/docs/pico-css-variables.md` for customization with CSS variables - USE THIS FOR ALL STYLING
- **Implementation Patterns**: Read `/docs/pico-implementation-guide.md` for semantic HTML patterns and best practices
- **Template Syntax**: Read `/docs/buffalo-template-syntax.md` for Plush templating patterns
- **HTMX Best Practices**: Read `/docs/htmx-best-practices.md` for HTMX navigation, progressive enhancement, and official patterns

**Styling Change Process:**
1. **Check `/docs/pico-css-variables.md`** - Find the appropriate Pico variable to modify
2. **Check `/docs/pico-implementation-guide.md`** - Use semantic HTML patterns instead of custom CSS
3. **Use CSS variables only** - Modify `--pico-*` variables, never write custom CSS rules
4. **Test in both themes** - Ensure changes work in light and dark modes

### Key Implementation Strategies

1. **Semantic HTML First**: Use proper HTML elements (`<nav>`, `<article>`, `<section>`, `<details>`)
2. **Minimal CSS Classes**: Prefer `role="button"`, `class="secondary"`, `class="dropdown"` over custom styles
3. **CSS Variables for Customization**: Use `--pico-primary`, `--pico-background-color`, etc. instead of hardcoded values
4. **Theme Support**: Always test both light and dark modes using `[data-theme="dark"]` selectors
5. **Responsive by Default**: Trust Pico.css responsive behavior, avoid custom breakpoints unless necessary
6. **Documentation First**: Always consult `/docs/` before making any styling changes

### Authentication Patterns

- Use `<details class="dropdown">` for user menus instead of JavaScript dropdowns
- Implement theme switching with `localStorage.setItem('picoPreferredColorScheme', theme)`
- Style CTAs with `role="button"` and appropriate classes (`secondary`, `contrast`, `outline`)

### Anti-Patterns to Avoid

- Don't use utility classes like Tailwind CSS (`bg-blue-500`, `text-white`, etc.)
- Don't override Pico.css with excessive inline styles
- Don't use Alpine.js or JavaScript for basic interactions that Pico.css handles
- Don't hardcode colors - use CSS variables for theme compatibility
- Don't write custom CSS without first checking if Pico variables can achieve the same result

### Documentation and Communication Guidelines

### Tone and Language Requirements

**ALWAYS maintain a factual, matter-of-fact tone in all documentation and communication:**

- **Avoid promotional language**: Never use words like "comprehensive", "professional", "robust", "powerful", "seamless", "cutting-edge"
- **Avoid exaggerated claims**: Don't claim features are "production-ready", "enterprise-grade", or "industry-standard" unless verified
- **Be specific about functionality**: Instead of "complete CRUD operations", say "basic CRUD operations" or list specific functions
- **Avoid marketing speak**: Don't use phrases like "enhances perceived performance" - just state what it does
- **Remove unnecessary qualifiers**: Instead of "secure session management", just say "session management"
- **State actual capabilities**: Only document features that actually exist and work

### Documentation Best Practices

- **Feature descriptions**: Describe what the feature actually does, not how amazing it is
- **Technical accuracy**: Only claim technical capabilities that are implemented and tested
- **Realistic scope**: Don't oversell the template's capabilities or production readiness
- **Clear limitations**: Be honest about what's missing or needs work
- **Simple language**: Use clear, direct language without unnecessary adjectives

### Examples of Good vs. Bad Documentation

**❌ Bad (Overly promotional):**
> "Comprehensive role-based admin management system with full CRUD operations and professional dashboard"

**✅ Good (Factual):**
> "Role-based admin management system with basic CRUD operations and admin dashboard"

**❌ Bad (Exaggerated claims):**
> "Production-ready foundation with enterprise-grade security and scalable architecture"

**✅ Good (Honest scope):**
> "Template with basic authentication, role-based access control, and admin panel"

### When Writing or Updating Documentation

1. **Focus on functionality**: What does it actually do?
2. **Avoid superlatives**: Remove words like "best", "most", "ultimate"
3. **Be concrete**: Use specific technical terms rather than vague descriptors
4. **Test claims**: Only document features you can verify work
5. **Keep it simple**: Straightforward language is more trustworthy

## General Guidelines

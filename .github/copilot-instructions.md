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

## General Guidelines

- Never attempt to read `pico.min.css` files - they are minified and will only cause failures
- Trust that Pico.css is properly installed and working in the project
- Focus on using semantic HTML with minimal CSS classes - Pico.css provides styling automatically
- Use semantic HTML elements and follow modern web development practices with accessibility in mind

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

### HTMX Development Notes
- **Content loaded via HTMX**: Changes to partial templates auto-reload
- **JavaScript changes**: May require browser hard refresh (Ctrl+F5)
- **Modal forms**: Test by triggering modals, don't assume full page reload needed
- **🚨 CRITICAL: HTMX Template Structure**: Partial templates loaded into `#htmx-content` should NOT include `<main class="container">` wrapper since the target div already has this structure. This prevents nested main elements and rendering issues.

**HTMX Template Best Practices:**
- **Full page templates** (for direct loads): Include complete HTML with nav, main, footer
- **Partial templates** (for HTMX): Only include content sections without main wrapper
- **Avoid nested main elements**: Partial templates go inside existing `<main id="htmx-content">`
- **Test both scenarios**: Direct page load vs HTMX navigation should both work correctly

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

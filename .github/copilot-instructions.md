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
- **Migrations**: Run `buffalo pop migrate` only when adding new migrations

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

### Common Commands
- `make dev` - Start everything (use once)
- `make test` - Run comprehensive test suite (recommended command)
- `make test-fast` - Run tests quickly (assumes database running)
- `buffalo test` - Run all tests directly (NEVER use `go test` directly)
- `buffalo test -v` - Run tests with verbose output
- `podman-compose ps` - Check database status
- `buffalo pop migrate` - Run new database migrations
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

## Project-Specific Notes

- This is a Go Buffalo SaaS template project
- Templates use Plush templating engine (.plush.html files)
- Styling is handled through Pico.css - a semantic CSS framework with automatic theming
- Custom styles can be added with CSS variables to maintain Pico.css design consistency
- Dark/light mode switching is built-in with localStorage persistence

## Implementation Guidelines

**IMPORTANT**: Always refer to the documentation in `/docs/` folder for implementation strategies:

- **Pico.css CSS Variables**: Read `/docs/pico-css-variables.md` for customization with CSS variables
- **Implementation Patterns**: Read `/docs/pico-implementation-guide.md` for semantic HTML patterns and best practices
- **Template Syntax**: Read `/docs/buffalo-template-syntax.md` for Plush templating patterns

### Key Implementation Strategies

1. **Semantic HTML First**: Use proper HTML elements (`<nav>`, `<article>`, `<section>`, `<details>`)
2. **Minimal CSS Classes**: Prefer `role="button"`, `class="secondary"`, `class="dropdown"` over custom styles
3. **CSS Variables for Customization**: Use `--pico-primary`, `--pico-background-color`, etc. instead of hardcoded values
4. **Theme Support**: Always test both light and dark modes using `[data-theme="dark"]` selectors
5. **Responsive by Default**: Trust Pico.css responsive behavior, avoid custom breakpoints unless necessary

### Authentication Patterns

- Use `<details class="dropdown">` for user menus instead of JavaScript dropdowns
- Implement theme switching with `localStorage.setItem('picoPreferredColorScheme', theme)`
- Style CTAs with `role="button"` and appropriate classes (`secondary`, `contrast`, `outline`)

### Anti-Patterns to Avoid

- Don't use utility classes like Tailwind CSS (`bg-blue-500`, `text-white`, etc.)
- Don't override Pico.css with excessive inline styles
- Don't use Alpine.js or JavaScript for basic interactions that Pico.css handles
- Don't hardcode colors - use CSS variables for theme compatibility
6. **🚨 CRITICAL**: Assume Buffalo is running and working unless proven otherwise

### Database Management
- **PostgreSQL**: Runs in a Podman container on port 5432
- **Use `podman-compose ps`** to check container status
- **Database persists** between restarts via Docker volumes
- **Migrations**: Run `buffalo pop migrate` only when adding new migrations

### Testing Changes
- **Templates**: Auto-reload on save, just refresh the browser - NO RESTART NEEDED
- **Go code**: Auto-compiles and restarts Buffalo server automatically
- **Static assets**: Auto-reload via Buffalo's asset pipeline
- **Database changes**: Require migration runs but Buffalo stays running
- **🚨 IMPORTANT**: Let Buffalo handle all reloading - manual intervention not needed

### Common Commands
- `make dev` - Start everything (use once)
- `podman-compose ps` - Check database status
- `buffalo pop migrate` - Run new database migrations
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

## Project-Specific Notes

- This is a Go Buffalo SaaS template project
- Templates use Plush templating engine (.plush.html files)
- Styling is handled through Pico.css - a semantic CSS framework with automatic theming
- Custom styles can be added with CSS variables to maintain Pico.css design consistency
- Dark/light mode switching is built-in with localStorage persistence

## Implementation Guidelines

**IMPORTANT**: Always refer to the documentation in `/docs/` folder for implementation strategies:

- **Pico.css CSS Variables**: Read `/docs/pico-css-variables.md` for customization with CSS variables
- **Implementation Patterns**: Read `/docs/pico-implementation-guide.md` for semantic HTML patterns and best practices
- **Template Syntax**: Read `/docs/buffalo-template-syntax.md` for Plush templating patterns

### Key Implementation Strategies

1. **Semantic HTML First**: Use proper HTML elements (`<nav>`, `<article>`, `<section>`, `<details>`)
2. **Minimal CSS Classes**: Prefer `role="button"`, `class="secondary"`, `class="dropdown"` over custom styles
3. **CSS Variables for Customization**: Use `--pico-primary`, `--pico-background-color`, etc. instead of hardcoded values
4. **Theme Support**: Always test both light and dark modes using `[data-theme="dark"]` selectors
5. **Responsive by Default**: Trust Pico.css responsive behavior, avoid custom breakpoints unless necessary

### Authentication Patterns

- Use `<details class="dropdown">` for user menus instead of JavaScript dropdowns
- Implement theme switching with `localStorage.setItem('picoPreferredColorScheme', theme)`
- Style CTAs with `role="button"` and appropriate classes (`secondary`, `contrast`, `outline`)

### Anti-Patterns to Avoid

- Don't use utility classes like Tailwind CSS (`bg-blue-500`, `text-white`, etc.)
- Don't override Pico.css with excessive inline styles
- Don't use Alpine.js or JavaScript for basic interactions that Pico.css handles
- Don't hardcode colors - use CSS variables for theme compatibility

# Copilot Instructions

## General Guidelines

- Never attempt to read `pico.min.css` files - they are minified and will only cause failures
- Trust that Pico.css is properly installed and working in the project
- Focus on using semantic HTML with minimal CSS classes - Pico.css provides styling automatically
- Use semantic HTML elements and follow modern web development practices with accessibility in mind

## Buffalo Development Environment Guidelines

### Buffalo Development Server
- **Buffalo runs on port 3000** and automatically reloads on file changes
- **DO NOT kill the Buffalo process** when testing changes - it has hot reload built-in
- **Check for existing Buffalo instances** before starting a new one:
  - Use `ps aux | grep buffalo` or `lsof -i :3000` to check for running instances
  - Buffalo dev server should be left running in a background terminal
  - Changes to Go files, templates, and assets will auto-reload

### Development Workflow
1. **Start once**: Use `make dev` to start PostgreSQL + Buffalo
2. **Keep running**: Leave Buffalo running in the background
3. **Make changes**: Edit files and let Buffalo auto-reload
4. **Test changes**: Refresh browser or use the running instance
5. **Only restart if**: There are compilation errors or you need to reset the database

### Database Management
- **PostgreSQL**: Runs in a Podman container on port 5432
- **Use `podman-compose ps`** to check container status
- **Database persists** between restarts via Docker volumes
- **Migrations**: Run `buffalo pop migrate` only when adding new migrations

### Testing Changes
- **Templates**: Auto-reload on save, just refresh the browser
- **Go code**: Auto-compiles and restarts Buffalo server
- **Static assets**: Auto-reload via Buffalo's asset pipeline
- **Database changes**: Require migration runs but Buffalo stays running

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

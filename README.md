# My Go SaaS Template

A Buffalo-based SaaS application template with containerized PostgreSQL database, complete authentication system, and a modern HTMX-driven UI.

## ✅ Setup Checklist

- [x] **Buffalo application generated** - Basic Buffalo app structure created
- [x] **PostgreSQL containerized** - Docker Compose setup with PostgreSQL 15
- [x] **Database configuration** - All databases created (development, test, production)
- [x] **Database migrations** - Schema up to date and working
- [x] **Application running** - Buffalo dev server successfully connecting to database
- [x] **Authentication system** - Complete user registration, login, logout with session management
- [x] **User dashboard** - Protected dashboard with user dropdown menu
- [x] **Template system** - Plush templates with Pico.css semantic styling
- [x] **HTMX Integration** - Core navigation and forms use HTMX for dynamic content loading
- [x] **Persistent UI Shell** - Main header and footer persist, content swaps via HTMX
- [x] **Modal Authentication** - Login and Sign Up forms presented in modals
- [x] **SEO optimization** - Search engine friendly with meta tags, Open Graph, and sitemap
- [ ] User profile management
- [ ] Billing/subscription features
- [ ] Email services
- [ ] Production deployment

## 🚀 Quick Start

### Prerequisites
- Go 1.19+
- Docker and Docker Compose
- Buffalo CLI

### Development Mode (Recommended)

The easiest way to get started is using the Makefile shortcuts:

```console
# First time setup (creates database and runs migrations)
make setup

# Start development mode (database + Buffalo dev server)
make dev
```

Visit [http://127.0.0.1:3000](http://127.0.0.1:3000) to see your application. The homepage content is loaded dynamically via an HTMX call on page load.

### Manual Setup

If you prefer to run commands manually:

```console
# Start PostgreSQL container
docker-compose up -d

# Run database migrations (if needed)
buffalo pop migrate

# Start the development server
buffalo dev
```

### Available Make Commands

- **`make dev`** - Start database and Buffalo development server
- **`make setup`** - Initial setup (database + migrations)
- **`make db-up`** - Start only the database
- **`make test`** - Run tests with database
- **`make clean`** - Stop all services and clean up

## 🔐 Authentication Features

### User Registration & Login
- **Registration**: `/users/new` - Create new user accounts (form loads in a modal via HTMX)
- **Login**: `/auth/new` - Sign in with email/password (form loads in a modal via HTMX)
- **Dashboard**: `/dashboard` - Protected area for authenticated users (content loads via HTMX)
- **Logout**: Available via user dropdown menu (uses HTMX POST)

### User Interface
- **Persistent Header/Footer**: The main site header (with navigation, theme toggle, profile/auth links) and footer are defined in `templates/home/index.plush.html` and persist across page views using `hx-preserve="true"`.
- **Dynamic Content Area**: The `<main id="htmx-content">` in `index.plush.html` is where page-specific content is dynamically loaded by HTMX.
- **Modal Authentication Forms**: "Login" and "Sign Up" buttons in the header trigger Pico.css modals. The respective forms are loaded into these modals via HTMX.
- **Landing Page**: Marketing page with conditional CTAs. Login/Signup CTAs open modals.
- **User Dropdown**: Professional dropdown menu in the persistent header for authenticated users with:
  - User avatar (initials with gradient background)
  - User name display
  - Profile Settings (placeholder, loads via HTMX)
  - Account Settings (placeholder, loads via HTMX)  
  - Sign Out functionality (HTMX POST request)

### Authentication Flow
1. Unauthenticated users see the landing page. "Login" and "Sign Up" buttons in the header open modals with the respective forms, loaded via HTMX.
2. After successful login/signup from a modal, the server typically responds with an `HX-Refresh: true` header, causing a full page refresh. This updates the header to the logged-in state and closes the modal.
3. Authenticated users see the persistent header with their profile dropdown. Navigating to areas like `/dashboard` loads the content into the `#htmx-content` area.
4. Logout (via an HTMX POST request from the dropdown) clears the session and typically triggers an `HX-Refresh: true` or `HX-Redirect` to the landing page.

## ✨ HTMX Integration

This template heavily utilizes HTMX for a modern, single-page application feel without complex JavaScript frameworks.

- **Core Principle**: The main layout (`templates/home/index.plush.html`) acts as a persistent shell. Navigation links and form submissions use HTMX attributes (`hx-get`, `hx-post`, `hx-target`, `hx-swap`) to fetch HTML fragments from the server and swap them into the `#htmx-content` div.
- **Initial Page Load**: The homepage (`/`) loads `index.plush.html`, and the `<main id="htmx-content">` tag has an `hx-trigger="load"` attribute that immediately makes an HTMX request to fetch the initial homepage content (`_index_content.plush.html`).
- **Server-Side Handling**:
    - Go handlers in the `actions` package detect HTMX requests (using `IsHTMX(c.Request())` from `actions/render.go`).
    - For HTMX requests, handlers use a specific render engine (`rHTMX`) that employs a minimal layout (`templates/htmx.plush.html`, which is just `<%= yield %>`). This ensures only the necessary HTML fragment is sent to the client.
    - Standard (non-HTMX) requests render full pages using the default engine (`r`) and `templates/application.plush.html` (which now primarily serves `index.plush.html` for the main app view).
- **Benefits**: Reduced page flicker, faster perceived load times for content changes, and simpler server-rendered HTML.

##  SEO & Performance Features

### Search Engine Optimization
- **Search Engine Friendly**: Fixed robots.txt to allow crawling while protecting private areas. Initial content for the homepage is loaded via HTMX on page load, ensuring it's available.
- **Dynamic Meta Tags**: Page-specific titles, descriptions, and keywords (managed in `application.plush.html` and passed from handlers).
- **Open Graph**: Rich social media previews for Facebook, Twitter, and LinkedIn.
- **Structured Data**: JSON-LD schema markup for SaaS applications.
- **Canonical URLs**: Prevent duplicate content issues.
- **XML Sitemap**: Auto-generated sitemap for search engines.

### Performance & Accessibility
- **Semantic HTML**: Proper HTML5 structure with Pico.css semantic styling.
- **HTMX for Dynamic Updates**: Enhances perceived performance by only updating necessary page parts.
- **Mobile-First**: Responsive design with proper viewport settings.
- **Theme Support**: Dark/light/auto modes with system preference detection, functional within the persistent HTMX-driven UI.
- **Fast Loading**: Minimal CSS/JS footprint.
- **Accessibility**: WCAG-compliant markup and keyboard navigation.

## 📊 Architecture

- **Backend**: Buffalo (Go web framework)
- **Database**: PostgreSQL 15 (containerized)
- **Frontend**:
    - Plush templates
    - Pico.css (semantic styling)
    - HTMX (dynamic content loading, AJAX interactions)
- **Authentication**: Session-based with bcrypt password hashing, forms presented in modals.
- **Background Jobs**: Buffalo workers (setup not detailed here but available in Buffalo).

## 🛠️ Development

### Template Development

This project uses Buffalo's Plush templating engine, Pico.css for styling, and HTMX for dynamic interactions.

- **Main Shell**: `templates/home/index.plush.html` is the primary persistent layout containing the header, footer, and the `<main id="htmx-content">` target.
- **Content Partials**: Most page-specific content is in separate partial files (e.g., `templates/home/_index_content.plush.html`, `templates/home/dashboard.plush.html`). These are loaded into `#htmx-content`.
- **HTMX Fragments Layout**: `templates/htmx.plush.html` (containing just `<%= yield %>`) is used by the `rHTMX` render engine for HTMX responses.
- **Plush Syntax**: See `/docs/buffalo-template-syntax.md`.
- **Pico.css**: See `/docs/pico-implementation-guide.md` and `/docs/pico-css-variables.md`.
- **Modals**: Pico.css `<dialog>` elements are used for login/signup, triggered by JavaScript and populated by HTMX.
- **Theme Switching**: Built-in dark/light/auto mode support, works with the persistent header.

### Database Management
The application includes a PostgreSQL container configured via `docker-compose.yml`:
- **Development DB**: `my_go_saas_template_development`
- **Test DB**: `my_go_saas_template_test`
- **Production DB**: `my_go_saas_template_production`

### Common Commands
```console
# Development shortcuts (recommended)
make dev                       # Start everything for development
make setup                     # First-time setup
make test                      # Run tests
make clean                     # Stop and cleanup

# Manual commands
# Database operations
buffalo pop create -a          # Create databases
buffalo pop migrate            # Run migrations
buffalo pop generate migration # Create new migration

# Development
buffalo dev                    # Start dev server with hot reload
buffalo build                  # Build production binary
buffalo test                   # Run tests

# Docker
docker-compose up -d           # Start database
docker-compose down            # Stop database
```
User management endpoints (`/users`, `/auth`) are still the same, but interactions are now primarily via HTMX from modals or links.

## 🤖 Bot Instructions

When working with this Buffalo SaaS template:

### Template Development & HTMX
1.  **Understand the Shell**: `templates/home/index.plush.html` is the persistent shell. Most new content should be a partial loaded into `#htmx-content`.
2.  **HTMX Attributes**: Use `hx-get`, `hx-post`, `hx-target="#htmx-content"`, `hx-swap` for navigation and forms. For modals, target the modal's content div.
3.  **Server-Side**:
    *   Check for HTMX requests using `IsHTMX(c.Request())`.
    *   Use `rHTMX.HTML("path/to/partial.plush.html")` for HTMX responses.
    *   Use `r.HTML("home/index.plush.html")` for initial full page loads of the main app view.
4.  **Plush Syntax**: Refer to `/docs/buffalo-template-syntax.md`. Avoid Go-style operations.
5.  **Modals**: Login/signup forms are in modals. Ensure HTMX attributes on trigger buttons target modal content divs. Server responses for modal forms (e.g., validation errors) should re-render the form fragment. Successful modal submissions often use `HX-Refresh: true`.

### Styling with Pico.css
1.  **Semantic HTML**: Key for Pico.css.
2.  **Modals**: Use `<dialog>` and `<article>` structure.
3.  **Theme Support**: Works with the persistent header.

### Authentication
1.  **Modal Forms**: Login/signup are via modals loaded with HTMX.
2.  **Session Management**: `current_user_id` in session.
3.  **Post-Login/Signup**: Usually `HX-Refresh: true` from server.

### Common Patterns
- **Persistent Elements**: Header/footer in `index.plush.html` use `hx-preserve="true"`.
- **Conditional Content**: Check `current_user` for auth-specific content, often within partials.
- **Form Handling**: Standard Buffalo form helpers can be used, but HTMX attributes handle submission.

### Troubleshooting
- **500 errors**: Often Plush syntax. Check Buffalo logs.
- **HTMX Issues**: Use browser dev tools (Network tab) to inspect HTMX requests and responses. Check `HX-Request` headers and what HTML fragments are being returned. Ensure `hx-target` and `hx-swap` are correct.

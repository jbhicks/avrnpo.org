# My Go SaaS Template

A production-ready Buffalo-based SaaS application template with containerized PostgreSQL database, complete authentication system, comprehensive role-based admin management, and a modern HTMX-driven UI using Pico.css semantic styling.

## ✅ Features Checklist

### Core Infrastructure
- [x] **Buffalo application** - Modern Go web framework with hot reload development
- [x] **PostgreSQL containerized** - Podman Compose setup with PostgreSQL 15
- [x] **Database configuration** - All environments configured (development, test, production)
- [x] **Database migrations** - Complete schema with role-based user system
- [x] **Robust development workflow** - Enhanced make commands with health checks and error handling

### Authentication & Authorization
- [x] **Complete authentication system** - User registration, login, logout with secure session management
- [x] **Role-based access control** - User and admin roles with proper authorization middleware
- [x] **User profile management** - Edit profile information with admin role management
- [x] **Password security** - bcrypt hashing with proper validation

### Admin Management System
- [x] **Admin dashboard** - Statistics and overview panel for administrators
- [x] **User management CRUD** - Complete user listing, editing, and deletion with safety controls
- [x] **Role assignment** - Promote/demote users between user and admin roles
- [x] **Admin promotion system** - Automated first-user admin promotion via grift tasks
- [x] **Authorization middleware** - Secure admin-only route protection
- [x] **Safety controls** - Prevent self-deletion and unauthorized access

### User Interface & Experience
- [x] **Template system** - Plush templates with semantic Pico.css styling
- [x] **HTMX integration** - Dynamic content loading without page refreshes
- [x] **Persistent UI shell** - Header and footer persist, content area updates dynamically
- [x] **Modal authentication** - Login and Sign Up forms in responsive modals
- [x] **Theme system** - Dark/light/auto modes with system preference detection
- [x] **Responsive design** - Mobile-first approach with semantic HTML
- [x] **Professional UI components** - User dropdowns, admin tables, and management interfaces

### SEO & Performance
- [x] **SEO optimization** - Meta tags, Open Graph, structured data, and XML sitemap
- [x] **Performance optimization** - Minimal CSS/JS footprint with efficient HTMX updates
- [x] **Accessibility** - WCAG-compliant markup and keyboard navigation
- [x] **Search engine friendly** - Proper robots.txt and canonical URLs

### Development & Testing
- [x] **Hot reload development** - Buffalo dev server with automatic recompilation
- [x] **Testing framework** - Buffalo testing suite with database integration
- [x] **Database health checks** - Automated PostgreSQL readiness verification
- [x] **Error handling** - Comprehensive error handling in make commands and scripts

### Pending Features
- [ ] **Billing/subscription features** - Payment processing and subscription management
- [ ] **Email services** - Transactional emails and notifications
- [ ] **Production deployment** - Docker containers and cloud deployment guides

## 🚀 Quick Start

### Prerequisites
- **Go 1.19+** - [Download Go](https://golang.org/dl/)
- **Podman/Docker** - For PostgreSQL container ([Install Podman](https://podman.io/getting-started/installation))
- **Buffalo CLI** - `go install github.com/gobuffalo/cli/cmd/buffalo@latest`

### One-Command Setup

The fastest way to get started with a fully functional SaaS application:

```console
# Clone the repository
git clone <your-repo-url>
cd my-go-saas-template

# Complete setup (database + migrations + first run)
make setup

# Start development mode
make dev
```

After setup, visit [http://127.0.0.1:3000](http://127.0.0.1:3000) to see your application running.

## 🔥 Buffalo Auto-Reload Development

**Important**: Buffalo has built-in hot reload that automatically handles all file changes. Once you run `make dev`, the server stays running and automatically reloads when you make changes.

### How Auto-Reload Works
- **Go code changes** → Buffalo automatically recompiles and restarts the server
- **Template changes** → Templates reload instantly without server restart
- **Static assets** → CSS/JS changes update automatically via the asset pipeline
- **Database migrations** → Run migrations with `buffalo pop migrate` while server runs

### Development Best Practices
- **Start once**: Run `make dev` at the beginning of your development session
- **Keep running**: Leave Buffalo running in the background throughout development
- **Just code**: Make your changes and refresh the browser to see updates
- **No manual restarts**: Buffalo handles all recompilation automatically

### When to Restart Buffalo
- **Compilation errors**: If Go code has syntax errors preventing compilation
- **Database issues**: If you need to reset the database state
- **Explicit request**: Only restart if specifically needed for debugging

**🚨 Never kill the Buffalo process during normal development** - it's designed to handle all changes automatically!

### First Admin User

To set up your first admin user:

1. **Register a user account** through the web interface
2. **Promote to admin** with one command:
   ```console
   make admin
   ```

This automatically promotes the first registered user to admin role, giving them access to the admin panel.

## 👑 Admin Management System

This template includes a comprehensive role-based admin management system with full CRUD operations, safety controls, and a professional interface.

### Admin System Features

#### User Management
- **Complete CRUD Operations** - Create, read, update, and delete users
- **Role Assignment** - Promote/demote users between admin and user roles
- **Bulk Operations** - Efficient user management with pagination
- **Safety Controls** - Admins cannot delete their own accounts
- **Audit Trail** - Track role changes and admin actions

#### Admin Interface
- **Professional Dashboard** - Overview with user statistics and system metrics
- **User Management Table** - Sortable, paginated list with search capabilities
- **Role Management Forms** - Easy role assignment with visual feedback
- **Responsive Design** - Works seamlessly on desktop and mobile devices

#### Security Features
- **Authorization Middleware** - All admin routes protected with `AdminRequired` middleware
- **Role-Based Access** - Dynamic UI based on user permissions
- **Session Security** - Secure session management with role verification
- **Input Validation** - Comprehensive validation for all admin operations

### Setting Up Admin Access

#### Automatic Admin Promotion
```console
# Promote the first registered user to admin
make admin
```

This grift task automatically finds the first user (by creation date) and promotes them to admin role.

#### Manual Admin Promotion
```console
# Using Buffalo task directly
buffalo task db:promote_admin

# Or promote a specific user via database
psql -d my_go_saas_template_development -c "UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';"
```

### Admin Routes & API

| Route | Method | Description | Access Level |
|-------|--------|-------------|--------------|
| `/admin` | GET | Admin dashboard with statistics | Admin Only |
| `/admin/users` | GET | User management list (paginated) | Admin Only |
| `/admin/users/{id}` | GET | Edit user form | Admin Only |
| `/admin/users/{id}` | POST | Update user (including role) | Admin Only |
| `/admin/users/{id}` | DELETE | Delete user (with safety checks) | Admin Only |

### Role System Details

#### User Roles
- **`user`** (default) - Standard application access
  - Profile management
  - Dashboard access
  - Standard features

- **`admin`** - Full administrative privileges
  - All user permissions
  - Admin panel access (`/admin`)
  - User management capabilities
  - Role assignment permissions
  - System administration

#### Role Enforcement
- **Database Level** - Role field with proper constraints and validation
- **Middleware Level** - `AdminRequired` middleware protects admin routes
- **Template Level** - Conditional rendering based on user role
- **UI Level** - Dynamic navigation and feature visibility

### Admin Development Patterns

#### Adding New Admin Features
```go
// In actions/app.go - Add new admin routes
adminGroup := app.Group("/admin")
adminGroup.Use(AdminRequired)
adminGroup.GET("/new-feature", AdminNewFeatureHandler)
```

#### Template Access Control
```html
<!-- In templates - Check admin role -->
<%= if (current_user.Role == "admin") { %>
  <a href="/admin">Admin Panel</a>
<% } %>
```

#### Safety Checks Example
```go
// Prevent self-deletion
if userToDelete.ID == currentUser.ID {
    return c.Error(400, errors.New("cannot delete your own account"))
}
```

## 🛠️ Development Commands

### Quick Reference

| Command | Purpose | Description |
|---------|---------|-------------|
| `make setup` | First-time setup | Creates database, runs migrations |
| `make dev` | Development mode | Starts database + Buffalo dev server |
| `make admin` | Admin setup | Promotes first user to admin role |
| `make test` | Run tests | Executes Buffalo test suite with database |
| `make test-fast` | Quick tests | Runs Buffalo tests (assumes DB running) |
| `make clean` | Cleanup | Stops services and cleans containers |
| `make db-status` | Health check | Shows database container status |

### Development Workflow

#### First Time Setup
```console
# Clone and setup the project
git clone <your-repo-url>
cd my-go-saas-template
make setup

# Create your first user account via the web interface
# Then promote to admin
make admin
```

#### Daily Development
```console
# Start development (runs database + Buffalo dev server)
make dev

# Buffalo automatically reloads on file changes
# Visit http://127.0.0.1:3000 to see your changes
```

#### Testing & Quality Assurance
```console
# Run all tests with Buffalo (recommended)
make test

# Quick test run (assumes database is running)
make test-fast

# Run Buffalo tests directly
buffalo test

# Check database health
make db-status

# Clean up after development
make clean
```

### Advanced Commands

#### Database Operations
```console
# Manual database management
make db-up                     # Start database only
make db-down                   # Stop database
make db-reset                  # Reset database (drop/create/migrate)
make migrate                   # Run migrations only

# Buffalo database commands
buffalo pop create -a          # Create all databases
buffalo pop migrate            # Run migrations
buffalo pop generate migration # Create new migration
buffalo pop drop -e development # Drop development database
```

#### Admin Management
```console
# Admin user management
buffalo task db:promote_admin  # Promote first user to admin
make admin                     # Same as above (via make)

# Manual role assignment via database
psql -d my_go_saas_template_development \
  -c "UPDATE users SET role = 'admin' WHERE email = 'user@example.com';"
```

#### Building & Production
```console
# Build for production
make build                     # Creates binary in bin/
buffalo build                  # Direct Buffalo build
buffalo build --static        # Static binary build

# Production database setup
buffalo pop create -e production
buffalo pop migrate -e production
```

### Development Tips

#### Buffalo Development Server
- **Automatic reload** - Buffalo watches files and reloads automatically
- **Port 3000** - Default development port
- **Hot reload** - Template and Go code changes reload automatically
- **Keep running** - Leave Buffalo running, it handles recompilation

#### Database Development
- **Container persistence** - Database data persists between restarts
- **Health checks** - Make commands wait for database readiness
- **Multiple environments** - Development, test, and production databases
- **Migration tracking** - Buffalo tracks applied migrations automatically

#### Template Development
- **HTMX integration** - Templates support dynamic content loading
- **Pico.css styling** - Semantic HTML with automatic styling
- **Plush templating** - Buffalo's template engine with Go-like syntax
- **Live reload** - Template changes appear immediately

### Troubleshooting

#### Common Issues

**Database Connection Issues**
```console
# Check container status
make db-status
podman-compose ps

# Check logs
podman-compose logs postgres

# Reset database if corrupted
make db-reset
```

**Buffalo Issues**
```console
# Check if Buffalo is running
ps aux | grep buffalo
lsof -i :3000

# Restart Buffalo if needed
# Ctrl+C to stop, then: make dev
```

**Port Conflicts**
```console
# Check what's using port 3000
lsof -i :3000

# Kill process if needed
kill -9 $(lsof -t -i:3000)
```

**Template Errors**
- Check Buffalo console output for Plush syntax errors
- Ensure proper variable names and template structure
- Verify HTMX attributes and targets are correct

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

## 📊 Architecture & Technology Stack

### Backend Architecture
- **Framework**: Buffalo (Go web framework) - Modern, productive web development
- **Database**: PostgreSQL 15 (containerized) - Reliable, ACID-compliant database
- **Authentication**: Session-based with bcrypt password hashing
- **Authorization**: Role-based access control (RBAC) with middleware protection
- **Background Jobs**: Buffalo workers (available for async processing)
- **Testing**: Buffalo testing framework with database integration

### Frontend Architecture
- **Templating**: Plush templates - Buffalo's template engine with Go-like syntax
- **Styling**: Pico.css - Semantic CSS framework with automatic theming
- **Interactions**: HTMX - Dynamic content loading without complex JavaScript
- **Theme System**: Dark/light/auto modes with localStorage persistence
- **Responsive Design**: Mobile-first approach with semantic HTML

### Database Schema

#### Users Table
```sql
users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user' -- 'user' or 'admin'
);
```

#### Key Features
- **UUID Primary Keys** - Secure, non-enumerable identifiers
- **Timestamps** - Automatic created_at/updated_at tracking
- **Email Uniqueness** - Prevents duplicate accounts
- **Password Security** - bcrypt hashing with proper salt rounds
- **Role System** - Extensible role-based permissions

### Application Structure

```
my-go-saas-template/
├── actions/          # HTTP handlers and middleware
│   ├── app.go       # Main application and routing
│   ├── auth.go      # Authentication handlers
│   ├── admin.go     # Admin management system
│   ├── users.go     # User profile management
│   └── render.go    # Template rendering utilities
├── models/          # Database models and validation
│   └── user.go      # User model with role support
├── templates/       # Plush template files
│   ├── application.plush.html  # Main layout
│   ├── home/        # Homepage and dashboard
│   ├── auth/        # Authentication forms
│   ├── users/       # User profile management
│   └── admin/       # Admin panel templates
├── migrations/      # Database migration files
├── grifts/         # Background tasks and utilities
├── public/         # Static assets (CSS, JS, images)
└── docs/           # Documentation and guides
```

### HTMX Integration Architecture

#### Core Concept
The application uses a persistent shell architecture where the main layout stays loaded and content areas are dynamically updated via HTMX.

#### Key Components
1. **Persistent Shell** (`templates/home/index.plush.html`)
   - Header with navigation and user menu
   - Footer with site information
   - `<main id="htmx-content">` target area

2. **Content Partials** (Various template files)
   - Dashboard content (`templates/home/dashboard.plush.html`)
   - Admin interfaces (`templates/admin/*.plush.html`)
   - User forms (`templates/users/*.plush.html`)

3. **HTMX Response Engine** (`actions/render.go`)
   - Detects HTMX requests via headers
   - Uses minimal layout for partial responses
   - Maintains full page rendering for direct access

#### Benefits
- **Faster Navigation** - Only content area updates, header/footer persist
- **Better UX** - No page flicker, smoother transitions
- **SEO Friendly** - Full pages still render for search engines
- **Progressive Enhancement** - Works without JavaScript as fallback

### Security Architecture

#### Authentication Security
- **Session Management** - Secure session cookies with proper expiration
- **Password Hashing** - bcrypt with appropriate cost factor
- **CSRF Protection** - Built-in Buffalo CSRF middleware
- **Input Validation** - Comprehensive validation on all user inputs

#### Authorization Security
- **Role-Based Access** - Middleware-enforced role checking
- **Route Protection** - Admin routes protected with `AdminRequired` middleware
- **Template Security** - Role-based conditional rendering
- **API Security** - Proper authorization checks on all endpoints

#### Database Security
- **Prepared Statements** - All queries use proper parameterization
- **Connection Pooling** - Secure database connection management
- **Migration Tracking** - Database schema version control
- **Data Validation** - Model-level validation before database operations

### Performance Optimizations

#### Frontend Performance
- **Minimal JavaScript** - HTMX provides dynamic behavior with minimal JS
- **Semantic CSS** - Pico.css provides styling without utility class bloat
- **Efficient Templates** - Plush templates compiled for performance
- **Static Asset Optimization** - Minified CSS and optimized images

#### Backend Performance
- **Compiled Go Binary** - High-performance compiled application
- **Connection Pooling** - Efficient database connection management
- **Session Optimization** - Efficient session storage and retrieval
- **Template Caching** - Plush templates cached in production

#### Database Performance
- **Indexed Queries** - Proper indexing on frequently queried columns
- **Query Optimization** - Efficient queries with minimal N+1 problems
- **Connection Limits** - Proper connection pool sizing
- **Migration Efficiency** - Non-blocking migrations where possible

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
make admin                     # Promote first user to admin

# Manual commands
# Database operations
buffalo pop create -a          # Create databases
buffalo pop migrate            # Run migrations
buffalo pop generate migration # Create new migration

# Admin management
buffalo task db:promote_admin  # Promote first user to admin role

# Development
buffalo dev                    # Start dev server with hot reload
buffalo build                  # Build production binary
buffalo test                   # Run tests (always use this instead of 'go test')

# Container management (Podman/Docker)
podman-compose up -d           # Start database
podman-compose down            # Stop database
podman-compose ps              # Check container status
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

### Authentication & Authorization
1.  **Modal Forms**: Login/signup are via modals loaded with HTMX.
2.  **Session Management**: `current_user_id` in session, `current_user` in templates.
3.  **Role-based Access**: Check `current_user.Role` for admin functionality.
4.  **Admin Middleware**: Use `AdminRequired` middleware for admin-only routes.
5.  **Post-Login/Signup**: Usually `HX-Refresh: true` from server.

### Admin System Patterns
1.  **Admin Routes**: Group under `/admin` with `AdminRequired` middleware.
2.  **Role Checks**: Use `current_user.Role == "admin"` in templates for conditional content.
3.  **User Management**: CRUD operations follow Buffalo conventions with proper validation.
4.  **Safety Checks**: Always prevent users from deleting themselves or escalating beyond their permissions.

### Common Patterns
- **Persistent Elements**: Header/footer in `index.plush.html` use `hx-preserve="true"`.
- **Conditional Content**: Check `current_user` for auth-specific content, often within partials.
- **Form Handling**: Standard Buffalo form helpers can be used, but HTMX attributes handle submission.

### Troubleshooting
- **500 errors**: Often Plush syntax. Check Buffalo logs.
- **HTMX Issues**: Use browser dev tools (Network tab) to inspect HTMX requests and responses. Check `HX-Request` headers and what HTML fragments are being returned. Ensure `hx-target` and `hx-swap` are correct.

## 🤖 Development Assistant Instructions

When working with this Buffalo SaaS template, follow these patterns and guidelines:

#### Template Development & HTMX Integration
1. **Understand the Shell Architecture**: `templates/home/index.plush.html` is the persistent shell. Most new content should be a partial template loaded into `#htmx-content`.

2. **HTMX Response Patterns**: 
   - Use `hx-get`, `hx-post`, `hx-target="#htmx-content"`, `hx-swap` for navigation and forms
   - For modals, target the modal's content div specifically
   - Server responses use `rHTMX.HTML("path/to/partial.plush.html")` for HTMX requests

3. **Template Engine Guidelines**:
   - Check for HTMX requests using `IsHTMX(c.Request())` in handlers
   - Use `r.HTML("home/index.plush.html")` for initial full page loads
   - Reference `/docs/buffalo-template-syntax.md` for Plush syntax patterns

#### Styling with Pico.css Framework
1. **Semantic HTML First**: Use proper HTML elements (`<nav>`, `<article>`, `<section>`, `<details>`)
2. **Minimal CSS Classes**: Prefer `role="button"`, `class="secondary"`, `class="dropdown"` over custom styles
3. **Theme Compatibility**: Use CSS variables (`--pico-primary`, `--pico-background-color`) instead of hardcoded colors
4. **Responsive Design**: Trust Pico.css responsive behavior, avoid custom breakpoints unless necessary

#### Authentication & Authorization Patterns
1. **Modal Authentication**: Login/signup forms load via HTMX into modal dialogs
2. **Session Management**: Use `current_user_id` in session, `current_user` available in templates
3. **Role-Based Access**: Check `current_user.Role` for admin functionality in templates
4. **Admin Middleware**: Always use `AdminRequired` middleware for admin-only routes
5. **Post-Authentication**: Use `HX-Refresh: true` header for successful modal form submissions

#### Admin System Development
1. **Route Structure**: Group admin routes under `/admin` with `AdminRequired` middleware protection
2. **Role Checking**: Use `current_user.Role == "admin"` in templates for conditional content
3. **CRUD Operations**: Follow Buffalo conventions with proper validation and error handling
4. **Safety Controls**: Always prevent users from deleting themselves or escalating beyond permissions

#### Database & Migration Patterns
1. **Migration Safety**: Use non-blocking migrations where possible for production deployments
2. **Model Validation**: Implement comprehensive validation at the model level
3. **Role System**: Use `role` field with proper constraints and default values
4. **UUID Primary Keys**: Maintain UUID usage for security and scalability

#### Common Development Patterns
- **Persistent Elements**: Header/footer use `hx-preserve="true"` to maintain state
- **Conditional Content**: Check `current_user` for authentication-specific content rendering
- **Form Handling**: Use HTMX attributes for submission while maintaining Buffalo form helpers
- **Error Handling**: Provide comprehensive error messages and proper HTTP status codes

#### Testing & Quality Assurance
- **Test Coverage**: Maintain tests for all authentication and admin functionality
- **Database Testing**: Use test database environment for isolated test runs
- **HTMX Testing**: Test both HTMX and direct URL access for all routes
- **Role Testing**: Verify proper authorization for all role-based features

#### Troubleshooting Guidelines
- **500 Errors**: Usually Plush template syntax issues - check Buffalo console output
- **HTMX Issues**: Use browser dev tools Network tab to inspect HTMX requests/responses
- **Database Issues**: Use `make db-status` and `make db-logs` for diagnostics
- **Permission Issues**: Verify middleware application and role assignments

## 📁 Project File Structure

```
my-go-saas-template/
├── 🗄️  Database & Configuration
│   ├── database.yml              # Database configuration for all environments
│   ├── docker-compose.yml        # PostgreSQL container configuration
│   └── migrations/               # Database migration files
│       ├── *_create_users.up.fizz   # Initial user table creation
│       └── *_add_role_to_users.*.fizz # Role system addition
│
├── 🏗️  Application Core
│   ├── main.go                   # Application entry point
│   ├── app                       # Buffalo application instance
│   ├── actions/                  # HTTP handlers and middleware
│   │   ├── app.go               # Main routing and application setup
│   │   ├── auth.go              # Authentication handlers (login/logout)
│   │   ├── users.go             # User profile management
│   │   ├── admin.go             # Admin management system (CRUD)
│   │   ├── home.go              # Homepage and dashboard handlers
│   │   └── render.go            # Template rendering utilities
│   │
│   ├── models/                   # Database models and validation
│   │   ├── models.go            # Database connection and base models
│   │   └── user.go              # User model with role support
│   │
│   └── grifts/                   # Background tasks and utilities
│       └── db.go                # Admin promotion and database tasks
│
├── 🎨 Frontend & Templates
│   ├── templates/                # Plush template files
│   │   ├── application.plush.html    # Base application layout
│   │   ├── home/                     # Homepage and dashboard templates
│   │   │   ├── index.plush.html      # Persistent shell (header/footer)
│   │   │   └── dashboard.plush.html  # User dashboard with admin section
│   │   ├── auth/                     # Authentication form templates
│   │   │   ├── new.plush.html        # Login form (modal)
│   │   │   └── landing.plush.html    # Marketing landing page
│   │   ├── users/                    # User management templates
│   │   │   ├── profile.plush.html    # User profile editing
│   │   │   └── new.plush.html        # User registration (modal)
│   │   └── admin/                    # Admin panel templates
│   │       ├── users.plush.html      # User management table
│   │       └── user_edit.plush.html  # User editing form
│   │
│   └── public/                   # Static assets
│       ├── css/
│       │   ├── pico.min.css     # Pico.css framework (semantic styling)
│       │   └── custom.css       # Custom CSS variables and overrides
│       ├── js/
│       │   ├── theme.js         # Dark/light mode switching
│       │   └── icons.js         # Icon system utilities
│       └── images/              # Static images and favicon
│
├── 🛠️  Development & Deployment
│   ├── Makefile                  # Robust development commands
│   ├── scripts/
│   │   └── wait-for-postgres.sh # Database health check script
│   ├── go.mod                   # Go module dependencies
│   ├── go.sum                   # Dependency checksums
│   └── bin/                     # Compiled binaries (created by build)
│
├── 🧪 Testing & Quality
│   ├── *_test.go                # Go test files throughout project
│   ├── fixtures/                # Test data fixtures
│   └── tmp/                     # Temporary build files
│
└── 📖 Documentation
    ├── README.md                # This comprehensive guide
    └── docs/                    # Additional documentation
        ├── buffalo-template-syntax.md     # Plush templating guide
        ├── pico-implementation-guide.md   # Semantic CSS patterns
        ├── pico-css-variables.md          # CSS customization guide
        └── seo-implementation.md          # SEO optimization guide
```

### Key File Descriptions

#### Core Application Files
- **`actions/app.go`** - Main application setup, routing configuration, and middleware stack
- **`actions/admin.go`** - Complete admin management system with CRUD operations and safety controls  
- **`models/user.go`** - User model with role support, validation, and authentication methods
- **`templates/home/index.plush.html`** - Persistent application shell with HTMX content area

#### Database Files
- **`migrations/*.fizz`** - Database schema evolution with role-based user system
- **`database.yml`** - Multi-environment database configuration
- **`grifts/db.go`** - Administrative tasks including user promotion to admin role

#### Frontend Architecture
- **`public/css/pico.min.css`** - Semantic CSS framework providing automatic styling
- **`public/js/theme.js`** - Theme switching functionality with localStorage persistence
- **`templates/admin/*.plush.html`** - Professional admin interface templates

#### Development Infrastructure  
- **`Makefile`** - Comprehensive development commands with health checks and error handling
- **`scripts/wait-for-postgres.sh`** - Database readiness verification for reliable automation
- **`docker-compose.yml`** - PostgreSQL container configuration for development environment

This file structure supports a scalable, maintainable SaaS application with clear separation of concerns and professional development practices.

## 📚 Additional Resources

### Documentation
- **Buffalo Framework**: [https://gobuffalo.io/documentation](https://gobuffalo.io/documentation)
- **Pico.css Documentation**: [https://picocss.com/docs](https://picocss.com/docs)
- **HTMX Documentation**: [https://htmx.org/docs](https://htmx.org/docs)
- **PostgreSQL Documentation**: [https://www.postgresql.org/docs](https://www.postgresql.org/docs)

### Project-Specific Documentation
- **Template Syntax Guide**: `/docs/buffalo-template-syntax.md`
- **Pico.css Implementation**: `/docs/pico-implementation-guide.md`
- **CSS Customization**: `/docs/pico-css-variables.md`
- **SEO Implementation**: `/docs/seo-implementation.md`

### Development Resources
- **Go Documentation**: [https://golang.org/doc](https://golang.org/doc)
- **Podman Documentation**: [https://docs.podman.io](https://docs.podman.io)
- **Database Migrations**: Buffalo Pop documentation

---

**🎉 Ready to build your SaaS application!** Start with `make setup` and follow this guide for a comprehensive, production-ready foundation.

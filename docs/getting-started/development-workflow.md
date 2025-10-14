# Development Workflow

## Starting the Development Server

The project uses Air for hot reloading during development.

### Primary Development Command

```bash
make dev
```

This command:
- Starts the PocketBase server on port 8090
- Automatically watches for file changes
- Recompiles and restarts when Go files or templates change
- Serves static assets from `public/`

### What Runs During Development

- **PocketBase Server**: Embedded database + API + admin UI
- **Custom Routes**: Donation system, contact form, blog
- **Hot Reload**: Via Air configuration in `.air.toml`
- **Template Generation**: Templ files auto-compile to `.go` files

### Configuration

Development configuration is read from:
- `.env` - Environment variables (admin credentials, API keys)
- `.air.toml` - Hot reload configuration
- `main.go` - PocketBase initialization and custom routes

### Port Configuration

Default port is 8090. To change:

```bash
# Via command line flag
./avrnpo serve --http=0.0.0.0:3000

# Or set in .env
HTTP_PORT=3000
```

### Development Best Practices

1. **Use `make dev`** - Never run `./avrnpo serve` directly (it blocks your shell)
2. **Let Air handle restarts** - File changes trigger automatic rebuilds
3. **Check logs in terminal** - Air shows compilation errors and runtime logs
4. **Run tests in separate terminal** - Don't stop dev server to test

## Stopping the Server

- **Graceful shutdown**: `Ctrl+C` in the terminal running `make dev`
- **Force kill**: `pkill -f "avrnpo serve"` (if process is stuck)

## Troubleshooting Development Server

### Port Already in Use

If you get "address already in use" errors:

```bash
# Check what's using port 8090
lsof -ti:8090

# Kill specific process if needed
pkill -f "avrnpo serve"
```

### Server Not Reloading

1. Check `.air.toml` configuration
2. Verify you're in the project root directory
3. Check for syntax errors in Go files or Templ templates
4. Run `templ generate` manually if template changes aren't detected

### Database Issues

```bash
# Reset database (WARNING: Deletes all data)
rm -rf pb_data/
make dev  # Recreates database with migrations

# Check database manually
sqlite3 pb_data/data.db "SELECT * FROM users;"
```

### Template Compilation Errors

```bash
# Manually regenerate Templ templates
templ generate

# Check for syntax errors in .templ files
templ fmt templates/
```

## Common Development Tasks

### Working with Templates

```bash
# Regenerate all templates after editing .templ files
templ generate

# Format Templ files
templ fmt templates/

# Check for Templ syntax errors
templ generate -path templates/
```

### Database Operations

**PocketBase uses SQLite with automatic migrations:**

```bash
# Create new migration (Go-based)
# Create file: pb_migrations/TIMESTAMP_description.go

# View database contents
sqlite3 pb_data/data.db
> .tables
> SELECT * FROM users;
> .quit

# Reset database (deletes all data)
rm -rf pb_data/
make dev
```

### Testing Workflow

```bash
# Run unit tests (in separate terminal while dev server runs)
go test ./...

# Run specific test
go test -v -run TestContactHandler

# Run E2E tests
E2E_TESTS=1 go test -v -run E2E

# Check test coverage
go test -cover ./...
```

### Building for Production

```bash
# Build production binary
go build -o avrnpo

# Run production server
./avrnpo serve --http=0.0.0.0:8090

# Build with optimizations
go build -ldflags="-s -w" -o avrnpo
```

## File Watching

Air watches these file types (configured in `.air.toml`):
- `.go` files (handlers, middleware, services)
- `.templ` files (automatically runs `templ generate`)
- `.env` file (triggers reload)

**Not watched:**
- Static assets in `public/` (no restart needed)
- `pb_data/` directory (database files)

## Environment Variables

Key environment variables for development (set in `.env`):

```bash
# Admin credentials
PB_ADMIN_EMAIL=admin@example.com
PB_ADMIN_PASSWORD=changeme

# Payment integration (optional for local dev)
HELCIM_API_TOKEN=your_token
HELCIM_TERMINAL_ID=your_terminal

# Email (optional, use mock in tests)
RESEND_API_KEY=your_key
```

## Working with PocketBase Admin UI

The admin UI is available at `/_/` during development:

1. Navigate to http://127.0.0.1:8090/_/
2. Login with PB_ADMIN_EMAIL and PB_ADMIN_PASSWORD from `.env`
3. Manage collections: users, posts, donations
4. View API logs and requests
5. Test API endpoints using built-in API preview

## Logs and Debugging

Development server shows:
- HTTP requests and responses
- PocketBase database queries
- Template compilation results
- Air rebuild triggers
- Go compilation errors
- Runtime panics and errors

**Debugging Tips:**
```go
// Add debug logging in handlers
app.Logger().Info("Debug message", "data", someVar)

// Check PocketBase logs
// Logs are shown in terminal during `make dev`

// Enable PocketBase debug mode
app.Bootstrap().(*core.BaseApp).Debug = true
```
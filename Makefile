.PHONY: help dev setup db-up db-down db-reset test clean build admin migrate db-status db-logs health check-deps install-deps update-deps
# Add clean-caches target for clearing Go and gopls caches
.PHONY: clean-caches

# Default target
help:
	@echo "🚀 My Go SaaS Template - Development Commands"
	@echo ""
	@echo "Recommended: Run './dev.sh' for full dev workflow (server + log monitoring in tmux)"
	@echo "Requires: tmux, multitail installed"
	@echo ""

	@echo ""
	@echo "Quick Start:"
	@echo "  setup      - 🔧 Initial setup: start database, run migrations, install deps"
	@echo "  dev        - 🏃 Start database and run Buffalo development server"
	@echo "  admin      - 👑 Promote first user to admin role"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-up      - 🗄️  Start PostgreSQL database with Docker/Podman"
	@echo "  db-down    - ⬇️  Stop PostgreSQL database"
	@echo "  db-reset   - 🔄 Reset database (drop, create, migrate)"
	@echo "  db-status  - 📊 Check database container status"
	@echo "  db-logs    - 📋 Show database container logs"
	@echo "  migrate    - 🔀 Run database migrations"
	@echo ""
	@echo "Development:"
	@echo "  test            - 🧪 Run all tests with Buffalo (recommended)"
	@echo "  test-fast       - ⚡ Run Buffalo tests without database setup"
	@echo "  test-resilient  - 🛡️  Run tests with automatic database startup"
	@echo "  test-integration - 🔒 Run CSRF integration tests (tests real middleware)"
	@echo "  validate-templates - 🔍 Enhanced template validation with variable checking"
	@echo "  validate-templates-verbose - 🔍 Enhanced template validation with detailed output"
	@echo "  build           - 🔨 Build the application for production"
	@echo "  health          - 🏥 Check system health (dependencies, database, etc.)"
	@echo "  clean           - 🧹 Stop all services and clean up containers"
	@echo "  clean-caches    - 🧹 Clear Go build, module, and gopls caches"
	@echo ""
	@echo "Dependencies:"
	@echo "  check-deps  - ✅ Check if all required dependencies are installed"
	@echo "  install-deps - 📦 Install missing dependencies (where possible)"
	@echo "  update-deps - 🔄 Update all frontend dependencies (JS/CSS) to latest versions"

# Check if all required dependencies are installed
check-deps:
	@echo "🔍 Checking required dependencies..."
	@error_count=0; \
	if ! command -v go >/dev/null 2>&1; then \
		echo "❌ Go is not installed. Please install Go 1.19+ from https://golang.org/dl/"; \
		error_count=$$((error_count + 1)); \
	else \
		echo "✅ Go is installed: $$(go version)"; \
	fi; \
	if ! command -v buffalo >/dev/null 2>&1; then \
		echo "❌ Buffalo CLI is not installed. Run: go install github.com/gobuffalo/cli/cmd/buffalo@latest"; \
		error_count=$$((error_count + 1)); \
	else \
		echo "✅ Buffalo CLI is installed: $$(buffalo version)"; \
	fi; \
	if ! command -v podman-compose >/dev/null 2>&1; then \
		if ! command -v docker-compose >/dev/null 2>&1 && ! (command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1); then \
			echo "❌ No container orchestration found. Please install Podman (recommended) or Docker with Compose."; \
			error_count=$$((error_count + 1)); \
		else \
			if command -v docker-compose >/dev/null 2>&1; then \
				echo "✅ Docker Compose (v1) is installed: $$(docker-compose version)"; \
			else \
				echo "✅ Docker Compose (v2) is installed: $$(docker compose version)"; \
			fi; \
		fi; \
	else \
		echo "✅ Podman Compose is installed: $$(podman-compose version)"; \
	fi; \
	if [ $$error_count -gt 0 ]; then \
		echo ""; \
		echo "❌ $$error_count dependencies are missing. Please install them before continuing."; \
		echo "Run 'make install-deps' to install dependencies where possible."; \
		exit 1; \
	else \
		echo ""; \
		echo "✅ All dependencies are installed and ready!"; \
	fi

# Install missing dependencies where possible
install-deps:
	@echo "📦 Installing missing dependencies..."
	@if ! command -v buffalo >/dev/null 2>&1; then \
		echo "Installing Buffalo CLI..."; \
		go install github.com/gobuffalo/cli/cmd/buffalo@latest || echo "Failed to install Buffalo CLI"; \
	fi
	@echo "✅ Dependency installation complete. Run 'make check-deps' to verify."

# Start database and development server with improved resilience
dev: check-deps validate-templates
	@echo "🏃 Starting development environment..."
	@echo "🔍 Checking for conflicting PostgreSQL containers..."
	@# Kill any existing PostgreSQL containers that might be using our port
	@if command -v docker >/dev/null 2>&1; then \
		CONFLICTING_CONTAINERS=$$(docker ps -q --filter "publish=5432" --filter "ancestor=postgres" 2>/dev/null); \
		if [ -n "$$CONFLICTING_CONTAINERS" ]; then \
			echo "🔪 Found conflicting PostgreSQL containers using port 5432:"; \
			docker ps --filter "publish=5432" --filter "ancestor=postgres" --format "table {{.Names}}\t{{.Image}}\t{{.Ports}}"; \
			echo "🗡️  Stopping conflicting containers..."; \
			docker stop $$CONFLICTING_CONTAINERS >/dev/null 2>&1; \
			docker rm $$CONFLICTING_CONTAINERS >/dev/null 2>&1; \
			echo "✅ Conflicting containers removed"; \
		fi; \
	fi; \
	if command -v podman >/dev/null 2>&1; then \
		CONFLICTING_CONTAINERS=$$(podman ps -q --filter "publish=5432" --filter "ancestor=postgres" 2>/dev/null); \
		if [ -n "$$CONFLICTING_CONTAINERS" ]; then \
			echo "🔪 Found conflicting PostgreSQL containers using port 5432:"; \
			podman ps --filter "publish=5432" --filter "ancestor=postgres" --format "table {{.Names}}\t{{.Image}}\t{{.Ports}}"; \
			echo "🗡️  Stopping conflicting containers..."; \
			podman stop $$CONFLICTING_CONTAINERS >/dev/null 2>&1; \
			podman rm $$CONFLICTING_CONTAINERS >/dev/null 2>&1; \
			echo "✅ Conflicting containers removed"; \
		fi; \
	fi
	@echo "🔍 Checking database status..."
	@# Check if database is already running and ready
	@DB_READY=false; \
	if command -v docker-compose >/dev/null 2>&1; then \
		if docker-compose ps 2>/dev/null | grep -q "postgres.*Up\|postgres.*running\|postgres.*healthy"; then \
			echo "✅ Database container is running (Docker)"; \
			if docker-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
				echo "✅ PostgreSQL is ready and accepting connections"; \
				DB_READY=true; \
			else \
				echo "⚠️  Database container is running but PostgreSQL is not ready"; \
			fi; \
		else \
			echo "🐳 Database container not running, starting with Docker Compose..."; \
		fi; \
	elif command -v podman-compose >/dev/null 2>&1; then \
		if podman-compose ps 2>/dev/null | grep -q "postgres.*Up\|postgres.*running\|postgres.*healthy"; then \
			echo "✅ Database container is running (Podman)"; \
			if podman-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
				echo "✅ PostgreSQL is ready and accepting connections"; \
				DB_READY=true; \
			else \
				echo "⚠️  Database container is running but PostgreSQL is not ready"; \
			fi; \
		else \
			echo "🔷 Database container not running, starting with Podman Compose..."; \
		fi; \
	else \
		echo "❌ Neither docker-compose nor podman-compose found."; \
		echo "Please install Docker (recommended) or Podman."; \
		exit 1; \
	fi; \
	\
	if [ "$$DB_READY" = "false" ]; then \
		echo "🗄️  Ensuring database is running..."; \
		if command -v docker-compose >/dev/null 2>&1; then \
			docker-compose up -d postgres || (echo "❌ Failed to start database with Docker Compose" && exit 1); \
		elif command -v podman-compose >/dev/null 2>&1; then \
			podman-compose up -d postgres || (echo "❌ Failed to start database with Podman Compose" && exit 1); \
		fi; \
		\
		echo "🔍 Waiting for database to be ready..."; \
		MAX_WAIT=30; \
		WAIT_COUNT=0; \
		while [ $$WAIT_COUNT -lt $$MAX_WAIT ]; do \
			if command -v docker-compose >/dev/null 2>&1; then \
				if docker-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
					echo "✅ PostgreSQL is ready!"; \
					break; \
				fi; \
			elif command -v podman-compose >/dev/null 2>&1; then \
				if podman-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
					echo "✅ PostgreSQL is ready!"; \
					break; \
				fi; \
			fi; \
			echo "Waiting for PostgreSQL... ($$((WAIT_COUNT + 1))/$$MAX_WAIT)"; \
			sleep 1; \
			WAIT_COUNT=$$((WAIT_COUNT + 1)); \
		done; \
		\
		if [ $$WAIT_COUNT -ge $$MAX_WAIT ]; then \
			echo "❌ PostgreSQL failed to become ready within $$MAX_WAIT seconds"; \
			echo "Database container status:"; \
			if command -v docker-compose >/dev/null 2>&1; then \
				docker-compose ps; \
				echo "Container logs:"; \
				docker-compose logs postgres --tail 20; \
			elif command -v podman-compose >/dev/null 2>&1; then \
				podman-compose ps; \
				echo "Container logs:"; \
				podman-compose logs postgres --tail 20; \
			fi; \
			echo "⚠️  Database startup failed, but continuing to try Buffalo..."; \
			echo "💡 You may need to run 'make db-reset' if there are database issues."; \
		fi; \
	fi
	@echo "🚀 Starting Buffalo development server..."
	@echo "📱 Visit http://127.0.0.1:3001 to see your application"
	@echo "🔥 Hot reload is enabled - changes will be reflected automatically"
	@buffalo dev || (echo "❌ Buffalo failed to start. Check the output above for errors." && exit 1)

# Initial setup with comprehensive checks
setup: check-deps db-up migrate
	@echo "🎉 Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make dev' to start the development server"
	@echo "  2. Visit http://127.0.0.1:3001 to see your application"
	@echo "  3. Create a user account through the web interface"
	@echo "  4. Run 'make admin' to promote your user to admin"
	@echo ""
	@echo "🔧 Development commands available: make help"

# Promote first user to admin with better error handling
admin: db-up
	@echo "👑 Setting up admin user..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database is not ready. Cannot promote user to admin."; \
		exit 1; \
	fi
	@echo "🔍 Looking for users to promote..."
	@if buffalo task db:promote_admin 2>/dev/null; then \
		echo "✅ User successfully promoted to admin role!"; \
		echo "🎯 You can now access the admin panel at http://127.0.0.1:3001/admin"; \
	else \
		echo "⚠️  No users found to promote. Please:"; \
		echo "   1. Create a user account through the web interface first"; \
		echo "   2. Then run 'make admin' again"; \
	fi

# Run database migrations with better error handling
migrate: db-up
	@echo "🔀 Running database migrations..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database is not ready. Cannot run migrations."; \
		exit 1; \
	fi
	@echo "📊 Checking migration status..."
	@if buffalo pop migrate 2>/dev/null || soda migrate 2>/dev/null; then \
		echo "✅ Migrations completed successfully!"; \
	else \
		echo "❌ Migration failed. Check database connection and migration files."; \
		exit 1; \
	fi

# Start PostgreSQL database with comprehensive checks
db-up:
	@echo "🗄️  Starting PostgreSQL database..."
	@# Check if database is already running
	@if command -v docker-compose >/dev/null 2>&1; then \
		if docker-compose ps 2>/dev/null | grep -q "postgres.*Up\|postgres.*running\|postgres.*healthy"; then \
			echo "✅ Database is already running (Docker)"; \
			if docker-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
				echo "✅ PostgreSQL is ready and accepting connections"; \
			else \
				echo "🔄 Database container is running but PostgreSQL is not ready, restarting..."; \
				docker-compose restart postgres; \
			fi; \
		else \
			echo "🐳 Starting database with Docker Compose..."; \
			docker-compose up -d postgres || (echo "❌ Failed to start database with Docker Compose" && exit 1); \
		fi; \
	elif command -v podman-compose >/dev/null 2>&1; then \
		if podman-compose ps 2>/dev/null | grep -q "postgres.*Up\|postgres.*running\|postgres.*healthy"; then \
			echo "✅ Database is already running (Podman)"; \
			if podman-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
				echo "✅ PostgreSQL is ready and accepting connections"; \
			else \
				echo "🔄 Database container is running but PostgreSQL is not ready, restarting..."; \
				podman-compose restart postgres; \
			fi; \
		else \
			echo "🔷 Starting database with Podman Compose..."; \
			podman-compose up -d postgres || (echo "❌ Failed to start database with Podman Compose" && exit 1); \
		fi; \
	else \
		echo "❌ Neither docker-compose nor podman-compose found."; \
		echo "Please install Docker (recommended) or Podman."; \
		echo "Docker: https://docs.docker.com/get-docker/"; \
		echo "Podman: https://podman.io/getting-started/installation"; \
		exit 1; \
	fi
	@echo "✅ Database container started successfully."

# Stop PostgreSQL database
db-down:
	@echo "⬇️  Stopping PostgreSQL database..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose down || echo "Database was not running."; \
	elif command -v podman-compose >/dev/null 2>&1; then \
		podman-compose down || echo "Database was not running."; \
	else \
		echo "❌ No compose command found."; \
	fi
	@echo "✅ Database stopped."

# Check database status with detailed information
db-status:
	@echo "📊 Database container status:"
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose ps postgres 2>/dev/null || echo "❌ Database container not found (Docker)"; \
		echo ""; \
		echo "📡 Container health:"; \
		if docker-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready and accepting connections"; \
		else \
			echo "❌ PostgreSQL is not ready"; \
		fi; \
	elif command -v podman-compose >/dev/null 2>&1; then \
		podman-compose ps postgres 2>/dev/null || echo "❌ Database container not found (Podman)"; \
		echo ""; \
		echo "📡 Container health:"; \
		if podman-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready and accepting connections"; \
		else \
			echo "❌ PostgreSQL is not ready"; \
		fi; \
	else \
		echo "❌ No compose command found."; \
	fi

# Show database logs
db-logs:
	@echo "📋 Database container logs (last 50 lines):"
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose logs postgres --tail 50 || echo "❌ Cannot access database logs"; \
	elif command -v podman-compose >/dev/null 2>&1; then \
		podman-compose logs postgres --tail 50 || echo "❌ Cannot access database logs"; \
	else \
		echo "❌ No compose command found."; \
	fi

# Reset database with safety confirmations
db-reset: 
	@echo "🔄 Database Reset - This will DELETE ALL DATA!"
	@echo "Are you sure you want to reset the database? [y/N]" && read ans && [ $${ans:-N} = y ]
	@echo "🗄️  Starting database..."
	@$(MAKE) db-up
	@echo "⏳ Waiting for database to be ready..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database failed to start. Cannot reset."; \
		exit 1; \
	fi
	@echo "� Terminating active database connections..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose exec postgres psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='avrnpo_development' AND pid <> pg_backend_pid();" 2>/dev/null || echo "No active connections to terminate"; \
	elif command -v podman-compose >/dev/null 2>&1; then \
		podman-compose exec postgres psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='avrnpo_development' AND pid <> pg_backend_pid();" 2>/dev/null || echo "No active connections to terminate"; \
	fi
	@echo "�🗑️  Dropping development database..."
	@buffalo pop drop -e development 2>/dev/null || echo "Database drop failed (may not exist)"
	@echo "🏗️  Creating development database..."
	@buffalo pop create -e development || (echo "❌ Database create failed" && exit 1)
	@echo "🔀 Running migrations..."
	@buffalo pop migrate -e development || soda migrate || (echo "❌ Migration failed" && exit 1)
	@echo "✅ Database reset complete!"
	@echo "🎯 You can now run 'make dev' to start the development server"

# Run tests with comprehensive setup
test: check-deps db-up
	@echo "🧪 Running test suite with Go test (Buffalo suite)..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database is not ready. Cannot run tests."; \
		exit 1; \
	fi
	@echo "🔄 Setting up test database..."
	@GO_ENV=test soda create -a >/dev/null 2>&1 || true
	@GO_ENV=test soda migrate up >/dev/null 2>&1 || true
	@echo "🏃 Executing tests..."
	@if GO_ENV=test go test ./actions -v -vet=printf; then \
		echo "✅ All tests passed!"; \
	else \
		echo "❌ Some tests failed. Check the output above for details."; \
		exit 1; \
	fi

# Run Buffalo tests quickly (assumes database is already running)
test-fast: check-deps
	@echo "⚡ Running tests (fast mode)..."
	@echo "🏃 Executing tests..."
	@# Run tests and only show output on failure. Passing output is suppressed.
	@HELCIM_PRIVATE_API_KEY=test_key_for_testing GO_ENV=test bash -c '\
		tmp=$$(mktemp); \
		if go test ./actions -vet=printf >"$$tmp" 2>&1; then \
			echo "✅ All tests passed!"; \
			rm -f "$$tmp"; \
		else \
			cat "$$tmp"; rm -f "$$tmp"; exit 1; \
		fi'


# Resilient test command that handles database startup automatically
test-resilient: check-deps
	@echo "🔄 Running resilient test suite..."
	@echo "🔍 Checking if database is running..."
	@if ! docker-compose ps | grep -q "postgres.*Up" 2>/dev/null && ! podman-compose ps | grep -q "postgres.*Up" 2>/dev/null; then \
		echo "🗄️  Database not running, starting it..."; \
		$(MAKE) db-up; \
		sleep 3; \
	else \
		echo "✅ Database is already running"; \
	fi
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database failed to start or become ready. Cannot run tests."; \
		exit 1; \
	fi
	@echo "🔄 Setting up test database..."
	@GO_ENV=test soda create -a >/dev/null 2>&1 || true
	@GO_ENV=test soda migrate up >/dev/null 2>&1 || true
	@echo "🏃 Executing Buffalo tests..."
	@if buffalo test; then \
		echo "✅ All tests passed!"; \
	else \
		echo "❌ Some tests failed. Check the output above for details."; \
		exit 1; \
	fi

# CSRF Integration tests - tests with CSRF middleware enabled
test-integration: check-deps db-up
	@echo "🔒 Running CSRF integration tests..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database is not ready. Cannot run integration tests."; \
		exit 1; \
	fi
	@echo "🔄 Setting up integration test database..."
	@GO_ENV=integration buffalo pop create -a >/dev/null 2>&1 || true
	@GO_ENV=integration buffalo pop migrate up >/dev/null 2>&1 || true
	@echo "🏃 Executing CSRF integration tests..."
	@if GO_ENV=integration go test ./actions -run "TestCSRF" -v; then \
		echo "✅ All CSRF integration tests passed!"; \
		echo "🔒 CSRF middleware is working correctly"; \
	else \
		echo "❌ CSRF integration tests failed. Check the output above for details."; \
		echo "💡 These tests verify that CSRF protection works with real middleware"; \
		exit 1; \
	fi

# Template validation with variable checking
validate-templates:
	@echo "🔍 Running enhanced template validation..."
	@go run scripts/validate-templates-fast.go

# Template validation with verbose output
validate-templates-verbose:
	@echo "🔍 Running enhanced template validation (verbose)..."
	@go run scripts/validate-templates-fast.go --verbose

# Build the application for production with validation
build: validate-templates
	@echo "🔨 Building application for production..."
	@if buffalo build; then \
		echo "✅ Build completed successfully!"; \
	else \
		echo "❌ Build failed. Check the output above for errors."; \
		exit 1; \
	fi

# Clear Go build, module, and gopls caches
clean-caches:
	@echo "🧹 Clearing Go and language server caches..."
	@go clean -cache || echo "Go cache already clean"
	@go clean -modcache || echo "Module cache already clean" 
	@echo "💡 If VS Code still shows errors, restart the Go language server:"
	@echo "   Ctrl+Shift+P -> 'Go: Restart Language Server'"
	@echo "✅ Cache cleanup complete!"

# Clean up everything with confirmation
clean:
	@echo "🧹 Cleaning up development environment..."
	@echo "This will stop all services and remove containers. Continue? [y/N]" && read ans && [ $${ans:-N} = y ]
	@echo "🛑 Stopping all services..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose down || echo "Services were not running."; \
		echo "🗑️  Cleaning up containers and volumes..."; \
		docker system prune -f --volumes 2>/dev/null || echo "Cleanup completed with warnings."; \
	elif command -v podman-compose >/dev/null 2>&1; then \
		podman-compose down || echo "Services were not running."; \
		echo "🗑️  Cleaning up containers and volumes..."; \
		podman system prune -f --volumes 2>/dev/null || echo "Cleanup completed with warnings."; \
	else \
		echo "❌ No compose command found."; \
	fi
	@echo "✅ Clean complete!"

# Update all frontend dependencies to latest versions
update-deps:
	@echo "🔄 Updating frontend dependencies to latest versions..."
	@echo ""
	
	# Check for required tools
	@if ! command -v curl >/dev/null 2>&1; then \
		echo "❌ curl is required but not installed."; \
		exit 1; \
	fi
	
	@echo "📦 Checking latest versions..."
	
	# Get latest Quill.js version
	@echo "🔍 Checking Quill.js..."
	@QUILL_VERSION=$$(curl -s "https://registry.npmjs.org/quill/latest" | grep '"version"' | head -1 | sed 's/.*"version":"\([^"]*\)".*/\1/'); \
	echo "   Latest Quill.js version: $$QUILL_VERSION"; \
	echo "   📥 Downloading Quill.js $$QUILL_VERSION..."; \
	curl -s -o public/css/quill.snow.css "https://cdn.jsdelivr.net/npm/quill@$$QUILL_VERSION/dist/quill.snow.css" && \
	curl -s -o public/js/quill.min.js "https://cdn.jsdelivr.net/npm/quill@$$QUILL_VERSION/dist/quill.js" && \
	echo "   ✅ Quill.js updated to $$QUILL_VERSION"
	
	# Get latest HTMX version
	@echo "🔍 Checking HTMX..."
	@HTMX_VERSION=$$(curl -s "https://api.github.com/repos/bigskysoftware/htmx/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": "v\([^"]*\)".*/\1/'); \
	echo "   Latest HTMX version: $$HTMX_VERSION"; \
	echo "   📥 Downloading HTMX $$HTMX_VERSION..."; \
	curl -s -o public/js/htmx.min.js "https://unpkg.com/htmx.org@$$HTMX_VERSION/dist/htmx.min.js" && \
	echo "   ✅ HTMX updated to $$HTMX_VERSION"
	
	# Get latest Pico.css version
	@echo "🔍 Checking Pico.css..."
	@PICO_VERSION=$$(curl -s "https://api.github.com/repos/picocss/pico/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": "v\([^"]*\)".*/\1/'); \
	echo "   Latest Pico.css version: $$PICO_VERSION"; \
	echo "   📥 Downloading Pico.css $$PICO_VERSION..."; \
	curl -s -o public/css/pico.min.css "https://cdn.jsdelivr.net/npm/@picocss/pico@$$PICO_VERSION/css/pico.min.css" && \
	echo "   ✅ Pico.css updated to $$PICO_VERSION"
	
	@echo ""
	@echo "🎉 All frontend dependencies updated successfully!"
	@echo "📝 Updated files:"
	@echo "   - public/css/quill.snow.css"
	@echo "   - public/css/pico.min.css" 
	@echo "   - public/js/quill.min.js"
	@echo "   - public/js/htmx.min.js"
	@echo ""
	@echo "💡 Tip: Restart Buffalo dev server to see changes: make dev"

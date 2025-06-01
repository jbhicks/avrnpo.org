.PHONY: help dev setup db-up db-down db-reset test clean build admin migrate db-status db-logs health check-deps install-deps

# Default target
help:
	@echo "🚀 My Go SaaS Template - Development Commands"
	@echo ""
	@echo "Quick Start:"
	@echo "  setup      - 🔧 Initial setup: start database, run migrations, install deps"
	@echo "  dev        - 🏃 Start database and run Buffalo development server"
	@echo "  admin      - 👑 Promote first user to admin role"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-up      - 🗄️  Start PostgreSQL database with Podman"
	@echo "  db-down    - ⬇️  Stop PostgreSQL database"
	@echo "  db-reset   - 🔄 Reset database (drop, create, migrate)"
	@echo "  db-status  - 📊 Check database container status"
	@echo "  db-logs    - 📋 Show database container logs"
	@echo "  migrate    - 🔀 Run database migrations"
	@echo ""
	@echo "Development:"
	@echo "  test       - 🧪 Run all tests with database"
	@echo "  build      - 🔨 Build the application for production"
	@echo "  health     - 🏥 Check system health (dependencies, database, etc.)"
	@echo "  clean      - 🧹 Stop all services and clean up containers"
	@echo ""
	@echo "Dependencies:"
	@echo "  check-deps - ✅ Check if all required dependencies are installed"
	@echo "  install-deps - 📦 Install missing dependencies (where possible)"

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
		if ! command -v docker-compose >/dev/null 2>&1; then \
			echo "❌ Neither podman-compose nor docker-compose found. Please install Podman or Docker."; \
			error_count=$$((error_count + 1)); \
		else \
			echo "✅ Docker Compose is installed: $$(docker-compose version)"; \
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

# Start database and development server with full health checks
dev: check-deps db-up
	@echo "🔍 Waiting for database to be ready..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database failed to start. Check 'make db-logs' for details."; \
		exit 1; \
	fi
	@echo "🚀 Starting Buffalo development server..."
	@echo "📱 Visit http://127.0.0.1:3000 to see your application"
	@echo "🔥 Hot reload is enabled - changes will be reflected automatically"
	@buffalo dev || (echo "❌ Buffalo failed to start. Check the output above for errors." && exit 1)

# Initial setup with comprehensive checks
setup: check-deps db-up migrate
	@echo "🎉 Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make dev' to start the development server"
	@echo "  2. Visit http://127.0.0.1:3000 to see your application"
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
		echo "🎯 You can now access the admin panel at http://127.0.0.1:3000/admin"; \
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
	@if ! command -v podman-compose >/dev/null 2>&1; then \
		if ! command -v docker-compose >/dev/null 2>&1; then \
			echo "❌ Neither podman-compose nor docker-compose found."; \
			echo "Please install Podman (recommended) or Docker."; \
			echo "Podman: https://podman.io/getting-started/installation"; \
			echo "Docker: https://docs.docker.com/get-docker/"; \
			exit 1; \
		else \
			echo "🐳 Using Docker Compose..."; \
			docker-compose up -d postgres || (echo "❌ Failed to start database with Docker Compose" && exit 1); \
		fi; \
	else \
		echo "🔷 Using Podman Compose..."; \
		podman-compose up -d postgres || (echo "❌ Failed to start database with Podman Compose" && exit 1); \
	fi
	@echo "✅ Database container started successfully."

# Stop PostgreSQL database
db-down:
	@echo "⬇️  Stopping PostgreSQL database..."
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose down || echo "Database was not running."; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose down || echo "Database was not running."; \
	else \
		echo "❌ No compose command found."; \
	fi
	@echo "✅ Database stopped."

# Check database status with detailed information
db-status:
	@echo "📊 Database container status:"
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose ps postgres 2>/dev/null || echo "❌ Database container not found (Podman)"; \
		echo ""; \
		echo "📡 Container health:"; \
		if podman-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready and accepting connections"; \
		else \
			echo "❌ PostgreSQL is not ready"; \
		fi; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose ps postgres 2>/dev/null || echo "❌ Database container not found (Docker)"; \
		echo ""; \
		echo "📡 Container health:"; \
		if docker-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
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
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose logs postgres --tail 50 || echo "❌ Cannot access database logs"; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose logs postgres --tail 50 || echo "❌ Cannot access database logs"; \
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
	@echo "🗑️  Dropping development database..."
	@buffalo pop drop -e development 2>/dev/null || echo "Database drop failed (may not exist)"
	@echo "🏗️  Creating development database..."
	@buffalo pop create -e development || (echo "❌ Database create failed" && exit 1)
	@echo "🔀 Running migrations..."
	@buffalo pop migrate -e development || soda migrate || (echo "❌ Migration failed" && exit 1)
	@echo "✅ Database reset complete!"
	@echo "🎯 You can now run 'make dev' to start the development server"

# Run tests with comprehensive setup
test: check-deps db-up
	@echo "🧪 Running test suite..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database is not ready. Cannot run tests."; \
		exit 1; \
	fi
	@echo "🔄 Ensuring test database is ready..."
	@buffalo pop create -e test >/dev/null 2>&1 || true
	@buffalo pop migrate -e test >/dev/null 2>&1 || true
	@echo "🏃 Executing tests..."
	@if go test ./...; then \
		echo "✅ All tests passed!"; \
	else \
		echo "❌ Some tests failed. Check the output above for details."; \
		exit 1; \
	fi

# System health check
health: check-deps
	@echo "🏥 System Health Check"
	@echo ""
	@echo "📊 Database Status:"
	@$(MAKE) db-status 2>/dev/null || echo "❌ Database health check failed"
	@echo ""
	@echo "🔍 Buffalo Status:"
	@if pgrep -f "buffalo.*dev" >/dev/null; then \
		echo "✅ Buffalo development server is running"; \
		echo "🌐 Application should be available at http://127.0.0.1:3000"; \
	else \
		echo "❌ Buffalo development server is not running"; \
		echo "💡 Run 'make dev' to start the development server"; \
	fi
	@echo ""
	@echo "📁 Project Structure:"
	@if [ -f "go.mod" ]; then echo "✅ go.mod exists"; else echo "❌ go.mod missing"; fi
	@if [ -f "database.yml" ]; then echo "✅ database.yml exists"; else echo "❌ database.yml missing"; fi
	@if [ -d "templates" ]; then echo "✅ templates directory exists"; else echo "❌ templates directory missing"; fi
	@if [ -d "actions" ]; then echo "✅ actions directory exists"; else echo "❌ actions directory missing"; fi

# Clean up everything with confirmation
clean:
	@echo "🧹 Cleaning up development environment..."
	@echo "This will stop all services and remove containers. Continue? [y/N]" && read ans && [ $${ans:-N} = y ]
	@echo "🛑 Stopping all services..."
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose down || echo "Services were not running."; \
		echo "🗑️  Cleaning up containers and volumes..."; \
		podman system prune -f --volumes 2>/dev/null || echo "Cleanup completed with warnings."; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose down || echo "Services were not running."; \
		echo "🗑️  Cleaning up containers and volumes..."; \
		docker system prune -f --volumes 2>/dev/null || echo "Cleanup completed with warnings."; \
	else \
		echo "❌ No compose command found."; \
	fi
	@echo "✅ Clean complete!"

# Build the application with version info
build:
	@echo "🔨 Building application..."
	@echo "📦 Compiling Go binary..."
	@if buffalo build; then \
		echo "✅ Build complete!"; \
		echo "📁 Binary created: bin/my-go-saas-template"; \
		echo "🚀 Run with: ./bin/my-go-saas-template"; \
	else \
		echo "❌ Build failed. Check the output above for errors."; \
		exit 1; \
	fi

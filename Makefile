.PHONY: help dev setup db-up db-down db-reset test clean build

# Default target
help:
	@echo "Available commands:"
	@echo "  dev        - Start database and run Buffalo development server"
	@echo "  setup      - Initial setup: start database and run migrations"
	@echo "  db-up      - Start PostgreSQL database with Docker"
	@echo "  db-down    - Stop PostgreSQL database"
	@echo "  db-reset   - Reset database (drop, create, migrate)"
	@echo "  test       - Run all tests"
	@echo "  clean      - Stop all services and clean up"
	@echo "  build      - Build the application"

# Start database and development server
dev: db-up
	@echo "Waiting for database to be ready..."
	@./scripts/wait-for-postgres.sh
	@echo "Starting Buffalo development server..."
	buffalo dev

# Initial setup
setup: db-up
	@echo "Waiting for database to be ready..."
	@./scripts/wait-for-postgres.sh
	@echo "Running database migrations..."
	buffalo pop migrate
	@echo "Setup complete! Run 'make dev' to start development server."

# Start PostgreSQL database
db-up:
	@echo "Starting PostgreSQL database..."
	@podman-compose up -d postgres
	@echo "Database starting..."

# Stop PostgreSQL database
db-down:
	@echo "Stopping PostgreSQL database..."
	podman-compose down

# Reset database
db-reset: db-up
	@echo "Waiting for database to be ready..."
	@./scripts/wait-for-postgres.sh
	@echo "Resetting database..."
	buffalo pop drop -e development
	buffalo pop create -e development
	buffalo pop migrate -e development
	@echo "Database reset complete!"

# Run tests
test: db-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Running tests..."
	go test ./...

# Clean up everything
clean:
	@echo "Stopping all services..."
	docker-compose down
	@echo "Cleaning up..."
	@docker system prune -f
	@echo "Clean complete!"

# Build the application
build:
	@echo "Building application..."
	buffalo build
	@echo "Build complete!"

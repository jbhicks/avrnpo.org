.PHONY: help dev stop build clean clean-caches update-deps install

# Default target
help:
	@echo "🚀 AVRNPO - Development Commands"
	@echo ""
	@echo "Recommended: Run './dev.sh' for full dev workflow"
	@echo ""

	@echo ""
	@echo "Quick Start:"
	@echo "  install    - 📦 Install development tools (Air, Templ)"
	@echo "  dev        - 🏃 Start PocketBase development server (background)"
	@echo "  stop       - 🛑 Stop development server"
	@echo "  build      - 🔨 Build the application for production"
	@echo ""
	@echo "Development:"
	@echo "  clean-caches    - 🧹 Clear Go build, module, and gopls caches"
	@echo ""
	@echo "Dependencies:"
	@echo "  update-deps - 📦 Update frontend dependencies (HTMX, Pico.css)"



# Install development tools
install:
	@echo "📦 Installing development tools..."
	@echo ""
	@echo "🔥 Installing Air (hot reload)..."
	@go install github.com/air-verse/air@latest
	@echo "✅ Air installed"
	@echo ""
	@echo "📝 Installing Templ (template engine)..."
	@go install github.com/a-h/templ/cmd/templ@latest
	@echo "✅ Templ installed"
	@echo ""
	@echo "📥 Downloading Go module dependencies..."
	@go mod download
	@echo "✅ Go modules downloaded"
	@echo ""
	@echo "🎉 All development tools installed successfully!"
	@echo ""
	@echo "💡 Next steps:"
	@echo "   1. Copy .env.example to .env and configure"
	@echo "   2. Run 'make dev' to start the development server"

# Start database and development server with auto-reload
dev:
	@echo "🏃 Starting PocketBase development environment with auto-reload..."
	@echo "🔥 Air will watch for changes and rebuild automatically"
	@echo "📱 Visit http://127.0.0.1:8090 to see your application"
	@echo "📱 Admin UI: http://127.0.0.1:8090/_/ (first run will prompt for admin creation)"
	@echo "📋 Logs: /tmp/avrnpo-dev.log"
	@echo ""
	@air > /tmp/avrnpo-dev.log 2>&1 & echo $$! > /tmp/avrnpo-dev.pid
	@sleep 2
	@if kill -0 $$(cat /tmp/avrnpo-dev.pid 2>/dev/null) 2>/dev/null; then \
		echo "✅ Server started successfully (PID: $$(cat /tmp/avrnpo-dev.pid))"; \
		echo "📋 View logs: tail -f /tmp/avrnpo-dev.log"; \
		echo "🛑 Stop server: make stop"; \
	else \
		echo "❌ Air failed to start. Check logs: tail /tmp/avrnpo-dev.log"; \
		exit 1; \
	fi

# Stop development server
stop:
	@echo "🛑 Stopping development server..."
	@if [ -f /tmp/avrnpo-dev.pid ]; then \
		kill $$(cat /tmp/avrnpo-dev.pid) 2>/dev/null && \
		rm /tmp/avrnpo-dev.pid && \
		echo "✅ Server stopped"; \
	else \
		echo "⚠️  No PID file found. Trying to kill by process name..."; \
		pkill -f "avrnpo serve" && echo "✅ Server stopped" || echo "❌ No server process found"; \
	fi



# Template validation with variable checking (deprecated - templ used instead)
validate-templates:
	@echo "⚠️  Template validation is not needed with templ"

# Template validation with verbose output (deprecated - templ used instead)
validate-templates-verbose:
	@echo "⚠️  Template validation is not needed with templ"

# Build the application for production with validation
build:
	@echo "🔨 Building PocketBase application for production..."
	@if go build -o avrnpo; then \
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
	@echo "🧹 Cleaning up build artifacts..."
	@rm -f avrnpo avrnpo.org app test-binary
	@echo "✅ Clean complete!"

# Update all frontend dependencies to latest versions
update-deps:
	@echo "🔄 Updating frontend dependencies to latest versions..."
	@echo ""
	
	@if ! command -v curl >/dev/null 2>&1; then \
		echo "❌ curl is required but not installed."; \
		exit 1; \
	fi
	
	@echo "📦 Checking latest versions..."
	
	@echo "🔍 Checking HTMX..."
	@HTMX_VERSION=$$(curl -s "https://api.github.com/repos/bigskysoftware/htmx/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": "v\([^"]*\)".*/\1/'); \
	echo "   Latest HTMX version: $$HTMX_VERSION"; \
	echo "   📥 Downloading HTMX $$HTMX_VERSION..."; \
	curl -s -o public/assets/js/htmx.min.js "https://unpkg.com/htmx.org@$$HTMX_VERSION/dist/htmx.min.js" && \
	echo "   ✅ HTMX updated to $$HTMX_VERSION"
	
	@echo "🔍 Checking Pico.css..."
	@PICO_VERSION=$$(curl -s "https://api.github.com/repos/picocss/pico/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": "v\([^"]*\)".*/\1/'); \
	echo "   Latest Pico.css version: $$PICO_VERSION"; \
	echo "   📥 Downloading Pico.css $$PICO_VERSION..."; \
	curl -s -o public/assets/css/pico.min.css "https://cdn.jsdelivr.net/npm/@picocss/pico@$$PICO_VERSION/css/pico.min.css" && \
	echo "   ✅ Pico.css updated to $$PICO_VERSION"
	
	@echo ""
	@echo "🎉 All frontend dependencies updated successfully!"
	@echo "📝 Updated files:"
	@echo "   - public/assets/css/pico.min.css" 
	@echo "   - public/assets/js/htmx.min.js"
	@echo ""
	@echo "💡 Tip: Restart development server to see changes: make dev"

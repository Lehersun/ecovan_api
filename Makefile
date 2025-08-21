# Get GOPATH
GOPATH := $(shell go env GOPATH)

.PHONY: help tools lint fmt test test-integration build run clean db test-db env-setup dev dev-stop dev-reset

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Install development tools
tools: ## Install development tools
	@echo "Installing development tools..."
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully"

# Run linter
lint: ## Run linter
	@echo "Running linter..."
	$(GOPATH)/bin/golangci-lint run ./...

# Format code
fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...

# Run tests
test: ## Run unit tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report generated: coverage.html"

# Run integration tests
test-integration: test-db ## Run integration tests
	@echo "Running integration tests..."
	@if [ -f .env.test ]; then \
		export $$(cat .env.test | grep -v '^#' | xargs); \
	fi; \
	go test -tags=integration -v ./...

# Build the application
build: ## Build the application
	@echo "Building application..."
	go build -o bin/eco-van-api ./cmd/api

# Run the application
run: db ## Run the application
	@echo "Running application..."
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | xargs); \
	fi; \
	go run ./cmd/api

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache

# Development database
db: ## Start development database
	@echo "Starting development database..."
	@if [ "$(shell docker ps -q -f name=eco-van-db)" ]; then \
		echo "Development database already running"; \
	else \
		docker run --name eco-van-db \
			-e POSTGRES_DB=eco_van_db \
			-e POSTGRES_USER=app \
			-e POSTGRES_PASSWORD=app \
			-p 5432:5432 \
			-d postgres:16-alpine; \
		echo "Waiting for database to be ready..."; \
		sleep 5; \
		echo "Applying schema..."; \
		psql "postgres://app:app@localhost:5432/eco_van_db?sslmode=disable" -f db/migrations/001_initial_schema.sql; \
		echo "Database ready at postgres://app:app@localhost:5432/eco_van_db"; \
	fi

# Test database
test-db: ## Start test database
	@echo "Starting test database..."
	@if [ "$(shell docker ps -q -f name=eco-van-test-db)" ]; then \
		echo "Test database already running"; \
	else \
		docker run --name eco-van-test-db \
			-e POSTGRES_DB=waste_test \
			-e POSTGRES_USER=app \
			-e POSTGRES_PASSWORD=app \
			-p 5433:5432 \
			-d postgres:16-alpine; \
		echo "Waiting for database to be ready..."; \
		sleep 5; \
		echo "Applying schema..."; \
		psql "postgres://app:app@localhost:5433/waste_test?sslmode=disable" -f db/migrations/001_initial_schema.sql; \
		echo "Test database ready at postgres://app:app@localhost:5433/waste_test"; \
	fi

# Stop databases
db-stop: ## Stop development database
	@if [ "$(shell docker ps -q -f name=eco-van-db)" ]; then \
		docker stop eco-van-db && docker rm eco-van-db; \
		echo "Development database stopped"; \
	else \
		echo "Development database not running"; \
	fi

# Complete development environment setup and start
dev: db ## Complete development setup: database + migrations + admin seed + application
	@echo "🚀 Starting complete development environment..."
	@echo "✅ Database: Running and migrated"
	@echo "✅ Admin user: Will be seeded on first application start"
	@echo "✅ Application: Starting in background..."
	@echo ""
	@echo "📋 Development environment includes:"
	@echo "   • PostgreSQL database (port 5432)"
	@echo "   • All migrations applied (comprehensive schema)"
	@echo "   • Admin user auto-seeding"
	@echo "   • Application server (port 8080) running in background"
	@echo "   • Environment variables from .env"
	@echo ""
	@echo "🔐 Admin credentials:"
	@echo "   • Email: admin@example.com"
	@echo "   • Password: from ADMIN_PASSWORD env var (or default: admin123456)"
	@echo ""
	@echo "🌐 Application endpoints:"
	@echo "   • Health: http://localhost:8080/healthz"
	@echo "   • Ready: http://localhost:8080/readyz"
	@echo "   • API: http://localhost:8080/api/v1"
	@echo "   • Login: http://localhost:8080/api/v1/auth/login"
	@echo ""
	@if [ -f .env ]; then \
		echo "📁 Loading environment from .env file..."; \
		export $$(cat .env | grep -v '^#' | xargs); \
		echo "🚀 Starting application in background..."; \
		screen -dmS eco-van-api bash -c "go run ./cmd/api"; \
		echo "✅ Application started in background (screen session: eco-van-api)"; \
		echo "💡 Use 'screen -r eco-van-api' to attach to the session"; \
		echo "💡 Use 'make dev-stop' to stop the development environment"; \
	else \
		echo "⚠️  Warning: .env file not found. Using default environment variables."; \
		echo "💡 Run 'make env-setup' to create environment files."; \
		echo "🚀 Starting application in background..."; \
		screen -dmS eco-van-api bash -c "go run ./cmd/api"; \
		echo "✅ Application started in background (screen session: eco-van-api)"; \
	fi

# Stop development environment
dev-stop: ## Stop development environment (database + application)
	@echo "🛑 Stopping development environment..."
	@echo "📱 Stopping application..."
	@screen -S eco-van-api -X quit 2>/dev/null || echo "No application screen session found"
	@echo "🗄️  Stopping development database..."
	@make db-stop
	@echo "✅ Development environment stopped"

# Reset development environment (fresh start)
dev-reset: dev-stop db build dev## Reset development environment (stop + fresh start)
	@echo "🔄 Resetting development environment..."
	@echo "🗑️  Removing old database container..."
	@make db-stop
	@echo "🔫 Killing any existing application processes on port 8080..."
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No processes found on port 8080"
	@echo "🚀 Starting fresh development environment..."
	@make
	@make dev

test-db-stop: ## Stop test database
	@if [ "$(shell docker ps -q -f name=eco-van-test-db)" ]; then \
		docker stop eco-van-test-db && docker rm eco-van-test-db; \
		echo "Test database stopped"; \
	else \
		echo "Test database not running"; \
	fi
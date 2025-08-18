# Get GOPATH
GOPATH := $(shell go env GOPATH)

.PHONY: help tools lint test test-integration test-integration-race build run migrate-up migrate-down gen clean test-db test-db-stop test-db-reset db db-stop db-reset-docker env-setup

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Install development tools
tools: ## Install development tools
	@echo "Installing development tools..."
	go mod tidy
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v2@latest
	@echo "Tools installed successfully"

# Run linter
lint: ## Run linter
	@echo "Running linter..."
	$(GOPATH)/bin/golangci-lint run ./...

# Run tests
test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report generated: coverage.html"

# Run integration tests
test-integration: test-db ## Run integration tests (depends on test database)
	@echo "Running integration tests..."
	@if [ -f .env.test ]; then \
		echo "Loading test environment from .env.test file..."; \
		export $$(cat .env.test | grep -v '^#' | xargs); \
		go test -tags=integration -v ./...; \
	else \
		echo "Warning: .env.test file not found. Using default test environment variables."; \
		go test -tags=integration -v ./...; \
	fi

# Run integration tests with race detection
test-integration-race: test-db ## Run integration tests with race detection (depends on test database)
	@echo "Running integration tests with race detection..."
	@if [ -f .env.test ]; then \
		echo "Loading test environment from .env.test file..."; \
		export $$(cat .env.test | grep -v '^#' | xargs); \
		go test -tags=integration -race -v ./...; \
	else \
		echo "Warning: .env.test file not found. Using default test environment variables."; \
		go test -tags=integration -race -v ./...; \
	fi

# Build the application
build: ## Build the application
	@echo "Building application..."
	go build -o bin/eco-van-api ./cmd/api

# Run the application
run: db ## Run the application (depends on database)
	@echo "Running application..."
	@if [ -f .env ]; then \
		echo "Loading environment from .env file..."; \
		export $$(cat .env | grep -v '^#' | xargs); \
		go run ./cmd/api; \
	else \
		echo "Warning: .env file not found. Using default environment variables."; \
		go run ./cmd/api; \
	fi

# Run migrations up
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	$(GOPATH)/bin/migrate -path db/migrations -database "postgres://localhost:5432/eco_van_db?sslmode=disable" up

# Run migrations down
migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	$(GOPATH)/bin/migrate -path db/migrations -database "postgres://localhost:5432/eco_van_db?sslmode=disable" down

# Generate code (sqlc, mocks, etc.)
gen: ## Generate code using sqlc and mockery
	@echo "Generating code..."
	$(GOPATH)/bin/sqlc generate
	$(GOPATH)/bin/mockery --all --output=internal/mocks --outpkg=mocks
	@echo "Code generation completed"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache

# Environment setup
env-setup: ## Setup environment files (.env and .env.test)
	@echo "Setting up environment files..."
	@./setup-env.sh

# Development setup
dev-setup: tools gen env-setup ## Setup development environment
	@echo "Development environment setup completed"

# Pre-commit checks
pre-commit: lint test ## Run pre-commit checks
	@echo "Pre-commit checks completed successfully"

# Install dependencies
deps: ## Install Go dependencies
	@echo "Installing Go dependencies..."
	go mod download
	go mod verify

# Format code
fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...
	$(GOPATH)/bin/goimports -w .

# Vet code
vet: ## Vet Go code
	@echo "Vetting Go code..."
	go vet ./...

# Security check
security: ## Run security checks
	@echo "Running security checks..."
	gosec ./...

# Benchmark tests
bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./...

# Docker operations
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t eco-van-api .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 eco-van-api

# Database operations
db-create: ## Create database
	@echo "Creating database..."
	createdb eco_van_db

db-drop: ## Drop database
	@echo "Dropping database..."
	dropdb eco_van_db

db-reset: db-drop db-create migrate-up ## Reset database
	@echo "Database reset completed"

# Test database (for integration tests)
test-db: ## Create PostgreSQL test database using Docker
	@echo "Creating PostgreSQL test database..."
	@if [ "$(shell docker ps -q -f name=eco-van-test-db)" ]; then \
		echo "Test database container already running, reinitializing..."; \
		make test-db-stop; \
	fi
	docker run --name eco-van-test-db \
		-e POSTGRES_DB=waste_test \
		-e POSTGRES_USER=app \
		-e POSTGRES_PASSWORD=app \
		-p 5433:5432 \
		-d postgres:16-alpine
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Applying migrations to test database..."
	@psql "postgres://app:app@localhost:5433/waste_test?sslmode=disable" -f migrations/0001_init.sql
	@psql "postgres://app:app@localhost:5433/waste_test?sslmode=disable" -f migrations/0002_create_tables.sql
	@psql "postgres://app:app@localhost:5433/waste_test?sslmode=disable" -f migrations/0002_users.sql
	@echo "Test database created and migrated at postgres://app:app@localhost:5433/waste_test?sslmode=disable"

test-db-stop: ## Stop and remove test database container
	@echo "Stopping test database container..."
	@if [ "$(shell docker ps -q -f name=eco-van-test-db)" ]; then \
		docker stop eco-van-test-db && docker rm eco-van-test-db; \
		echo "Test database container removed"; \
	else \
		echo "Test database container not running"; \
	fi

test-db-reset: test-db-stop test-db ## Reset test database container

# Local development database (for frontend testing)
db: ## Create PostgreSQL database for local development using Docker
	@echo "Creating PostgreSQL database for local development..."
	@if [ "$(shell docker ps -q -f name=eco-van-db)" ]; then \
		echo "Development database container already running"; \
	else \
		docker run --name eco-van-db \
			-e POSTGRES_DB=eco_van_db \
			-e POSTGRES_USER=app \
			-e POSTGRES_PASSWORD=app \
			-p 5432:5432 \
			-d postgres:16-alpine; \
		echo "Waiting for database to be ready..."; \
		sleep 5; \
		echo "Applying migrations to development database..."; \
		psql "postgres://app:app@localhost:5432/eco_van_db?sslmode=disable" -f migrations/0001_init.sql; \
		psql "postgres://app:app@localhost:5432/eco_van_db?sslmode=disable" -f migrations/0002_create_tables.sql; \
		psql "postgres://app:app@localhost:5432/eco_van_db?sslmode=disable" -f migrations/0002_users.sql; \
		echo "Development database created and migrated at postgres://app:app@localhost:5432/eco_van_db?sslmode=disable"; \
	fi

db-stop: ## Stop and remove development database container
	@echo "Stopping development database container..."
	@if [ "$(shell docker ps -q -f name=eco-van-db)" ]; then \
		docker stop eco-van-db && docker rm eco-van-db; \
		echo "Development database container removed"; \
	else \
		echo "Development database container not running"; \
	fi

db-reset-docker: db-stop db ## Reset development database container

# Complete development environment setup and start
dev: db ## Complete development setup: database + migrations + admin seed + application
	@echo "🚀 Starting complete development environment..."
	@echo "✅ Database: Running and migrated"
	@echo "✅ Admin user: Will be seeded on first application start"
	@echo "✅ Application: Starting..."
	@echo ""
	@echo "📋 Development environment includes:"
	@echo "   • PostgreSQL database (port 5432)"
	@echo "   • All migrations applied (init, tables, users)"
	@echo "   • Admin user auto-seeding"
	@echo "   • Application server (port 8080)"
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
		echo "🚀 Starting application..."; \
		go run ./cmd/api; \
	else \
		echo "⚠️  Warning: .env file not found. Using default environment variables."; \
		echo "💡 Run 'make env-setup' to create environment files."; \
		echo "🚀 Starting application..."; \
		go run ./cmd/api; \
	fi

# Stop development environment
dev-stop: ## Stop development environment (database + application)
	@echo "🛑 Stopping development environment..."
	@echo "📱 Stopping application (if running)..."
	@pkill -f "go run ./cmd/api" || echo "No application process found"
	@echo "🗄️  Stopping development database..."
	@make db-stop
	@echo "✅ Development environment stopped"

# Reset development environment (fresh start)
dev-reset: dev-stop ## Reset development environment (stop + fresh start)
	@echo "🔄 Resetting development environment..."
	@echo "🗑️  Removing old database container..."
	@make db-reset-docker
	@echo "🚀 Starting fresh development environment..."
	@make dev

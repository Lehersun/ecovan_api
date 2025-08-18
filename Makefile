# Get GOPATH
GOPATH := $(shell go env GOPATH)

.PHONY: help tools lint test test-integration test-integration-race build run migrate-up migrate-down gen clean

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
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -tags=integration -v ./...

# Run integration tests with race detection
test-integration-race: ## Run integration tests with race detection
	@echo "Running integration tests with race detection..."
	go test -tags=integration -race -v ./...

# Build the application
build: ## Build the application
	@echo "Building application..."
	go build -o bin/eco-van-api ./cmd/api

# Run the application
run: ## Run the application
	@echo "Running application..."
	go run ./cmd/api

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

# Development setup
dev-setup: tools gen ## Setup development environment
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

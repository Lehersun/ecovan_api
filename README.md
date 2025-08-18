# Eco Van API

Backend API for the Eco Van waste management system ("Спецэкогруз").

## Features

- **Clients Management** - Company and object management
- **Equipment Management** - Waste bins and containers
- **Transport Management** - Vehicle fleet management
- **Drivers Management** - Personnel management
- **Orders Management** - Waste collection request processing
- **Photo Management** - File upload and storage

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL
- **ORM**: SQLC (SQL Compiler)
- **Migrations**: golang-migrate
- **Linting**: golangci-lint
- **Testing**: Go testing package
- **Mocking**: Mockery

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Make (for build automation)

## Local Development

### 1. Clone and Setup

```bash
git clone <repository-url>
cd eco-van-api
```

### 2. Environment Setup (Recommended)

```bash
# Setup environment files (.env and .env.test)
make env-setup

# Or complete development setup (tools + environment)
make dev-setup
```

This creates:
- `.env` - Local development configuration
- `.env.test` - Integration test configuration

### 3. Install Development Tools

```bash
make tools
```

This will install:
- `sqlc` - SQL compiler for type-safe database queries
- `migrate` - Database migration tool
- `golangci-lint` - Linting tool
- `mockery` - Mock generation tool

### 4. Database Setup

#### **Option A: Docker (Recommended)**
```bash
# Start development database with automatic migrations
make db

# Start test database with automatic migrations
make test-db
```

#### **Option B: Manual PostgreSQL Setup**
```bash
# Create database
make db-create

# Run migrations
make migrate-up
```

**Note**: Update the database connection string in the Makefile if your PostgreSQL setup differs.

### 5. Install Dependencies

```bash
make deps
```

### 6. Generate Code

```bash
make gen
```

This generates:
- SQLC code from SQL queries
- Mock interfaces for testing

### 7. Run the Application

```bash
# Development mode (automatically starts DB + loads .env)
make run

# Or build and run
make build
./bin/eco-van-api
```

## Development Commands

### Code Quality

```bash
# Run linter
make lint

# Format code
make fmt

# Vet code
make vet

# Run tests
make test

# Run security checks
make security

# Run benchmarks
make bench
```

### Database Operations

#### **Docker-based (Recommended)**
```bash
# Development database
make db                    # Start development DB with migrations
make db-stop              # Stop development DB container
make db-reset-docker      # Reset development DB container

# Test database
make test-db              # Start test DB with migrations
make test-db-stop         # Stop test DB container
make test-db-reset        # Reset test DB container
```

#### **Manual PostgreSQL**
```bash
# Run migrations up
make migrate-up

# Run migrations down
make migrate-down

# Reset database (drop, create, migrate)
make db-reset
```

### Build Operations

```bash
# Build application
make build

# Clean build artifacts
make clean

# Docker operations
make docker-build
make docker-run
```

## Enhanced Makefile Commands

### **Quick Start Commands**
```bash
# Complete development setup (tools + environment + code generation)
make dev-setup

# Start application with automatic database setup
make run

# Run integration tests with automatic test database setup
make test-integration
```

### **Database Management**
```bash
# Development database (port 5432)
make db              # Start with migrations
make db-stop         # Stop container
make db-reset-docker # Reset container

# Test database (port 5433)
make test-db         # Start with migrations
make test-db-stop    # Stop container
make test-db-reset   # Reset container
```

### **Environment Management**
```bash
# Setup environment files
make env-setup       # Create .env and .env.test

# View all available commands
make help
```

### **Integration Testing**
```bash
# Run integration tests (automatically manages test database)
make test-integration      # Normal mode
make test-integration-race # With race detection
```

**Features:**
- ✅ Automatic database container management
- ✅ Automatic migration application
- ✅ Environment variable loading from files
- ✅ Container reinitialization for fresh test data
- ✅ Port separation (dev: 5432, test: 5433)
- ✅ Transaction-per-test isolation

### Pre-commit Checks

```bash
# Run all pre-commit checks
make pre-commit
```

## Project Structure

```
eco-van-api/
├── cmd/
│   └── api/           # Application entry point
├── internal/           # Private application code
│   ├── api/           # HTTP handlers
│   ├── config/        # Configuration
│   ├── database/      # Database layer
│   ├── models/        # Data models
│   ├── services/      # Business logic
│   └── mocks/         # Generated mocks
├── db/                # Database related files
│   ├── migrations/    # Database migrations
│   ├── queries/       # SQL queries for SQLC
│   └── sqlc.yaml      # SQLC configuration
├── pkg/               # Public packages
├── scripts/           # Build and deployment scripts
├── Makefile           # Build automation
├── tools.go           # Tool dependencies
└── README.md          # This file
```

## Configuration

The application uses environment variables for configuration. Two environment files are supported:

### **`.env` - Local Development**
```env
# Server Configuration
HTTP_ADDR=:8080
HTTP_READ_TIMEOUT=15s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
HTTP_MAX_BODY=10485760
CORS_ORIGINS=*

# Database Configuration
DB_DSN=postgres://app:app@localhost:5432/eco_van_db?sslmode=disable
DB_MAX_CONNS=10
DB_MIN_CONNS=2
DB_MAX_CONN_LIFETIME=30m
DB_MAX_CONN_IDLE=10m

# Authentication
JWT_SECRET=your-secret-key-change-in-production
ACCESS_TTL=15m
REFRESH_TTL=720h

# Telemetry
LOG_LEVEL=info
OTLP_ENDPOINT=
ENABLE_METRICS=true
ENABLE_TRACING=false

# Photos
PHOTOS_DIR=/photos
```

### **`.env.test` - Integration Tests**
```env
# Test Database Configuration
DB_DSN=postgres://app:app@localhost:5433/waste_test?sslmode=disable
DB_MAX_CONNS=5
DB_MIN_CONNS=1
DB_MAX_CONN_LIFETIME=5m
DB_MAX_CONN_IDLE=1m

# Test Authentication
JWT_SECRET=test-secret-key

# Test Telemetry
LOG_LEVEL=warn
ENABLE_METRICS=false
ENABLE_TRACING=false
```

### **Automatic Setup**
```bash
# Create both environment files
make env-setup

# Or include in complete setup
make dev-setup
```

## Testing

### Unit Tests
```bash
# Run all tests
make test

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test ./internal/services/...

# Run tests with race detection
go test -race ./...
```

### Integration Tests
```bash
# Run integration tests (automatically starts test DB + loads .env.test)
make test-integration

# Run integration tests with race detection
make test-integration-race

# View test coverage
go tool cover -html=coverage.out
```

**Note**: Integration tests use ephemeral PostgreSQL containers and automatically:
- Start a fresh test database
- Apply all migrations
- Run tests in isolated transactions
- Clean up containers after completion

## Linting

```bash
# Run linter
make lint

# Auto-fix some issues
golangci-lint run --fix ./...
```

## Code Generation

```bash
# Generate SQLC code
sqlc generate

# Generate mocks
mockery --all --output=internal/mocks --outpkg=mocks
```

## Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir db/migrations -seq migration_name

# Run migrations up
make migrate-up

# Run migrations down
make migrate-down

# Check migration status
migrate -path db/migrations -database "postgres://localhost:5432/eco_van_db?sslmode=disable" version
```

## Complete Development Workflow

### **First Time Setup**
```bash
# 1. Clone and navigate
git clone <repository-url>
cd eco-van-api

# 2. Complete setup (tools + environment + code generation)
make dev-setup

# 3. Start development database
make db

# 4. Start application
make run
```

### **Daily Development**
```bash
# Start application (automatically starts DB if needed)
make run

# Run tests
make test

# Run integration tests (automatically starts test DB)
make test-integration

# Stop development database when done
make db-stop
```

### **Integration Testing**
```bash
# Run integration tests (fresh test database each time)
make test-integration

# Run with race detection
make test-integration-race

# Clean up test database
make test-db-stop
```

## Troubleshooting

### Common Issues

1. **Tools not found**: Run `make tools` to install development tools
2. **Database connection failed**: 
   - For Docker: Run `make db` or `make test-db`
   - For manual: Check PostgreSQL is running and credentials are correct
3. **Migrations fail**: 
   - For Docker: Containers automatically apply migrations
   - For manual: Ensure database exists and user has proper permissions
4. **Code generation fails**: Check SQLC configuration and SQL syntax
5. **Port conflicts**: 
   - Development DB uses port 5432
   - Test DB uses port 5433
   - Use `make db-stop` and `make test-db-stop` to free ports

### Getting Help

- Check the Makefile targets: `make help`
- Review error messages and logs
- Ensure all prerequisites are installed
- Check database connection and permissions
- View available commands: `make help`
- Check environment files: `.env` and `.env.test`
- Verify Docker containers: `docker ps`

### Environment File Structure

The project uses two environment files for different purposes:

- **`.env`** - Local development configuration
  - Loaded automatically by `make run`
  - Contains production-like settings
  - Database: `eco_van_db` on port 5432

- **`.env.test`** - Integration test configuration
  - Loaded automatically by `make test-integration`
  - Contains test-specific settings
  - Database: `waste_test` on port 5433

**Note**: Both files are created automatically by `make env-setup` and contain all necessary configuration variables.

## New Features & Enhancements

### **🚀 Enhanced Makefile Commands**
- **Automatic Database Management**: `make db` and `make test-db` now include migrations
- **Environment Integration**: `make run` automatically loads `.env` and starts database
- **Test Automation**: `make test-integration` automatically manages test database
- **Container Reinitialization**: Fresh test data for each integration test run

### **🐳 Docker-based Development**
- **Development Database**: PostgreSQL on port 5432 with automatic migrations
- **Test Database**: PostgreSQL on port 5433 with automatic migrations
- **Container Lifecycle**: Easy start/stop/reset commands
- **Port Separation**: No conflicts between development and testing

### **🔧 Environment Management**
- **Automatic Setup**: `make env-setup` creates configuration files
- **File-based Config**: `.env` for development, `.env.test` for testing
- **Variable Loading**: Automatic environment variable loading in commands

### **🧪 Integration Testing**
- **Ephemeral Databases**: Fresh PostgreSQL containers for each test run
- **Transaction Isolation**: Each test runs in its own transaction
- **Automatic Cleanup**: Containers removed after test completion
- **Race Detection**: Support for concurrent testing

## Contributing

1. Follow Go best practices and conventions
2. Write tests for new functionality
3. Ensure linting passes: `make lint`
4. Run tests before committing: `make test`
5. Run integration tests: `make test-integration`
6. Use conventional commit messages
7. Test with both unit and integration tests

## License

[Add your license information here]

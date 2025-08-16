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

### 2. Install Development Tools

```bash
make tools
```

This will install:
- `sqlc` - SQL compiler for type-safe database queries
- `migrate` - Database migration tool
- `golangci-lint` - Linting tool
- `mockery` - Mock generation tool

### 3. Database Setup

```bash
# Create database
make db-create

# Run migrations
make migrate-up
```

**Note**: Update the database connection string in the Makefile if your PostgreSQL setup differs.

### 4. Install Dependencies

```bash
make deps
```

### 5. Generate Code

```bash
make gen
```

This generates:
- SQLC code from SQL queries
- Mock interfaces for testing

### 6. Run the Application

```bash
# Development mode
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

The application uses environment variables for configuration. Create a `.env` file:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=eco_van_db
DB_USER=postgres
DB_PASSWORD=password
DB_SSL_MODE=disable

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Environment
ENV=development
LOG_LEVEL=debug
```

## Testing

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

## Troubleshooting

### Common Issues

1. **Tools not found**: Run `make tools` to install development tools
2. **Database connection failed**: Check PostgreSQL is running and credentials are correct
3. **Migrations fail**: Ensure database exists and user has proper permissions
4. **Code generation fails**: Check SQLC configuration and SQL syntax

### Getting Help

- Check the Makefile targets: `make help`
- Review error messages and logs
- Ensure all prerequisites are installed
- Check database connection and permissions

## Contributing

1. Follow Go best practices and conventions
2. Write tests for new functionality
3. Ensure linting passes: `make lint`
4. Run tests before committing: `make test`
5. Use conventional commit messages

## License

[Add your license information here]

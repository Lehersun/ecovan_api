# Development Rules

This document outlines the development rules and best practices for the Eco Van API project.

## 🚀 Development Workflow

### 1. Testing Commands
**Always use make commands for testing:**
```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run both tests
make test && make test-integration
```

### 2. Development Environment
**Only start the development environment via make dev:**
```bash
# Start complete development environment
make dev

# Stop development environment
make dev-stop

# Reset development environment (fresh start)
make dev-reset
```

**Never manually start the application or database outside of make commands.**

### 3. Database Management
**Never change the database manually. Only use migrations:**
- All schema changes must go through migration files
- Never run SQL commands directly on the database
- Never modify table structures manually
- All changes must be version controlled in `migrations/` directory

### 4. Pre-Commit Checklist
**Before committing, always:**
1. **Run all tests:**
   ```bash
   make test && make test-integration
   ```

2. **Test development environment with simple curls (success cases only):**
   ```bash
   # Start dev environment
   make dev
   
   # Wait for startup, then test endpoints:
   
   # Health check
   curl -s http://localhost:8080/healthz
   
   # Ready check
   curl -s http://localhost:8080/readyz
   
   # Login (get token)
   curl -s -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"admin123456"}'
   
   # Users endpoint (with token)
   TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"admin123456"}' | \
     grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)
   
   curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/users
   ```

3. **Verify all endpoints return successful responses**

4. **Only commit if all tests pass and all endpoints work**

## 📋 Available Make Commands

### Core Development
- `make dev` - Complete development environment (DB + migrations + app)
- `make dev-stop` - Stop development environment
- `make dev-reset` - Reset and restart development environment

### Testing
- `make test` - Run unit tests
- `make test-integration` - Run integration tests
- `make test-db` - Start test database
- `make test-db-stop` - Stop test database

### Database
- `make db` - Start development database with migrations
- `make db-stop` - Stop development database

### Code Quality
- `make lint` - Run golangci-lint
- `make fmt` - Format Go code
- `make build` - Build the application

## ⚠️ Important Notes

- **Never commit without running tests first**
- **Never commit without testing the dev environment**
- **Never modify the database outside of migrations**
- **Always use make commands for development tasks**
- **Keep the development environment clean and consistent**

## 🔧 Troubleshooting

If something goes wrong:
1. Stop the environment: `make dev-stop`
2. Reset the environment: `make dev-reset`
3. Check logs: `screen -r eco-van-api`
4. Verify database: `make db` then `make test-db`

## 📚 Additional Resources

- Check the Makefile for all available commands
- Review `migrations/` directory for schema changes
- Use `screen -r eco-van-api` to attach to running application
- Environment variables are loaded from `.env` file

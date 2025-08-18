#!/bin/bash

# Setup environment files for eco-van-api

echo "Setting up environment files for eco-van-api..."

# Create .env file for local development
if [ ! -f .env ]; then
    cat > .env << 'EOF'
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

# Admin User (created automatically on first run)
ADMIN_PASSWORD=admin123456

# Telemetry
LOG_LEVEL=info
OTLP_ENDPOINT=
ENABLE_METRICS=true
ENABLE_TRACING=false

# Photos
PHOTOS_DIR=/photos
EOF
    echo "Created .env file for local development"
else
    echo ".env file already exists"
fi

# Create .env.test file for integration tests
if [ ! -f .env.test ]; then
    cat > .env.test << 'EOF'
# Test Database Configuration
DB_DSN=postgres://app:app@localhost:5433/waste_test?sslmode=disable
DB_MAX_CONNS=5
DB_MIN_CONNS=1
DB_MAX_CONN_LIFETIME=5m
DB_MAX_CONN_IDLE=1m

# Test Authentication
JWT_SECRET=test-secret-key

# Test Admin User
ADMIN_PASSWORD=test-admin-password

# Test Telemetry
LOG_LEVEL=warn
ENABLE_METRICS=false
ENABLE_TRACING=false
EOF
    echo "Created .env.test file for integration tests"
else
    echo ".env.test file already exists"
fi

echo ""
echo "Environment files setup complete!"
echo ""
echo "To start the application:"
echo "  make run"
echo ""
echo "To run integration tests:"
echo "  make test-integration"
echo ""
echo "To manage databases:"
echo "  make db          # Start development database"
echo "  make test-db     # Start test database"
echo "  make db-stop     # Stop development database"
echo "  make test-db-stop # Stop test database"

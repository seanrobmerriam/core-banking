# Core Banking Microservices - Go Monorepo

A production-ready microservices architecture for core banking operations built with Go, featuring centralized configuration, structured logging, and PostgreSQL connectivity.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Core Banking Platform                     │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  Customer    │  │   Account    │  │  Transaction     │  │
│  │   Service    │  │   Service    │  │     Service      │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘  │
│         │                 │                    │             │
│         └─────────────────┼────────────────────┘             │
│                           │                                  │
│                    ┌──────▼──────┐                          │
│                    │    NATS     │                          │
│                    │   Message   │                          │
│                    │    Bus      │                          │
│                    └──────┬──────┘                          │
│                           │                                  │
│         ┌─────────────────┼────────────────────┐            │
│         │                 │                    │            │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌──────────▼─────────┐  │
│  │  PostgreSQL │  │   Redis     │  │  External Systems  │  │
│  │  Database   │  │   Cache     │  │  (Payment Gateways)│  │
│  └─────────────┘  └─────────────┘  └────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
core-banking/
├── Makefile                    # Build automation
├── README.md                   # This file
├── docker-compose.yml          # Local development infrastructure
├── go.mod                      # Go module definition
├── .env.example                # Environment template
├── .gitignore                  # Git ignore rules
│
├── pkg/                        # Shared packages
│   ├── config/                 # Configuration management
│   ├── database/               # PostgreSQL connection utilities
│   ├── errors/                 # Custom error types
│   ├── logger/                 # Structured logging
│   └── middleware/             # HTTP middleware
│
└── services/                   # Microservices
    ├── customer-service/       # Customer management
    │   └── cmd/api/
    ├── account-service/        # Account management (placeholder)
    │   └── cmd/api/
    └── transaction-service/    # Transaction processing (placeholder)
        └── cmd/api/
```

## Features

- **Structured Logging**: JSON-formatted logs with request correlation IDs using zerolog
- **Configuration Management**: Environment-based configuration with validation
- **Database Layer**: PostgreSQL connection pooling with health checks
- **Graceful Shutdown**: Proper signal handling for zero-downtime deployments
- **HTTP Middleware**: Request logging, panic recovery, and CORS support
- **Docker Support**: Complete containerized development environment

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15 (via Docker)
- NATS messaging server (via Docker)

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd core-banking
```

### 2. Configure Environment

Copy the example environment file and customize:

```bash
cp .env.example .env
```

Review and modify `.env` according to your local setup:

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
LOG_LEVEL=debug

# PostgreSQL Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=banking_user
DB_PASSWORD=secure_password
DB_NAME=core_banking
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# NATS Messaging
NATS_URL=nats://localhost:4222
```

### 3. Start Infrastructure

```bash
make docker-up
```

This starts PostgreSQL 15 and NATS in Docker containers.

### 4. Build and Run

```bash
# Build all services
make build

# Run customer service (main service)
make run-customer

# In another terminal, run additional services
make run-account
make run-transaction
```

### 5. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","database":"connected","timestamp":"..."}
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make build` | Build all microservices |
| `make run-customer` | Run customer service |
| `make run-account` | Run account service |
| `make run-transaction` | Run transaction service |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make docker-up` | Start Docker containers |
| `make docker-down` | Stop Docker containers |
| `make docker-logs` | View Docker container logs |
| `make clean` | Clean build artifacts |
| `make lint` | Run linter (if golangci-lint installed) |

## API Endpoints

### Customer Service (Port 8080)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check with database status |
| GET | `/api/v1/customers` | List all customers |
| POST | `/api/v1/customers` | Create new customer |
| GET | `/api/v1/customers/:id` | Get customer by ID |
| PUT | `/api/v1/customers/:id` | Update customer |
| DELETE | `/api/v1/customers/:id` | Delete customer |

### Example Usage

```bash
# Health check
curl http://localhost:8080/health

# Create customer
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890"
  }'

# List customers
curl http://localhost:8080/api/v1/customers
```

## Shared Packages

### Configuration (`pkg/config`)

Provides environment-based configuration with type safety:

```go
import "github.com/core-banking/pkg/config"

type Config struct {
    ServerHost string `envconfig:"SERVER_HOST" default:"0.0.0.0"`
    ServerPort int    `envconfig:"SERVER_PORT" default:"8080"`
    // ... more fields
}

cfg, err := config.Load[Config]()
```

### Database (`pkg/database`)

PostgreSQL connection pooling and health checks:

```go
import "github.com/core-banking/pkg/database"

db, err := database.NewDatabase(cfg.Database)
if err != nil {
    return err
}
defer db.Close()

// Health check
if err := db.Ping(ctx); err != nil {
    return err
}
```

### Logger (`pkg/logger`)

Structured logging with request correlation:

```go
import "github.com/core-banking/pkg/logger"

log := logger.New("customer-service")
log.Info().Msg("Service started")

// With context for request correlation
log.Info().Ctx(ctx).Str("customer_id", "123").Msg("Customer created")
```

### Middleware (`pkg/middleware`)

HTTP middleware stack:

```go
router := chi.NewRouter()
router.Use(middleware.RequestLogger)
router.Use(middleware.Recovery)
router.Use(middleware.CORS)
```

### Errors (`pkg/errors`)

Standardized error handling:

```go
import "github.com/core-banking/pkg/errors"

return errors.NewBadRequest("invalid_customer_id", "Customer ID must be a number")
```

## Docker Development

### Manual Docker Commands

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Stop with volume removal
docker-compose down -v
```

### Container Status

```bash
# Check container health
docker-compose ps

# PostgreSQL connection test
docker exec -it core-banking-postgres-1 psql -U banking_user -d core_banking -c "\dt"
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./pkg/config/... -v
go test ./pkg/errors/... -v
```

## Graceful Shutdown

The services implement graceful shutdown handling:

1. Receives SIGINT/SIGTERM signals
2. Stops accepting new connections
3. Waits for active requests to complete (configurable timeout)
4. Closes database connections
5. Logs shutdown completion

```bash
# Test graceful shutdown
# Press Ctrl+C while service is running
# Expected: "Shutting down server gracefully..."
```

## Production Deployment

### Environment Variables

For production, ensure these environment variables are set:

```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
LOG_LEVEL=info
DB_HOST=your-production-db-host
DB_PORT=5432
DB_USER=banking_user
DB_PASSWORD=strong-password
DB_NAME=core_banking
DB_MAX_OPEN_CONNS=50
NATS_URL=nats://nats-cluster:4222
```

### Building Docker Images

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/api ./services/customer-service/cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/api /app/api
EXPOSE 8080
CMD ["/app/api"]
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Ensure all tests pass (`make test`)
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions or issues, please open a GitHub issue or contact the development team.

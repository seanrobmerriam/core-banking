.PHONY: all build test clean docker-up docker-down docker-logs run-customer run-account run-transaction lint test-coverage help

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build -buildvcs=false
GOTEST=$(GOCMD) test
GOMOD=github.com/core-banking
SERVICES=customer-service account-service transaction-service

# Default target
all: help

# Build all services
build: 
	@echo "Building all services..."
	$(GOBUILD) -o bin/customer-service ./services/customer-service/cmd/api
	$(GOBUILD) -o bin/account-service ./services/account-service/cmd/api
	$(GOBUILD) -o bin/transaction-service ./services/transaction-service/cmd/api
	@echo "Build complete! Binaries located in ./bin/"

# Build customer service
build-customer:
	@echo "Building customer service..."
	$(GOBUILD) -o bin/customer-service ./services/customer-service/cmd/api
	@echo "Customer service built successfully!"

# Build account service
build-account:
	@echo "Building account service..."
	$(GOBUILD) -o bin/account-service ./services/account-service/cmd/api
	@echo "Account service built successfully!"

# Build transaction service
build-transaction:
	@echo "Building transaction service..."
	$(GOBUILD) -o bin/transaction-service ./services/transaction-service/cmd/api
	@echo "Transaction service built successfully!"

# Run customer service
run-customer: build
	@echo "Starting customer service..."
	./bin/customer-service

# Run account service
run-account: build
	@echo "Starting account service..."
	./bin/account-service

# Run transaction service
run-transaction: build
	@echo "Starting transaction service..."
	./bin/transaction-service

# Run all services in background
run-all: build
	@echo "Starting all services..."
	./bin/customer-service &
	./bin/account-service &
	./bin/transaction-service &
	@echo "All services started! PID: $$(jobs -p)"

# Run tests
test:
	@echo "Running all tests..."
	$(GOTEST) -v ./pkg/...
	$(GOTEST) -v ./services/...
	@echo "All tests passed!"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./pkg/...
	$(GOTEST) -coverprofile=services_coverage.out ./services/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	$(GOCMD) tool cover -html=services_coverage.out -o services_coverage.html
	@echo "Coverage reports generated: coverage.html, services_coverage.html"

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -race ./pkg/...
	$(GOTEST) -race ./services/...
	@echo "Race detector tests passed!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out services_coverage.out coverage.html services_coverage.html
	rm -f *.test
	@echo "Cleanup complete!"

# Start Docker containers
docker-up:
	@echo "Starting Docker containers (PostgreSQL 15 + NATS)..."
	docker-compose up -d
	@echo "Containers started! Waiting for services to be ready..."
	@sleep 5
	@echo "PostgreSQL is ready on localhost:5432"
	@echo "NATS is ready on localhost:4222"

# Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down
	@echo "Containers stopped!"

# Stop Docker containers and remove volumes
docker-down-v:
	@echo "Stopping Docker containers and removing volumes..."
	docker-compose down -v
	@echo "Containers and volumes removed!"

# View Docker container logs
docker-logs:
	@echo "Showing Docker container logs..."
	docker-compose logs -f

# Follow specific service logs
docker-logs-postgres:
	@echo "Showing PostgreSQL logs..."
	docker-compose logs -f postgres

docker-logs-nats:
	@echo "Showing NATS logs..."
	docker-compose logs -f nats

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./pkg/... ./services/...; \
	else \
		echo "golangci-lint not found. Install from https://golangci-lint.run/"; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOCMD) mod download
	@echo "Dependencies downloaded!"

# Tidy go.mod
tidy:
	@echo "Tidying go.mod..."
	$(GOCMD) mod tidy
	@echo "go.mod tidied!"

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GOCMD) mod verify
	@echo "All dependencies verified!"

# Initialize database schema
db-init:
	@echo "Initializing database schema..."
	@if docker ps | grep -q postgres; then \
		docker exec -i $$(docker-compose ps -q postgres) psql -U banking_user -d core_banking < sql/schema.sql; \
		echo "Database schema initialized!"; \
	else \
		echo "PostgreSQL container not running. Run 'make docker-up' first."; \
	fi

# Show help
help:
	@echo "Core Banking Microservices - Available Commands"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build              - Build all services"
	@echo "  make build-customer     - Build customer service only"
	@echo "  make build-account      - Build account service only"
	@echo "  make build-transaction  - Build transaction service only"
	@echo ""
	@echo "Run Commands:"
	@echo "  make run-customer       - Build and run customer service"
	@echo "  make run-account        - Build and run account service"
	@echo "  make run-transaction    - Build and run transaction service"
	@echo "  make run-all            - Build and run all services"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test               - Run all tests"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo "  make test-race          - Run tests with race detector"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-up          - Start PostgreSQL and NATS containers"
	@echo "  make docker-down        - Stop Docker containers"
	@echo "  make docker-down-v      - Stop containers and remove volumes"
	@echo "  make docker-logs        - View all container logs"
	@echo "  make docker-logs-postgres - View PostgreSQL logs"
	@echo "  make docker-logs-nats   - View NATS logs"
	@echo ""
	@echo "Database Commands:"
	@echo "  make db-init            - Initialize database schema"
	@echo ""
	@echo "Maintenance Commands:"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make deps               - Download dependencies"
	@echo "  make tidy               - Tidy go.mod"
	@echo "  make verify             - Verify dependencies"
	@echo "  make lint               - Run linter"
	@echo "  make help               - Show this help message"

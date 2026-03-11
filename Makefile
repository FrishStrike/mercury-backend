# ─────────────────────────────────────────────────────────────────────────────
# Mercury Backend - Makefile
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: help build run test clean migrate-up migrate-down docker-up docker-down logs

# ─────────────────────────────────────────────────────────────────────────────
# Load environment variables from .env file
# ─────────────────────────────────────────────────────────────────────────────
include .env
export

# ─────────────────────────────────────────────────────────────────────────────
# Build database URL from .env variables
# ─────────────────────────────────────────────────────────────────────────────
DATABASE_URL = postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

MIGRATIONS_PATH = migrations/sql

# ─────────────────────────────────────────────────────────────────────────────
# Help
# ─────────────────────────────────────────────────────────────────────────────
help:
	@echo "=============================================="
	@echo "  Mercury Backend - Available Commands"
	@echo "=============================================="
	@echo ""
	@echo "  Infrastructure:"
	@echo "    make docker-up       - Start infrastructure (Postgres, Redis, Kafka)"
	@echo "    make docker-down     - Stop infrastructure"
	@echo "    make docker-clean    - Stop + remove volumes (full cleanup)"
	@echo "    make logs            - View all logs"
	@echo "    make logs-kafka      - View Kafka logs"
	@echo "    make logs-postgres   - View PostgreSQL logs"
	@echo ""
	@echo "  Database:"
	@echo "    make migrate-up      - Apply all migrations"
	@echo "    make migrate-down    - Rollback all migrations"
	@echo "    make migrate-version - Show current migration version"
	@echo "    make db-shell        - Open psql shell"
	@echo ""
	@echo "  Starting Services"
	@echo "    make run-catalog     - Run catalog service"
	@echo "    make run-order       - Run order service"
	@echo "    make run-payment     - Run payment service"
	@echo ""
	@echo "  Build & Run:"
	@echo "    make build           - Build all services"
	@echo "    make run             - Run all services locally"
	@echo "    make clean           - Clean build artifacts"
	@echo ""
	@echo "  Proto:"
	@echo "    make proto-update    - Update git submodule"
	@echo "    make proto-gen       - Generate files"
	@echo ""
	@echo "  Tests:"
	@echo "    make test            - Run unit tests"
	@echo "    make test-coverage   - Run tests with coverage"
	@echo ""
	@echo "  Code Quality:"
	@echo "    make fmt             - Format Go code"
	@echo "    make lint            - Run linter (requires golangci-lint)"
	@echo ""
	@echo "  Setup:"
	@echo "    make setup           - First-time setup (docker + migrations)"
	@echo ""
	@echo "=============================================="

# ─────────────────────────────────────────────────────────────────────────────
# Docker / Infrastructure
# ─────────────────────────────────────────────────────────────────────────────
docker-up:
	@echo "Starting infrastructure..."
	docker-compose up -d postgres redis zookeeper kafka kafka-ui
	@echo "Waiting for services to be ready..."
	timeout /t 10 /nobreak >nul 2>&1 || sleep 10
	@echo "Infrastructure is up!"
	@echo ""
	@echo "Access points:"
	@echo "   PostgreSQL: localhost:$(POSTGRES_PORT)"
	@echo "   Redis:      localhost:$(REDIS_PORT)"
	@echo "   Kafka:      ${KAFKA_BROKERS_HOST}"
	@echo "   Kafka UI:   http://localhost:$(KAFKA_UI_PORT)"

docker-down:
	@echo "Stopping infrastructure..."
	docker-compose down
	@echo "Infrastructure stopped"

docker-clean:
	@echo "Cleaning infrastructure (removing volumes)..."
	docker-compose down -v
	docker system prune -f
	@echo "Clean complete"

logs:
	docker-compose logs -f

logs-kafka:
	docker-compose logs -f kafka

logs-postgres:
	docker-compose logs -f postgres

# ─────────────────────────────────────────────────────────────────────────────
# Database / Migrations
# ─────────────────────────────────────────────────────────────────────────────
migrate-up:
	@echo "Applying migrations..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up
	@echo "Migrations applied"

migrate-down:
	@echo "Rolling back migrations..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down
	@echo "Migrations rolled back"

migrate-version:
	@echo "Current migration version:"
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version

db-shell:
	@echo "Connecting to PostgreSQL..."
	docker exec -it mercury-postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

# ─────────────────────────────────────────────────────────────────────────────
# Run Services
# ─────────────────────────────────────────────────────────────────────────────
run-catalog:
	@echo "Starting Catalog Service..."
	go run ./cmd/catalog-service

run-order:
	@echo "Starting Order Service..."
	go run ./cmd/order-service

run-payment:
	@echo "Starting Payment Service..."
	go run ./cmd/payment-service

# ─────────────────────────────────────────────────────────────────────────────
# Build & Run
# ─────────────────────────────────────────────────────────────────────────────
build:
	@echo "Building services..."
	go build -o bin/api-gateway ./cmd/api-gateway
	go build -o bin/catalog-service ./cmd/catalog-service
	go build -o bin/order-service ./cmd/order-service
	go build -o bin/payment-service ./cmd/payment-service
	go build -o bin/notification-service ./cmd/notification-service
	@echo "Build complete!"

run:
	@echo "Starting services..."
	start "" go run ./cmd/api-gateway
	start "" go run ./cmd/catalog-service
	start "" go run ./cmd/order-service
	start "" go run ./cmd/payment-service
	start "" go run ./cmd/notification-service
	@echo "Services started in new windows. Press Ctrl+C in each to stop."

clean:
	@echo "Cleaning build artifacts..."
	rmdir /s /q bin 2>nul || true
	go clean -cache
	@echo "Clean complete"


# ─────────────────────────────────────────────────────────────────────────────
# Proto
# ─────────────────────────────────────────────────────────────────────────────
proto-update:
	@echo "Updating proto submodule..."
	git submodule update --remote --merge
	@echo "Proto submodule updated"

proto-gen:
	@echo "Generating proto code..."
	cd api/proto && buf generate && cd ../..
	@echo "Proto generation complete"
	@echo "Proto generation complete"

# ─────────────────────────────────────────────────────────────────────────────
# Tests
# ─────────────────────────────────────────────────────────────────────────────
test:
	@echo "Running unit tests..."
	go test -v -race ./internal/... ./pkg/...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ─────────────────────────────────────────────────────────────────────────────
# Code Quality
# ─────────────────────────────────────────────────────────────────────────────
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete"

lint:
	@echo "Running linter..."
	golangci-lint run ./...
	@echo "Lint complete"

# ─────────────────────────────────────────────────────────────────────────────
# Setup (First Time)
# ─────────────────────────────────────────────────────────────────────────────
setup:
	@echo "Setting up project..."
	go mod tidy
	git submodule update --init --recursive
	$(MAKE) docker-up
	$(MAKE) migrate-up
	@echo ""
	@echo "Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "   1. make build"
	@echo "   2. make run"
	@echo "   3. Open http://localhost:$(KAFKA_UI_PORT) for Kafka UI"
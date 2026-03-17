# =============================================================================
# Seva — Service Marketplace Platform - Root Makefile
# =============================================================================

.PHONY: dev backend-run backend-build backend-test backend-lint \
        web-dev web-build web-preview \
        db-migrate-up db-migrate-down db-migrate-create \
        api-generate docker-up docker-down docker-build \
        setup clean help

# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------
DOCKER_COMPOSE   := docker compose -f infrastructure/docker/docker-compose.yml
BACKEND_DIR      := backend
WEB_DIR          := web
MIGRATIONS_DIR   := $(BACKEND_DIR)/migrations
DATABASE_URL     ?= postgres://seva:seva_dev@localhost:5432/seva?sslmode=disable
GOBIN            ?= $(shell go env GOPATH)/bin

# ---------------------------------------------------------------------------
# Development
# ---------------------------------------------------------------------------

## dev: Start infrastructure services and run backend + web in parallel
dev: docker-infra
	@echo "Starting backend and web dev servers..."
	@trap 'kill 0' EXIT; \
		$(MAKE) backend-run & \
		$(MAKE) web-dev & \
		wait

## docker-infra: Start only infrastructure services (postgres, redis, meilisearch)
docker-infra:
	@echo "Starting infrastructure services..."
	$(DOCKER_COMPOSE) up -d postgres redis meilisearch
	@echo "Waiting for services to be healthy..."
	@$(DOCKER_COMPOSE) exec postgres sh -c 'until pg_isready -U seva; do sleep 1; done' 2>/dev/null
	@echo "Infrastructure services are ready."

# ---------------------------------------------------------------------------
# Backend (Go)
# ---------------------------------------------------------------------------

## backend-run: Run the Go backend with live-reload (requires air)
backend-run:
	@echo "Starting Go backend..."
	@if command -v air > /dev/null 2>&1; then \
		cd $(BACKEND_DIR) && air; \
	else \
		echo "air not found, running without live-reload..."; \
		cd $(BACKEND_DIR) && go run ./cmd/server; \
	fi

## backend-build: Build the Go backend binary
backend-build:
	@echo "Building backend..."
	cd $(BACKEND_DIR) && go build -ldflags="-w -s" -o bin/server ./cmd/server
	@echo "Backend binary built: $(BACKEND_DIR)/bin/server"

## backend-test: Run all backend tests
backend-test:
	@echo "Running backend tests..."
	cd $(BACKEND_DIR) && go test -v -race -coverprofile=coverage.out ./...
	@echo "Generating coverage report..."
	cd $(BACKEND_DIR) && go tool cover -func=coverage.out

## backend-lint: Lint the Go backend (requires golangci-lint)
backend-lint:
	@echo "Linting backend..."
	cd $(BACKEND_DIR) && golangci-lint run ./...

# ---------------------------------------------------------------------------
# Web (SvelteKit)
# ---------------------------------------------------------------------------

## web-dev: Start the SvelteKit dev server
web-dev:
	@echo "Starting SvelteKit dev server..."
	cd $(WEB_DIR) && npm run dev

## web-build: Build the SvelteKit application for production
web-build:
	@echo "Building web application..."
	cd $(WEB_DIR) && npm run build

## web-preview: Preview the production build locally
web-preview:
	@echo "Previewing production build..."
	cd $(WEB_DIR) && npm run preview

# ---------------------------------------------------------------------------
# Database Migrations
# ---------------------------------------------------------------------------

## db-migrate-up: Run all pending database migrations
db-migrate-up:
	@echo "Running database migrations (up)..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
	@echo "Migrations applied successfully."

## db-migrate-down: Rollback the last database migration
db-migrate-down:
	@echo "Rolling back last migration..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "Migration rolled back."

## db-migrate-create: Create a new migration file (usage: make db-migrate-create NAME=create_users)
db-migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make db-migrate-create NAME=create_users"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@mkdir -p $(MIGRATIONS_DIR)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "Migration files created in $(MIGRATIONS_DIR)/"

# ---------------------------------------------------------------------------
# API / Code Generation
# ---------------------------------------------------------------------------

## api-generate: Generate API clients from OpenAPI spec
api-generate:
	@echo "Generating API clients from OpenAPI spec..."
	@if [ ! -f api/openapi.yaml ]; then \
		echo "Error: api/openapi.yaml not found"; \
		exit 1; \
	fi
	@mkdir -p generated/ts-client
	@mkdir -p generated/dart-client
	@echo "Generating TypeScript client..."
	npx @openapitools/openapi-generator-cli generate \
		-i api/openapi.yaml \
		-g typescript-fetch \
		-o generated/ts-client \
		--additional-properties=supportsES6=true,typescriptThreePlus=true
	@echo "Generating Dart client..."
	npx @openapitools/openapi-generator-cli generate \
		-i api/openapi.yaml \
		-g dart \
		-o generated/dart-client
	@echo "API clients generated successfully."

# ---------------------------------------------------------------------------
# Docker
# ---------------------------------------------------------------------------

## docker-up: Start all Docker services (infra + app)
docker-up:
	@echo "Starting all Docker services..."
	$(DOCKER_COMPOSE) up -d
	@echo "All services started."

## docker-down: Stop and remove all Docker services
docker-down:
	@echo "Stopping all Docker services..."
	$(DOCKER_COMPOSE) down
	@echo "All services stopped."

## docker-build: Build all Docker images
docker-build:
	@echo "Building Docker images..."
	$(DOCKER_COMPOSE) build
	@echo "Docker images built."

# ---------------------------------------------------------------------------
# Setup & Clean
# ---------------------------------------------------------------------------

## setup: Install all dependencies and set up local environment
setup:
	@echo "=== Setting up Seva development environment ==="
	@echo ""
	@echo "--- Checking Go installation ---"
	@go version || (echo "Error: Go is not installed. Install from https://go.dev/dl/" && exit 1)
	@echo ""
	@echo "--- Checking Node.js installation ---"
	@node --version || (echo "Error: Node.js is not installed. Install from https://nodejs.org/" && exit 1)
	@echo ""
	@echo "--- Installing Go tools ---"
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo ""
	@echo "--- Installing backend Go dependencies ---"
	@if [ -f $(BACKEND_DIR)/go.mod ]; then \
		cd $(BACKEND_DIR) && go mod download; \
	else \
		echo "Skipping: $(BACKEND_DIR)/go.mod not found (initialize with 'go mod init')"; \
	fi
	@echo ""
	@echo "--- Installing web dependencies ---"
	@if [ -f $(WEB_DIR)/package.json ]; then \
		cd $(WEB_DIR) && npm ci; \
	else \
		echo "Skipping: $(WEB_DIR)/package.json not found"; \
	fi
	@echo ""
	@echo "--- Creating .env from example if not present ---"
	@if [ ! -f .env ] && [ -f .env.example ]; then \
		cp .env.example .env; \
		echo "Created .env from .env.example - please review and update values."; \
	elif [ ! -f .env ]; then \
		echo "No .env.example found. You may need to create a .env file manually."; \
	else \
		echo ".env already exists, skipping."; \
	fi
	@echo ""
	@echo "--- Checking Docker installation ---"
	@docker --version || echo "Warning: Docker is not installed. Required for running infrastructure services."
	@echo ""
	@echo "=== Setup complete! Run 'make dev' to start developing. ==="

## clean: Remove all build artifacts, generated files, and temp directories
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BACKEND_DIR)/bin/
	rm -rf $(BACKEND_DIR)/tmp/
	rm -rf $(BACKEND_DIR)/coverage.out
	rm -rf $(WEB_DIR)/build/
	rm -rf $(WEB_DIR)/.svelte-kit/
	rm -rf generated/
	rm -rf dist/
	rm -rf tmp/
	@echo "Clean complete."

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------

## help: Show this help message
help:
	@echo "Seva — Service Marketplace Platform - Available Commands"
	@echo "======================================================="
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
	@echo ""

# Default target
.DEFAULT_GOAL := help

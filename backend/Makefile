.PHONY: run build test lint migrate-up migrate-down migrate-create sqlc-generate clean

# Application
APP_NAME    := seva-api
BUILD_DIR   := ./bin
MAIN_PATH   := ./cmd/api

# Tools
GO          := go
GOTEST      := $(GO) test
GOBUILD     := $(GO) build
MIGRATE     := migrate
SQLC        := sqlc
GOLINT      := golangci-lint

# --- Build & Run ---

run:
	$(GO) run $(MAIN_PATH)/main.go

build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)/main.go

clean:
	rm -rf $(BUILD_DIR)

# --- Testing ---

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-short:
	$(GOTEST) -v -short ./...

coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html

# --- Database Migrations ---

MIGRATE_DIR  := ./migrations
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/seva?sslmode=disable

migrate-up:
	$(MIGRATE) -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	$(MIGRATE) create -ext sql -dir $(MIGRATE_DIR) -seq $$name

# --- Code Generation ---

sqlc-generate:
	$(SQLC) generate

# --- Linting ---

lint:
	$(GOLINT) run ./...

# --- Docker (convenience) ---

docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker compose up -d

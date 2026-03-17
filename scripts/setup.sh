#!/usr/bin/env bash
set -euo pipefail

echo "=== Seva — Development Setup ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_command() {
    if command -v "$1" &> /dev/null; then
        echo -e "${GREEN}✓${NC} $1 found: $($1 --version 2>/dev/null | head -1)"
        return 0
    else
        echo -e "${RED}✗${NC} $1 not found"
        return 1
    fi
}

echo "Checking prerequisites..."
echo ""

# Required tools
MISSING=0

check_command "go" || {
    echo "  Install Go: https://go.dev/dl/ (1.23+)"
    MISSING=1
}

check_command "node" || {
    echo "  Install Node.js: https://nodejs.org/ (22+)"
    MISSING=1
}

check_command "npm" || {
    echo "  npm comes with Node.js"
    MISSING=1
}

check_command "docker" || {
    echo "  Install Docker: https://docs.docker.com/get-docker/"
    MISSING=1
}

check_command "flutter" || {
    echo "  Install Flutter: https://docs.flutter.dev/get-started/install"
    echo "  (Optional — only needed for mobile development)"
}

echo ""

if [ "$MISSING" -eq 1 ]; then
    echo -e "${YELLOW}Some required tools are missing. Install them and re-run this script.${NC}"
    exit 1
fi

echo "All required tools found!"
echo ""

# Backend setup
echo "Setting up backend..."
cd "$(dirname "$0")/../backend"

if [ ! -f ".env" ]; then
    cp .env.example .env
    echo -e "${GREEN}✓${NC} Created backend/.env from .env.example"
else
    echo -e "${YELLOW}→${NC} backend/.env already exists, skipping"
fi

echo "Installing Go dependencies..."
go mod download
echo -e "${GREEN}✓${NC} Go dependencies installed"

# Install Go tools
echo "Installing Go tools..."
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/air-verse/air@latest
echo -e "${GREEN}✓${NC} Go tools installed (sqlc, migrate, air)"

cd ..

# Web setup
echo ""
echo "Setting up web..."
cd web

npm install
echo -e "${GREEN}✓${NC} Web dependencies installed"

cd ..

# Start infrastructure
echo ""
echo "Starting infrastructure (PostgreSQL, Redis, Meilisearch)..."
cd infrastructure/docker
docker compose up -d postgres redis meilisearch
echo -e "${GREEN}✓${NC} Infrastructure started"

# Wait for postgres
echo "Waiting for PostgreSQL..."
for i in {1..30}; do
    if docker compose exec -T postgres pg_isready -U seva &> /dev/null; then
        echo -e "${GREEN}✓${NC} PostgreSQL is ready"
        break
    fi
    if [ "$i" -eq 30 ]; then
        echo -e "${RED}✗${NC} PostgreSQL failed to start"
        exit 1
    fi
    sleep 1
done

cd ../..

# Run migrations
echo ""
echo "Running database migrations..."
cd backend
migrate -path migrations -database "$DATABASE_URL" up 2>/dev/null || {
    # Fall back to reading from .env
    source .env 2>/dev/null || true
    migrate -path migrations -database "${DATABASE_URL:-postgres://seva:seva@localhost:5432/seva?sslmode=disable}" up
}
echo -e "${GREEN}✓${NC} Migrations applied"

cd ..

echo ""
echo -e "${GREEN}=== Setup complete! ===${NC}"
echo ""
echo "To start development:"
echo "  Backend:  cd backend && air          (hot reload)"
echo "  Web:      cd web && npm run dev      (SvelteKit dev server)"
echo "  Both:     make dev                   (runs both)"
echo ""
echo "Infrastructure:"
echo "  Start:    cd infrastructure/docker && docker compose up -d"
echo "  Stop:     cd infrastructure/docker && docker compose down"
echo "  Logs:     cd infrastructure/docker && docker compose logs -f"
echo ""

.PHONY: up down build migrate seed test lint logs

# Start all services in the background
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# Build all Docker images
build:
	docker compose build

# Run database migrations (via API startup)
migrate:
	docker compose up -d postgres redis
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	docker compose run --rm api ./api

# Seed the database with sample data
seed:
	docker compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	docker compose exec -T postgres psql -U zenvikar -d zenvikar < scripts/seed.sql

# Run Go tests
test:
	cd apps/api && go test ./...

# Run linters
lint:
	cd apps/api && go vet ./...

# Tail logs from all services
logs:
	docker compose logs -f

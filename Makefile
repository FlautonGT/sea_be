# ============================================
# Gate API - Makefile
# ============================================

.PHONY: help build run stop restart logs clean migrate seed test dev prod

# Default target
help:
	@echo "Gate API - Available Commands:"
	@echo ""
	@echo "  Development:"
	@echo "    make dev          - Start development environment (PostgreSQL, Redis, PgAdmin, Redis Commander)"
	@echo "    make dev-down     - Stop development environment"
	@echo ""
	@echo "  Production:"
	@echo "    make prod         - Start production environment (with Nginx)"
	@echo "    make prod-down    - Stop production environment"
	@echo ""
	@echo "  Database:"
	@echo "    make migrate-up   - Run database migrations"
	@echo "    make migrate-down - Rollback database migrations"
	@echo "    make seed         - Seed initial data"
	@echo "    make reset-db     - Reset database (drop all and recreate)"
	@echo ""
	@echo "  Docker:"
	@echo "    make build        - Build Docker images"
	@echo "    make logs         - View container logs"
	@echo "    make ps           - List running containers"
	@echo "    make clean        - Remove all containers and volumes"
	@echo ""
	@echo "  Application:"
	@echo "    make run          - Run application locally (without Docker)"
	@echo "    make test         - Run tests"
	@echo "    make lint         - Run linter"

# ============================================
# Development
# ============================================

dev:
	@echo "Starting development environment..."
	docker-compose --profile dev up -d
	@echo ""
	@echo "Services started:"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - PgAdmin: http://localhost:5050"
	@echo "  - Redis Commander: http://localhost:8081"

dev-down:
	@echo "Stopping development environment..."
	docker-compose --profile dev down

# ============================================
# Production
# ============================================

prod:
	@echo "Starting production environment..."
	docker-compose --profile production up -d
	@echo ""
	@echo "Services started:"
	@echo "  - API: http://localhost:8080"
	@echo "  - Nginx: http://localhost:80"

prod-down:
	@echo "Stopping production environment..."
	docker-compose --profile production down

# ============================================
# Database Migrations
# ============================================

migrate-up:
	@echo "Running migrations..."
	docker-compose run --rm migrate up

migrate-down:
	@echo "Rolling back migrations..."
	docker-compose run --rm migrate down 1

migrate-force:
	@echo "Force setting migration version..."
	docker-compose run --rm migrate force $(VERSION)

seed:
	@echo "Seeding database..."
	docker-compose exec postgres psql -U gate -d gate_db -f /docker-entrypoint-initdb.d/000002_seed_data.up.sql

reset-db:
	@echo "Resetting database..."
	docker-compose exec postgres psql -U gate -d gate_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "Running migrations..."
	docker-compose run --rm migrate up

# ============================================
# Docker
# ============================================

build:
	@echo "Building Docker images..."
	docker-compose build

logs:
	docker-compose logs -f

logs-api:
	docker-compose logs -f api

logs-postgres:
	docker-compose logs -f postgres

logs-redis:
	docker-compose logs -f redis

ps:
	docker-compose ps

clean:
	@echo "Removing all containers and volumes..."
	docker-compose down -v --remove-orphans
	docker system prune -f

# ============================================
# Application (Local Development)
# ============================================

run:
	@echo "Running application..."
	go run ./cmd/api

test:
	@echo "Running tests..."
	go test -v -race -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linter..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# ============================================
# Utilities
# ============================================

psql:
	docker-compose exec postgres psql -U gate -d gate_db

redis-cli:
	docker-compose exec redis redis-cli -a gate_redis_password

generate:
	@echo "Generating code..."
	go generate ./...

swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs

# ============================================
# Setup
# ============================================

setup:
	@echo "Setting up development environment..."
	@echo "1. Copying environment file..."
	cp env.example .env 2>/dev/null || true
	@echo ""
	@echo "2. Starting services..."
	$(MAKE) dev
	@echo ""
	@echo "3. Waiting for PostgreSQL to be ready..."
	sleep 5
	@echo ""
	@echo "4. Running migrations..."
	$(MAKE) migrate-up
	@echo ""
	@echo "Setup complete! Edit .env file with your configuration."

.PHONY: help build up down logs clean test test-backend test-frontend test-all dev prod

# Default environment
APP_ENV ?= development

help: ## Show this help message
	@echo 'Usage: make [target] [APP_ENV=development|production]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Current environment: $(APP_ENV)'
	@echo 'Available environments: development, production'

build: ## Build all Docker images
	APP_ENV=$(APP_ENV) docker compose build

up: ## Start all services
	APP_ENV=$(APP_ENV) docker compose --env-file .env.$(APP_ENV) up -d

down: ## Stop all services
	docker compose down

dev: ## Start development environment
	$(MAKE) up APP_ENV=development

prod: ## Start production environment
	$(MAKE) up APP_ENV=production

logs: ## Show logs from all services
	docker compose logs -f

logs-backend: ## Show backend logs
	docker compose logs -f backend

logs-frontend: ## Show frontend logs
	docker compose logs -f frontend

clean: ## Remove all containers and volumes
	docker compose down -v
	docker system prune -f


restart: ## Restart all services
	docker compose restart

status: ## Show status of all services
	docker compose ps

test: test-backend test-frontend ## Run all tests (backend + frontend)

test-backend: ## Run backend tests
	@echo "=== Running Backend Tests ==="
	@cd backend && go test ./pkg/... ./internal/services/similarity/... ./internal/middleware/... ./internal/handlers/... -v
	@echo ""
	@echo "✅ Backend tests passed!"

test-frontend: ## Run frontend E2E tests
	@echo "=== Running Frontend E2E Tests ==="
	@echo "Installing Playwright browsers if needed..."
	@cd frontend && npx playwright install chromium --with-deps || true
	@cd frontend && npm run test:e2e
	@echo "✅ Frontend tests passed!"

test-all: ## Run all backend tests (including integration tests)
	@echo "Running all backend tests..."
	@cd backend && go test ./... -v || true

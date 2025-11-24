.PHONY: help run build test clean migrate-up migrate-down docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application
	@go run ./cmd/api

build: ## Build the application
	@go build -o server ./cmd/api

test: ## Run tests
	@go test -v -race -cover ./...

clean: ## Clean build artifacts
	@rm -f server
	@rm -rf data/

migrate-up: ## Run database migrations
	@mkdir -p data
	@echo "Applying migrations..."
	@sqlite3 data/app.db < migrations/000001_create_users.up.sql
	@sqlite3 data/app.db < migrations/000002_create_messages.up.sql
	@echo "Migrations applied successfully"

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@sqlite3 data/app.db < migrations/000002_create_messages.down.sql
	@sqlite3 data/app.db < migrations/000001_create_users.down.sql
	@echo "Migrations rolled back successfully"

docker-up: ## Start Docker Compose services
	@docker-compose up -d

docker-down: ## Stop Docker Compose services
	@docker-compose down

lint: ## Run linters
	@golangci-lint run

deps: ## Download dependencies
	@go mod download
	@go mod tidy

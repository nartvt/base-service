# Base Service Makefile
# Common development and deployment tasks

# Load environment variables from .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Variables (can be overridden by .env or command line)
# If APP_DATABASE_* variables are set, use them; otherwise use defaults
DB_HOST ?= $(or $(APP_DATABASE_HOST),localhost)
DB_PORT ?= $(or $(APP_DATABASE_PORT),5432)
DB_USER ?= $(or $(APP_DATABASE_USERNAME),postgres)
DB_NAME ?= $(or $(APP_DATABASE_DBNAME),orders)
PROFILE ?= $(or $(APP_SERVER_PROFILE),local)

MIGRATION_DIR = internal/database/migrations
BINARY_NAME = server

# Colors for output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m
COLOR_BLUE = \033[34m

##@ General
.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: all
all: deps docs build ## Install dependencies, generate docs, and build

##@ Development

.PHONY: deps
deps: ## Install dependencies
	@echo "$(COLOR_BLUE)Installing dependencies...$(COLOR_RESET)"
	go mod download
	go mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies installed$(COLOR_RESET)"

.PHONY: vendor
vendor: ## Vendor dependencies
	@echo "$(COLOR_BLUE)Vendoring dependencies...$(COLOR_RESET)"
	go mod vendor
	@echo "$(COLOR_GREEN)✓ Dependencies vendored$(COLOR_RESET)"

.PHONY: build
build: ## Build the application
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	go mod tidy && go mod vendor && go build -o $(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✓ Build complete: ./$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-prod
build-prod: ## Build for production (optimized)
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME) for production...$(COLOR_RESET)"
	go build -ldflags="-s -w" -o $(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✓ Production build complete: ./$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: run
run: build ## Run the application
	@echo "$(COLOR_BLUE)Starting the server...$(COLOR_RESET)"
	./server -config config -env $(PROFILE)
	# go run main.go

.PHONY: dev
dev: docs run ## Generate docs and run the application

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	rm -f $(BINARY_NAME)
	rm -rf vendor/
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

##@ Documentation

.PHONY: docs
docs: ## Generate Swagger documentation
	@echo "$(COLOR_BLUE)Generating Swagger Documentation...$(COLOR_RESET)"
	swag init --parseDependency --parseInternal
	@echo "$(COLOR_GREEN)✓ Swagger Documentation Generated$(COLOR_RESET)"

##@ Database

.PHONY: db-init
db-init: ## Initialize database with schema (creates DB and applies migrations)
	@echo "$(COLOR_BLUE)Initializing database: $(DB_NAME)$(COLOR_RESET)"
	@./scripts/migrate.sh init
	@echo "$(COLOR_GREEN)✓ Database initialized$(COLOR_RESET)"

.PHONY: db-migrate
db-migrate: ## Apply all pending migrations
	@echo "$(COLOR_BLUE)Applying migrations to: $(DB_NAME)$(COLOR_RESET)"
	@./scripts/migrate.sh up
	@echo "$(COLOR_GREEN)✓ Migrations applied$(COLOR_RESET)"

.PHONY: db-rollback
db-rollback: ## Rollback last migration
	@echo "$(COLOR_YELLOW)Rolling back last migration from: $(DB_NAME)$(COLOR_RESET)"
	@./scripts/migrate.sh down

.PHONY: db-status
db-status: ## Show migration status
	@./scripts/migrate.sh status

.PHONY: db-reset
db-reset: ## Reset database (⚠️ DELETES ALL DATA)
	@echo "$(COLOR_YELLOW)⚠️  WARNING: This will DELETE ALL DATA in: $(DB_NAME)$(COLOR_RESET)"
	@./scripts/migrate.sh reset

.PHONY: db-create
db-create: ## Create database only (no schema)
	@echo "$(COLOR_BLUE)Creating database: $(DB_NAME)$(COLOR_RESET)"
	psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);"
	@echo "$(COLOR_GREEN)✓ Database created$(COLOR_RESET)"

.PHONY: db-drop
db-drop: ## Drop database (⚠️ DELETES ALL DATA)
	@echo "$(COLOR_YELLOW)⚠️  Dropping database: $(DB_NAME)$(COLOR_RESET)"
	psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "$(COLOR_GREEN)✓ Database dropped$(COLOR_RESET)"

.PHONY: db-connect
db-connect: ## Connect to database with psql
	psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-dump
db-dump: ## Backup database to file
	@echo "$(COLOR_BLUE)Backing up database: $(DB_NAME)$(COLOR_RESET)"
	pg_dump -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -F c -f backup_$(DB_NAME)_$(shell date +%Y%m%d_%H%M%S).dump
	@echo "$(COLOR_GREEN)✓ Database backed up$(COLOR_RESET)"

.PHONY: db-restore
db-restore: ## Restore database from latest backup (usage: make db-restore FILE=backup.dump)
	@if [ -z "$(FILE)" ]; then \
		echo "$(COLOR_YELLOW)Usage: make db-restore FILE=backup.dump$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BLUE)Restoring database from: $(FILE)$(COLOR_RESET)"
	pg_restore -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c $(FILE)
	@echo "$(COLOR_GREEN)✓ Database restored$(COLOR_RESET)"

##@ Code Generation

.PHONY: sqlc
sqlc: ## Generate SQLC code
	@echo "$(COLOR_BLUE)Generating SQLC code...$(COLOR_RESET)"
	sqlc generate
	@echo "$(COLOR_GREEN)✓ SQLC code generated$(COLOR_RESET)"

.PHONY: generate
generate: sqlc docs ## Generate all code (SQLC + Swagger docs)
	@echo "$(COLOR_GREEN)✓ All code generated$(COLOR_RESET)"

##@ Testing

.PHONY: test
test: ## Run tests
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	go test -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: coverage.html$(COLOR_RESET)"

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(COLOR_BLUE)Running tests with race detector...$(COLOR_RESET)"
	go test -race -v ./...

##@ Code Quality

.PHONY: fmt
fmt: ## Format code
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	go fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: lint
lint: ## Run linter
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	golangci-lint run ./...

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	go vet ./...

.PHONY: check
check: fmt vet test ## Format, vet, and test

##@ Docker (Optional)

.PHONY: dockerup
dockerup: ## Start Docker containers
	@echo "$(COLOR_BLUE)Starting Docker containers...$(COLOR_RESET)"
	docker compose -f docker/docker-compose.yml up -d --build
	@echo "$(COLOR_GREEN)✓ Docker containers started$(COLOR_RESET)"

.PHONY: dockerdown
dockerdown: ## Stop Docker containers
	@echo "$(COLOR_BLUE)Stopping Docker containers...$(COLOR_RESET)"
	docker compose -f docker/docker-compose.yml down -v --remove-orphans
	@echo "$(COLOR_GREEN)✓ Docker containers stopped$(COLOR_RESET)"

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	docker build -t base-service:latest .
	@echo "$(COLOR_GREEN)✓ Docker image built$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(COLOR_BLUE)Running Docker container...$(COLOR_RESET)"
	docker run -p 8081:8081 --env-file .env base-service:latest

.PHONY: docker-db
docker-db: ## Start PostgreSQL in Docker
	@echo "$(COLOR_BLUE)Starting PostgreSQL container...$(COLOR_RESET)"
	docker run -d \
		--name postgres-base-service \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=root \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		postgres:14-alpine
	@echo "$(COLOR_GREEN)✓ PostgreSQL container started$(COLOR_RESET)"

.PHONY: docker-db-stop
docker-db-stop: ## Stop PostgreSQL Docker container
	@echo "$(COLOR_BLUE)Stopping PostgreSQL container...$(COLOR_RESET)"
	docker stop postgres-base-service
	docker rm postgres-base-service
	@echo "$(COLOR_GREEN)✓ PostgreSQL container stopped$(COLOR_RESET)"

##@ Setup

.PHONY: setup
setup: deps db-init generate ## Complete project setup (deps + db + generate)
	@echo "$(COLOR_GREEN)========================================$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)✓ Setup complete!$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)========================================$(COLOR_RESET)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Copy env.example to .env and configure"
	@echo "  2. Run: make run"
	@echo "  3. Visit: http://localhost:8081/swagger/"
	@echo ""

.PHONY: setup-docker
setup-docker: docker-db ## Setup with Docker PostgreSQL
	@echo "$(COLOR_BLUE)Waiting for PostgreSQL to be ready...$(COLOR_RESET)"
	@sleep 3
	@$(MAKE) db-init
	@$(MAKE) generate
	@echo "$(COLOR_GREEN)✓ Docker setup complete!$(COLOR_RESET)"

##@ Information

.PHONY: info
info: ## Show project information
	@echo "$(COLOR_BLUE)Project Information$(COLOR_RESET)"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Database: $(DB_NAME)"
	@echo "  DB User: $(DB_USER)"
	@echo "  DB Host: $(DB_HOST):$(DB_PORT)"
	@echo ""
	@echo "$(COLOR_BLUE)Go Information$(COLOR_RESET)"
	@go version
	@echo ""
	@echo "$(COLOR_BLUE)Database Status$(COLOR_RESET)"
	@$(MAKE) db-status || echo "  Database not initialized"

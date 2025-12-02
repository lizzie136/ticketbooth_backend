.PHONY: help build run test clean deps fmt vet lint install docker-build docker-run db-init db-seed

# Variables
BINARY_NAME=ticketbooth-backend
MAIN_PATH=./main.go
BUILD_DIR=./bin
SCHEMA_FILE=./schema.sql
MOCK_DATA_FILE=./insert_mock_data.sql
DB_HOST ?= localhost
DB_USER ?= root
DB_PASSWORD ?= ""
DB_NAME ?= ticketbooth

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Run the application (loads .env if present)
	@echo "Running $(BINARY_NAME)..."
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs); \
	fi; \
	go run $(MAIN_PATH)

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: fmt vet ## Run linters (fmt + vet)

install: deps ## Install dependencies and build
	@$(MAKE) deps
	@$(MAKE) build

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 4000:4000 --env-file .env $(BINARY_NAME):latest

db-init: ## Initialize database schema (set DB_HOST, DB_USER, DB_PASSWORD as env vars)
	@echo "Initializing database schema..."
	@if [ -z "$(DB_PASSWORD)" ]; then \
		mysql -h $(DB_HOST) -u $(DB_USER) -p < $(SCHEMA_FILE); \
	else \
		mysql -h $(DB_HOST) -u $(DB_USER) -p$(DB_PASSWORD) < $(SCHEMA_FILE); \
	fi
	@echo "Database schema initialized successfully"

db-seed: ## Insert mock data into database (set DB_HOST, DB_USER, DB_PASSWORD, DB_NAME as env vars)
	@echo "Inserting mock data..."
	@if [ -z "$(DB_PASSWORD)" ]; then \
		mysql -h $(DB_HOST) -u $(DB_USER) -p $(DB_NAME) < $(MOCK_DATA_FILE); \
	else \
		mysql -h $(DB_HOST) -u $(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) < $(MOCK_DATA_FILE); \
	fi
	@echo "Mock data inserted successfully"


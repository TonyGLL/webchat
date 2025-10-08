.PHONY: build run test watch up-dev up-prod down

# Build the Go application
build:
	@echo "Building..."
	@go build -o main cmd/server/main.go

# Run the Go application for development
depends-on: build
run:
	@echo "Running in development mode..."
	@CONFIG_FILE=dev.env go run cmd/server/main.go

# Run tests
test:
	@echo "Testing..."
	@go test -v ./...

# Live Reload
watch:
	@if command -v air > /dev/null; then \
	    CONFIG_FILE=local.env air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/air-verse/air@latest; \
	        CONFIG_FILE=local.env air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

database-up:
	@echo "Starting PostgreSQL database..."
	@CONFIG_FILE=dev.env docker compose up -d db

# Docker-compose commands
up-dev:
	@echo "Starting development environment..."
	@CONFIG_FILE=dev.env docker compose up --build -d

up-prod:
	@echo "Starting production environment..."
	@CONFIG_FILE=prod.env docker compose up --build -d

down:
	@echo "Stopping docker compose environment..."
	@docker compose down
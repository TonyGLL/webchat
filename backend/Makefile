.PHONY: build run test watch up-dev up-prod down

# Build the Go application
build:
	@echo "Building..."
	@go build -o ../main main.go

# Run the Go application for development
run:
	@echo "Running in development mode..."
	@CONFIG_FILE=dev.env go run .

# Run tests
test:
	@echo "Testing..."
	@go test -v ./...

# Live Reload with Air for development
watch:
	@if ! command -v air > /dev/null; then \
		echo "air is not installed. Installing it now..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@echo "Watching for changes..."
	@CONFIG_FILE=dev.env air

# Docker-compose commands
up-dev:
	@echo "Starting development environment..."
	@CONFIG_FILE=dev.env docker-compose up --build -d

up-prod:
	@echo "Starting production environment..."
	@CONFIG_FILE=prod.env docker-compose up --build -d

down:
	@echo "Stopping docker-compose environment..."
	@docker-compose down
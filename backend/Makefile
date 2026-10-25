.PHONY: build run test watch up-dev up-prod down admin-list admin-grant admin-revoke

# Build the Go application
build:
	@echo "Building..."
	@go build -o main cmd/server/main.go

# Run the Go application for development
depends-on: build
run:
	@echo "Running in development mode..."
	@APP_ENV=dev CONFIG_FILE=.env.local go run cmd/server/main.go

# Run tests
test:
	@echo "Testing..."
	@go test -v ./...

# Live Reload (uses .air.toml). `air` is looked up on the PATH and in Go's bin dir,
# since `go install` puts it in $(go env GOBIN) or $(go env GOPATH)/bin.
GO_BIN_DIR := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
AIR = $(or $(shell command -v air),$(GO_BIN_DIR)/air)

watch:
	@if [ ! -x "$(AIR)" ]; then \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" = "n" ] || [ "$$choice" = "N" ]; then \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	    go install github.com/air-verse/air@latest || exit 1; \
	fi
	@echo "Watching..."
	@APP_ENV=dev CONFIG_FILE=.env.local $(AIR)

# Docker lives at the repository root (docker-compose.yml); these targets just delegate.
database-up:
	@$(MAKE) --no-print-directory -C .. infra-up

up-dev:
	@$(MAKE) --no-print-directory -C .. up-dev

up-prod:
	@$(MAKE) --no-print-directory -C .. up-prod

down:
	@$(MAKE) --no-print-directory -C .. down

# Global roles (never granted through the API). Examples:
#   make admin-list
#   make admin-grant EMAIL=someone@example.com ROLE=admin     (admin | moderator)
#   make admin-revoke EMAIL=someone@example.com
admin-list:
	@CONFIG_FILE=.env.local go run ./cmd/admin list

admin-grant:
	@CONFIG_FILE=.env.local go run ./cmd/admin grant --email "$(EMAIL)" --role "$(ROLE)"

admin-revoke:
	@CONFIG_FILE=.env.local go run ./cmd/admin revoke --email "$(EMAIL)"

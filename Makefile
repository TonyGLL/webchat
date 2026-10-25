# Root Makefile: run everything from here, no need to cd into backend/ or frontend/.
# `make` or `make help` lists the targets. It delegates to backend/Makefile (Go, Docker
# compose) and to pnpm (frontend), so their own commands keep working too.

.DEFAULT_GOAL := help

BACKEND  := backend
FRONTEND := frontend

# pnpm: the installed one if there is any, otherwise the version pinned in package.json via npx.
PNPM_VERSION := $(shell node -p "require('./$(FRONTEND)/package.json').packageManager.split('@')[1]" 2>/dev/null)
PNPM := $(or $(shell command -v pnpm),npx --yes pnpm@$(PNPM_VERSION))

# `make dev` starts the frontend only once the backend answers /health (it compiles first, and a
# browser tab that is already open would otherwise hit "Network Error" in that gap).
BACKEND_URL ?= http://localhost:8081
WAIT_BACKEND_TIMEOUT ?= 120

# Throwaway services for `make test-integration` (host ports chosen not to clash with the dev stack).
IT_PG_PORT    ?= 55432
IT_REDIS_PORT ?= 56379

.PHONY: help setup infra-up infra-down infra-ps infra-logs dev backend-run backend-watch \
        wait-backend frontend-dev frontend-dev-after-backend build test test-integration lint fmt check up-dev up-prod down \
        admin-list admin-grant admin-revoke clean

help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5)} /^[a-zA-Z0-9_-]+:.*## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ First time

setup: ## Install frontend dependencies and download Go modules
	cd $(FRONTEND) && $(PNPM) install --frozen-lockfile
	cd $(BACKEND) && go mod download

##@ Local development (backend and frontend on your machine, Postgres and Redis in Docker)

infra-up: ## Start Postgres and Redis (docker compose)
	cd $(BACKEND) && docker compose up -d db redis

infra-down: ## Stop Postgres and Redis (data is kept)
	cd $(BACKEND) && docker compose stop db redis

infra-ps: ## Show the state of the containers
	cd $(BACKEND) && docker compose ps

infra-logs: ## Follow the Postgres and Redis logs
	cd $(BACKEND) && docker compose logs -f db redis

dev: infra-up ## Start everything: Postgres + Redis, backend with live reload, then the frontend once the backend is up
	@$(MAKE) --no-print-directory -j2 backend-watch frontend-dev-after-backend

frontend-dev-after-backend: wait-backend frontend-dev

wait-backend: ## Wait until the backend answers /health (up to WAIT_BACKEND_TIMEOUT seconds)
	@if ! command -v curl >/dev/null 2>&1; then echo "curl not found: not waiting for the backend"; exit 0; fi; \
	echo "Waiting for the backend at $(BACKEND_URL)/health ..."; \
	i=0; until curl -fs -o /dev/null -m 2 $(BACKEND_URL)/health; do \
	  i=$$((i+1)); \
	  if [ $$i -ge $(WAIT_BACKEND_TIMEOUT) ]; then \
	    echo "The backend did not answer in $(WAIT_BACKEND_TIMEOUT)s (is it compiling, or did it fail?). Starting the frontend anyway."; exit 0; \
	  fi; \
	  sleep 1; \
	done; echo "Backend is up."

backend-run: ## Run the backend once, without live reload (http://localhost:8081)
	$(MAKE) -C $(BACKEND) run

backend-watch: ## Run the backend with live reload (air)
	$(MAKE) -C $(BACKEND) watch

frontend-dev: ## Run the frontend in development mode (http://localhost:3100)
	cd $(FRONTEND) && $(PNPM) dev

##@ Quality

build: ## Compile the backend and build the frontend
	cd $(BACKEND) && go build ./...
	cd $(FRONTEND) && $(PNPM) run build

test: ## Unit tests: backend (go test -race) and frontend (vitest)
	cd $(BACKEND) && go test -race ./...
	cd $(FRONTEND) && $(PNPM) run test

test-integration: ## Backend tests against a throwaway Postgres and Redis (migrations, presence)
	@docker rm -f webchat-it-pg webchat-it-redis >/dev/null 2>&1; \
	docker run -d --rm --name webchat-it-pg -e POSTGRES_PASSWORD=ci -e POSTGRES_DB=ci \
	  -p 127.0.0.1:$(IT_PG_PORT):5432 postgres:16-alpine >/dev/null && \
	docker run -d --rm --name webchat-it-redis -p 127.0.0.1:$(IT_REDIS_PORT):6379 redis:8-alpine >/dev/null && \
	until docker exec webchat-it-pg pg_isready -U postgres -d ci >/dev/null 2>&1; do sleep 1; done; \
	( cd $(BACKEND) && \
	  TEST_DATABASE_URL='postgres://postgres:ci@127.0.0.1:$(IT_PG_PORT)/ci?sslmode=disable' \
	  TEST_REDIS_ADDR=127.0.0.1:$(IT_REDIS_PORT) go test -race -count=1 ./... ); status=$$?; \
	docker rm -f webchat-it-pg webchat-it-redis >/dev/null 2>&1; \
	exit $$status

lint: ## go vet, gofmt check, tsc and eslint
	cd $(BACKEND) && go vet ./...
	@cd $(BACKEND) && test -z "$$(gofmt -l .)" || { echo "gofmt: unformatted files, run 'make fmt'"; gofmt -l .; exit 1; }
	cd $(FRONTEND) && $(PNPM) exec tsc --noEmit && $(PNPM) run lint

fmt: ## Format the Go code
	cd $(BACKEND) && gofmt -w .

check: lint test build ## Everything CI checks, except the integration tests

##@ Docker (full stack, backend in a container)

up-dev: ## Start the whole dev stack in Docker (backend/.env.dev)
	$(MAKE) -C $(BACKEND) up-dev

up-prod: ## Start the production stack in Docker (backend/.env.prod)
	$(MAKE) -C $(BACKEND) up-prod

down: ## Stop the docker compose stack
	$(MAKE) -C $(BACKEND) down

##@ Administration (global roles are only granted from the server)

admin-list: ## List users with a global role
	$(MAKE) -C $(BACKEND) admin-list

admin-grant: ## Grant a role: make admin-grant EMAIL=you@example.com ROLE=admin (admin | moderator)
	$(MAKE) -C $(BACKEND) admin-grant EMAIL="$(EMAIL)" ROLE="$(ROLE)"

admin-revoke: ## Revoke a role: make admin-revoke EMAIL=you@example.com
	$(MAKE) -C $(BACKEND) admin-revoke EMAIL="$(EMAIL)"

##@ Cleanup

clean: ## Remove build output (backend binary, Next.js .next)
	rm -f $(BACKEND)/main
	rm -rf $(FRONTEND)/.next

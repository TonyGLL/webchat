# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Hard rule: never read credential files

Under no circumstances read, cat, grep, open, or otherwise expose files related to credentials or secrets: `.env`, `*.env`, `dev.env`, `prod.env`, `local.env`, `*.env.example` variants, key/cert/token files, or anything similar. This applies always, even if it seems needed for debugging; ask the user instead. Reading the Go code that consumes env vars (e.g. `internal/shared/config`) is fine.

Full step-by-step setup (env vars, DB schema, ports, troubleshooting) is in `docs/LEVANTAR.md`; work plan in `docs/ROADMAP.md` and open issues in `docs/HALLAZGOS.md` (move items to Completados with commit hash when resolved).

## Overview

Monorepo with two independent apps: `backend/` (Go + Gin, PostgreSQL, Redis, Gorilla WebSocket) and `frontend/` (Next.js 16 App Router, React 19, TypeScript, Tailwind 3). There is no root-level build; run commands from each subdirectory. There are currently no tests in either app.

## Commands

Backend (run from `backend/`):
- `make watch` — run with `air` live reload (loads `CONFIG_FILE=local.env`)
- `make run` — `go run cmd/server/main.go` with `CONFIG_FILE=local.env`
- `make build` / `make test` (`go test -v ./...`); single test: `go test -v ./internal/<pkg> -run TestName`
- `make database-up` — start only the `db` service; `make up-dev` / `make up-prod` / `make down` — full docker compose stack (`dev.env` / `prod.env`)
- Docker naming standard: `<project>-<service>` in kebab-case (e.g. `webchat-api`, `webchat-db`, `webchat-redis`); volumes `webchat-<service>-data`; compose project `name: webchat`. Compose service keys stay short (`app`, `db`, `redis`).
- Ports (chosen to avoid clashing with other local projects): frontend 3100, backend 8081, Postgres 5433, Redis 6380 (host side; all published on 127.0.0.1). Keep `PORT`, `CORS_ALLOWED_ORIGINS` and `NEXT_PUBLIC_API_URL` in the env files consistent with these.
- The compose file is `backend/docker-compose.yml` (the root README mentions a root-level one, which does not exist).

Frontend (run from `frontend/`, uses Bun): `bun install`, `bun dev`, `bun run build`, `bun run lint`.

Config: `config.NewConfig` reads env vars, optionally pre-loading the file named by `CONFIG_FILE` via godotenv. `DATABASE_URL` (or `DB_USER/DB_HOST/DB_PORT/DB_NAME`) is required; `REDIS_URL` can likewise be built from `REDIS_*` parts. Schema lives in `backend/migrations/schema.sql` (applied manually, no migration tool). Frontend uses `NEXT_PUBLIC_API_URL` (default `http://localhost:8081/api/v1`).

## Backend architecture

Clean Architecture, one package per module under `backend/internal/`: `auth`, `users`, `room`, `message`, plus `websocket` and `shared`. Each module has `domain` (entities, repository interfaces), `application` (use cases, DTOs), `persistence` (pgx/Redis implementations), `presentation` (Gin controllers), and a `module.go` exposing `RegisterModule(...)`.

- **Wiring is manual in `cmd/server/main.go`**: it builds config, DB pool, Redis, JWT/mailer services, all repositories, and the WebSocket hub, then calls each module's `RegisterModule`, which constructs that module's use cases/controller and registers routes under `/api/v1`. New dependencies must be threaded through `main.go` and the module's `RegisterModule` signature.
- `shared/` holds cross-cutting pieces: `config`, `http` (Gin server, `JWTMiddleware`, request/response helpers), `infra` (db, redis, JWT and mailer implementations), `application` (service interfaces like `JwtService`, `MailerService`, `Store`), `domain` (shared errors, transaction types). `Store` (`sql_store.go`) provides transactional execution for use cases.
- **WebSocket**: `/api/v1/ws` is registered directly in `main.go` and authenticates via a `?token=` query param (JWT). `websocket.Hub` runs an event loop and tracks clients per room; the `message` module receives the hub as a `WebsocketBroadcaster` and broadcasts on message create/update/reaction changes. The `room` module's `RoomListerAdapter` lets the websocket layer look up a user's rooms without importing the room module (avoids cross-module dependency cycles).
- Auth: JWT access + refresh tokens (refresh tokens stored in Redis via `TokenRepository`), email verification flow sent through the SMTP mailer.

## Frontend architecture

Next.js App Router in `frontend/src/`: route groups `(auth)` (login, register, email verification) and `(private)` (dashboard, profile; guarded by `PrivateRoute`). Data access goes through `services/*Service.ts` (axios + zod schemas), wrapped by hooks in `hooks/` (`useAuth`, `useMessages`, `useRoom`, `useSocket`, `useUser`). `contexts/SocketContext.tsx` opens a single WebSocket using the access token, and hooks consume it for live message/reaction updates. UI primitives are in `components/ui` (shadcn-style, Radix + cva).

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Hard rule: never read credential files

Under no circumstances read, cat, grep, open, or otherwise expose files related to credentials or secrets: `.env`, `*.env`, `.env.dev`, `.env.prod`, `.env.local`, `*.env.example` variants, key/cert/token files, or anything similar. This applies always, even if it seems needed for debugging; ask the user instead. Reading the Go code that consumes env vars (e.g. `internal/shared/config`) is fine.

Full step-by-step setup (env vars, DB schema, ports, troubleshooting) is in `docs/LEVANTAR.md`; bilingual (en/es) documentation with diagrams and design decisions in `docs/en/` and `docs/es/` (same file names, language switch on line 1: when you change one language, change the other, and add an entry to `decisions.md` for any real design choice); work plan in `docs/ROADMAP.md` and open issues in `docs/HALLAZGOS.md` (move items to Completados with commit hash when resolved).

## Overview

Monorepo with two independent apps: `backend/` (Go + Gin, PostgreSQL, Redis, Gorilla WebSocket) and `frontend/` (Next.js 16 App Router, React 19, TypeScript, Tailwind 3). There is no root-level build; run commands from each subdirectory. Tests: backend `go test -race ./...`, frontend Vitest (`bun run test`); both run in CI.

## Commands

Backend (run from `backend/`):
- `make watch` — run with `air` live reload (loads `CONFIG_FILE=.env.local`)
- `make run` — `go run cmd/server/main.go` with `CONFIG_FILE=.env.local`
- `make build` / `make test` (`go test -v ./...`); single test: `go test -v ./internal/<pkg> -run TestName`
- `make database-up` — start only the `db` service; `make up-dev` / `make up-prod` / `make down` — full docker compose stack (`.env.dev` / `.env.prod`)
- Env files: committed templates are `backend/.env.example` and `frontend/.env.example`; real files are `.env.local` (run on host), `.env.dev` (Docker dev), `.env.prod`, all git-ignored and never to be read (see hard rule). Keep `backend/.env.example` in sync when adding config variables.
- Docker naming standard: `<project>-<service>` in kebab-case (e.g. `webchat-api`, `webchat-db`, `webchat-redis`); volumes `webchat-<service>-data`; compose project `name: webchat`. Compose service keys stay short (`app`, `db`, `redis`).
- Ports (chosen to avoid clashing with other local projects): frontend 3100, backend 8081, Postgres 5433, Redis 6380 (host side; all published on 127.0.0.1). Keep `PORT`, `CORS_ALLOWED_ORIGINS` and `NEXT_PUBLIC_API_URL` in the env files consistent with these.
- The compose file is `backend/docker-compose.yml` (the root README mentions a root-level one, which does not exist).

Frontend (run from `frontend/`, uses Bun): `bun install`, `bun dev`, `bun run build`, `bun run lint`, `bun run test`.

Config: `config.NewConfig` reads env vars, optionally pre-loading the file named by `CONFIG_FILE` via godotenv. `DATABASE_URL` (or `DB_USER/DB_HOST/DB_PORT/DB_NAME`) is required; `REDIS_URL` can likewise be built from `REDIS_*` parts. Schema lives in `backend/migrations/` as golang-migrate files (`NNNNNN_name.up.sql`/`.down.sql`, embedded and applied on startup by `migrations.Up`; legacy hand-built databases are adopted at version 1). Permissions: every authorization rule lives in `internal/shared/permissions` (rank-based matrix; room roles member<moderator<admin<owner + global roles in `users.role`, granted only via `cmd/admin`/`make admin-grant`). New permission checks must go through it, not ad-hoc role string comparisons. Moderation actions are audited in `audit_log`. Direct messages are private rooms with `is_direct`, one per user pair (`direct_rooms`); error classification helpers for the `lib/pq` driver live in `shared/infra/db/errors.go`. Never edit an applied migration — add a new numbered pair. Frontend uses `NEXT_PUBLIC_API_URL` (default `http://localhost:8081/api/v1`).

## Backend architecture

Clean Architecture, one package per module under `backend/internal/`: `auth`, `users`, `room`, `message`, plus `websocket` and `shared`. Each module has `domain` (entities, repository interfaces), `application` (use cases, DTOs), `persistence` (pgx/Redis implementations), `presentation` (Gin controllers), and a `module.go` exposing `RegisterModule(...)`.

- **Wiring is manual in `cmd/server/main.go`**: it builds config, DB pool, Redis, JWT/mailer services, all repositories, and the WebSocket hub, then calls each module's `RegisterModule`, which constructs that module's use cases/controller and registers routes under `/api/v1`. New dependencies must be threaded through `main.go` and the module's `RegisterModule` signature.
- `shared/` holds cross-cutting pieces: `config`, `http` (Gin server, `JWTMiddleware`, request/response helpers), `infra` (db, redis, JWT and mailer implementations), `application` (service interfaces like `JwtService`, `MailerService`, `Store`), `domain` (shared errors, transaction types). `Store` (`sql_store.go`) provides transactional execution for use cases.
- **API contract**: `backend/api/openapi.yaml` (hand-written OpenAPI 3.1, embedded with `go:embed`) is served as `/openapi.yaml` and rendered by Scalar at `/docs` (`api.RegisterDocs` in `main.go`). Any route/request/response change must update it in the same commit; `go test ./api/` checks its consistency.
- **WebSocket**: `/api/v1/ws` is registered directly in `main.go` and authenticates with a single-use 30 s ticket (`POST /api/v1/ws-ticket` with the JWT, then `/ws?ticket=`) and checks the Origin against `CORS_ALLOWED_ORIGINS`. `websocket.Hub` runs an event loop, tracks clients per room and relays events between backend instances through Redis pub/sub (`bus.go`); the `message` module receives the hub as a `WebsocketBroadcaster` and broadcasts on message create/update/reaction changes. The `room` module's `RoomListerAdapter` lets the websocket layer look up a user's rooms without importing the room module (avoids cross-module dependency cycles).
- Auth: access JWT (1 h, returned in the body) + single-use rotating refresh JWT (jti stored in Redis via `TokenRepository`, delivered as an httpOnly `refresh_token` cookie scoped to `/api/v1/auth`; `POST /auth/logout` revokes it). Email verification code exists but is paused.

## Frontend architecture

Next.js App Router in `frontend/src/`: route groups `(auth)` (login, register, email verification) and `(private)` (dashboard, profile; guarded by `PrivateRoute`). Data access goes through `services/*Service.ts`, all using the shared axios instance in `lib/api.ts` (attaches the access token, silently refreshes it on 401 via `POST /auth/refresh-token`). The access token lives only in memory (`lib/tokenStorage.ts`); the refresh token is an httpOnly cookie restored on reload by `AuthContext`. Session state is a single `AuthContext` (`contexts/AuthContext.tsx`, exposed by `useAuth`/`useUser`); other hooks: `useMessages`, `useRoom`, `useSocket`. `contexts/SocketContext.tsx` opens a single WebSocket using a ticket from `/ws-ticket`, and hooks consume it for live message/reaction updates. Styling: minimalist neutral theme driven by CSS variables in `src/app/globals.css` (light/dark: the `.dark` class on `<html>` is set before first paint by an inline script from `lib/theme.ts`, follows the OS by default and is switched with `ThemeToggle`, choice kept in localStorage). i18n (es/en) is a dependency-free layer in `lib/i18n`: typed dictionaries `en.ts` (source of truth for keys) and `es.ts`, the `useT()` hook, language kept in localStorage and defaulting to the browser language. Never hard-code UI text: add the key to both dictionaries (a test checks placeholders match) and use `t(...)`; zod schemas that carry messages are factories taking `t` (`makeRegisterSchema`). Use semantic Tailwind tokens (`bg-background`, `text-muted-foreground`, `border-border`, `bg-primary`...) — never hard-coded colors — so dark mode keeps working. UI primitives are in `components/ui` (shadcn-style, Radix + cva).

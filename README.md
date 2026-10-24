# WebChat

Real-time chat application: a Go backend (Gin, PostgreSQL, Redis, WebSockets) and a Next.js frontend. Minimal UI, JWT authentication with silent token refresh, chat rooms with invitations, message editing and emoji reactions.

## Features

- Real-time messaging over WebSockets (single-use ticket authentication, origin checks).
- Registration and login with JWT: 1-hour access token plus a single-use refresh token.
- Public and private rooms; invitations created by room owners/admins.
- Message editing and emoji reactions, visible live to everyone in the room.
- Rate limiting, password policy, structured logs, health/readiness endpoints.
- Responsive, themeable UI (light theme active, dark palette ready).

## Tech stack

| | |
|---|---|
| Backend | Go 1.24, Gin, PostgreSQL 16, Redis, Gorilla WebSocket, JWT |
| Frontend | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS, React Hook Form + Zod, Axios |
| Tooling | Docker Compose, Make, `air` (live reload), GitHub Actions CI |

## Getting started

The complete, step-by-step guide (environment variables, database schema, ports, troubleshooting) is in **[docs/LEVANTAR.md](docs/LEVANTAR.md)** (Spanish). The short version:

```bash
# 1. Database and cache
cd backend && docker compose up -d db redis

# 2. Backend (it applies the database migrations on startup): cp .env.example .env.local, fill in the values, then
make watch            # http://localhost:8081

# 3. Frontend: cp .env.example .env.local, then
cd ../frontend && bun install && bun dev    # http://localhost:3100
```

| Service | Port (localhost) |
|---|---|
| Frontend | 3100 |
| Backend | 8081 |
| PostgreSQL | 5433 |
| Redis | 6380 |

## Project structure

```
backend/    Go API (Clean Architecture, one package per module)
  cmd/server/        entry point and dependency wiring
  internal/auth|users|room|message/   domain, application, persistence, presentation
  internal/websocket/                  hub, clients, single-use tickets
  internal/shared/                     config, HTTP server/middleware, infra (db, redis, services)
  migrations/                          versioned SQL migrations (applied on startup)
frontend/   Next.js app
  src/app/            routes ((auth) and (private) groups)
  src/lib/            API client (token refresh) and token storage
  src/contexts/       AuthContext, SocketContext
  src/hooks, src/services, src/components
docs/       setup guide, roadmap, use cases, findings tracker
```

## API (prefix `/api/v1`, except health)

- `GET /health`, `GET /ready` (outside the prefix): liveness and readiness.
- **Auth:** `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh-token`, `POST /auth/send-verify-email/:user_id`, `POST /auth/verify-email/:token`.
- **Users:** `GET /users/profile`, `DELETE /users/deactivate`.
- **Rooms:** `POST /rooms`, `GET /rooms`, `POST /rooms/join/:room_id`, `POST /rooms/direct/:user_id` (open a one-to-one chat), `GET /rooms/members/:room_id` (members + online status), `POST /rooms/leave/:room_id`, `POST /rooms/read/:room_id` (mark as read; `GET /rooms` includes `unread_count`), `POST /rooms/invite/:room_id`, `POST /rooms/accept-invite/:code`.
- **Messages:** `GET /messages/:roomId/messages?limit=&before=` (cursor pagination), `GET /messages/search?q=&room_id=` (full-text, accent-insensitive, only your rooms), `POST /messages` (optional `reply_to_message_id`), `PATCH /messages/:id`, `DELETE /messages/:id` (author only, soft delete), `PUT|DELETE /messages/:id/pin` + `GET /messages/:roomId/pinned` (owner/admin pin), `POST|DELETE /messages/:id/reactions`.
- **WebSocket:** `POST /ws-ticket` (authenticated) returns a 30-second single-use ticket; connect to `GET /ws?ticket=...`.

Everything except register, login, refresh and email verification requires `Authorization: Bearer <access token>`.

## Development

```bash
# backend (from backend/)
make test             # go test -v ./...
make watch            # live reload with air
# frontend (from frontend/)
bunx tsc --noEmit && bun run lint && bun run test && bun run build
```

CI (`.github/workflows/ci.yml`) runs formatting, vet, build and tests for the backend, and type check, lint and build for the frontend.

## Project planning

- [docs/ROADMAP.md](docs/ROADMAP.md): phases to a production-grade product.
- [docs/CASOS_DE_USO.md](docs/CASOS_DE_USO.md): use cases, existing and new.
- [docs/HALLAZGOS.md](docs/HALLAZGOS.md): tracked findings and their status.

## License

See [LICENSE](LICENSE).

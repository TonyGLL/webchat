**🌐 English** · [Español](../es/architecture.md)

# Architecture

## 1. System context

WebChat is a **modular monolith** with two independently deployable apps in one repository.

```mermaid
flowchart TB
    subgraph Client
        B[Browser<br/>Next.js 16 · React 19]
    end
    subgraph Server["backend/ (Go 1.27)"]
        H[HTTP API<br/>Gin · /api/v1]
        W[WebSocket hub<br/>/api/v1/ws]
    end
    PG[(PostgreSQL 16)]
    R[(Redis)]
    B -- "REST + JWT" --> H
    B <-- "events" --> W
    H --> PG
    H --> R
    W --> R
    W -. "room lookups through an adapter" .-> H
```

| Piece | Responsibility | Port (host) |
|---|---|---|
| Frontend | UI, session state, live updates | 3100 |
| Backend | REST API, WebSocket hub, migrations on startup | 8081 |
| PostgreSQL | Durable data: users, rooms, messages, audit | 5433 |
| Redis | Short-lived state: refresh-token ids, WS tickets, presence, rate counters, cross-instance event bus | 6380 |

## 2. Backend: Clean Architecture, one package per module

Each business module under `backend/internal/` has the same four layers. Dependencies point **inwards only**: presentation → application → domain, and persistence implements interfaces that the domain declares.

```mermaid
flowchart LR
    P[presentation<br/>Gin controllers] --> A[application<br/>use cases · DTOs]
    A --> D[domain<br/>entities · repository interfaces]
    PS[persistence<br/>SQL / Redis] -. implements .-> D
```

Modules: `auth`, `users`, `room`, `message`, `moderation`, plus `websocket` and `shared`.

```mermaid
flowchart TB
    main["cmd/server/main.go<br/>(manual wiring)"]
    main --> auth & users & room & message & moderation & websocket
    message -- "WebsocketBroadcaster (interface)" --> websocket
    moderation -- "Realtime (interface)" --> websocket
    websocket -- "RoomLister (interface)" --> room
    moderation --> shared
    auth --> shared
    users --> shared
    room --> shared
    message --> shared
    shared["shared/<br/>config · http · infra · permissions · domain"]
```

`shared/` holds cross-cutting code: `config`, `http` (server, JWT middleware, rate limiter, health), `infra` (database, Redis, JWT and mail services), `permissions` (the authorization matrix) and `domain` (shared errors, transaction types).

### Wiring

`cmd/server/main.go` builds the config, database pool, Redis client, services and repositories, then calls each module's `RegisterModule(...)`, which constructs the module's use cases and controller and registers its routes. Adding a dependency means threading it through `main.go` and the module's `RegisterModule` signature.

### Transactions

Use cases that must change several tables atomically run through `Store` (`shared/infra/db/sql_store.go`), which wraps a function in a database transaction. Controllers never touch the database.

## 3. Request lifecycle

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gin router
    participant RL as Rate limiter
    participant J as JWT middleware
    participant UC as Use case
    participant DB as PostgreSQL
    participant WS as WebSocket hub
    C->>G: POST /api/v1/messages
    G->>RL: count by IP (Redis)
    RL-->>G: allowed
    G->>J: validate access token
    J-->>G: user id
    G->>UC: CreateMessage(input)
    UC->>UC: permissions check (membership, ban, rank)
    UC->>DB: INSERT message
    UC->>WS: Broadcast("new_message", room)
    UC-->>C: 201 + message
    WS-->>C: event to every member of the room
```

## 4. Frontend

```mermaid
flowchart TB
    subgraph Next["Next.js App Router · frontend/src"]
        R["app/<br/>(auth) · (private)"]
        CMP[components]
        HK["hooks<br/>useMessages · useRoom · useRoomActivity"]
        CTX["contexts<br/>AuthContext · SocketContext"]
        SV["services/*Service.ts"]
        LIB["lib/api.ts<br/>axios + silent refresh"]
        R --> CMP --> HK
        HK --> SV --> LIB
        HK --> CTX
    end
    LIB -->|REST| API[(Backend)]
    CTX -->|WebSocket| API
```

See [Frontend](frontend.md) for details.

## 5. Deployment topology

```mermaid
flowchart LR
    LB[Load balancer] --> A1[backend instance 1]
    LB --> A2[backend instance 2]
    A1 <--> R[(Redis pub/sub)]
    A2 <--> R
    A1 --> PG[(PostgreSQL)]
    A2 --> PG
```

Any number of backend instances can serve the same room because events travel through Redis. See [Real-time](realtime.md). Deployment automation (images, environments, backups) is outside the scope of this documentation.

## 6. Configuration

`config.NewConfig` reads environment variables, optionally pre-loading the file named by `CONFIG_FILE`. `DATABASE_URL` (or `DB_USER`/`DB_HOST`/`DB_PORT`/`DB_NAME`) is required, and `REDIS_URL` can likewise be assembled from `REDIS_*` parts. The server refuses to start without `JWT_SECRET`. The committed templates are `backend/.env.example` and `frontend/.env.example`; real env files are git-ignored.

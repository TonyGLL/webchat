[English](../en/architecture.md) · **🌐 Español**

# Arquitectura

## 1. Contexto del sistema

WebChat es un **monolito modular** con dos aplicaciones desplegables de forma independiente en un mismo repositorio.

```mermaid
flowchart TB
    subgraph Cliente
        B[Navegador<br/>Next.js 16 · React 19]
    end
    subgraph Server["backend/ (Go 1.26)"]
        H[API HTTP<br/>Gin · /api/v1]
        W[Hub WebSocket<br/>/api/v1/ws]
    end
    PG[(PostgreSQL 16)]
    R[(Redis)]
    B -- "REST + JWT" --> H
    B <-- "eventos" --> W
    H --> PG
    H --> R
    W --> R
    W -. "consulta de salas mediante un adaptador" .-> H
```

| Pieza | Responsabilidad | Puerto (host) |
|---|---|---|
| Frontend | Interfaz, estado de sesión, actualizaciones en vivo | 3100 |
| Backend | API REST, hub WebSocket, migraciones al arrancar | 8081 |
| PostgreSQL | Datos duraderos: usuarios, salas, mensajes, auditoría | 5433 |
| Redis | Estado efímero: ids de refresh token, tickets WS, presencia, contadores de límite, bus de eventos entre instancias | 6380 |

## 2. Backend: Clean Architecture, un paquete por módulo

Cada módulo de negocio bajo `backend/internal/` tiene las mismas cuatro capas. Las dependencias apuntan **solo hacia dentro**: presentation → application → domain, y persistence implementa las interfaces que declara el dominio.

```mermaid
flowchart LR
    P[presentation<br/>controladores Gin] --> A[application<br/>casos de uso · DTOs]
    A --> D[domain<br/>entidades · interfaces de repositorio]
    PS[persistence<br/>SQL / Redis] -. implementa .-> D
```

Módulos: `auth`, `users`, `room`, `message`, `moderation`, más `websocket` y `shared`.

```mermaid
flowchart TB
    main["cmd/server/main.go<br/>(cableado manual)"]
    main --> auth & users & room & message & moderation & websocket
    message -- "WebsocketBroadcaster (interfaz)" --> websocket
    moderation -- "Realtime (interfaz)" --> websocket
    websocket -- "RoomLister (interfaz)" --> room
    moderation --> shared
    auth --> shared
    users --> shared
    room --> shared
    message --> shared
    shared["shared/<br/>config · http · infra · permissions · domain"]
```

`shared/` contiene el código transversal: `config`, `http` (servidor, middleware JWT, limitador de peticiones, health), `infra` (base de datos, Redis, servicios de JWT y correo), `permissions` (la matriz de autorización) y `domain` (errores compartidos, tipos de transacción).

### Cableado

`cmd/server/main.go` construye la configuración, el pool de base de datos, el cliente de Redis, los servicios y los repositorios, y luego llama al `RegisterModule(...)` de cada módulo, que construye los casos de uso y el controlador del módulo y registra sus rutas. Añadir una dependencia implica pasarla por `main.go` y por la firma de `RegisterModule` del módulo.

### Transacciones

Los casos de uso que deben modificar varias tablas de forma atómica pasan por `Store` (`shared/infra/db/sql_store.go`), que envuelve una función en una transacción de base de datos. Los controladores nunca tocan la base de datos.

## 3. Ciclo de vida de una petición

```mermaid
sequenceDiagram
    participant C as Cliente
    participant G as Router Gin
    participant RL as Limitador
    participant J as Middleware JWT
    participant UC as Caso de uso
    participant DB as PostgreSQL
    participant WS as Hub WebSocket
    C->>G: POST /api/v1/messages
    G->>RL: contar por IP (Redis)
    RL-->>G: permitido
    G->>J: validar access token
    J-->>G: id de usuario
    G->>UC: CreateMessage(input)
    UC->>UC: comprobar permisos (membresía, ban, rango)
    UC->>DB: INSERT del mensaje
    UC->>WS: Broadcast("new_message", sala)
    UC-->>C: 201 + mensaje
    WS-->>C: evento a todos los miembros de la sala
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
        LIB["lib/api.ts<br/>axios + renovación silenciosa"]
        R --> CMP --> HK
        HK --> SV --> LIB
        HK --> CTX
    end
    LIB -->|REST| API[(Backend)]
    CTX -->|WebSocket| API
```

Más detalle en [Frontend](frontend.md).

## 5. Topología de despliegue

```mermaid
flowchart LR
    LB[Balanceador] --> A1[instancia backend 1]
    LB --> A2[instancia backend 2]
    A1 <--> R[(Redis pub/sub)]
    A2 <--> R
    A1 --> PG[(PostgreSQL)]
    A2 --> PG
```

Cualquier número de instancias puede atender la misma sala porque los eventos viajan por Redis. Ver [Tiempo real](realtime.md). La automatización del despliegue (imágenes, entornos, copias de seguridad) queda fuera del alcance de esta documentación.

## 6. Configuración

`config.NewConfig` lee variables de entorno y, opcionalmente, precarga el archivo indicado en `CONFIG_FILE`. Se exige `DATABASE_URL` (o `DB_USER`/`DB_HOST`/`DB_PORT`/`DB_NAME`) y `REDIS_URL` también puede armarse con las partes `REDIS_*`. El servidor se niega a arrancar sin `JWT_SECRET`. Las plantillas versionadas son `backend/.env.example` y `frontend/.env.example`; los archivos de entorno reales están ignorados por git.

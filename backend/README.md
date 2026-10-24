# Backend (Go)

API de WebChat: Go + Gin, PostgreSQL, Redis y WebSockets, con Arquitectura Limpia (un paquete por módulo bajo `internal/`: `auth`, `users`, `room`, `message`, además de `websocket` y `shared`).

La guía completa para levantarlo (variables de entorno, esquema, puertos y solución de problemas) está en [`../docs/LEVANTAR.md`](../docs/LEVANTAR.md).

## Comandos

```bash
make watch        # recarga automática con air (usa .env.local)
make run          # sin recarga (usa .env.local)
make test         # go test -v ./...
make build
make database-up  # solo PostgreSQL
make up-dev       # stack completo en Docker (usa .env.dev)
make up-prod      # idem con .env.prod
make down
```

## Configuración

Copia la plantilla y rellena los valores (`.env.example` documenta cada variable):

```bash
cp .env.example .env.local
```

Obligatoria: `JWT_SECRET` (mínimo 32 caracteres) y la conexión a PostgreSQL y a Redis. Puertos en tu máquina: API `8081`, PostgreSQL `5433`, Redis `6380`.

## Endpoints

Prefijo `/api/v1`, excepto `/health` y `/ready`. Salvo registro, login, refresh y verificación de email, todas requieren `Authorization: Bearer <access token>`.

| Método y ruta | Descripción |
|---|---|
| `GET /health` · `GET /ready` | Estado del proceso · estado de PostgreSQL y Redis |
| `POST /auth/register` | Crea el usuario (contraseña de 8 a 72 caracteres, no común) |
| `POST /auth/login` | Devuelve `token` (1 h) y `refreshToken` (7 días, un solo uso) |
| `POST /auth/refresh-token` | Cambia un refresh token por un par nuevo |
| `GET /users/profile` · `DELETE /users/deactivate` | Perfil propio · desactivar cuenta |
| `POST /rooms` · `GET /rooms` | Crear sala (`is_private`) · salas del usuario |
| `POST /rooms/join/:room_id` | Unirse a una sala pública |
| `POST /rooms/read/:room_id` | Marcar la sala como leída (el contador `unread_count` viene en `GET /rooms`) |
| `GET /rooms/members/:room_id` | Miembros de la sala con estado en línea (solo miembros) |
| `POST /rooms/leave/:room_id` | Salir de una sala (el dueño no puede; 409) |
| `POST /rooms/invite/:room_id` | Crear invitación (solo dueño o admin) |
| `POST /rooms/accept-invite/:code` | Aceptar una invitación |
| `GET /messages/:roomId/messages` | Historial paginado por cursor (`limit`, `before`=id del mensaje más antiguo cargado); solo miembros |
| `POST /messages` · `PATCH /messages/:id` · `DELETE /messages/:id` | Enviar (con `reply_to_message_id` opcional para responder) · editar · borrar (solo el autor; borrado lógico) |
| `PUT /messages/:id/pin` · `DELETE /messages/:id/pin` · `GET /messages/:roomId/pinned` | Fijar/desfijar (solo dueño o admin de la sala) · listar fijados |
| `POST/DELETE /messages/:id/reactions` | Añadir · quitar reacción |
| `POST /ws-ticket` → `GET /ws?ticket=` | Ticket de un solo uso (30 s) y conexión WebSocket |

Eventos que el servidor emite por WebSocket a la sala: `new_message`, `update_message`, `reaction_added`, `reaction_removed`.

## Seguridad y operación

- Límite de peticiones por IP en login, registro, refresh, tickets y envío de mensajes (respuesta `429` con `Retry-After`).
- El WebSocket solo acepta orígenes de `CORS_ALLOWED_ORIGINS`.
- Logs estructurados: JSON si `APP_ENV=prod`, texto en los demás casos; cada petición lleva `X-Request-ID`.
- Imagen Docker mínima (distroless, usuario no root).

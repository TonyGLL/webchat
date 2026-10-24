[English](../en/data-model.md) · **🌐 Español**

# Modelo de datos

PostgreSQL 16. El esquema lo definen los archivos de `backend/migrations/` y se aplica al arrancar.

```mermaid
erDiagram
    users ||--o| passwords : "tiene"
    users ||--o{ rooms : "es propietario de"
    users ||--o{ room_members : "se une"
    rooms ||--o{ room_members : "tiene"
    rooms ||--o{ messages : "contiene"
    users ||--o{ messages : "escribe"
    messages ||--o{ messages : "responde a"
    messages ||--o{ reactions : "recibe"
    users ||--o{ reactions : "añade"
    rooms ||--o| direct_rooms : "es un DM"
    rooms ||--o{ invites : "tiene"
    rooms ||--o{ bans : "tiene"
    users ||--o{ bans : "recibe"
    messages ||--o{ reports : "denunciado en"
    users ||--o{ reports : "presenta"
    users ||--o{ audit_log : "actúa en"
    rooms ||--o{ audit_log : "delimita"
    messages ||--o{ attachments : "reservado"

    users {
        int id PK
        string username UK
        string email UK
        string role "user | moderator | admin"
        bool deleted
        timestamptz email_verified_at
    }
    rooms {
        uuid id PK
        int owner_id FK
        string name
        bool is_private
        bool is_direct
    }
    room_members {
        uuid room_id PK
        int user_id PK
        enum role "member | moderator | admin | owner"
        timestamptz last_read_at
    }
    messages {
        uuid id PK
        uuid room_id FK
        int author_id FK "NULL tras borrar la cuenta"
        text content
        uuid reply_to_message_id FK
        timestamptz edited_at
        timestamptz deleted_at
        timestamptz pinned_at
    }
    reactions {
        uuid message_id PK
        int user_id PK
        string emoji PK
    }
    direct_rooms {
        int user_a PK "user_a < user_b"
        int user_b PK
        uuid room_id UK
    }
    bans {
        int user_id FK
        uuid room_id FK
        timestamptz expires_at "NULL = permanente"
    }
    reports {
        int id PK
        uuid reported_message_id FK
        enum status "pending | resolved | ignored"
    }
    audit_log {
        bigint id PK
        string action
        jsonb details
    }
```

## Notas de diseño

| Tema | Decisión |
|---|---|
| Identificadores | `users` usa enteros seriales; `rooms`, `messages` y `attachments` usan UUID para poder crearse y referenciarse desde el cliente sin revelar cantidades |
| Borrado | `messages.deleted_at` es un borrado lógico. `author_id` es `ON DELETE SET NULL`, lo que convierte una cuenta borrada en "usuario eliminado" |
| Contadores de no leídos | `room_members.last_read_at`; el contador es el número de mensajes más nuevos en esa sala |
| Mensajes directos | Una sala normal con `is_direct`, más `direct_rooms`, que garantiza una por pareja (`CHECK (user_a < user_b)`) |
| Mensajes fijados | `pinned_at` más un índice parcial que solo cubre las filas fijadas |
| Búsqueda | Índice GIN de texto completo sobre `messages.content`, sin distinguir acentos |
| Rol global | `users.role` con un `CHECK`, modificado solo por `cmd/admin` |
| Refresh tokens | Sus ids viven en Redis. La tabla `refresh_tokens` de la primera migración existe pero no se usa |
| Adjuntos | La tabla está reservada para la futura función de adjuntos; por ahora nada escribe en ella |

## Índices principales

| Índice | Para qué sirve |
|---|---|
| `messages (room_id, created_at DESC)` | paginación del historial |
| `messages (room_id, pinned_at DESC) WHERE pinned_at IS NOT NULL` | lista de fijados |
| `messages USING GIN (to_tsvector('simple', content))` | búsqueda |
| `invites (room_id, expires_at)` | invitaciones vigentes |
| `audit_log (room_id, created_at DESC)` | auditoría por sala |

## Historial de migraciones

| # | Cambio |
|---|---|
| 1 | Esquema inicial |
| 2 | Seguimiento de no leídos (`last_read_at`) |
| 3 | Mensajes fijados |
| 4 | Mensajes directos |
| 5–6 | Búsqueda de texto completo, luego sin distinguir acentos |
| 7 | Rol de sala `moderator` |
| 8 | Roles globales y `audit_log` |
| 9 | Una denuncia por persona y mensaje |
| 10 | Soporte para borrar la cuenta |

Regla: nunca se edita una migración ya aplicada. Se añade un nuevo par numerado de archivos `.up.sql` y `.down.sql`.

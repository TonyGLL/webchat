# Hallazgos

Registro de problemas detectados. Cuando se resuelve uno, se mueve de **Pendientes** a **Completados** y se anota fecha y commit/PR.

**Severidad:** 🔴 crítica · 🟠 alta · 🟡 media · 🟢 baja
**Verificación:** *Confirmado* = visto directamente en el código · *Por confirmar* = sospecha que hay que comprobar antes de actuar.
**Fase:** ver [ROADMAP.md](ROADMAP.md).

Plantilla para añadir uno nuevo:

```
### H-XXX Título
- Severidad / Fase / Área:
- Verificación:
- Descripción:
- Ubicación:
- Propuesta:
```

---

## Pendientes

### Seguridad

### H-004 Tokens guardados en `localStorage`
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** frontend · **Verificación:** Confirmado
- **Descripción:** cualquier XSS puede robar access y refresh token.
- **Ubicación:** `frontend/src/hooks/useAuth.ts` y otros 3 puntos que leen `authTokens`.
- **Propuesta:** cookies httpOnly + SameSite, o al menos centralizar el acceso en un único módulo.

### H-008 Endpoint de envío de verificación sin autenticación
- **Nota:** la verificación de email está pausada (decisión del 2026-10-23); retomar junto con este hallazgo.
- **Severidad:** 🟡 · **Fase:** 1 · **Área:** backend · **Verificación:** Confirmado (la ruta `POST /auth/send-verify-email/:user_id` es pública)
- **Descripción:** cualquiera puede disparar correos a un usuario por su id.
- **Propuesta:** rate limiting y evitar que el id sea enumerable; revisar si debe exigir autenticación.

### H-010 Imagen Docker sobre `scratch` y como root
- **Severidad:** 🟢 · **Fase:** 5 · **Área:** backend · **Verificación:** Confirmado
- **Descripción:** sin certificados CA (problemas con SMTP/TLS saliente) ni usuario no root; copia el binario dos veces.
- **Ubicación:** `backend/Dockerfile`
- **Propuesta:** `distroless/static` o alpine con certificados y `USER` no root.

### Calidad y pruebas

### H-020 No hay tests
- **Severidad:** 🔴 · **Fase:** 0 · **Área:** front + back · **Verificación:** Confirmado
- **Descripción:** no existen archivos `_test.go` ni tests en el frontend.
- **Propuesta:** empezar por use cases de auth y mensajes; Vitest en frontend.

### H-021 No hay CI
- **Severidad:** 🟠 · **Fase:** 0 · **Área:** repo · **Verificación:** Confirmado (no existe `.github/`)
- **Propuesta:** workflow con vet, test, lint y build.

### H-022 Tipado débil en el frontend
- **Severidad:** 🟡 · **Fase:** 2 · **Área:** frontend · **Verificación:** Confirmado
- **Descripción:** `useState<any>` para el usuario; (el id de usuario es `SERIAL`, así que `number` es correcto en ese caso).
- **Ubicación:** `frontend/src/hooks/useAuth.ts`, `frontend/src/services/authService.ts`
- **Propuesta:** tipos generados desde OpenAPI; activar reglas de lint contra `any`.

### H-023 Estado de auth duplicado por hook
- **Severidad:** 🟡 · **Fase:** 1 · **Área:** frontend · **Verificación:** Confirmado
- **Descripción:** `useAuth` es un hook con estado local, no un contexto. Cada componente que lo usa crea su propia copia y vuelve a pedir el perfil.
- **Propuesta:** `AuthContext` único.

### H-024 Sin renovación automática del token
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** frontend · **Verificación:** Confirmado (no hay interceptores de axios)
- **Descripción:** el backend ofrece refresh pero el frontend lo ignora; al expirar el access token se cierra la sesión.
- **Propuesta:** interceptor con cola de peticiones durante la renovación.

### Datos y rendimiento

### H-030 Migraciones manuales
- **Severidad:** 🟠 · **Fase:** 0 · **Área:** backend · **Verificación:** Confirmado
- **Descripción:** solo existe `migrations/schema.sql`; no hay control de versiones del esquema.
- **Propuesta:** golang-migrate o goose, ejecutadas en el arranque o en un paso de despliegue.

### H-031 Paginación del historial
- **Severidad:** 🟡 · **Fase:** 2 · **Área:** front + back · **Verificación:** Por confirmar
- **Descripción:** verificar que `GET /messages/:roomId/messages` pagina por cursor y que el frontend carga más al hacer scroll.
- **Ubicación:** `backend/internal/message/application/list_messages_use_case.go`, `frontend/src/hooks/useMessages.ts`

### H-032 Hub de WebSocket en memoria
- **Severidad:** 🟡 · **Fase:** 5 · **Área:** backend · **Verificación:** Confirmado
- **Descripción:** el estado de salas y clientes vive en un solo proceso; no se puede escalar horizontalmente.
- **Propuesta:** Redis pub/sub entre instancias.

### H-033 Mutex y canal sin buffer en el hub
- **Severidad:** 🟡 · **Fase:** 5 · **Área:** backend · **Verificación:** Por confirmar
- **Descripción:** `broadcast`, `register` y `unregister` son canales sin buffer y hay un `sync.Mutex` junto a un bucle de eventos; revisar posibles bloqueos con clientes lentos y que `rooms` se proteja de forma uniforme.
- **Ubicación:** `backend/internal/websocket/hub.go`

### H-034 Límite de 512 bytes por mensaje entrante del cliente
- **Severidad:** 🟢 · **Fase:** 3 · **Área:** backend · **Verificación:** Confirmado
- **Descripción:** `maxMessageSize = 512` aplica a lo que el cliente envía por WebSocket. Hoy no se procesa nada, pero habrá que ampliarlo si se añaden eventos (typing, ack).
- **Ubicación:** `backend/internal/websocket/client.go`

### Operación y documentación

### H-051 Contraseñas escritas directamente en docker-compose.yml
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** infra · **Verificación:** Confirmado (las claves `POSTGRES_PASSWORD` y `REDIS_PASSWORD` tienen valor literal en un archivo versionado; no se revisó el valor)
- **Ubicación:** `backend/docker-compose.yml`
- **Propuesta:** leerlas de variables de entorno (`${POSTGRES_PASSWORD}`) definidas en un archivo ignorado por git, y rotar las actuales porque ya están en el historial.

### H-054 El script `lint` del frontend no funciona
- **Severidad:** 🟡 · **Fase:** 0 · **Área:** frontend · **Verificación:** Confirmado (`bun run lint` falla)
- **Descripción:** `next lint` ya no existe en Next 16; el script intenta tratar `lint` como un directorio.
- **Ubicación:** `frontend/package.json`
- **Propuesta:** cambiar a `eslint .` (ya hay `eslint.config.mjs`) y añadirlo al CI.

### H-050 Rutas del frontend que no coinciden con el backend
- **Severidad:** 🟠 · **Fase:** 0 · **Área:** front + back · **Verificación:** Confirmado en el código (falta probarlo en ejecución)
- **Descripción:** el frontend llama `POST /auth/refresh`, pero el backend expone `POST /auth/refresh-token` (y espera `refreshToken` en el cuerpo). La ruta de verificación de email ya se corrigió en H-053. En `roomService.ts` también hay rutas que no existen en el backend (`/rooms/:id/join`, `/rooms/:id/invites`, `/rooms/invites/:token/accept`; el backend usa `/rooms/join/:room_id`, `/rooms/invite/:room_id`, `/rooms/accept-invite/:code`).
- **Ubicación:** `frontend/src/services/authService.ts`, `backend/internal/auth/module.go`
- **Propuesta:** alinear las rutas y generar el cliente desde OpenAPI para que no se desincronicen.

### H-040 README raíz y de backend desactualizados
- **Severidad:** 🟡 · **Fase:** 0 · **Área:** docs · **Verificación:** Confirmado
- **Descripción:** el README raíz menciona un `docker-compose.yml` en la raíz que no existe (está en `backend/`); el del backend describe `POST /messages` con `authorId`/`channelId` y el puerto 3000, mientras la app usa 8080 por defecto.
- **Propuesta:** reescribir ambos y mantener la lista de endpoints en OpenAPI.

### H-042 Sin health check ni métricas
- **Severidad:** 🟡 · **Fase:** 0 y 5 · **Área:** backend · **Verificación:** Confirmado (no hay ruta `/health` registrada en `main.go`)
- **Propuesta:** `/health` y `/ready`, luego métricas Prometheus.

### H-043 Logs sin estructura
- **Severidad:** 🟢 · **Fase:** 0 · **Área:** backend · **Verificación:** Confirmado (`log.Printf` y `gin.Default()`)
- **Propuesta:** `slog` con JSON y request id.

### H-044 Tablas sin uso (attachments, bans, reports)
- **Severidad:** 🟢 · **Fase:** 3 y 4 · **Área:** backend · **Verificación:** Por confirmar
- **Descripción:** existen en el esquema pero no encontré módulos que las usen. Implementarlas o retirarlas.

---

## En progreso

_(nada por ahora)_

---

## Completados

| ID | Título | Fecha | Commit / PR |
|----|--------|-------|-------------|
| H-025 | Estilo global inyectado desde el dashboard (y fondo `/whatsapp-bg.png` inexistente) | 2026-10-23 | 3422373 |
| H-026 | Navegación móvil incompleta: ahora se alterna entre lista de salas y chat | 2026-10-23 | 3422373 |
| H-005 | Sin rate limiting: ahora login (10/min), registro (20/10 min), refresh (30/min), reenvío de verificación (5/10 min), ws-ticket y mensajes, por IP, con Redis y respuesta 429 | 2026-10-23 | _(este commit)_ |
| H-009 | Política de contraseña: 8–72 caracteres, rechaza contraseñas comunes y las que contienen tu usuario o email (validado en backend y frontend) | 2026-10-23 | _(este commit)_ |
| H-007 | Salas y acceso: unirse a salas públicas fallaba con 500; las salas "privadas" del frontend quedaban públicas (`is_public` vs `is_private`); cualquiera podía crear invitaciones de salas ajenas (ahora solo dueño/admin); invitaciones repetidas/inexistentes dan 409/404 | 2026-10-23 | 7cada01 |
| H-006 | Autorización por sala: enviar, leer, editar y reaccionar exigen ser miembro; editar exige ser el autor (403 si no). Verificado en vivo y con tests | 2026-10-23 | 8d991fd |
| H-003 | Token JWT en la URL del WebSocket; ahora se canjea por un ticket de un solo uso (30 s) vía `POST /ws-ticket` | 2026-10-23 | b1ca39e |
| H-002 | WebSocket aceptaba cualquier origen; ahora solo los de `CORS_ALLOWED_ORIGINS` (y respuesta 401 clara sin token) | 2026-10-23 | b2791fd |
| H-055 | Registro dejaba usuarios huérfanos sin contraseña (el escaneo de `last_access` NULL fallaba y la "transacción" no usaba la conexión transaccional) | 2026-10-23 | 8e79c4c |
| H-052 | El registro del frontend no enviaba los campos que exige el backend (name, last_name, phone, contraseña de 8+); además nunca se pedía el envío del correo de verificación | 2026-10-23 | 71d50fb |
| H-053 | Enlace de verificación de email sin dominio; ahora usa `FRONTEND_URL/verify/<token>` y la página llama a la ruta correcta del backend | 2026-10-23 | 71d50fb |
| H-041 | Puertos inconsistentes (ahora: front 3100, back 8081, db 5433, redis 6380) | 2026-10-23 | a1411ec |
| H-001 | Secreto JWT con valor por defecto | 2026-10-23 | 8e8d3ce |
| H-000 | Falta `CLAUDE.md` con guía del repositorio | 2026-10-23 | 8b94309 |

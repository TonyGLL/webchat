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

### H-001 Secreto JWT con valor por defecto
- **Severidad:** 🔴 · **Fase:** 1 · **Área:** backend · **Verificación:** Confirmado
- **Descripción:** si `JWT_SECRET` no está definido se usa `"default-secret"`, lo que permite falsificar tokens.
- **Ubicación:** `backend/internal/shared/config/config.go`
- **Propuesta:** devolver error al arrancar si falta o es demasiado corto.

### H-002 WebSocket acepta cualquier origen
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** backend · **Verificación:** Confirmado
- **Descripción:** `CheckOrigin` devuelve siempre `true` (comentario en el código: "in production, you should check the origin").
- **Ubicación:** `backend/internal/websocket/client.go`
- **Propuesta:** validar contra `CORS_ALLOWED_ORIGINS`.

### H-003 Token JWT en la URL del WebSocket
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** front + back · **Verificación:** Confirmado
- **Descripción:** el token se envía como `?token=...`, queda en logs de proxies y en el historial.
- **Ubicación:** `frontend/src/contexts/SocketContext.tsx`, `backend/internal/websocket`
- **Propuesta:** ticket de un solo uso y vida corta pedido por REST, o cookie httpOnly.

### H-004 Tokens guardados en `localStorage`
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** frontend · **Verificación:** Confirmado
- **Descripción:** cualquier XSS puede robar access y refresh token.
- **Ubicación:** `frontend/src/hooks/useAuth.ts` y otros 3 puntos que leen `authTokens`.
- **Propuesta:** cookies httpOnly + SameSite, o al menos centralizar el acceso en un único módulo.

### H-005 Sin rate limiting
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** backend · **Verificación:** Confirmado (no hay middleware de límite)
- **Descripción:** login, registro, reenvío de verificación y envío de mensajes no tienen límite.
- **Propuesta:** middleware con Redis (token bucket) por IP y por usuario.

### H-006 Autorización por membresía de sala
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** backend · **Verificación:** Por confirmar
- **Descripción:** comprobar que enviar, listar, editar mensajes y reaccionar exigen ser miembro de la sala, y que editar exige ser el autor.
- **Ubicación:** `backend/internal/message/application/*`
- **Propuesta:** revisar cada use case y cubrirlo con tests.

### H-007 Unirse a cualquier sala por id
- **Severidad:** 🟡 · **Fase:** 1 · **Área:** backend · **Verificación:** Por confirmar
- **Descripción:** `POST /rooms/join/:room_id` podría permitir entrar a salas privadas sin invitación.
- **Ubicación:** `backend/internal/room/application/join_room_use_case.go`
- **Propuesta:** distinguir salas públicas/privadas.

### H-008 Endpoint de envío de verificación sin autenticación
- **Severidad:** 🟡 · **Fase:** 1 · **Área:** backend · **Verificación:** Confirmado (la ruta `POST /auth/send-verify-email/:user_id` es pública)
- **Descripción:** cualquiera puede disparar correos a un usuario por su id.
- **Propuesta:** rate limiting y evitar que el id sea enumerable; revisar si debe exigir autenticación.

### H-009 Política de contraseña débil
- **Severidad:** 🟢 · **Fase:** 1 · **Área:** front + back · **Verificación:** Confirmado (mínimo 6 en el esquema del frontend)
- **Propuesta:** mínimo 10, comprobación contra contraseñas comunes, validar también en el backend.

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
- **Descripción:** `useState<any>` para el usuario; `User.id` tipado como `number` mientras el backend usa `uuid` en otros módulos (por confirmar el tipo real del id de usuario).
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

### H-025 Estilo global inyectado desde el dashboard
- **Severidad:** 🟢 · **Fase:** 2 · **Área:** frontend · **Verificación:** Confirmado
- **Descripción:** `<style jsx global>` aplica fondo a todos los `main`.
- **Ubicación:** `frontend/src/app/(private)/dashboard/page.tsx`
- **Propuesta:** mover a una clase de Tailwind o CSS del componente.

### H-026 Navegación móvil incompleta
- **Severidad:** 🟡 · **Fase:** 2 · **Área:** frontend · **Verificación:** Por confirmar
- **Descripción:** el layout usa `md:grid-cols-[350px_1fr]`; en pantallas pequeñas no se ve cómo se alterna entre lista y chat.
- **Ubicación:** `frontend/src/app/(private)/dashboard/page.tsx`

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

### H-040 README raíz y de backend desactualizados
- **Severidad:** 🟡 · **Fase:** 0 · **Área:** docs · **Verificación:** Confirmado
- **Descripción:** el README raíz menciona un `docker-compose.yml` en la raíz que no existe (está en `backend/`); el del backend describe `POST /messages` con `authorId`/`channelId` y el puerto 3000, mientras la app usa 8080 por defecto.
- **Propuesta:** reescribir ambos y mantener la lista de endpoints en OpenAPI.

### H-041 Puertos inconsistentes
- **Severidad:** 🟢 · **Fase:** 0 · **Área:** backend/infra · **Verificación:** Confirmado
- **Descripción:** el `Dockerfile` expone 3000, la configuración usa 8080 por defecto y el frontend apunta a 8080.

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
| H-000 | Falta `CLAUDE.md` con guía del repositorio | 2026-10-23 | 8b94309 |

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

### H-008 Endpoint de envío de verificación sin autenticación
- **Nota:** la verificación de email está pausada (decisión del 2026-10-23); retomar junto con este hallazgo.
- **Severidad:** 🟡 · **Fase:** 1 · **Área:** backend · **Verificación:** Confirmado (la ruta `POST /auth/send-verify-email/:user_id` es pública)
- **Descripción:** cualquiera puede disparar correos a un usuario por su id.
- **Propuesta:** rate limiting y evitar que el id sea enumerable; revisar si debe exigir autenticación.

### Calidad y pruebas

### Datos y rendimiento

### Operación y documentación

### H-051 Contraseñas escritas directamente en docker-compose.yml
- **Severidad:** 🟠 · **Fase:** 1 · **Área:** infra · **Verificación:** Confirmado (las claves `POSTGRES_PASSWORD` y `REDIS_PASSWORD` tienen valor literal en un archivo versionado; no se revisó el valor)
- **Ubicación:** `backend/docker-compose.yml`
- **Propuesta:** leerlas de variables de entorno (`${POSTGRES_PASSWORD}`) definidas en un archivo ignorado por git, y rotar las actuales porque ya están en el historial.

---

## En progreso

_(nada por ahora)_

---

## Completados

| ID | Título | Fecha | Commit / PR |
|----|--------|-------|-------------|
| H-025 | Estilo global inyectado desde el dashboard (y fondo `/whatsapp-bg.png` inexistente) | 2026-10-23 | 3422373 |
| H-026 | Navegación móvil incompleta: ahora se alterna entre lista de salas y chat | 2026-10-23 | 3422373 |
| H-064 | (Roadmap Fase 2) Mensajes optimistas: al enviar aparece al instante ("Sending…"); si falla se marca "Not sent" con Retry/Discard; si el WebSocket ya lo entregó no se duplica | 2026-10-23 | _(este commit)_ |
| H-063 | (Roadmap Fase 2) Reconexión automática del WebSocket con espera exponencial y jitter (1 s → 30 s) y recarga de mensajes tras reconectar para recuperar los que se perdieron | 2026-10-23 | 231c885 |
| H-044 | Tablas `attachments`, `bans` y `reports`: confirmado que ningún módulo las usa; se conservan a propósito porque las fases 3 y 4 del ROADMAP (adjuntos, moderación) las necesitan | 2026-10-23 | 6071ea6 |
| H-004 | Tokens fuera de `localStorage`: el refresh token va en cookie `httpOnly` + `SameSite=Strict` (solo se envía a `/api/v1/auth`), el access token vive solo en memoria y la sesión se restaura al recargar con la cookie; nuevo `POST /auth/logout` que revoca el token | 2026-10-23 | 17a6b6c |
| H-020 | Pruebas automáticas: backend (config, política de contraseñas, autorización de mensajes/salas/invitaciones, rate limit, health, origen y tickets de WebSocket, hub con `-race`, bus entre instancias) y frontend con Vitest (almacenamiento de tokens y renovación del token); ambas corren en CI | 2026-10-23 | 838956c |
| H-032 | Hub de WebSocket multi-instancia: los eventos y las suscripciones viajan por Redis pub/sub, así que varias instancias del backend pueden servir las mismas salas (probado con un bus en memoria entre dos hubs y en vivo con Redis) | 2026-10-23 | b4604f6 |
| H-031 | Historial por cursor (`before` = id del mensaje más antiguo) en vez de página/offset, que duplicaba mensajes con tráfico en vivo; oculta mensajes borrados; botón "Load older messages" que conserva el scroll | 2026-10-23 | 555e434 |
| H-030 | Migraciones versionadas con golang-migrate, aplicadas al arrancar; las bases creadas a mano se adoptan como versión 1 | 2026-10-23 | 42e1b61 |
| H-033 | Hub de WebSocket: se bloqueaba para siempre con un cliente lento (re-tomaba su propio candado), el canal de eventos no tenía búfer y había una carrera al registrar clientes. Reescrito y probado con `-race` | 2026-10-23 | a7758b5 |
| H-062 | Las salas creadas/unidas después de conectar no recibían mensajes en vivo hasta reconectar: ahora el hub suscribe al usuario al crear sala, unirse o aceptar invitación | 2026-10-23 | a7758b5 |
| H-034 | Límite de 512 bytes del cliente por WebSocket: documentado en el código (hoy el cliente no envía nada por ahí) | 2026-10-23 | a7758b5 |
| H-010 | Imagen Docker: ahora distroless con certificados CA, usuario no root, binario compilado sin CGO (42 MB) | 2026-10-23 | 13b60f9 |
| H-021 | CI en GitHub Actions: formato, vet, build y tests (backend); tipos, lint y build (frontend) | 2026-10-23 | 13b60f9 |
| H-040 | README raíz y de backend reescritos con puertos, rutas y comandos reales | 2026-10-23 | 13b60f9 |
| H-042 | `GET /health` y `GET /ready` (comprueba PostgreSQL y Redis) | 2026-10-23 | 13b60f9 |
| H-043 | Logs estructurados con `slog` (JSON en producción) y un `X-Request-ID` por petición | 2026-10-23 | 13b60f9 |
| H-061 | Login devolvía un "access token" de 7 días (en realidad un refresh token) y un refresh token que era solo un UUID: la renovación nunca funcionó. Ahora access 1 h + refresh JWT de un solo uso, guardado en Redis; los refresh tokens ya no sirven como access tokens | 2026-10-23 | 4a81494 |
| H-023 | Estado de auth duplicado por hook: ahora un único `AuthContext` (un solo pedido de perfil para toda la app) | 2026-10-23 | 4a81494 |
| H-024 | Renovación automática del token: cliente axios único con interceptor que renueva una vez y reintenta | 2026-10-23 | 4a81494 |
| H-050 | Rutas del frontend que no coincidían con el backend (refresh, verify, rooms join/invite/accept, users/deactivate) y eventos WebSocket con otros nombres (reacciones y edición no llegaban en vivo) | 2026-10-23 | 4a81494 |
| H-022 | Tipado del frontend: sin `any` en código propio; tipos de User, Room, mensajes y eventos WebSocket | 2026-10-23 | 4a81494 |
| H-054 | Script `lint` del frontend: ahora `eslint .` con ESLint 9 + eslint-config-next; 0 errores | 2026-10-23 | 4a81494 |
| H-005 | Sin rate limiting: ahora login (10/min), registro (20/10 min), refresh (30/min), reenvío de verificación (5/10 min), ws-ticket y mensajes, por IP, con Redis y respuesta 429 | 2026-10-23 | 9870410 |
| H-009 | Política de contraseña: 8–72 caracteres, rechaza contraseñas comunes y las que contienen tu usuario o email (validado en backend y frontend) | 2026-10-23 | 9870410 |
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

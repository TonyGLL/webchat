[English](../en/decisions.md) · **🌐 Español**

# Decisiones de diseño

Cada entrada recoge el **contexto**, la **decisión**, las **alternativas** que se evaluaron y las **consecuencias** que aceptamos. Cuando una decisión cambia, se añade una entrada nueva en lugar de reescribir la historia.

| # | Decisión | Estado |
|---|---|---|
| D1 | Monolito modular con Clean Architecture | Aceptada |
| D2 | Go + Gin para el backend | Aceptada |
| D3 | PostgreSQL como fuente de verdad, Redis para estado efímero | Aceptada |
| D4 | Migraciones SQL versionadas (golang-migrate), aplicadas al arrancar | Aceptada |
| D5 | Access JWT corto + refresh token rotativo en cookie httpOnly | Aceptada |
| D6 | El access token solo en memoria en el cliente | Aceptada |
| D7 | WebSocket autenticado con un ticket de un solo uso | Aceptada |
| D8 | Redis pub/sub como bus de eventos entre instancias | Aceptada |
| D9 | Una única matriz de permisos basada en rangos | Aceptada |
| D10 | Los mensajes directos son salas privadas | Aceptada |
| D11 | Paginación por cursor | Aceptada |
| D12 | Búsqueda de texto completo dentro de PostgreSQL | Aceptada |
| D13 | Borrado lógico de mensajes, anonimización de cuentas | Aceptada |
| D14 | Roles globales otorgados solo mediante una CLI | Aceptada |
| D15 | i18n sin dependencias y temas con variables CSS | Aceptada |
| D16 | Envío optimista con estados explícitos | Aceptada |
| D17 | Dirección visual "Foco" | Aceptada |

---

## D1. Monolito modular con Clean Architecture

- **Contexto:** un equipo pequeño, un único producto desplegable, pero un dominio (auth, salas, mensajes, moderación) que seguirá creciendo.
- **Decisión:** un solo proceso Go con un paquete por módulo; cada módulo tiene capas `domain`, `application`, `persistence` y `presentation`. Los módulos se comunican mediante interfaces pequeñas (por ejemplo `WebsocketBroadcaster`, `RoomLister`), nunca importando los paquetes de otro módulo.
- **Alternativas:** microservicios (coste operativo sin beneficio a este tamaño); una estructura plana de "handlers + modelos" (rápida al inicio, pero las reglas de negocio se filtran al código HTTP y dejan de ser testeables).
- **Consecuencias:** los casos de uso se prueban con dobles y sin base de datos. Las interfaces cuestan algo de código repetitivo y el cableado es manual en `main.go`. Si algún módulo hay que extraerlo, las costuras ya existen.

## D2. Go + Gin

- **Contexto:** un servicio dominado por muchas conexiones concurrentes, casi siempre inactivas.
- **Decisión:** Go por sus goroutines baratas y un binario estático único; Gin para rutas y middleware.
- **Alternativas:** Node.js (comparte lenguaje con el frontend, pero el trabajo intensivo de CPU y la memoria por conexión escalan peor); un framework más pesado (más funciones que no necesitamos).
- **Consecuencias:** frontend y backend no comparten tipos, así que los contratos de la API se mantienen a mano (un contrato OpenAPI está en el roadmap).

## D3. PostgreSQL como fuente de verdad, Redis para estado efímero

- **Decisión:** todo lo que debe sobrevivir a un reinicio vive en PostgreSQL. Todo lo que tiene TTL o se comparte entre instancias vive en Redis: ids de refresh token, tickets WebSocket, presencia, contadores de límite y el bus de eventos.
- **Alternativas:** guardar la presencia en el proceso (se rompe con más de una instancia); guardar las sesiones en PostgreSQL (más escrituras para datos que están pensados para caducar).
- **Consecuencias:** dos dependencias con estado que operar. El limitador **falla abierto**: si Redis cae, se permiten las peticiones y se registra el problema, porque aquí pesa más la disponibilidad que la rigurosidad.

## D4. Migraciones SQL versionadas

- **Contexto:** la primera versión creaba las tablas a mano desde un `schema.sql`.
- **Decisión:** archivos de golang-migrate (`NNNNNN_nombre.up.sql` / `.down.sql`) embebidos en el binario y aplicados al arrancar. Una base creada a mano se adopta en la versión 1.
- **Reglas:** nunca se edita una migración ya aplicada; se añade un nuevo par numerado.
- **Consecuencias:** un clon limpio, la CI y producción convergen en el mismo esquema. Arrancar una instancia nueva puede ejecutar migraciones un instante, aceptable a esta escala.

## D5. Access JWT corto + refresh token rotativo

```mermaid
sequenceDiagram
    participant C as Cliente
    participant A as API
    participant R as Redis
    C->>A: POST /auth/login
    A->>R: guardar jti del refresh
    A-->>C: access JWT (1 h) en el cuerpo + refresh JWT en cookie httpOnly
    Note over C: el access token caduca
    C->>A: POST /auth/refresh-token (cookie)
    A->>R: consumir el jti anterior (un solo uso)
    A->>R: guardar el nuevo jti
    A-->>C: nuevo access JWT + nueva cookie de refresh
```

- **Decisión:** un access token de 1 hora devuelto en el cuerpo, y un refresh token de 7 días cuyo id (`jti`) se guarda en Redis y se puede usar **una sola vez**. Al usarlo se emite un par nuevo.
- **Por qué rotar:** un refresh token robado deja de servir en cuanto el cliente legítimo renueva, y uno reutilizado es detectable. `POST /auth/logout` revoca el id vigente.
- **Alternativas:** access tokens de larga duración (una filtración vale días); sesiones en servidor (revocación más simple, pero exige estado en cada petición).

## D6. El access token solo en memoria en el cliente

- **Decisión:** el navegador guarda el access token en una variable de JavaScript (`lib/tokenStorage.ts`), nunca en `localStorage`. El refresh token es una cookie httpOnly, `SameSite=Strict` y limitada a `/api/v1/auth`. Al recargar, `AuthContext` restaura la sesión llamando a refresh.
- **Por qué:** cualquier XSS puede leer `localStorage`; no puede leer una cookie httpOnly, y el token en memoria desaparece con la página.
- **Consecuencias:** una petición extra al cargar la página, y el interceptor de axios debe renovar de forma transparente ante un `401`.

## D7. WebSocket autenticado con un ticket de un solo uso

```mermaid
sequenceDiagram
    participant C as Cliente
    participant A as API
    participant R as Redis
    C->>A: POST /ws-ticket (Bearer JWT)
    A->>R: SET ticket (TTL 30 s)
    A-->>C: ticket
    C->>A: GET /ws?ticket=…  (se valida el Origin)
    A->>R: GETDEL ticket
    A-->>C: conexión actualizada
```

- **Contexto:** los navegadores no pueden enviar la cabecera `Authorization` en un WebSocket, así que el token terminaría en la URL, donde lo registran proxies y logs.
- **Decisión:** canjear el JWT por un ticket de 30 segundos y un solo uso. El servidor además valida la cabecera `Origin` contra `CORS_ALLOWED_ORIGINS`.
- **Consecuencias:** una URL filtrada no sirve pasados unos segundos y no se puede reutilizar.

## D8. Redis pub/sub como bus de eventos

- **Decisión:** el hub publica cada evento en un canal de Redis; cada instancia lo entrega a sus propios clientes conectados. La presencia es un sorted set de Redis compartido por todas las instancias.
- **Alternativas:** sesiones pegajosas (carga desigual y, aun así, los miembros de una sala pueden estar en instancias distintas); Kafka/NATS (más infraestructura de la que el problema necesita).
- **Consecuencias:** la entrega es como máximo una vez. Los clientes reconectan con backoff y recargan el historial reciente para recuperar lo perdido.

## D9. Una única matriz de permisos basada en rangos

- **Contexto:** moderación, fijado, invitaciones y borrado tenían cada uno su propia comparación de roles, que es donde se esconden los errores de autorización.
- **Decisión:** `internal/shared/permissions` es el único lugar que decide quién puede hacer qué. Los roles de sala (`member < moderator < admin < owner`) y los globales se convierten en un rango numérico, y las reglas dicen "quien actúa debe **superar estrictamente** al objetivo".
- **Regla para quien contribuye:** las comprobaciones nuevas pasan por ese paquete, nunca por comparaciones de texto ad hoc. El frontend replica la matriz en `lib/permissions.ts` solo para decidir qué *mostrar*; el servidor siempre decide.

## D10. Los mensajes directos son salas privadas

- **Decisión:** un chat 1 a 1 es una sala privada con `is_direct = true`, una por pareja de usuarios (`direct_rooms`, con `user_a < user_b` forzado por un `CHECK`).
- **Por qué:** historial, contadores de no leídos, búsqueda, reacciones, presencia y entrega por WebSocket funcionan **sin código nuevo**.
- **Consecuencias:** algunas funciones se bifurcan según `is_direct` (sin lista de miembros, sin denuncias).

## D11. Paginación por cursor

- **Decisión:** el historial se pagina por cursor (`?limit=&before=`) sobre `(room_id, created_at DESC)`.
- **Alternativas:** la paginación con `OFFSET` salta o repite filas cuando llegan mensajes nuevos y se vuelve más lenta con la profundidad.

## D12. Búsqueda de texto completo dentro de PostgreSQL

- **Decisión:** un índice GIN sobre `to_tsvector('simple', content)` con coincidencia sin distinguir acentos (`unaccent`), limitado a las salas a las que pertenece el usuario.
- **Alternativas:** Elasticsearch/Meilisearch ofrecen mejor ordenación pero añaden un servicio que operar y sincronizar. Se revisará si la búsqueda pasa a ser una función central.

## D13. Borrado lógico de mensajes, anonimización de cuentas

- **Decisión:** borrar un mensaje fija `deleted_at`, de modo que hilos, denuncias y auditoría siguen siendo coherentes. Borrar una cuenta traspasa las salas propias a otro miembro (o las elimina si el propietario está solo) y conserva los mensajes mostrados como "usuario eliminado" (`author_id` pasa a `NULL`).
- **Por qué:** respeta el derecho de una persona a borrar sus datos sin romper las conversaciones de los demás. `GET /users/export` ofrece la exportación de datos correspondiente.

## D14. Roles globales solo mediante una CLI

- **Decisión:** `users.role` (`user | moderator | admin`) nunca se puede escribir por la API HTTP. Se otorga con `cmd/admin` (`make admin-grant`).
- **Por qué:** no hay endpoint que atacar para escalar privilegios; quien pueda llegar a la base de datos decide quién es admin.

## D15. i18n sin dependencias y temas con variables CSS

- **Decisión:** diccionarios tipados (`en.ts` es la fuente de verdad de las claves y `es.ts` debe cumplir el mismo tipo) y un hook `useT()`; temas claro y oscuro mediante variables CSS y tokens semánticos de Tailwind.
- **Por qué:** la app tiene dos idiomas y unos cientos de textos, lo que no justifica un framework de i18n ligado al enrutado. Una traducción que falta es un error de compilación, y un test comprueba que los placeholders coinciden en ambos idiomas.
- **Consecuencias:** el idioma se elige en el cliente, así que el HTML renderizado en el servidor sale en inglés y puede parpadear un instante antes de hidratar. Si aparecen páginas públicas relevantes para SEO, se revisará con rutas con prefijo de idioma.

## D16. Envío optimista con estados explícitos

- **Decisión:** el mensaje aparece al instante como *enviando* y, si falla, pasa a *no enviado* con **Reintentar** y **Descartar**.
- **Por qué:** las redes inestables son normales en móvil; ocultar los fallos pierde mensajes en silencio.

## D17. Dirección visual "Foco"

- **Contexto:** la primera interfaz era un marcador neutro. Se prototiparon tres direcciones con el mismo contenido (chat de escritorio, móvil y acceso): **Foco** (densa, herramienta de equipo), **Calidez** (redondeada, burbujas, comunidad) y **Redacción** (editorial, serif, monocroma).
- **Decisión:** Foco. Tarjetas de mensaje con borde, con autor y rol, lista de salas a la izquierda con prefijos `#`, panel lateral de miembros y un único acento índigo.
- **Por qué:** el centro del producto son las salas, los roles y la moderación, que necesitan densidad de información y etiquetas de rol claras más que personalidad. Las tarjetas muestran autor, rol y hora de forma compacta, y un solo acento mantiene sin ambigüedad los contadores de no leídos y la selección en ambos temas.
- **Alternativas:** Calidez ofrece el móvil más amable pero gasta ancho en burbujas; Redacción es la más distintiva pero es más difícil de mantener coherente según crecen las funciones.
- **Consecuencias:** todo el color sale de tokens semánticos (`primary`, `soft`, `success`...), así que la dirección puede reajustarse en un solo archivo (`globals.css`).

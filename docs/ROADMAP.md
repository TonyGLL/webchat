# Roadmap WebChat

Objetivo: llevar WebChat de MVP funcional a un producto profesional, seguro, eficiente y escalable.

Documentos relacionados:
- [CASOS_DE_USO.md](CASOS_DE_USO.md): qué debe hacer el sistema (existente a mejorar + nuevo).
- [HALLAZGOS.md](HALLAZGOS.md): problemas detectados, con estado (pendiente / en progreso / completado).

Cada fase tiene un criterio de salida. No se pasa a la siguiente sin cumplirlo (salvo trabajo en paralelo explícito).

## Fase 0: Línea base (1 semana)
Dejar el repo en un estado donde se pueda cambiar con confianza.
- CI en GitHub Actions: `go vet`, `go test`, `bun run lint`, `bun run build`.
- Esqueleto de tests en backend (use cases con repos mockeados) y frontend (Vitest + Testing Library).
- README raíz y de backend actualizados (puertos, docker-compose, endpoints reales).
- `docker-compose.yml` unificado en la raíz (front, back, db, redis).
- Herramienta de migraciones (golang-migrate o goose) en lugar del `schema.sql` manual.
- Endpoint `/health` y logging estructurado (`slog`).

**Salida:** `docker compose up` levanta todo; CI verde en cada PR; migraciones versionadas.

## Fase 1: Seguridad y robustez (1-2 semanas)
- `JWT_SECRET` obligatorio (fallar al arrancar si falta).
- WebSocket: validar `Origin`, quitar el token de la URL (ticket de un solo uso de vida corta).
- Rate limiting en `/auth/*` y en envío de mensajes (Redis).
- Refresh token con rotación y revocación; sesiones por dispositivo.
- Frontend: `AuthContext` único, interceptor axios con renovación automática, tokens fuera de `localStorage` (cookies httpOnly).
- Autorización: verificar membresía de sala en cada endpoint de mensajes y reacciones.
- Validación y límites de tamaño en todos los DTOs; cabeceras de seguridad.

**Salida:** checklist OWASP básico revisado; tests de autorización por endpoint.

## Fase 2: Calidad del chat actual (2 semanas)
- Paginación por cursor en historial y scroll infinito.
- Estado de mensajes (enviado / entregado / leído) y contador de no leídos por sala.
- Reconexión automática del WebSocket con backoff y reenvío de eventos perdidos.
- Cliente optimista: el mensaje aparece al instante y se confirma después.
- Tipado estricto en frontend (eliminar `any`, tipos compartidos con la API, contrato OpenAPI).
- Accesibilidad y responsive (la vista móvil hoy no tiene navegación entre lista y chat).

**Salida:** chat estable con red inestable; sin `any` en el código del frontend.

## Fase 3: Funcionalidades de producto (3-4 semanas)
- Adjuntos e imágenes (almacenamiento S3-compatible, URLs firmadas; ya existe la tabla `attachments`).
- Indicador de "escribiendo" y presencia en línea.
- Mensajes directos 1 a 1 y gestión de miembros/roles de sala.
- Búsqueda de mensajes (Postgres full-text).
- Responder a un mensaje, borrar, fijar.
- Notificaciones (Web Push) y preferencias de silencio por sala.

**Salida:** paridad razonable con apps de chat de equipos pequeños.

## Fase 4: Moderación y administración (2 semanas)
- Reportes de mensajes y usuarios, bans por sala (tablas `reports` y `bans` ya existen).
- Rol admin/moderador con panel básico.
- Registro de auditoría de acciones sensibles.
- Eliminación y exportación de datos del usuario.

## Fase 5: Escala y operación (2-3 semanas)
- Hub de WebSocket distribuido con Redis pub/sub para correr varias instancias.
- Métricas (Prometheus), trazas (OpenTelemetry) y alertas.
- Pruebas de carga (k6) con objetivo de conexiones concurrentes definido.
- Caché de lectura en Redis para salas y perfiles; revisión de índices y planes de consulta.
- Despliegue reproducible (imágenes versionadas, entornos staging/prod, backups de Postgres).
- Imagen Docker con usuario no root y certificados CA (hoy usa `scratch`).

**Salida:** dos instancias del backend sirviendo la misma sala; dashboards y alertas activos.

## Fase 6: Pulido (continuo)
- Temas claro/oscuro, i18n (es/en), onboarding.
- Tests end-to-end (Playwright) de los flujos críticos.
- Documentación de API publicada y guía de contribución. ✅ Hecho en `docs/en/` y `docs/es/` (arquitectura, decisiones, seguridad, tiempo real, modelo de datos, API, frontend, primeros pasos, contribuir). Pendiente: contrato OpenAPI formal.

## Métricas de éxito
- Cobertura de tests en use cases del backend ≥ 70 %.
- Latencia de entrega de mensaje p95 < 200 ms.
- 0 hallazgos críticos o altos abiertos en `HALLAZGOS.md`.
- Despliegue a producción con un solo comando y rollback documentado.

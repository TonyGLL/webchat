# Casos de uso

Actores: **Visitante** (sin sesión), **Usuario** (autenticado), **Miembro** (usuario dentro de una sala), **Admin de sala**, **Moderador/Admin global**, **Sistema**.

Estado: ✅ existe y funciona · 🔧 existe, mejorar · 🆕 nuevo.
Fase: ver [ROADMAP.md](ROADMAP.md).

## A. Casos de uso existentes a mejorar

### CU-01 Registro de usuario 🔧 (Fase 1)
- **Actor:** Visitante.
- **Flujo:** envía usuario, email y contraseña → el sistema crea la cuenta y manda un correo de verificación.
- **Mejoras:** política de contraseña más fuerte (hoy mínimo 6), rate limiting, aviso claro si el email ya existe sin filtrar si la cuenta existe, reenvío de verificación con límite.

### CU-02 Verificación de email 🔧 (Fase 1)
- **Actor:** Visitante.
- **Flujo:** abre el enlace con token → la cuenta queda verificada.
- **Mejoras:** token de un solo uso con expiración corta, pantalla de reenvío, bloquear login o acciones hasta verificar.

### CU-03 Inicio de sesión 🔧 (Fase 1)
- **Flujo:** credenciales → access token + refresh token.
- **Mejoras:** rate limiting y bloqueo temporal ante intentos fallidos, tokens en cookie httpOnly, lista de sesiones activas, "cerrar sesión en todos los dispositivos".
- **Nota:** el formulario valida el campo `user` como email; revisar si debe aceptar también el nombre de usuario.

### CU-04 Renovación de sesión 🔧 (Fase 1)
- **Flujo actual:** existe `POST /auth/refresh-token`, pero el frontend no lo usa y al expirar cierra la sesión.
- **Mejoras:** interceptor que renueve en silencio y reintente la petición; rotación del refresh token y detección de reutilización.

### CU-05 Ver y gestionar perfil 🔧 (Fase 2)
- **Flujo actual:** consultar perfil y desactivar cuenta.
- **Mejoras:** editar nombre, avatar y teléfono; cambiar contraseña; eliminar cuenta definitivamente y exportar datos (Fase 4).

### CU-06 Crear sala 🔧 (Fase 2)
- **Mejoras:** salas públicas o privadas, descripción e imagen, el creador queda como admin.

### CU-07 Unirse a una sala / aceptar invitación 🔧 (Fase 1 y 3)
- **Flujo actual:** unirse por id o aceptar invitación por código.
- **Mejoras:** unirse por id solo si la sala es pública; invitaciones con expiración y límite de usos visibles; revocar invitaciones.

### CU-08 Listar mis salas 🔧 (Fase 2)
- **Mejoras:** último mensaje, hora, contador de no leídos, orden por actividad, búsqueda.

### CU-09 Enviar mensaje 🔧 (Fase 2)
- **Flujo actual:** `POST /messages` y difusión por WebSocket a la sala.
- **Mejoras:** envío optimista con estados (enviando / enviado / error / reintentar), validación de membresía y de longitud, rate limiting, idempotencia (id de cliente) para evitar duplicados al reintentar.

### CU-10 Ver historial de mensajes 🔧 (Fase 2)
- **Mejoras:** paginación por cursor, scroll infinito hacia arriba, mantener la posición al cargar más, salto al primer no leído.

### CU-11 Editar mensaje 🔧 (Fase 2)
- **Mejoras:** solo el autor, marca "editado", ventana de tiempo configurable, evento WebSocket a la sala.

### CU-12 Reaccionar a un mensaje 🔧 (Fase 2)
- **Mejoras:** una reacción por emoji y usuario, agrupar por emoji con contador y lista de quién reaccionó, actualización en tiempo real sin recargar.

### CU-13 Recibir mensajes en tiempo real 🔧 (Fase 2 y 5)
- **Mejoras:** reconexión automática con backoff, recuperar eventos perdidos desde el último id recibido, validación de origen, soporte multi-instancia con Redis pub/sub.

## B. Casos de uso nuevos

### CU-20 Adjuntar archivos e imágenes 🆕 (Fase 3)
- **Actor:** Miembro.
- **Flujo:** elige archivo → el backend entrega una URL de subida firmada → el cliente sube → el mensaje referencia el adjunto.
- **Reglas:** tipos y tamaño máximos, escaneo/validación de tipo real, miniaturas, URLs de descarga firmadas con expiración.

### CU-21 Indicador de "escribiendo" 🆕 (Fase 3)
- Evento efímero por WebSocket con límite de frecuencia; no se persiste.

### CU-22 Presencia en línea 🆕 (Fase 3)
- Estado en Redis con TTL renovado por ping; visible en lista de miembros y mensajes directos; respetar ajuste de privacidad.

### CU-23 Mensajes directos 🆕 (Fase 3)
- **Flujo:** abrir conversación 1 a 1 con otro usuario; se modela como sala privada de dos miembros, reutilizando la lógica existente.

### CU-24 Gestionar miembros y roles de sala 🆕 (Fase 3)
- Admin de sala puede expulsar, cambiar roles y transferir propiedad; el miembro puede salir de la sala.

### CU-25 Responder, borrar y fijar mensajes 🆕 (Fase 3)
- Responder citando un mensaje; borrado lógico (autor o moderador) con marca "mensaje eliminado"; fijar mensajes importantes (admin).

### CU-26 Marcar como leído y contador de no leídos 🆕 (Fase 2)
- Al abrir la sala o hacer scroll hasta el final se actualiza el último mensaje leído del miembro; el contador se calcula por sala.

### CU-27 Buscar mensajes 🆕 (Fase 3)
- Búsqueda de texto dentro de una sala o global entre las salas del usuario, solo sobre salas a las que pertenece.

### CU-28 Notificaciones 🆕 (Fase 3)
- Web Push y correo opcional para menciones y mensajes directos; silenciar sala; preferencias por usuario.

### CU-29 Reportar mensaje o usuario 🆕 (Fase 4)
- Miembro reporta con motivo → queda en cola de moderación (tabla `reports`).

### CU-30 Moderar contenido 🆕 (Fase 4)
- Moderador revisa reportes, borra mensajes, banea usuarios de una sala o globalmente (tabla `bans`); todo queda en auditoría.

### CU-31 Gestionar sesiones y dispositivos 🆕 (Fase 1)
- Usuario ve sus sesiones activas y puede revocarlas.

### CU-32 Recuperar contraseña 🆕 (Fase 1)
- Solicitud por email → enlace de un solo uso con expiración → nueva contraseña; invalida las sesiones existentes.

### CU-33 Exportar y eliminar mis datos 🆕 (Fase 4)
- El usuario descarga sus datos y puede eliminar la cuenta; los mensajes se anonimizan según política.

### CU-34 Observabilidad y operación 🆕 (Fase 0 y 5)
- **Actor:** Sistema / equipo.
- Health check, métricas, trazas, alertas y backups verificables.

[English](../en/security.md) · **🌐 Español**

# Seguridad

La seguridad va por capas: ningún control se considera suficiente por sí solo.

```mermaid
flowchart LR
    R[Petición] --> O[Origin / CORS]
    O --> RL[Límite de peticiones<br/>por IP · Redis]
    RL --> J[Validación del JWT]
    J --> P[Matriz de permisos<br/>por rangos]
    P --> V[Validación de entrada]
    V --> DB[(Datos)]
    P -. cada acción de moderación .-> AU[(audit_log)]
```

## 1. Autenticación

| Elemento | Comportamiento |
|---|---|
| Contraseñas | Hashes bcrypt en su propia tabla (`passwords`), nunca devueltos por la API |
| Política de contraseñas | 8–72 caracteres (72 es el límite de bcrypt); rechaza una lista corta de contraseñas muy comunes y las que contienen el usuario o la parte local del email |
| Access token | JWT de 1 hora, enviado como `Authorization: Bearer …` |
| Refresh token | JWT de 7 días, **un solo uso** y rotado dentro de una *familia* (una por inicio de sesión). Una repetición dentro de la gracia de 30 s recibe el mismo sucesor (no nace una segunda cadena); un token ya rotado presentado pasada la gracia revoca toda la familia; la sesión termina como máximo 30 días después del inicio sea cual sea su actividad; su id se guarda en Redis (los tokens emitidos antes de las familias abren la suya en su primera renovación); cookie httpOnly, `SameSite=Strict`, limitada a `/api/v1/auth` |
| Persistencia de sesión | cada renovación emite otros 7 días, así la sesión termina tras una semana **sin uso**; el navegador cierra sesión solo si el servidor rechaza el token, nunca por un error de red (D20) |
| Tipos de token y sesiones vivas | cada JWT lleva `typ` (`access`/`refresh`/`verify`) y los consumidores exigen el suyo; solo HS256. Un guardián de sesión rechaza en cada petición los tokens emitidos antes de una revocación y los de usuarios desactivados (D23). Desactivar exige la contraseña y termina sesiones y sockets al instante |
| Identidad | los nombres de usuario son `[A-Za-z0-9_.-]` (3–30); email y usuario son únicos sin distinguir mayúsculas; login y restablecimiento tardan lo mismo con cuentas desconocidas; bcrypt con coste 12 |
| Límites | cuerpos ≤ 1 MiB, timeouts del servidor, límites de tasa por IP en cada ruta que modifica estado, `TRUSTED_PROXIES` tras un balanceador (D27). Editar mensajes exige ser miembro y limita el contenido a 4000 caracteres; las reacciones tienen tope (10 distintas por persona, 30 por mensaje) |
| Endurecimiento de la API | toda respuesta lleva `X-Content-Type-Options: nosniff` y `/api/...` responde `Cache-Control: no-store` (los cuerpos con tokens y el JSON autenticado nunca se cachean); los 500 y los cuerpos inválidos devuelven mensajes genéricos (el detalle solo va al log); el JSON de más de 1 KiB se comprime con gzip; el preflight de CORS se cachea 2 h; `/docs` carga su script de Scalar fijado con un hash de Subresource Integrity; las llamadas a Redis en la ruta de petición vencen en 0,5 s (fail open) |
| Valores por defecto de producción | el flag `Secure` de la cookie de refresco y el modo release de Gin no dependen de una variable opcional: solo un `APP_ENV` explícito `dev`/`development`/`local`/`test` los relaja, y el servidor avisa en el log al arrancar salvo que `APP_ENV` sea `prod`/`production` |
| Navegador | cabeceras de seguridad y CSP en la aplicación web (D25); los errores del cliente se registran sin cuerpos de petición ni credenciales |
| Restablecer contraseña | `POST /auth/forgot-password` responde siempre igual, así que nunca revela quién tiene cuenta. El enlace del correo lleva un token aleatorio de 256 bits válido 30 minutos y de **un solo uso** (se canjea de forma atómica con `GETDEL`; un enlace nuevo sustituye al anterior); solo se guarda su hash SHA-256. El correo nunca incluye el nombre de la cuenta. Como máximo un correo por dirección y minuto, además de un límite por cliente. `POST /auth/reset-password` aplica la política de contraseñas del registro (una contraseña rechazada deja el enlace válido) y después **cierra todas las sesiones**: los refresh tokens emitidos antes se rechazan (los access tokens caducan en menos de una hora) (D21) |
| Cierre de sesión | `POST /auth/logout` revoca el id de refresh, borra la cookie y corta los sockets abiertos del usuario (las otras pestañas se reconectan solas) |
| Secreto | el servidor se niega a arrancar sin un `JWT_SECRET` de al menos 32 caracteres |
| Errores de login | el formulario muestra un único mensaje genérico, así no revela si un email existe |

El razonamiento de cada elección está en [Decisiones de diseño](decisions.md) (D5–D7, D20, D21).

## 2. WebSocket

- El JWT nunca aparece en una URL: `POST /ws-ticket` devuelve un ticket de 30 segundos y un solo uso guardado en Redis.
- La cabecera `Origin` debe coincidir con `CORS_ALLOWED_ORIGINS`.
- Un cliente solo puede anunciar `typing` en una sala a la que está suscrito, y no con más frecuencia que un intervalo fijo.
- Al expulsar o vetar a un miembro, su conexión se desuscribe de la sala (o se desconecta al borrar la cuenta).

## 3. Autorización: la matriz de permisos

Todo vive en `internal/shared/permissions`. Los roles de sala y los globales se reducen a un único **rango**:

```mermaid
flowchart LR
    N["0 · ninguno"] --> M["1 · miembro"] --> MO["2 · moderador"] --> AD["3 · admin"] --> OW["4 · propietario"] --> SU["5 · admin global"]
```

Un **moderador** global actúa como moderador en todas las salas; un **admin** global supera a todos los propietarios de sala.

| Acción | Rango mínimo |
|---|---|
| Leer una sala, enviar mensajes | 1 · miembro |
| Borrar mensajes ajenos, resolver denuncias | 2 · moderador |
| Expulsar o vetar a un usuario | 2 · moderador **y estrictamente por encima del objetivo** |
| Fijar mensajes, crear invitaciones (códigos o personales) | 3 · admin |
| Cambiar roles | 3 · admin, solo sobre miembros por debajo: un admin puede crear moderadores, solo los propietarios (o admins globales) pueden crear admins; el rol de propietario nunca se asigna |
| Leer el registro de auditoría | 3 · admin |

Como las reglas comparan rangos, un moderador nunca puede actuar sobre otro moderador, un admin o el propietario.

## 4. Protección contra abuso

Los límites son por IP de cliente y se guardan en Redis (`/api/v1`). Los clientes IPv6 se agrupan por su prefijo `/64`, así rotar direcciones dentro de una misma asignación no da un cupo nuevo. El contador es un pequeño script Lua (funciona en Redis 6.2+, no solo 7+):

| Ruta | Límite |
|---|---|
| `POST /auth/login` | 10 por minuto |
| `POST /auth/register` | 20 cada 10 minutos |
| `POST /auth/refresh-token` | 30 por minuto |
| `POST /auth/send-verify-email/:user_id` | 5 cada 10 minutos |
| `POST /ws-ticket` | 30 por minuto |
| `POST /messages` | 120 por minuto |
| `POST /rooms/read/:room_id` | 240 por minuto |
| `GET /messages/search` | 60 por minuto |
| `GET /users/export`, `DELETE /users/me` | 5 por hora cada una |

Al superarlos la API responde `429` con la cabecera `Retry-After`. Si Redis no está disponible, el limitador deja pasar el tráfico y registra el fallo. Detrás de un proxy inverso hay que configurar los proxies de confianza para que la IP del cliente no se pueda falsificar.

## 5. Moderación y auditoría

- Los miembros denuncian mensajes (una denuncia por persona, nunca del propio). Los moderadores revisan la cola y pueden borrar el mensaje, marcar la denuncia como resuelta o ignorarla.
- Los vetos son por sala, temporales o permanentes, e impiden volver a entrar por id o por invitación.
- Toda acción sensible se escribe en `audit_log` con quien actúa, el objetivo, la sala y los detalles. Solo los admins lo leen.

## 6. Privacidad

- **Directorio y descubrimiento:** cualquier usuario con sesión puede listar a otros usuarios activos (nombre, usuario, estado de conexión) y el nombre, tema y número de miembros de cada sala. Leer o escribir en una sala sigue exigiendo ser miembro, así que una sala privada es visible pero está cerrada. Las invitaciones personales solo las puede responder el usuario invitado (cualquier otro recibe "no encontrada") y al aceptar se vuelven a comprobar los vetos.
- `GET /users/export` descarga un JSON con el perfil, las salas, los mensajes y las reacciones del usuario.
- `DELETE /users/me` (confirmado con la contraseña) borra la cuenta: ver D13 en [Decisiones de diseño](decisions.md).

## 7. Límites conocidos

El directorio expone nombres y usuarios a cualquier persona con sesión: adecuado para un servidor de equipo, pero un despliegue público necesitaría un ajuste de visibilidad (ver D18 en [Decisiones de diseño](decisions.md)).

La verificación de email está implementada pero **pausada**, y `POST /auth/send-verify-email/:user_id` sigue siendo público (registrado como H-008 en el registro de hallazgos). Docker Compose todavía lleva contraseñas de base de datos en el archivo (H-051).

## 8. Cómo reportar una vulnerabilidad

Por favor no abras una incidencia pública. Contacta de forma privada con los mantenedores e incluye los pasos para reproducirla.

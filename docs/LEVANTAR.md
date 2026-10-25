# Cómo levantar WebChat

Guía completa para correr el proyecto en tu máquina sin ayuda externa. Si algo falla, ve a [Problemas comunes](#problemas-comunes) y a [Limitaciones conocidas](#limitaciones-conocidas-hoy).

## Resumen rápido

```
1. Docker:      cd backend && docker compose up -d db redis
2. Tablas:      (automático al arrancar el backend)
3. Backend:     crear backend/.env.local  →  make watch      (http://localhost:8081)
4. Frontend:    crear frontend/.env.local →  pnpm dev        (http://localhost:3100)
```

## Puertos (en tu máquina)

Elegidos para no chocar con otros proyectos locales. Todos se publican solo en `127.0.0.1`.

| Servicio | Puerto | Contenedor |
|---|---|---|
| Frontend (Next.js) | 3100 | corre fuera de Docker |
| Backend (Go) | 8081 | `webchat-api` (solo en modo Docker) |
| PostgreSQL | 5433 | `webchat-db` |
| Redis | 6380 | `webchat-redis` |

Convención de nombres: `<proyecto>-<servicio>` en minúsculas con guiones; volúmenes `webchat-<servicio>-data`.

## Requisitos

- Go 1.27.3 o superior
- Node 24+ y pnpm (la versión está fijada en `package.json`; `corepack enable` o `npm i -g pnpm`) para el frontend
- Docker y Docker Compose
- Make
- `air` (recarga automática del backend). `make watch` ofrece instalarlo; también: `go install github.com/air-verse/air@latest` (asegúrate de tener `~/go/bin` en el `PATH`).
- `openssl` (para generar secretos)

## Paso 1: Base de datos y Redis

```bash
cd backend
docker compose up -d db redis
docker compose ps
```

Debes ver `webchat-db` y `webchat-redis` en estado `running`.

Las contraseñas de Postgres y Redis están definidas en `backend/docker-compose.yml` (variables `POSTGRES_PASSWORD` y `REDIS_PASSWORD`). Las necesitarás en el paso 3. Ábrelo tú y cópialas; esta guía no las repite.

## Paso 2: Tablas de la base de datos (automático)

Ya no hay que aplicar el esquema a mano: el backend ejecuta las **migraciones** (`backend/migrations/*.sql`, versionadas con golang-migrate) al arrancar. En una base vacía crea todas las tablas; en una base que ya tenía las tablas de antes, la adopta como versión 1 sin tocar los datos.

Comprobar (después de arrancar el backend en el paso 3):

```bash
docker exec -it webchat-db psql -U root -d webchat -c '\dt'
docker exec -it webchat-db psql -U root -d webchat -c 'select * from schema_migrations'
```

Deben aparecer `users`, `passwords`, `rooms`, `room_members`, `messages`, `reactions`, etc., y la versión actual.

**Cambiar el esquema:** añade un par de archivos numerados en `backend/migrations/`, por ejemplo `000002_add_xxx.up.sql` y `000002_add_xxx.down.sql`. Nunca edites una migración ya aplicada.

Para empezar desde cero (borra **todos** los datos):

```bash
cd backend
docker compose down -v
docker compose up -d db redis
# y volver a arrancar el backend: recrea las tablas
```

## Paso 3: Configurar y levantar el backend

### 3.1 Crear `backend/.env.local`

Convención de archivos de entorno (todos, salvo `.env.example`, están en `.gitignore` y no se suben al repositorio):

| Archivo | Para qué | Quién lo carga |
|---|---|---|
| `backend/.env.example` | Plantilla con todas las variables y comentarios. **Se versiona.** | Nadie (solo se copia) |
| `backend/.env.local` | Backend corriendo en tu máquina | `make run`, `make watch` |
| `backend/.env.dev` | Backend dentro de Docker | `make up-dev`, `make database-up` |
| `backend/.env.prod` | Producción | `make up-prod` |
| `frontend/.env.example` | Plantilla del frontend. **Se versiona.** | Nadie (solo se copia) |
| `frontend/.env.local` | Frontend en desarrollo | `pnpm dev` (Next.js lo carga solo) |

Crea el tuyo copiando la plantilla y reemplazando todo lo que diga `<...>`:

```bash
cd backend
cp .env.example .env.local
```

La plantilla `backend/.env.example` es la fuente de verdad de las variables: ahí están todas, con su explicación. Si añades una variable nueva al código, agrégala también a la plantilla.

Notas sobre cada variable:

| Variable | Obligatoria | Detalle |
|---|---|---|
| `JWT_SECRET` | **Sí** | Mínimo 32 caracteres aleatorios. Si falta, es corta, es un texto de ejemplo (`change-me`...), un patrón repetido o usa menos de 8 caracteres distintos, el servidor no arranca (H-001). Genera uno con `openssl rand -base64 48`. |
| `DATABASE_URL` | Alternativa | Si la defines, se ignoran las `DB_*`. Formato: `postgres://usuario:clave@localhost:5433/webchat?sslmode=disable` |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME` | Sí (si no hay `DATABASE_URL`) | `DB_PASSWORD` según la contraseña del contenedor. |
| `REDIS_URL` | Alternativa | Si la defines, se ignoran las `REDIS_*`. Formato: `redis://:clave@localhost:6380/0` |
| `REDIS_HOST`, `REDIS_PORT` | Sí (si no hay `REDIS_URL`) | `REDIS_USER`, `REDIS_PASS`, `REDIS_DB` opcionales. |
| `CORS_ALLOWED_ORIGINS` | Sí en la práctica | Lista separada por comas. Debe incluir el origen del frontend. |
| `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME` | No | Pool de Postgres: 25, 25, `30m`, `5m` por defecto. |
| `TRUSTED_PROXIES` | Tras un balanceador | IPs/CIDRs cuyo `X-Forwarded-For` se acepta; vacío por defecto. |
| `PORT` | No | Por defecto `8081`. |
| `FRONTEND_URL` | No | URL pública del frontend para los enlaces de los correos. Por defecto `http://localhost:3100`. |
| `SMTP_*` | Solo para restablecer contraseña | `SMTP_HOST`, `SMTP_PORT`, `SMTP_FROM`, `SMTP_PASSWORD`. Sin ellos «¿Olvidaste tu contraseña?» no envía nada (el error solo queda en el log). Con Gmail necesitas una "contraseña de aplicación". Para probar en local sirve un servidor SMTP de pruebas como Mailpit. |
| `APP_ENV` | Recomendada | Seguro por defecto: solo `dev`, `development`, `local` o `test` desactivan el modo producción (cookie Secure, modo release). Pon `APP_ENV=dev` en local con http. Redis mínimo: 6.2. |

### 3.2 Arrancar

```bash
cd backend
make watch      # con recarga automática; usa .env.local
# o bien
make run        # sin recarga; usa .env.local
```

Deberías ver `Server running on port: 8081`.

Pruebas automáticas del backend:

```bash
make test
```

### Alternativa: backend dentro de Docker

Usa `.env.dev` (o `.env.prod`) en lugar de `.env.local`. **Dentro de Docker los hosts y puertos son los internos**, no los de tu máquina:

```env
DB_HOST=db
DB_PORT=5432
REDIS_HOST=redis
REDIS_PORT=6379
```

El resto de variables igual que arriba (con `PORT=8081`; el compose también lo fija). Después:

```bash
cd backend
make up-dev     # levanta app + db + redis
make down       # detiene todo
```

## Paso 4: Configurar y levantar el frontend

### 4.1 Crear `frontend/.env.local`

```bash
cd frontend
cp .env.example .env.local
```

Contiene una sola variable: `NEXT_PUBLIC_API_URL=http://localhost:8081/api/v1`.

(`.env*.local` está en `.gitignore`.) Si no lo creas, el WebSocket usa ese mismo valor por defecto, pero las llamadas REST no lo tienen y fallarán.

### 4.2 Arrancar

```bash
cd frontend
pnpm install
pnpm dev         # http://localhost:3100
```

Otros comandos: `pnpm run build` (compilar), `pnpm run lint`.

## Paso 5: Comprobar que todo funciona

### 5.1 Crear un usuario

Desde el navegador: `http://localhost:3100/register`. El formulario pide nombre, apellido, usuario (3-30 caracteres), email, teléfono y contraseña (mínimo 8). Al terminar te lleva al login.

Por la API (alternativa):

```bash
curl -s -X POST http://localhost:8081/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Tony","last_name":"Gomez","username":"tony","email":"tu@correo.com","phone":"5551234567","password":"clave12345"}'
```

### 5.2 Verificación de email (pausada)

Por ahora **no se verifica el email**: el login no lo exige y el registro ya no envía correo, así que para eso no necesitas `SMTP_*` (sí hacen falta para «¿Olvidaste tu contraseña?», ver arriba). El código de verificación (rutas `/auth/send-verify-email` y `/auth/verify-email`, páginas `/verify-email` y `/verify/[token]`) sigue en el repositorio, sin usarse, para retomarlo más adelante.

### 5.3 Iniciar sesión

```bash
curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"user":"tu@correo.com","password":"clave12345"}'
```

Devuelve `token` (access, 1 h) en el cuerpo y pone el refresh token en una cookie `httpOnly` (`Set-Cookie: refresh_token=...`), que JavaScript no puede leer. Con `curl -i -c jar.txt` ves la cabecera y guardas la cookie; `curl -b jar.txt -X POST .../auth/refresh-token` renueva y `.../auth/logout` la revoca.

### 5.4 Recorrido en el navegador

Abre `http://localhost:3100` e inicia sesión. Después:

1. Crear una sala.
2. Enviar un mensaje.
3. En una ventana de incógnito, crear un segundo usuario (pasos 5.1 y 5.2) y, desde el primero, invitarlo a la sala con el botón de invitar de la cabecera; el segundo acepta desde el bloque «Invitaciones» de la barra lateral. También se pueden escribir directamente desde el panel «Personas».
4. Comprobar que los mensajes llegan a ambas ventanas **sin recargar** (esto prueba el WebSocket).
5. Editar un mensaje y poner una reacción.

## Roles y administración

Dentro de cada sala los roles son dueño > admin > moderador > miembro (se cambian desde la lista de miembros de la sala). Los roles **globales** (moderador/admin de toda la aplicación) solo se otorgan desde el servidor:

```bash
cd backend
make admin-grant EMAIL=tu@correo.com ROLE=admin     # admin | moderator
make admin-list
make admin-revoke EMAIL=tu@correo.com
```

Detalle de la matriz de permisos en `backend/README.md` (sección Roles y permisos).

## Apagar todo

```bash
cd backend
docker compose down         # conserva los datos (volúmenes)
docker compose down -v      # borra también los datos
```

Para el backend y el frontend locales: `Ctrl+C` en cada terminal.

## Problemas comunes

| Síntoma | Causa probable | Solución |
|---|---|---|
| `JWT_SECRET must be set...` / `looks like a placeholder` / `not random enough` | Falta el secreto, es corto o adivinable | Ponlo en `.env.local` (`openssl rand -base64 48`). |
| `DATABASE_URL is not set and individual DB_...` | `.env.local` no existe o no se cargó | Ejecuta `make watch` desde `backend/` y revisa el nombre del archivo. |
| `unable to connect to database` | Contenedor apagado, puerto o contraseña errónea | `docker compose ps`; usa el puerto **5433** en local. |
| Error de Redis al arrancar | Puerto o contraseña erróneos | Usa el puerto **6380** en local y `REDIS_PASS` igual al del compose. |
| `bind: address already in use` | Otro proyecto usa el puerto | `ss -ltnp | grep <puerto>`; cambia el puerto del servicio (y actualiza esta guía). |
| El frontend carga pero no hay datos o error CORS | `CORS_ALLOWED_ORIGINS` no incluye `http://localhost:3100` | Corrígelo en `.env.local` y reinicia el backend. |
| Las peticiones del frontend van a `undefined/...` | Falta `frontend/.env.local` | Créalo y reinicia `pnpm dev`. |
| `air: command not found` | Versiones anteriores del Makefile; ahora `make watch` busca `air` en la carpeta de binarios de Go | Actualiza el repo; si aun así falla: `go install github.com/air-verse/air@latest` |

## Limitaciones conocidas hoy

Ver [HALLAZGOS.md](HALLAZGOS.md) (sección Pendientes). Cuando se resuelvan, actualiza esta guía.

## Dónde está cada cosa

| Qué | Dónde |
|---|---|
| Orquestación Docker (db, redis, api) | `backend/docker-compose.yml` |
| Comandos del backend | `backend/Makefile` |
| Migraciones SQL | `backend/migrations/` |
| Lectura de variables de entorno | `backend/internal/shared/config/config.go` |
| Arquitectura y convenciones | `CLAUDE.md` |
| Plan de trabajo | `docs/ROADMAP.md`, `docs/CASOS_DE_USO.md`, `docs/HALLAZGOS.md` |

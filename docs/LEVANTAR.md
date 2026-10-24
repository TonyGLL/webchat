# Cómo levantar WebChat

Guía completa para correr el proyecto en tu máquina sin ayuda externa. Si algo falla, ve a [Problemas comunes](#problemas-comunes) y a [Limitaciones conocidas](#limitaciones-conocidas-hoy).

## Resumen rápido

```
1. Docker:      cd backend && docker compose up -d db redis
2. Tablas:      docker exec -i webchat-db psql -U root -d webchat < migrations/schema.sql
3. Backend:     crear backend/.env.local  →  make watch      (http://localhost:8081)
4. Frontend:    crear frontend/.env.local →  bun dev        (http://localhost:3100)
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

- Go 1.24.3 o superior
- Bun (para el frontend) y Node 20+
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

## Paso 2: Crear las tablas

No hay herramienta de migraciones todavía (hallazgo H-030), así que el esquema se aplica a mano, **una sola vez** por base de datos nueva:

```bash
cd backend
docker exec -i webchat-db psql -U root -d webchat < migrations/schema.sql
```

Comprobar:

```bash
docker exec -it webchat-db psql -U root -d webchat -c '\dt'
```

Deben aparecer `users`, `passwords`, `rooms`, `room_members`, `messages`, `reactions`, etc. Si dice que las tablas ya existen, ya estaba hecho.

Para empezar desde cero (borra **todos** los datos):

```bash
cd backend
docker compose down -v
docker compose up -d db redis
# y repetir este paso 2
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
| `frontend/.env.local` | Frontend en desarrollo | `bun dev` (Next.js lo carga solo) |

Crea el tuyo copiando la plantilla y reemplazando todo lo que diga `<...>`:

```bash
cd backend
cp .env.example .env.local
```

La plantilla `backend/.env.example` es la fuente de verdad de las variables: ahí están todas, con su explicación. Si añades una variable nueva al código, agrégala también a la plantilla.

Notas sobre cada variable:

| Variable | Obligatoria | Detalle |
|---|---|---|
| `JWT_SECRET` | **Sí** | Mínimo 32 caracteres. Si falta o es corta, el servidor no arranca (H-001). Genera uno con `openssl rand -hex 32`. |
| `DATABASE_URL` | Alternativa | Si la defines, se ignoran las `DB_*`. Formato: `postgres://usuario:clave@localhost:5433/webchat?sslmode=disable` |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME` | Sí (si no hay `DATABASE_URL`) | `DB_PASSWORD` según la contraseña del contenedor. |
| `REDIS_URL` | Alternativa | Si la defines, se ignoran las `REDIS_*`. Formato: `redis://:clave@localhost:6380/0` |
| `REDIS_HOST`, `REDIS_PORT` | Sí (si no hay `REDIS_URL`) | `REDIS_USER`, `REDIS_PASS`, `REDIS_DB` opcionales. |
| `CORS_ALLOWED_ORIGINS` | Sí en la práctica | Lista separada por comas. Debe incluir el origen del frontend. |
| `PORT` | No | Por defecto `8081`. |
| `FRONTEND_URL` | No | URL pública del frontend para los enlaces de los correos. Por defecto `http://localhost:3100`. |
| `SMTP_*` | Solo para verificar emails | Con Gmail necesitas una "contraseña de aplicación". |
| `APP_ENV` | No | `prod` o `production` activa el modo producción. |

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
bun install
bun dev         # http://localhost:3100
```

Otros comandos: `bun run build` (compilar), `bun run lint`.

## Paso 5: Comprobar que todo funciona

### 5.1 Crear un usuario

Desde el navegador: `http://localhost:3100/register`. El formulario pide nombre, apellido, usuario (3-30 caracteres), email, teléfono y contraseña (mínimo 8). Al registrarte, el backend te envía un correo con el enlace de verificación (`FRONTEND_URL/verify/<token>`, válido 10 minutos).

Por la API (alternativa):

```bash
curl -s -X POST http://localhost:8081/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Tony","last_name":"Gomez","username":"tony","email":"tu@correo.com","phone":"5551234567","password":"clave12345"}'
# devuelve {"id": <n>, "email": "..."}; luego pide el correo de verificación:
curl -s -X POST http://localhost:8081/api/v1/auth/send-verify-email/<id> \
  -H 'Content-Type: application/json' -d '{"email":"tu@correo.com"}'
```

### 5.2 Verificar el email

Abre el enlace del correo. Si **no tienes SMTP configurado** (o no llegó), verifica a mano:

```bash
docker exec -it webchat-db psql -U root -d webchat \
  -c "UPDATE users SET email_verified_at = NOW() WHERE email = 'tu@correo.com';"
```

### 5.3 Iniciar sesión

```bash
curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"user":"tu@correo.com","password":"clave12345"}'
```

Debe devolver `token` y `refreshToken`.

### 5.4 Recorrido en el navegador

Abre `http://localhost:3100` e inicia sesión. Después:

1. Crear una sala.
2. Enviar un mensaje.
3. En una ventana de incógnito, crear un segundo usuario (pasos 5.1 y 5.2) y unirlo a la sala con una invitación.
4. Comprobar que los mensajes llegan a ambas ventanas **sin recargar** (esto prueba el WebSocket).
5. Editar un mensaje y poner una reacción.

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
| `JWT_SECRET must be set and at least 32 characters long` | Falta el secreto o es corto | Ponlo en `.env.local` (`openssl rand -hex 32`). |
| `DATABASE_URL is not set and individual DB_...` | `.env.local` no existe o no se cargó | Ejecuta `make watch` desde `backend/` y revisa el nombre del archivo. |
| `unable to connect to database` | Contenedor apagado, puerto o contraseña errónea | `docker compose ps`; usa el puerto **5433** en local. |
| Error de Redis al arrancar | Puerto o contraseña erróneos | Usa el puerto **6380** en local y `REDIS_PASS` igual al del compose. |
| `bind: address already in use` | Otro proyecto usa el puerto | `ss -ltnp | grep <puerto>`; cambia el puerto del servicio (y actualiza esta guía). |
| El frontend carga pero no hay datos o error CORS | `CORS_ALLOWED_ORIGINS` no incluye `http://localhost:3100` | Corrígelo en `.env.local` y reinicia el backend. |
| Las peticiones del frontend van a `undefined/...` | Falta `frontend/.env.local` | Créalo y reinicia `bun dev`. |
| `air: command not found` | `~/go/bin` no está en el `PATH` | `export PATH=$PATH:$(go env GOPATH)/bin` |
| No llega el correo de verificación | SMTP sin configurar o `FRONTEND_URL` incorrecta | Revisa las `SMTP_*`; usa el atajo SQL del paso 5.2. |

## Limitaciones conocidas hoy

Registradas en [HALLAZGOS.md](HALLAZGOS.md); cuando se resuelvan, actualiza esta guía.

- **H-050:** el frontend llama a `/auth/refresh`, pero el backend expone `/auth/refresh-token`.
- **H-054:** `bun run lint` no funciona en el frontend (usa `next lint`, que ya no existe en Next 16).
- **H-030:** el esquema de la base de datos se aplica a mano.

## Dónde está cada cosa

| Qué | Dónde |
|---|---|
| Orquestación Docker (db, redis, api) | `backend/docker-compose.yml` |
| Comandos del backend | `backend/Makefile` |
| Esquema SQL | `backend/migrations/schema.sql` |
| Lectura de variables de entorno | `backend/internal/shared/config/config.go` |
| Arquitectura y convenciones | `CLAUDE.md` |
| Plan de trabajo | `docs/ROADMAP.md`, `docs/CASOS_DE_USO.md`, `docs/HALLAZGOS.md` |

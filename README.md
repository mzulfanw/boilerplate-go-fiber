# Boilerplate Go Fiber

API boilerplate with Go + Fiber, production-ready setup: env-based config, JSON logging (logrus), security middleware, and graceful shutdown.

## Requirements

- Go 1.25+

## Quick Start

1) Copy environment file:
```
cp .env.example .env
```

2) Run the app:
```
make run
```

Health check:
```
curl http://localhost:3000/health
```

## Environment Variables

Core:
- `APP_NAME` (default: `boilerplate-go-fiber`)
- `APP_ENV` (default: `local`)
- `LOG_LEVEL` (default: `info`)
- `HTTP_ADDR` (default: `:3000`)
- `HTTP_READ_TIMEOUT` (default: `10s`)
- `HTTP_WRITE_TIMEOUT` (default: `10s`)
- `HTTP_IDLE_TIMEOUT` (default: `10s`)
- `SHUTDOWN_TIMEOUT` (default: `5s`)

CORS:
- `CORS_ALLOW_ORIGINS` (default: `*`)
- `CORS_ALLOW_METHODS` (default: `GET,POST,HEAD,PUT,DELETE,PATCH`)
- `CORS_ALLOW_HEADERS` (default: empty)
- `CORS_EXPOSE_HEADERS` (default: empty)
- `CORS_ALLOW_CREDENTIALS` (default: `false`)
- `CORS_MAX_AGE` (default: `0`)

Redis:
- `REDIS_ADDR` (default: `localhost:6379`)
- `REDIS_USERNAME` (default: empty)
- `REDIS_PASSWORD` (default: empty)
- `REDIS_DB` (default: `0`)
- `REDIS_DIAL_TIMEOUT` (default: `5s`)
- `REDIS_READ_TIMEOUT` (default: `3s`)
- `REDIS_WRITE_TIMEOUT` (default: `3s`)
- `REDIS_POOL_SIZE` (default: `0` = go-redis default)
- `REDIS_MIN_IDLE_CONNS` (default: `0`)
- `REDIS_DEFAULT_TTL` (default: `0s` = no expiration)
- `REDIS_TLS_ENABLED` (default: `false`)
- `REDIS_TLS_INSECURE_SKIP_VERIFY` (default: `false`)

Postgres:
- `POSTGRES_DSN` (default: empty)
- `POSTGRES_HOST` (default: `localhost`)
- `POSTGRES_PORT` (default: `5432`)
- `POSTGRES_USER` (default: `postgres`)
- `POSTGRES_PASSWORD` (default: `postgres`)
- `POSTGRES_DB` (default: `postgres`)
- `POSTGRES_SSLMODE` (default: `disable`)
- `POSTGRES_CONNECT_TIMEOUT` (default: `5s`)
- `POSTGRES_MAX_CONNS` (default: `10`)
- `POSTGRES_MIN_CONNS` (default: `0`)
- `POSTGRES_MAX_CONN_LIFETIME` (default: `0s`)
- `POSTGRES_MAX_CONN_IDLE_TIME` (default: `0s`)
- `POSTGRES_HEALTHCHECK_PERIOD` (default: `1m`)

Auth:
- `JWT_SECRET` (required)
- `JWT_ISSUER` (default: `boilerplate-go-fiber`)
- `ACCESS_TOKEN_TTL` (default: `15m`)
- `REFRESH_TOKEN_TTL` (default: `168h`)
- `REFRESH_TOKEN_CLEANUP_INTERVAL` (default: `1h`, set `0` to disable)
- `AUTH_MAX_LOGIN_ATTEMPTS` (default: `5`)
- `AUTH_LOCKOUT_DURATION` (default: `15m`)
- `AUTH_LOGIN_RATE_LIMIT` (default: `10`)
- `AUTH_LOGIN_RATE_WINDOW` (default: `1m`)

Swagger:
- `SWAGGER_ENABLED` (default: `false`)
- `SWAGGER_PATH` (default: `/docs`)

Observability:
- `METRICS_ENABLED` (default: `true`)
- `METRICS_PATH` (default: `/metrics`)
- `OTEL_ENABLED` (default: `false`)
- `OTEL_EXPORTER_OTLP_ENDPOINT` (default: `localhost:4317`)
- `OTEL_EXPORTER_OTLP_INSECURE` (default: `true`)
- `OTEL_SAMPLE_RATIO` (default: `1.0`)

Notes:
- Redis is initialized on startup; the app will fail to start if Redis is unreachable.
- Postgres is initialized on startup; the app will fail to start if Postgres is unreachable.
- `.env` is only loaded when `APP_ENV=local`.
- If `CORS_ALLOW_CREDENTIALS=true`, `CORS_ALLOW_ORIGINS` cannot be `*`.
- Refresh token cleanup runs every `REFRESH_TOKEN_CLEANUP_INTERVAL` when enabled.
- Login is rate-limited; repeated failures can trigger account lockout.

## API

GET `/health`:
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "status": "ok",
    "app": "boilerplate-go-fiber",
    "env": "local",
    "uptime": "12s",
    "time": "2025-01-01T12:00:00Z"
  }
}
```

GET `/metrics` (if enabled):
```json
{
  "uptime": "1m0s",
  "request": "123",
  "status": {
    "200": "120",
    "500": "3"
  }
}
```

## Swagger (OpenAPI)

Enable Swagger UI:
```
export SWAGGER_ENABLED=true
```

Open docs:
```
http://localhost:3000/docs
```

OpenAPI spec:
```
http://localhost:3000/docs/doc.json
```

Note: Swagger UI loads assets from a CDN. If you need offline docs, host the UI assets locally.

Generate docs with swag:
```
go install github.com/swaggo/swag/cmd/swag@latest
make swagger
```

## Auth API

POST `/auth/login`:
```json
{
  "email": "admin@example.com",
  "password": "secret"
}
```

POST `/auth/refresh`:
```json
{
  "refresh_token": "..."
}
```

POST `/auth/logout`:
```json
{
  "refresh_token": "..."
}
```

Successful login/refresh response:
```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "access_token": "jwt",
    "refresh_token": "opaque",
    "token_type": "bearer",
    "expires_in": 900
  }
}
```

## User API (Protected)

All user endpoints require `Authorization: Bearer <access_token>`.

- GET `/users` (permission: `user.read`)
- GET `/users/:id` (permission: `user.read`)
- POST `/users` (permission: `user.create`)
- PUT `/users/:id` (permission: `user.update`)
- DELETE `/users/:id` (permission: `user.delete`)
- GET `/users/:id/roles` (permission: `user.role.read`)
- PUT `/users/:id/roles` (permission: `user.role.update`)

## RBAC API (Protected)

All RBAC endpoints require `Authorization: Bearer <access_token>`.

Roles:
- GET `/rbac/roles` (permission: `role.read`)
- GET `/rbac/roles/:id` (permission: `role.read`)
- POST `/rbac/roles` (permission: `role.create`)
- PUT `/rbac/roles/:id` (permission: `role.update`)
- DELETE `/rbac/roles/:id` (permission: `role.delete`)
- GET `/rbac/roles/:id/permissions` (permission: `role.permission.read`)
- PUT `/rbac/roles/:id/permissions` (permission: `role.permission.update`)

Permissions:
- GET `/rbac/permissions` (permission: `permission.read`)
- GET `/rbac/permissions/:id` (permission: `permission.read`)
- POST `/rbac/permissions` (permission: `permission.create`)
- PUT `/rbac/permissions/:id` (permission: `permission.update`)
- DELETE `/rbac/permissions/:id` (permission: `permission.delete`)

Note: permission claims are embedded in access tokens at login/refresh time. If you update role permissions, users must refresh/re-login to get updated claims. User role changes bump `token_version` and invalidate existing access tokens.

## Middleware (HTTP)

- CORS
- Security headers (helmet)
- Request ID (`X-Request-ID`)
- Request logging (logrus JSON)
- Panic recovery (stack trace)
- OpenTelemetry tracing (optional)

## Project Structure

```
internal/
  app/
  config/
  logger/
  observability/
  transport/
    http/
      middleware.go
      router.go
      server.go
      auth/
      docs/
      health/
      rbac/
      response/
      user/
docs/
```

## Docker (Optional)

`docker-compose.yml` provides Redis, Postgres, Jaeger, and OpenTelemetry Collector:
```
docker-compose up -d
```
If you run the app inside Docker, set `REDIS_ADDR=redis:6379` and `POSTGRES_HOST=postgres`.

## Observability (Local)

Start tracing stack:
```
docker-compose up -d jaeger otel-collector
```
The collector exports traces to Jaeger over OTLP.

Enable OTEL exporter:
```
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

Open Jaeger UI:
```
http://localhost:16686
```

## Database Migrations

SQL migrations are in `migrations/` (each has `.up.sql` and `.down.sql`):
- `0001_auth_schema.up.sql`
- `0002_seed_rbac.up.sql`
- `0003_seed_rbac_management.up.sql`
- `0004_seed_user_role_permissions.up.sql`
- `0005_auth_security.up.sql`

You can run them using your preferred tool (e.g. `psql`, `golang-migrate`).

### Migrate Step by Step (Makefile)

1) Install migrate CLI (once):
```
brew install golang-migrate
```

2) Export DB connection:
```
export POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
```

3) Create a new migration:
```
make migrate-create name=create_users
```

4) Apply migrations:
```
make migrate-up
```

5) Check migration version:
```
make migrate-version
```

6) Roll back (default 1 step):
```
make migrate-down
```

7) Roll back N steps:
```
make migrate-down n=3
```

8) Force version (use with care):
```
make migrate-force version=2
```

## Create User (Manual)

You need at least one user before login works. Example with bcrypt:
```
htpasswd -bnBC 12 "" "secret" | tr -d ':\n'
```

Then insert:
```
INSERT INTO users (email, password_hash) VALUES ('admin@example.com', '<bcrypt-hash>');
```

## Build and Test

```
make build
make test
make lint
```

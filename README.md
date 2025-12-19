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

Notes:
- Redis is initialized on startup; the app will fail to start if Redis is unreachable.
- `.env` is only loaded when `APP_ENV=local`.
- If `CORS_ALLOW_CREDENTIALS=true`, `CORS_ALLOW_ORIGINS` cannot be `*`.

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

## Middleware (HTTP)

- CORS
- Security headers (helmet)
- Request ID (`X-Request-ID`)
- Request logging (logrus JSON)
- Panic recovery (stack trace)

## Project Structure

```
internal/
  app/
  config/
  logger/
  transport/
    http/
      middleware.go
      router.go
      server.go
      health/
      response/
```

## Docker (Optional)

`docker-compose.yml` currently provides Redis only:
```
docker-compose up -d
```
If you run the app inside Docker, set `REDIS_ADDR=redis:6379`.

## Build and Test

```
make build
make test
```

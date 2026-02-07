# How to start the proyect

Este es un proyecto de servicio de streaming construído en Go. Permite transmitir contenido multimedia utilizando AWS para la gestión de archivos y PostgreSQL para manejar la información de usuarios y videos. Además, soporta Docker para facilitar su despliegue en entornos locales y de producción.

## Tech Stack

**Client:** React, TailwindCSS

**Server:** Golang, Gin, AWS, Postgresql

## Architecture

![Architecture](docs/architecture.png)

### Request Flow
```
HTTP Request → Gin Router → Middlewares (CORS, Auth) → Controllers → Services → Data Layer (GORM/S3)
```

### Video Processing Pipeline
1. User uploads video via API (`POST /api/v1/streaming/upload`)
2. Server validates, saves locally, creates a Job (status: "pending") and enqueues task to RabbitMQ
3. Server responds immediately with `job_id` (HTTP 202)
4. Worker consumes task, converts to HLS (ffmpeg), generates thumbnail, uploads to S3/MinIO
5. Worker saves video metadata to PostgreSQL and updates job status to "completed"
6. Client queries job status (`GET /api/v1/jobs/:id`) and streams the video once ready

## Requerimientos

Para utilizar este proyecto, necesitas:

### Obligatorios

- **Git**: Para clonar el repositorio.
- **Go**: Para ejecutar y compilar el proyecto. (Versión recomendada: 1.24.0, utilizada en el `go.mod`).
- **Atlas**: Para gestionar migraciones de base de datos. [Instalacion](https://atlasgo.io/getting-started#installation).

### Opcionales

- **Docker**: Para contenerizar la aplicación.
- **Docker Compose**: Para orquestar servicios si se utiliza Docker.

## Installation

```bash
  git clone https://github.com/Unbot2313/go-streaming-service.git
  cd go-streaming-service/
```

## Usage/Examples

Para usarla con Go!:

```bash
    go mod tidy
    go run main.go
```

Con docker(incluye la instancia de postgresql en local):

```bash
    docker compose up --build
```

## Database Migrations

El proyecto usa [Atlas](https://atlasgo.io/) para migraciones declarativas. Atlas lee los modelos GORM y genera archivos SQL versionados automaticamente.

```bash
# Generar una migracion despues de modificar un modelo
make migrate-diff name=describe_change

# Aplicar migraciones pendientes
make migrate-apply DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"

# Ver estado de migraciones
make migrate-status DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```

## Contributing

Contributions are always welcome!

Please star a new fork, then make a pull request

In case of change the documentation make that:

```bash
    go install github.com/go-swagger/go-swagger/cmd/swagger@latest
    swag init #update the documentation from the swagger
```

## Rate Limiting

Uses the **Token Bucket** algorithm (`golang.org/x/time/rate`) to control request rates per client IP.

### How it works
- A "bucket" fills with tokens at a constant rate
- Each request consumes 1 token
- If the bucket is empty, the request is rejected with HTTP 429 (Too Many Requests)
- Tokens accumulate up to a maximum (burst), allowing short traffic spikes

### Current configuration

| Scope | Rate | Burst | Description |
|-------|------|-------|-------------|
| General (all routes) | 10 tokens/sec | 20 | Normal usage, allows short bursts |
| Auth (`/login`, `/register`) | 1 token every 20s | 3 | Brute force protection |

### Customization
Rate limits are configured in code:
- **General**: `main.go` - `middlewares.NewRateLimiter(10, 20)`
- **Auth**: `internal/routes/routes.go` - `middlewares.NewRateLimiter(rate.Every(20*time.Second), 3)`

To change limits, modify the two parameters: `rate` (tokens per second or interval) and `burst` (max accumulated tokens).

### Scaling note
The rate limiter stores counters in memory per server instance. If deploying multiple instances (e.g. Kubernetes), each instance tracks independently. For distributed rate limiting, replace with a Redis-backed solution like `github.com/ulule/limiter`.

## Features

- Manejar la carga de videos usando algun servicio de background jobs e implementar su visualizacion de progreso
- RefreshTokens
- Manejar transmision en vivos
- Terminar el README.md
- Optimizar el dockerFile

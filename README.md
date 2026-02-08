# Go Streaming Service

A multimedia streaming backend built with Go that enables video upload, asynchronous processing (HLS conversion via ffmpeg), and streaming. Uses RabbitMQ for job queuing, PostgreSQL for data persistence, and supports both AWS S3 and MinIO for storage.

## Tech Stack

- **Language:** Go 1.24
- **Framework:** Gin
- **Database:** PostgreSQL + GORM
- **Migrations:** Atlas
- **Queue:** RabbitMQ
- **Storage:** AWS S3 / MinIO
- **Auth:** JWT (HS256) + bcrypt
- **Media Processing:** ffmpeg / ffprobe
- **Monitoring:** Prometheus + Grafana
- **Logging:** slog (structured logging)
- **Docs:** Swagger (swag)

## Architecture

![Architecture](docs/architecture.png)

### Request Flow
```
HTTP Request -> Gin Router -> Middlewares (CORS, Auth, RateLimit) -> Controllers -> Services -> Data Layer (GORM/S3)
```

### Video Processing Pipeline
1. User uploads video via API (`POST /api/v1/streaming/upload`)
2. Server validates, saves locally, creates a Job (status: "pending") and enqueues task to RabbitMQ
3. Server responds immediately with `job_id` (HTTP 202)
4. Worker consumes task, converts to HLS (ffmpeg), generates thumbnail, uploads to S3/MinIO
5. Worker saves video metadata to PostgreSQL and updates job status to "completed"
6. Client queries job status (`GET /api/v1/jobs/:id`) and streams the video once ready

## Features

- Asynchronous video processing with RabbitMQ workers (HLS conversion + thumbnail generation)
- JWT authentication with refresh tokens and logout
- Video tagging system (many-to-many)
- Video search with pagination
- Rate limiting per IP (Token Bucket algorithm)
- Monitoring with Prometheus metrics and Grafana dashboards
- Structured logging with slog
- Health check and readiness endpoints
- Swagger API documentation
- Declarative database migrations with Atlas

## Requirements

- **Git**
- **Docker** and **Docker Compose** (recommended)

Or, to run without Docker:

- **Go** 1.24+
- **PostgreSQL** 15+
- **RabbitMQ** 3+
- **MinIO** or **AWS S3** account
- **ffmpeg** and **ffprobe** installed
- **Atlas** CLI - [Installation](https://atlasgo.io/getting-started#installation)

## Installation

```bash
git clone https://github.com/Unbot2313/go-streaming-service.git
cd go-streaming-service/
```

## Environment Variables

Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

Then edit `.env` with your values. Example for local development with Docker:

```env
PORT=3003
JWT_SECRET_KEY=your-secret-key-must-be-at-least-32-characters-long
LOCAL_STORAGE_PATH=./static/videos

CORS_ALLOWED_ORIGINS=http://localhost:3000

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=streaming_db

STORAGE_TYPE=minio

MINIO_ENDPOINT=localhost:9000
MINIO_BUCKET_NAME=streaming-videos
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VIDEO_QUEUE=video_processing
RABBITMQ_THUMBNAIL_QUEUE=thumbnail_generation

GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin
```

For AWS S3 storage or other configurations, see all available variables in `.env.example`.

### Validations

| Variable | Rule |
|----------|------|
| `JWT_SECRET_KEY` | **Required**. Must be at least 32 characters. The app will panic on startup if missing or too short. |
| `POSTGRES_PASSWORD` | Warns if set to default `postgres` |
| `RABBITMQ_PASSWORD` | Warns if set to default `guest` |
| `STORAGE_TYPE` | `minio` for local development, `s3` for production |
| `GRAFANA_*` | Only used by docker-compose, does not affect the Go app |

## Running with Docker (Recommended)

This starts PostgreSQL, RabbitMQ, MinIO, the API server, Prometheus, and Grafana:

```bash
docker compose up --build
```

The API server runs automatically. To also run the **video processing worker**, open a separate terminal:

```bash
docker exec -it go_streaming_service go run cmd/rabbitmq/consumer/main.go
```

## Running with Go

Make sure PostgreSQL, RabbitMQ, and MinIO (or S3) are running and accessible with the credentials in your `.env`.

```bash
go mod tidy
go run main.go
```

In a separate terminal, start the worker:

```bash
go run cmd/rabbitmq/consumer/main.go
```

Or using the Makefile:

```bash
make run      # API server
make worker   # Video processing worker
```

## Database Migrations

The project uses [Atlas](https://atlasgo.io/) for declarative database migrations. Atlas reads the GORM models and generates versioned SQL files.

### First time setup

If the database is empty (fresh install), apply all migrations:

```bash
make migrate-apply DATABASE_URL="postgres://postgres:postgres@localhost:5432/streaming_db?sslmode=disable"
```

If the database already has tables (e.g. from a previous version using AutoMigrate), use baseline to mark existing migrations as applied without running them:

```bash
make migrate-baseline version=20260207231247 DATABASE_URL="postgres://postgres:postgres@localhost:5432/streaming_db?sslmode=disable"
```

### After modifying a model

```bash
# Generate a new migration
make migrate-diff name=describe_change

# Review the generated SQL in migrations/

# Apply the new migration
make migrate-apply DATABASE_URL="postgres://postgres:postgres@localhost:5432/streaming_db?sslmode=disable"

# Check migration status
make migrate-status DATABASE_URL="postgres://postgres:postgres@localhost:5432/streaming_db?sslmode=disable"
```

## Local URLs

When running with Docker:

| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:3003 | - |
| Swagger Docs | http://localhost:3003/docs/index.html | - |
| Grafana | http://localhost:3001 | `GRAFANA_ADMIN_USER` / `GRAFANA_ADMIN_PASSWORD` from `.env` |
| Prometheus | http://localhost:9090 | No auth |
| RabbitMQ Management | http://localhost:15672 | `RABBITMQ_USER` / `RABBITMQ_PASSWORD` from `.env` |
| MinIO Console | http://localhost:9001 | `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` from `.env` |

## API Documentation

Interactive API docs are available at `/docs/index.html` when the server is running.

To update the documentation after modifying controller annotations:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init
```

Or using the Makefile:

```bash
make swagger
```

## CI

GitHub Actions runs automatically on every push and pull request to `main` and `develop`. The pipeline validates:

1. **Dependencies** - `go mod download`
2. **Lint** - `go vet ./...`
3. **Build** - `go build ./...`
4. **Tests** - `go test ./... -race -v`

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a branch based on `develop` (`git checkout -b feat/my-feature develop`)
3. Commit your changes
4. Push to your fork and open a Pull Request against `develop`

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-streaming-service is a multimedia streaming backend built with Go/Gin that enables video upload, asynchronous processing (HLS conversion via ffmpeg), and streaming. It uses RabbitMQ for job queuing and supports both AWS S3 and MinIO for storage.

## Common Commands

```bash
# Install dependencies
go mod tidy

# Run development server (port 3003)
go run main.go

# Run video processing worker
go run cmd/rabbitmq/consumer/main.go

# Run with Docker (includes PostgreSQL, RabbitMQ, MinIO)
docker compose up --build

# Regenerate Swagger documentation
swag init

# Generate a new migration after changing a model
make migrate-diff name=describe_change

# Apply pending migrations
make migrate-apply DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"

# Check migration status
make migrate-status DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```

## Architecture

The codebase follows a layered MVC pattern with interface-based dependency injection.

### Request Flow
```
HTTP Request → Gin Router → Middlewares (CORS, Auth) → Controllers → Services → Data Layer (GORM/S3)
```

### Key Directories
- `config/` - Singleton configuration (env, database, S3/MinIO clients)
- `internal/app/` - Dependency injection setup (`initializer.go`)
- `internal/controllers/` - Request handlers (user, auth, video, job, tag)
- `internal/services/` - Business logic layer
- `internal/services/storage/` - Storage abstraction (S3, MinIO)
- `internal/models/` - GORM database models
- `internal/middlewares/` - JWT auth, rate limiting, request logger
- `internal/routes/` - Route definitions
- `internal/mocks/` - Mock implementations for testing
- `cmd/rabbitmq/consumer/` - Video processing worker
- `cmd/atlas/` - Atlas GORM loader for migration generation
- `migrations/` - Versioned SQL migrations (managed by Atlas)
- `docs/` - Auto-generated Swagger documentation
- `static/` - Temporary video storage during processing

### Naming Conventions
- Controllers: `*Controller` interface + `*ControllerImpl` struct
- Services: `*Service` interface + `*ServiceImp` struct
- Models: `*Model` for database, `*Swagger` for API docs

### Video Processing Pipeline (Async)

**API Endpoint (immediate response):**
1. Upload validation (title required, extension, 100MB max)
2. Save to `./static/videos/`
3. Extract duration with ffprobe
4. Create Job with status "pending"
5. Publish task to RabbitMQ
6. Return job_id immediately

**Worker (background processing):**
1. Consume task from RabbitMQ
2. Update job status to "processing"
3. Convert to HLS format with ffmpeg (.m3u8 + .ts segments)
4. Generate WebP thumbnail at 8-second mark
5. Upload to storage (S3 or MinIO)
6. Save video metadata to PostgreSQL
7. Update job status to "completed"
8. Cleanup local temp files

### Configuration
Environment variables loaded from `.env` (see `.env.example`). Key configs:
- `JWT_SECRET_KEY` - JWT signing secret
- `POSTGRES_*` - Database connection
- `RABBITMQ_*` - Message queue connection
- `STORAGE_TYPE` - `minio` or `s3`
- `MINIO_*` - MinIO config (local development)
- `AWS_*` - S3 credentials (production)

### API Routes
Base path: `/api/v1`
- `/users/*` - User CRUD
- `/auth/*` - Login/register
- `/streaming/*` - Video upload (protected), retrieval, view counting
- `/jobs/:id` - Job status tracking
- `/static/*` - Serves streaming files
- `/docs/*` - Swagger UI

## Workers (RabbitMQ)

### Video Processing Worker
Procesa videos de forma asíncrona en background:
```bash
go run cmd/rabbitmq/consumer/main.go
```

**Responsabilidades:**
- Convierte video a HLS (ffmpeg)
- Genera thumbnail (ffmpeg)
- Sube archivos a storage (S3/MinIO)
- Actualiza estado del job en DB
- Limpia archivos temporales

### Colas
| Cola | Variable ENV | Propósito |
|------|--------------|-----------|
| video_processing | `RABBITMQ_VIDEO_QUEUE` | Procesamiento de videos |

## Storage

El proyecto soporta dos backends de storage intercambiables:

### MinIO (desarrollo local)
```env
STORAGE_TYPE=minio
MINIO_ENDPOINT=localhost:9000
MINIO_BUCKET_NAME=streaming-videos
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
```

### AWS S3 (producción)
```env
STORAGE_TYPE=s3
AWS_REGION=us-east-1
AWS_BUCKET_NAME=my-bucket
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=secret...
```

La abstracción está en `internal/services/storage/` con una interfaz común `StorageService`.

## Database Migrations (Atlas)

The project uses [Atlas](https://atlasgo.io/) with the GORM provider for declarative database migrations. Atlas reads the GORM structs directly and generates versioned SQL migration files.

### How it works
1. Models in `internal/models/` are the source of truth for the database schema
2. `cmd/atlas/main.go` is the loader that passes the models to Atlas
3. `atlas.hcl` configures Atlas to use the GORM loader and store migrations in `migrations/`
4. Atlas compares the current models against the migration history and generates the diff

### Workflow
```bash
# After modifying a model (adding a field, changing a type, etc.):
make migrate-diff name=add_user_avatar

# This generates a new .sql file in migrations/ with the exact DDL changes

# Apply migrations to the database:
make migrate-apply DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```

### Important
- Never edit migration files after they have been applied
- Always review the generated SQL before applying to production
- The `atlas.sum` file ensures migration integrity — commit it to git

## Dependencies

- **Web**: Gin with CORS
- **Database**: GORM + PostgreSQL driver
- **Migrations**: Atlas with GORM provider
- **Auth**: JWT (HS256, 24h expiration), bcrypt
- **Queue**: RabbitMQ (amqp091-go)
- **Storage**: AWS SDK v2 for S3, MinIO SDK for local
- **Media**: External ffmpeg/ffprobe (must be installed)

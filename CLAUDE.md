# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-streaming-service is a multimedia streaming backend built with Go/Gin that enables video upload, processing (HLS conversion via ffmpeg), and streaming. It uses AWS S3 for storage and PostgreSQL for persistence.

## Common Commands

```bash
# Install dependencies
go mod tidy

# Run development server (port 3003)
go run main.go

# Run with Docker (includes PostgreSQL)
docker compose up --build

# Regenerate Swagger documentation
swag init
```

## Architecture

The codebase follows a layered MVC pattern with interface-based dependency injection.

### Request Flow
```
HTTP Request → Gin Router → Middlewares (CORS, Auth) → Controllers → Services → Data Layer (GORM/S3)
```

### Key Directories
- `config/` - Singleton configuration (env, database, S3 clients)
- `internal/app/` - Dependency injection setup (`initializer.go`)
- `internal/controllers/` - Request handlers (user, auth, video)
- `internal/services/` - Business logic layer
- `internal/models/` - GORM database models
- `internal/middlewares/` - JWT auth middleware
- `internal/routes/` - Route definitions
- `docs/` - Auto-generated Swagger documentation
- `static/` - Temporary video storage during processing

### Naming Conventions
- Controllers: `*Controller` interface + `*ControllerImpl` struct
- Services: `*Service` interface + `*ServiceImp` struct
- Models: `*Model` for database, `*Swagger` for API docs

### Video Processing Pipeline
1. Upload validation (extension, 100MB max)
2. Save to `./static/videos/`
3. Extract duration with ffprobe
4. Convert to HLS format with ffmpeg (.m3u8 + .ts segments)
5. Generate WebP thumbnail at 8-second mark
6. Upload to S3
7. Save metadata to PostgreSQL
8. Cleanup local temp files

### Configuration
Environment variables loaded from `.env` (see `.env.example`). Key configs:
- `JWT_SECRET_KEY` - JWT signing secret
- `AWS_*` - S3 credentials and bucket
- `POSTGRES_*` - Database connection
- `DOCKER_MODE` - Container networking flag

### API Routes
Base path: `/api/v1`
- `/users/*` - User CRUD
- `/auth/*` - Login/register
- `/streaming/*` - Video upload (protected), retrieval, view counting
- `/static/*` - Serves streaming files
- `/docs/*` - Swagger UI

## Dependencies

- **Web**: Gin with CORS
- **Database**: GORM + PostgreSQL driver
- **Auth**: JWT (HS256, 24h expiration), bcrypt
- **AWS**: SDK v2 for S3
- **Media**: External ffmpeg/ffprobe (must be installed)

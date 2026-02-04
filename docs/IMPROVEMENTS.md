# Plan de Mejoras - Go Streaming Service

Plan de fixes y mejoras organizado en 6 fases por prioridad.

---

## FASE 1: Fixes Criticos de Seguridad

### 1.1 Proteger rutas sin autenticacion
**Rama:** `fix/security-routes-auth`
**Archivos:** `internal/routes/routes.go`, `internal/controllers/userController.go`, `internal/controllers/jobController.go`

- `DELETE /users/:id` y `POST /users/` no tienen auth middleware
- `DELETE /users/:id` debe verificar que el usuario autenticado es el dueno de la cuenta
- `GET /jobs/:jobid` debe validar que el job pertenece al usuario autenticado
- Mantener `GET /users/id/:id` y `GET /users/username/:username` como publicos (lectura)

### 1.2 Eliminar JWT secret por defecto
**Rama:** `fix/security-jwt-defaults`
**Archivos:** `config/env.go`

- Quitar default `"secretJwtKey"` de `getEnv("JWT_SECRET_KEY", "secretJwtKey")`
- Agregar funcion `validateConfig()` que haga panic si `JWT_SECRET_KEY` esta vacio o tiene menos de 32 caracteres

### 1.3 Restringir CORS
**Rama:** `fix/security-cors`
**Archivos:** `main.go`, `config/env.go`

- Agregar `CORS_ALLOWED_ORIGINS` a env config (comma-separated, default `http://localhost:3000`)
- Reemplazar `cors.Default()` por `cors.New(cors.Config{...})` con origenes, metodos y headers especificos

### 1.4 Sanitizar respuestas de error
**Rama:** `fix/security-error-sanitization`
**Archivos:** Todos los controllers, crear `internal/helpers/errors.go`

- Crear helper `HandleError(c, statusCode, userMessage, err)` que loguea el error completo pero retorna mensaje generico al cliente
- Reemplazar todos los `c.JSON(xxx, gin.H{"error": err.Error()})` en controllers

### 1.5 Agregar rate limiting
**Rama:** `feat/rate-limiting`
**Archivos:** Crear `internal/middlewares/ratelimit.go`, `main.go`, `internal/routes/routes.go`

- Dependencia: `golang.org/x/time/rate`
- Rate limit general: 10 req/s con burst de 20
- Rate limit estricto para `/auth/login`: 3 req/min (proteccion contra fuerza bruta)

### 1.6 Ocultar password hash en respuestas JSON
**Rama:** `fix/security-password-leak`
**Archivos:** `internal/models/user.go`

- Cambiar `json:"password"` a `json:"-"` en el campo Password del User struct
- Cambiar `json:"refresh_token"` a `json:"-"` en RefreshToken

---

## FASE 2: Completar Features Incompletos

### 2.1 Implementar endpoint Register
**Rama:** `feat/auth-register`
**Archivos:** `internal/controllers/authController.go`, `internal/services/authService.go`

- Implementar el stub vacio en `authController.Register`
- Validar: username (3-100 chars), password (min 8 chars), email (formato valido)
- Reutilizar `userService.CreateUser()` + `authService.GenerateToken()`
- Retornar 201 con token y datos del usuario
- Manejar duplicados con 409 Conflict

### 2.2 Implementar UpdateVideo y DeleteVideo
**Rama:** `feat/video-crud`
**Archivos:** `internal/services/databaseVideoService.go`, `internal/controllers/videoController.go`, `internal/routes/routes.go`

- `UpdateVideo`: usar GORM `.Updates()` para title/description
- `DeleteVideo`: soft delete + eliminar de storage
- Agregar rutas `PATCH /streaming/:videoid` y `DELETE /streaming/:videoid` (protegidas)
- Verificar ownership (solo el dueno puede modificar/eliminar)

### 2.3 Agregar paginacion a FindLatestVideos
**Rama:** `feat/pagination`
**Archivos:** `internal/services/databaseVideoService.go`, `internal/controllers/videoController.go`

- Cambiar `FindLatestVideos()` a `FindLatestVideos(page, pageSize int)`
- Agregar `.Limit(pageSize).Offset((page-1)*pageSize)` al query
- Parsear query params `page` (default 1) y `page_size` (default 20, max 100)
- Retornar respuesta con `{ data, page, page_size, total }`

### 2.4 Implementar RefreshToken
**Rama:** `feat/refresh-token`
**Archivos:** `internal/services/authService.go`, `internal/controllers/authController.go`, `internal/routes/routes.go`

- Agregar `GenerateRefreshToken()` y `ValidateRefreshToken()` al AuthService
- Refresh token con expiracion de 7 dias, almacenar hash en DB
- Agregar `POST /auth/refresh` (acepta refresh token, retorna nuevo par de tokens)
- Agregar `POST /auth/logout` (protegido, limpia refresh token en DB)

---

## FASE 3: Infraestructura y Confiabilidad

### 3.1 Graceful shutdown
**Rama:** `feat/graceful-shutdown`
**Archivos:** `main.go`, `cmd/rabbitmq/consumer/main.go`

- Reemplazar `r.Run(":3003")` con `http.Server` + manejo de senales (SIGINT, SIGTERM)
- Timeout de 30s para cerrar conexiones en progreso
- Cerrar conexiones de RabbitMQ y DB al apagar
- En el worker: manejar senales para esperar que el job actual termine

### 3.2 Health check endpoint
**Rama:** `feat/health-check`
**Archivos:** `internal/routes/routes.go`, `main.go`, `docker-compose.yml`

- `GET /health` -> `{"status": "ok"}` (liveness)
- `GET /ready` -> verificar DB y RabbitMQ (readiness, retorna 503 si alguno falla)
- Agregar healthcheck al servicio app en docker-compose

### 3.3 Dockerfile production-ready
**Rama:** `fix/dockerfile-production`
**Archivos:** `Dockerfile`, `docker-compose.yml`

- Multi-stage build:
  - Stage 1 (builder): `golang:1.24`, compilar binarios con `go build`
  - Stage 2 (runtime): `debian:bookworm-slim`, instalar solo ffmpeg, copiar binarios
- Imagen final ~200MB en vez de ~1.5GB
- Agregar servicio `worker` separado en docker-compose
- Corregir inconsistencia Go 1.23.1 (Dockerfile) vs Go 1.24.0 (go.mod)

### 3.4 Validacion de variables de entorno
**Rama:** `fix/env-validation`
**Archivos:** `config/env.go`

- Agregar campo `APP_ENV` (`development`/`production`)
- En produccion: rechazar defaults debiles (postgres password "postgres", rabbitmq "guest")
- En desarrollo: log de advertencia si se usan defaults

### 3.5 SSL para PostgreSQL
**Rama:** `fix/env-validation` (misma rama que 3.4)
**Archivos:** `config/db.go`, `config/env.go`

- Agregar `POSTGRES_SSLMODE` al config (default `disable` en dev, `require` en produccion)
- Usar el valor en el DSN

---

## FASE 4: Calidad de Codigo

### 4.1 Estandarizar formato de respuestas API
**Rama:** `refactor/code-quality`
**Archivos:** Todos los controllers

- Definir struct estandar: `{ success: bool, data: any, error: { code, message } }`
- Codigos HTTP consistentes: 400 validacion, 401 auth, 403 forbidden, 404 not found, 409 conflict, 500 server error

### 4.2 Corregir DI del AuthMiddleware
**Rama:** `refactor/code-quality`
**Archivos:** `internal/middlewares/auth.go`, `internal/routes/routes.go`, `internal/app/initializer.go`

- Actualmente `auth.go` crea su propio `authService` como variable de paquete (rompe el patron DI)
- Convertir a closure: `func AuthMiddleware(authService AuthService) gin.HandlerFunc`
- Pasar authService desde el initializer a traves de routes

### 4.3 Fix return type de NewUserController
**Rama:** `refactor/code-quality`
**Archivos:** `internal/controllers/userController.go`

- Retorna `*UserControllerImp` en vez de `UserController` (interface)

### 4.4 Optimizar conexion RabbitMQ en uploads
**Rama:** `refactor/code-quality`
**Archivos:** `internal/controllers/videoController.go`, `internal/app/initializer.go`

- Actualmente cada upload crea una nueva conexion RabbitMQ y la cierra con defer
- Usar una conexion persistente inyectada via DI

---

## FASE 5: Logging Estructurado

### 5.1 Migrar a slog
**Rama:** `feat/structured-logging`
**Archivos:** Todos los archivos que usan `log` package (37 instancias en 8 archivos)

- Usar `log/slog` (incluido en Go 1.21+, no requiere dependencia externa)
- Crear `internal/logger/logger.go` con logger configurado (JSON en produccion, texto en desarrollo)
- Reemplazar todos los `log.Printf`/`log.Println` con `slog.Info`/`slog.Error`/`slog.Warn`
- Agregar middleware de request logging (method, path, status, latency, IP)

---

## FASE 6: Testing y CI/CD

### 6.1 Tests unitarios
**Rama:** `feat/testing`
**Archivos a crear:** `*_test.go` en services/

- Crear mocks para las interfaces (AuthService, VideoService, StorageService, etc.)
- Tests prioritarios:
  - `authService_test.go`: GenerateToken, ValidateToken, Login, HashPassword
  - `jobService_test.go`: CreateJob, UpdateJobStatus
  - `videoService_test.go`: IsValidVideoExtension, SaveVideo

### 6.2 Tests de integracion
**Rama:** `feat/testing`
**Archivos a crear:** `*_test.go` en controllers/

- Usar `httptest.NewRecorder()` + Gin test mode
- Tests de flujo completo: auth requerido, validacion, casos de exito y error

### 6.3 CI/CD con GitHub Actions
**Rama:** `feat/ci-cd`
**Archivos a crear:** `.github/workflows/ci.yml`, `Makefile`

- Pipeline: checkout -> setup Go 1.24 -> `go vet` -> `go test -race` -> coverage
- Makefile con targets: `build`, `run`, `test`, `lint`, `docker-build`, `swagger`

---

## Estrategia de Ramas

Base: todas las ramas salen de `develop` y se mergean a `develop` via PR. Cuando `develop` este estable, se mergea a `main`.

```
main (produccion)
  |
  +-- develop (integracion)
        |
        +-- fix/security-routes-auth         (PR -> develop)
        +-- fix/security-jwt-defaults        (PR -> develop)
        +-- fix/security-cors                (PR -> develop)
        +-- fix/security-error-sanitization  (PR -> develop)
        +-- feat/rate-limiting               (PR -> develop)
        +-- fix/security-password-leak       (PR -> develop)
        |
        +-- feat/auth-register               (PR -> develop)
        +-- feat/video-crud                  (PR -> develop)
        +-- feat/pagination                  (PR -> develop)
        +-- feat/refresh-token               (PR -> develop)
        |
        +-- feat/graceful-shutdown           (PR -> develop)
        +-- feat/health-check               (PR -> develop)
        +-- fix/dockerfile-production        (PR -> develop)
        +-- fix/env-validation               (PR -> develop)
        |
        +-- refactor/code-quality            (PR -> develop)
        +-- feat/structured-logging          (PR -> develop)
        |
        +-- feat/testing                     (PR -> develop)
        +-- feat/ci-cd                       (PR -> develop)
```

### Reglas
- Los fixes de seguridad (fase 1) pueden mergearse en paralelo
- Las ramas de fase 2+ deben crearse despues de que fase 1 este mergeada en develop
- Cada PR debe incluir descripcion de cambios y pasos de verificacion
- No hacer force push a develop ni main

---

## Orden de Dependencias

```
Fase 1 (Seguridad)      <- sin dependencias, hacer primero
    |
Fase 2 (Features)       <- depende de Fase 1 para auth en nuevas rutas
    |
Fase 3 (Infra)          <- depende de Fase 1.2 para patron de validacion
    |
Fase 4 (Codigo)         <- depende de Fase 1.4 para error helper
    |
Fase 5 (Logging)        <- independiente, puede hacerse en paralelo con Fase 4
    |
Fase 6 (Testing)        <- depende de que todo lo anterior este estable
```

## Verificacion

Para validar cada fase:
1. **Fase 1**: Intentar requests sin token a rutas protegidas (esperar 401), verificar que errores no expongan internals
2. **Fase 2**: Probar register, CRUD de videos, paginacion, refresh token flow
3. **Fase 3**: Enviar SIGTERM durante procesamiento, verificar graceful shutdown; verificar health endpoints
4. **Fase 4**: Revisar consistencia de respuestas API con Postman
5. **Fase 5**: Verificar logs en formato JSON con campos estructurados
6. **Fase 6**: `go test ./... -race -coverprofile=coverage.out` pasa sin errores

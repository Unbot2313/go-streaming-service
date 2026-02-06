# Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilar ambos binarios (CGO deshabilitado para binario estático)
RUN CGO_ENABLED=0 go build -o /app/bin/server main.go
RUN CGO_ENABLED=0 go build -o /app/bin/worker cmd/rabbitmq/consumer/main.go

# Stage 2: Runtime
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copiar binarios compilados
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/worker /app/worker

# Crear directorio para archivos temporales de video
RUN mkdir -p /app/static/videos /app/static/temp

EXPOSE 3003

# Por defecto ejecuta el servidor API
CMD ["/app/server"]

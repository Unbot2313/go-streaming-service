# How to start the proyect

Este es un proyecto de servicio de streaming construído en Go. Permite transmitir contenido multimedia utilizando AWS para la gestión de archivos y PostgreSQL para manejar la información de usuarios y videos. Además, soporta Docker para facilitar su despliegue en entornos locales y de producción.

## Tech Stack

**Client:** React, TailwindCSS

**Server:** Golang, Gin, AWS, Postgresql

## Requerimientos

Para utilizar este proyecto, necesitas:

### Obligatorios

- **Git**: Para clonar el repositorio.
- **Go**: Para ejecutar y compilar el proyecto. (Versión recomendada: 1.23.1, utilizada en el `go.mod`).

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

## Contributing

Contributions are always welcome!

Please star a new fork, then make a pull request

In case of change the documentation make that:

```bash
    go install github.com/go-swagger/go-swagger/cmd/swagger@latest
    swag init #update the documentation from the swagger
```

## Features

- RefreshTokens
- Manejar transmision en vivos
- Terminar el README.md
- Optimizar el dockerFile

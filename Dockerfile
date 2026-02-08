# Usa una imagen base oficial de Go
FROM golang:1.24

# Instalar ffmpeg
RUN apt-get update && apt-get install -y ffmpeg

# Crea y define el directorio de trabajo
WORKDIR /app

# Copia los archivos del proyecto al contenedor
COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

# Expone el puerto en el que escucha tu aplicación
EXPOSE 3003

# Comando por defecto
CMD ["go", "run", "main.go"]

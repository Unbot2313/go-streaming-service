package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

func main() {
	// Cargar .env
	godotenv.Load()

	// Obtener configuración
	cfg := config.GetConfig()

	// Mensaje a enviar (puede venir de argumentos)
	message := bodyFrom(os.Args)

	// Crear servicio
	rabbitService := services.NewRabbitMQService()

	// Conectar
	err := rabbitService.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitService.Close()

	// Publicar mensaje a la cola de thumbnails
	err = rabbitService.Publish(cfg.RabbitMQThumbnailQueue, []byte(message))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[x] Mensaje enviado a '%s': %s", cfg.RabbitMQThumbnailQueue, message)
}

// bodyFrom construye el mensaje a partir de los argumentos
func bodyFrom(args []string) string {
	if len(args) < 2 || args[1] == "" {
		return "thumbnail_test"
	}
	return strings.Join(args[1:], " ")
}

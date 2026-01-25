package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

func main() {
	// Cargar .env
	godotenv.Load()

	// Obtener configuración
	cfg := config.GetConfig()

	// Crear servicio
	rabbitService := services.NewRabbitMQService()

	// Conectar
	err := rabbitService.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitService.Close()

	// Consumir mensajes de la cola de thumbnails
	err = rabbitService.Consume(cfg.RabbitMQThumbnailQueue, processThumbnail)
	if err != nil {
		log.Fatal(err)
	}

	// Mantener el programa corriendo
	log.Printf("[*] Thumbnail worker escuchando en '%s'. Presiona CTRL+C para salir", cfg.RabbitMQThumbnailQueue)
	select {}
}

// processThumbnail simula la generación de un thumbnail
func processThumbnail(message []byte) error {
	log.Printf("[>] Generando thumbnail para: %s", string(message))

	// Simular generación de thumbnail (5 segundos)
	time.Sleep(5 * time.Second)

	log.Printf("[✓] Thumbnail generado para: %s", string(message))
	return nil
}

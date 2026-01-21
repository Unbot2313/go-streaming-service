package main

import (
	"encoding/json"
	"log"

	"github.com/joho/godotenv"
	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

// Servicios globales para el worker
var (
	jobService           services.JobService
	videoService         services.VideoService
	databaseVideoService services.DatabaseVideoService
	filesService         services.FilesService
)

func main() {
	// Cargar .env
	godotenv.Load()

	// Obtener configuración
	cfg := config.GetConfig()

	// Inicializar servicios
	initServices()

	// Crear servicio RabbitMQ
	rabbitService := services.NewRabbitMQService()

	// Conectar
	err := rabbitService.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitService.Close()

	// Consumir mensajes de la cola de video
	err = rabbitService.Consume(cfg.RabbitMQVideoQueue, processVideoTask)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[*] Worker escuchando en '%s'. Presiona CTRL+C para salir", cfg.RabbitMQVideoQueue)
	select {} // Bloquea indefinidamente
}

// initServices inicializa los servicios necesarios para procesar videos
func initServices() {
	jobService = services.NewJobService()
	filesService = services.NewFilesService()
	S3configuration := services.GetS3Configuration()
	videoService = services.NewVideoService(S3configuration, filesService)
	databaseVideoService = services.NewDatabaseVideoService()

	log.Println("[*] Servicios inicializados")
}

// processVideoTask procesa una tarea de video recibida de RabbitMQ
func processVideoTask(message []byte) error {
	// 1. Parsear el mensaje JSON
	var task models.VideoTask
	if err := json.Unmarshal(message, &task); err != nil {
		log.Printf("[!] Error parseando mensaje: %s", err)
		return err
	}

	log.Printf("[>] Procesando job: %s, archivo: %s", task.JobID, task.UniqueName)

	// 2. Actualizar job a "processing"
	if err := jobService.UpdateJobStatus(task.JobID, "processing", ""); err != nil {
		log.Printf("[!] Error actualizando job a processing: %s", err)
		return err
	}

	// 3. Convertir video a HLS (ffmpeg)
	log.Printf("[.] Convirtiendo a HLS: %s", task.UniqueName)
	filesPath, err := videoService.FormatVideo(task.UniqueName)
	if err != nil {
		log.Printf("[!] Error en FormatVideo: %s", err)
		jobService.UpdateJobStatus(task.JobID, "failed", "Error convirtiendo video: "+err.Error())
		return err
	}

	// 4. Generar thumbnail
	log.Printf("[.] Generando thumbnail...")
	_, err = services.SaveThumbnail(task.LocalPath, filesPath)
	if err != nil {
		log.Printf("[!] Error en SaveThumbnail: %s", err)
		jobService.UpdateJobStatus(task.JobID, "failed", "Error generando thumbnail: "+err.Error())
		filesService.RemoveFolder(filesPath)
		return err
	}

	// 5. Subir a S3
	log.Printf("[.] Subiendo a S3...")
	savedDataInS3, baseFolder, err := videoService.UploadFilesFromFolderToS3(filesPath)
	if err != nil {
		log.Printf("[!] Error subiendo a S3: %s", err)
		jobService.UpdateJobStatus(task.JobID, "failed", "Error subiendo a S3: "+err.Error())
		filesService.RemoveFolder(filesPath)
		return err
	}

	// 6. Guardar video en base de datos
	log.Printf("[.] Guardando en base de datos...")
	videoData := &models.Video{
		Id:           task.JobID, // Usamos el mismo ID del job para el video
		Title:        task.Title,
		Description:  task.Description,
		M3u8FileURL:  savedDataInS3.M3u8FileURL,
		ThumbnailURL: savedDataInS3.ThumbnailURL,
	}

	_, err = databaseVideoService.CreateVideo(videoData, task.UserID)
	if err != nil {
		log.Printf("[!] Error guardando en DB: %s", err)
		jobService.UpdateJobStatus(task.JobID, "failed", "Error guardando en DB: "+err.Error())
		// Borrar de S3 si falla
		videoService.DeleteS3Folder(baseFolder + "/")
		filesService.RemoveFolder(filesPath)
		return err
	}

	// 7. Actualizar job a "completed"
	if err := jobService.UpdateJobCompleted(task.JobID, task.JobID); err != nil {
		log.Printf("[!] Error actualizando job a completed: %s", err)
		return err
	}

	// 8. Cleanup - Borrar archivos locales
	log.Printf("[.] Limpiando archivos locales...")
	filesService.RemoveFile(task.LocalPath)   // Video original
	filesService.RemoveFolder(filesPath)       // Carpeta con .ts y .m3u8

	log.Printf("[✓] Job completado: %s", task.JobID)
	return nil
}

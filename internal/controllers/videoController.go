package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/helpers"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type VideoController interface {
	GetLatestVideos(c *gin.Context)
	CreateVideo(c *gin.Context)
	GetVideoByID(c *gin.Context)
	IncrementViews(c *gin.Context)
}

// CreateVideoRequest valida los campos del formulario de upload
type CreateVideoRequest struct {
	Title       string `form:"title" binding:"required,min=1,max=100"`
	Description string `form:"description" binding:"max=500"`
}

// GetLatestVideos	godoc
// @Summary 		Get latest videos with pagination
// @Description 	Retrieve the latest videos ordered by creation date. Supports pagination via query params.
// @Tags 			streaming
// @Produce 		json
// @Param 			page query int false "Page number (default: 1)" default(1)
// @Param 			page_size query int false "Items per page (default: 10, max: 50)" default(10)
// @Success 		200 {object} services.PaginatedVideos{}
// @Failure 		400 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/streaming/ [get]
func (vc *VideoControllerImpl) GetLatestVideos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	result, err := vc.databaseVideoService.FindLatestVideos(page, pageSize)
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not retrieve videos", err)
		return
	}

	c.JSON(http.StatusOK, result)
}


// GetVideoByID		godoc
// @Summary 		Get a video by ID
// @Description 	Get a video by its ID
// @Tags 			streaming
// @Produce 		json
// @Param 			videoid path string true "Video ID"
// @Success 		200 {object} models.VideoSwagger{}
// @Failure 		400 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/streaming/id/{videoid} [get]
func (vc *VideoControllerImpl) GetVideoByID(c *gin.Context) {
	videoId := c.Param("videoid")

	video, err := vc.databaseVideoService.FindVideoByID(videoId)

	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "Video not found", err)
		return
	}

	c.JSON(http.StatusOK, video)
}

// IncrementViews		godoc
// @Summary 		Increment the views of a video
// @Description 	Increment the views of a video by 1
// @Tags 			streaming
// @Produce 		json
// @Param 			videoid path string true "Video ID"
// @Success 		200 {object} models.VideoSwagger{}	
// @Failure 		400 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/streaming/views/{videoid} [patch]
func (vc *VideoControllerImpl) IncrementViews(c *gin.Context) {
	videoId := c.Param("videoid")

	video, err := vc.databaseVideoService.IncrementViews(videoId)

	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not update views", err)
		return
	}

	c.JSON(http.StatusOK, video)
	
}

// CreateVideo godoc
// @Summary 		Upload a video for processing
// @Description 	Upload a video file and queue it for async processing. Returns a job ID to track progress.
// @Tags 			streaming
// @Accept 			multipart/form-data
// @Produce 		json
// @Param 			title formData string true "Video Title"
// @Param 			description formData string false "Video Description"
// @Param 			video formData file true "Video File"
// @Success 		202 {object} models.JobSwagger{}
// @Failure 		400 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/streaming/upload [post]
func (vc *VideoControllerImpl) CreateVideo(c *gin.Context) {
	cfg := config.GetConfig()

	// 1. Recuperar el usuario del contexto (del middleware JWT)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(500, gin.H{"error": "User not found in context"})
		return
	}

	authenticatedUser, ok := user.(*models.User)
	if !ok {
		c.JSON(500, gin.H{"error": "Failed to parse user data"})
		return
	}

	// 2. Validar campos requeridos (title obligatorio)
	var req CreateVideoRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title es requerido (máx 100 caracteres)"})
		return
	}

	// 3. Validar extensión del archivo
	if !vc.videoService.IsValidVideoExtension(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo no es un tipo de video válido."})
		return
	}

	// 4. Validar tamaño del archivo (máx 100MB)
	fileSize := c.Request.ContentLength
	const maxFileSize = 100 * 1024 * 1024
	if fileSize > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo excede el límite de tamaño permitido."})
		return
	}

	// 5. Guardar archivo en local (rápido)
	// TODO: Agregar compresión de video antes de encolar
	videoData, err := vc.videoService.SaveVideo(c.Request.Context(), c)
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not save video", err)
		return
	}

	// 6. Crear Job en DB con status "pending"
	job := &models.Job{
		Id:          videoData.Id,
		UserID:      authenticatedUser.Id,
		Status:      "pending",
		LocalPath:   videoData.LocalPath,
		UniqueName:  videoData.UniqueName,
		Title:       videoData.Title,
		Description: videoData.Description,
	}

	createdJob, err := vc.jobService.CreateJob(job)
	if err != nil {
		// Si falla crear el job, limpiar el video local
		vc.videoService.GetFilesService().RemoveFile(videoData.LocalPath)
		helpers.HandleError(c, http.StatusInternalServerError, "Could not create processing job", err)
		return
	}

	// 7. Conectar a RabbitMQ
	err = vc.rabbitMQService.Connect()
	if err != nil {
		vc.jobService.UpdateJobStatus(createdJob.Id, "failed", "Error conectando a RabbitMQ")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a cola de procesamiento"})
		return
	}
	defer vc.rabbitMQService.Close()

	// 8. Crear y serializar tarea para la cola
	videoTask := models.VideoTask{
		JobID:       createdJob.Id,
		UserID:      authenticatedUser.Id,
		LocalPath:   videoData.LocalPath,
		UniqueName:  videoData.UniqueName,
		Title:       videoData.Title,
		Description: videoData.Description,
		Duration:    videoData.Duration,
	}

	taskJSON, err := json.Marshal(videoTask)
	if err != nil {
		vc.jobService.UpdateJobStatus(createdJob.Id, "failed", "Error serializando tarea")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparando tarea"})
		return
	}

	// 9. Publicar tarea a la cola de video
	err = vc.rabbitMQService.Publish(cfg.RabbitMQVideoQueue, taskJSON)
	if err != nil {
		vc.jobService.UpdateJobStatus(createdJob.Id, "failed", "Error publicando a cola")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error encolando tarea"})
		return
	}

	log.Printf("[x] Video encolado: job_id=%s, file=%s", createdJob.Id, videoData.UniqueName)

	// 10. Responder inmediatamente con el job_id
	// NOTA: La limpieza de archivos locales la hace el WORKER después de procesar
	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  createdJob.Id,
		"status":  createdJob.Status,
		"message": "Video en cola de procesamiento. Consulta GET /jobs/" + createdJob.Id,
	})
}

type VideoControllerImpl struct {
	videoService         services.VideoService
	databaseVideoService services.DatabaseVideoService
	jobService           services.JobService
	rabbitMQService      services.RabbitMQService
}

func NewVideoController(videoService services.VideoService, databaseVideoService services.DatabaseVideoService, jobService services.JobService, rabbitMQService services.RabbitMQService) VideoController {
	return &VideoControllerImpl{
		videoService:         videoService,
		databaseVideoService: databaseVideoService,
		jobService:           jobService,
		rabbitMQService:      rabbitMQService,
	}
}
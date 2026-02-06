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
	UpdateVideo(c *gin.Context)
	DeleteVideo(c *gin.Context)
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

	helpers.Success(c, http.StatusOK, result)
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

	helpers.Success(c, http.StatusOK, video)
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

	helpers.Success(c, http.StatusOK, video)
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
		helpers.HandleError(c, http.StatusInternalServerError, "User not found in context", nil)
		return
	}

	authenticatedUser, ok := user.(*models.User)
	if !ok {
		helpers.HandleError(c, http.StatusInternalServerError, "Failed to parse user data", nil)
		return
	}

	// 2. Validar campos requeridos (title obligatorio)
	var req CreateVideoRequest
	if err := c.ShouldBind(&req); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "title es requerido (max 100 caracteres)", err)
		return
	}

	// 3. Validar extensión del archivo
	if !vc.videoService.IsValidVideoExtension(c) {
		helpers.HandleError(c, http.StatusBadRequest, "El archivo no es un tipo de video valido", nil)
		return
	}

	// 4. Validar tamaño del archivo (máx 100MB)
	fileSize := c.Request.ContentLength
	const maxFileSize = 100 * 1024 * 1024
	if fileSize > maxFileSize {
		helpers.HandleError(c, http.StatusBadRequest, "El archivo excede el limite de tamaño permitido", nil)
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

	// 7. Crear y serializar tarea para la cola
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
		helpers.HandleError(c, http.StatusInternalServerError, "Error preparando tarea", err)
		return
	}

	// 9. Publicar tarea a la cola de video
	err = vc.rabbitMQService.Publish(cfg.RabbitMQVideoQueue, taskJSON)
	if err != nil {
		vc.jobService.UpdateJobStatus(createdJob.Id, "failed", "Error publicando a cola")
		helpers.HandleError(c, http.StatusInternalServerError, "Error encolando tarea", err)
		return
	}

	log.Printf("[x] Video encolado: job_id=%s, file=%s", createdJob.Id, videoData.UniqueName)

	// 10. Responder inmediatamente con el job_id
	// NOTA: La limpieza de archivos locales la hace el WORKER después de procesar
	helpers.Success(c, http.StatusAccepted, gin.H{
		"job_id":  createdJob.Id,
		"status":  createdJob.Status,
		"message": "Video en cola de procesamiento. Consulta GET /jobs/" + createdJob.Id,
	})
}

// UpdateVideoRequest validates the fields for updating a video
type UpdateVideoRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateVideo godoc
// @Summary		Update a video's metadata
// @Description	Update title and description of a video. Only the owner can update.
// @Tags		streaming
// @Accept		json
// @Produce		json
// @Param		videoid path string true "Video ID"
// @Param		body body UpdateVideoRequest true "Updated video data"
// @Success		200 {object} models.VideoSwagger{}
// @Failure		400 {object} map[string]string
// @Failure		403 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/streaming/{videoid} [put]
func (vc *VideoControllerImpl) UpdateVideo(c *gin.Context) {
	videoId := c.Param("videoid")

	user, exists := c.Get("user")
	if !exists {
		helpers.HandleError(c, http.StatusInternalServerError, "User not found in context", nil)
		return
	}
	authenticatedUser := user.(*models.User)

	video, err := vc.databaseVideoService.FindVideoByID(videoId)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "Video not found", err)
		return
	}

	if video.UserID != authenticatedUser.Id {
		helpers.HandleError(c, http.StatusForbidden, "You are not the owner of this video", nil)
		return
	}

	var req UpdateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	video.Title = req.Title
	video.Description = req.Description

	updated, err := vc.databaseVideoService.UpdateVideo(video)
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not update video", err)
		return
	}

	helpers.Success(c, http.StatusOK, updated)
}

// DeleteVideo godoc
// @Summary		Delete a video
// @Description	Delete a video by ID. Only the owner can delete.
// @Tags		streaming
// @Produce		json
// @Param		videoid path string true "Video ID"
// @Success		200 {object} map[string]string
// @Failure		403 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/streaming/{videoid} [delete]
func (vc *VideoControllerImpl) DeleteVideo(c *gin.Context) {
	videoId := c.Param("videoid")

	user, exists := c.Get("user")
	if !exists {
		helpers.HandleError(c, http.StatusInternalServerError, "User not found in context", nil)
		return
	}
	authenticatedUser := user.(*models.User)

	video, err := vc.databaseVideoService.FindVideoByID(videoId)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "Video not found", err)
		return
	}

	if video.UserID != authenticatedUser.Id {
		helpers.HandleError(c, http.StatusForbidden, "You are not the owner of this video", nil)
		return
	}

	if err := vc.databaseVideoService.DeleteVideo(videoId); err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not delete video", err)
		return
	}

	helpers.Success(c, http.StatusOK, gin.H{"message": "Video deleted successfully"})
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
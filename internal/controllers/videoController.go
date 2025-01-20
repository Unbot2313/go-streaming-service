package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type VideoController interface {
	GetLatestVideos(c *gin.Context)
	CreateVideo(c *gin.Context)
	GetVideoByID(c *gin.Context)
	IncrementViews(c *gin.Context)
}

// SaveVideo		godoc
// @Summary 		Save a video
// @Description 	Upload a video file along with metadata (title and description) and save it to the AWS bucket.
// @Tags 			streaming
// @Produce 		json
// @Success 		200 {object} models.VideoSwagger{}
// @Failure 		400 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/streaming/ [get]
func (vc *VideoControllerImpl) GetLatestVideos(c *gin.Context) {
	videos, err := vc.databaseVideoService.FindLatestVideos()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, videos)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, video)
	
}

// SaveVideo		godoc
// @Summary 		Save a video
// @Description 	Upload a video file along with metadata (title and description) and save to the AWS bucket.
// @Tags 			streaming
// @Accept 			multipart/form-data
// @Produce 		json
// @Param 			title formData string true "Video Title"
// @Param 			description formData string false "Video Description"
// @Param 			video formData file true "Video File"
// @Success 		200 {object} models.VideoSwagger{}
// @Failure 		400 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/streaming/upload [post]
func (vc *VideoControllerImpl) CreateVideo(c *gin.Context) {

	// Recuperar el usuario del contexto
	user, exists := c.Get("user")
	if !exists {
		c.JSON(500, gin.H{"error": "User not found in context"})
		return
	}

	// Convertir a tipo User
	authenticatedUser, ok := user.(*models.User)
	if !ok {
		c.JSON(500, gin.H{"error": "Failed to parse user data"})
		return
	}

	// verificar si el archivo es válido
	if !vc.videoService.IsValidVideoExtension(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo no es un tipo de video válido."})
		return
	}

	fileSize := c.Request.ContentLength
	const maxFileSize = 100 * 1024 * 1024 // 100 MB
	if fileSize > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo excede el límite de tamaño permitido."})
		return
	}

	// guardar archivo en local
	videoData, err := vc.videoService.SaveVideo(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// borrar el archivo original
	defer vc.videoService.GetFilesService().RemoveFile(videoData.LocalPath)

	// comprimir el video
	// pendiente

	//pasar a archivos .ts y .m3u8 con ffmpeg y guardarlo en local
	filesPath, err := vc.videoService.FormatVideo(videoData.UniqueName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// borrar archivos locales .ts y .m3u8
	defer vc.videoService.GetFilesService().RemoveFolder(filesPath)

	// generar miniatura del segundo 1 del video
	_, err = services.SaveThumbnail(videoData.LocalPath, filesPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// subir el video a s3
	savedDataInS3, baseFolder, err := vc.videoService.UploadFilesFromFolderToS3(filesPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	videoData.M3u8FileURL = savedDataInS3.M3u8FileURL
	videoData.ThumbnailURL = savedDataInS3.ThumbnailURL
	
	// finalmente, guardar la url del video en la base de datos
	Video, err := vc.databaseVideoService.CreateVideo(videoData, authenticatedUser.Id)
	if err != nil {

		// como el video no se guardó en la base de datos, se debe borrar de s3
		// como folder/
		defer vc.videoService.DeleteS3Folder(baseFolder + "/")

		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, Video)
}

type VideoControllerImpl struct {
	videoService services.VideoService;
	databaseVideoService services.DatabaseVideoService
}

func NewVideoController(videoService services.VideoService, databaseVideoService services.DatabaseVideoService) VideoController {
	return &VideoControllerImpl{
		videoService: videoService,
		databaseVideoService: databaseVideoService,
	}
}
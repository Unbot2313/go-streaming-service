package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services/storage"
)

var (
	rawVideoPathFromWSL = "./static/videos/"
	saveFormatedVideoPath = "./static/temp/"
)

var validVideoExtensions = []string{
	".mp4", ".webm", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".3gp",
}

type VideoService interface {
	SaveVideo(ctx context.Context, c *gin.Context) (*models.Video, error)
	FormatVideo(ctx context.Context, videoName string) (string, error)
	UploadFolder(ctx context.Context, folder string) (storage.UploadResult, error)
	DeleteFolder(ctx context.Context, folderName string) error
	GenerateThumbnail(ctx context.Context, videoPath, outputDir string) (string, error)
	GetFilesService() FilesService
	IsValidVideoExtension(c *gin.Context) bool
}


func (vs *videoServiceImp) IsValidVideoExtension(c *gin.Context) bool {

	// Intentar obtener el archivo del request
	file, err := c.FormFile("video")
	if err != nil {
		return false // El archivo no existe o hubo un error
	}

	// Obtener la extensión del archivo en minúsculas
	extension := strings.ToLower(filepath.Ext(file.Filename))

	// Verificar si la extensión es válida
	for _, validExtension := range validVideoExtensions {
		if validExtension == extension {
			return true
		}
	}
	return false
}

func (vs *videoServiceImp) GetFilesService() FilesService {
	return vs.FilesService
}

func (vs *videoServiceImp) SaveVideo(ctx context.Context, c *gin.Context) (*models.Video, error) {
	if err := vs.FilesService.EnsureDir("static/videos"); err != nil {
		return nil, err
	}

	cfg := config.GetConfig()

	// 1. Obtener los campos de texto del formulario
	title := c.PostForm("title")
	description := c.PostForm("description")

	// 2. Obtener el archivo del formulario
	header, err := c.FormFile("video")
	if err != nil {
		return nil, fmt.Errorf("error al obtener el archivo: %w", err)
	}

	storagePath := cfg.LocalStoragePath
	id := uuid.New().String()
	uniqueName := fmt.Sprintf("%s_%s", id, header.Filename)

	// Guardar el archivo directamente con Gin
	savePath := filepath.Join(storagePath, uniqueName)
	if err := c.SaveUploadedFile(header, savePath); err != nil {
		return nil, fmt.Errorf("error al guardar el archivo: %w", err)
	}

	// Obtener la duración del video usando FFmpegService
	duration, err := vs.FFmpegService.ExtractDuration(ctx, savePath)
	if err != nil {
		return nil, fmt.Errorf("error al obtener la duración del video: %w", err)
	}

	videoData := &models.Video{
		Id:          id,
		Title:       title,
		Description: description,
		Video:       header.Filename,
		LocalPath:   savePath,
		UniqueName:  uniqueName,
		Duration:    duration,
	}

	return videoData, nil
}

func (vs *videoServiceImp) FormatVideo(ctx context.Context, videoName string) (string, error) {
	// Obtener el nombre del video sin la extensión
	stringName := strings.Split(videoName, ".")

	// Crear la carpeta donde se guardará el video formateado
	outputDir := saveFormatedVideoPath + stringName[0]
	err := vs.FilesService.CreateFolder("static/temp/" + stringName[0])
	if err != nil {
		return "", fmt.Errorf("error al crear la carpeta: %w", err)
	}

	videoPath := rawVideoPathFromWSL + videoName

	// Usar FFmpegService para convertir a HLS
	return vs.FFmpegService.ConvertToHLS(ctx, videoPath, outputDir)
}

func NewVideoService(storageService storage.StorageService, filesService FilesService, ffmpegService FFmpegService) VideoService {
	return &videoServiceImp{
		StorageService: storageService,
		FilesService:   filesService,
		FFmpegService:  ffmpegService,
	}
}

type videoServiceImp struct {
	StorageService storage.StorageService
	FilesService   FilesService
	FFmpegService  FFmpegService
}

// UploadFolder delega al StorageService para subir archivos
func (vs *videoServiceImp) UploadFolder(ctx context.Context, folder string) (storage.UploadResult, error) {
	return vs.StorageService.UploadFolder(ctx, folder)
}

// DeleteFolder delega al StorageService para eliminar archivos
func (vs *videoServiceImp) DeleteFolder(ctx context.Context, folderName string) error {
	return vs.StorageService.DeleteFolder(ctx, folderName)
}

// GenerateThumbnail delega al FFmpegService para generar miniatura
func (vs *videoServiceImp) GenerateThumbnail(ctx context.Context, videoPath, outputDir string) (string, error) {
	return vs.FFmpegService.GenerateThumbnail(ctx, videoPath, outputDir)
}


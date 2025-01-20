package services

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
)

var (
	rawVideoPathFromWSL = "./static/videos/"
	saveFormatedVideoPath = "./static/temp/"
)

var validVideoExtensions = []string{
	".mp4", ".webm", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".3gp",
}

type VideoService interface {
	SaveVideo(c *gin.Context) (*models.Video, error)
	FormatVideo(videoName string) (string, error) 
	UploadFilesFromFolderToS3(folder string) (importantFiles, string, error)
	DeleteS3Folder(folderName string) error
	GetFilesService() FilesService // Nuevo método para acceder a FilesService
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

func (vs *videoServiceImp) SaveVideo(c *gin.Context) (*models.Video, error) {
	if err := vs.FilesService.EnsureDir("static/videos"); err != nil {
		return nil, err
	}

	config := config.GetConfig()

	// 1. Obtener los campos de texto del formulario
	title := c.PostForm("title")
	description := c.PostForm("description")

	// 2. Obtener el archivo del formulario
	header, err := c.FormFile("video")

	if err != nil {
		return nil, fmt.Errorf("error al obtener el archivo: %w", err)
	}

	storagePath := config.LocalStoragePath

	uuid := uuid.New().String()

	uniqueName := fmt.Sprintf("%s_%s", uuid, header.Filename)

	// Guardar el archivo directamente con Gin
	savePath := filepath.Join(storagePath, uniqueName)
	if err := c.SaveUploadedFile(header, savePath); err != nil {
		return nil, fmt.Errorf("error al guardar el archivo: %w", err)
	}

	// Obtener la duración del video
	duration, err := getVideoDuration(savePath)
	if err != nil {
		return nil, fmt.Errorf("error al obtener la duración del video: %w", err)
	}

	videoData := &models.Video{
		Id: 			uuid,
		Title:    		title,
		Description:	description,
		Video: 	 		header.Filename,
		LocalPath: 	 	savePath,
		UniqueName: 	uniqueName,
		Duration: 		duration,
	}

	return videoData, nil
}

func (vs *videoServiceImp) FormatVideo(VideoName string) (string, error) {

	//obtener el nombre del video sin la extensión
	stringName := strings.Split(VideoName, ".")

	//crear la carpeta donde se guardará el video formateado
	err := vs.FilesService.CreateFolder("static/temp/" + stringName[0])

	if err != nil {
		return "", fmt.Errorf("error al crear la carpeta: %w", err)
	}

	saveFormatedPath := saveFormatedVideoPath + stringName[0] + "/output.m3u8"

	videoPath := rawVideoPathFromWSL + VideoName

	// ejecutar el comando ffmpeg para fragmentar el video y guardarlo en la carpeta ya creada para despues subirlo a s3
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-c", "copy", "-start_number", "0", "-hls_time", "10", "-hls_list_size", "0", "-f", "hls", saveFormatedPath)

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error al ejecutar el comando ffmpeg: %w", err)
	}

	ffmpegFilesPath := saveFormatedVideoPath + stringName[0]
	
	return ffmpegFilesPath, nil

}

func NewVideoService(S3Configuration S3Configuration, filesService FilesService) VideoService {
	return &videoServiceImp{
		S3configuration: S3Configuration,
		FilesService: filesService,
	}
}

type videoServiceImp struct{
	S3configuration S3Configuration
	FilesService FilesService
}

// Función para obtener la duración del video usando go-ffprobe
type FFProbeOutput struct {
    Format struct {
        Duration string `json:"duration"`
    } `json:"format"`
}

func formatDuration(seconds float64) string {
    // Si es menos de 60 segundos, retornar solo segundos
    if seconds < 60 {
        return fmt.Sprintf("%.0fs", seconds)
    }
    
    // Calcular minutos y segundos
    minutes := math.Floor(seconds / 60)
    remainingSeconds := math.Round(seconds - (minutes * 60))
    
    return fmt.Sprintf("%.0f:%.0f", minutes, remainingSeconds)
}

func getVideoDuration(videoPath string) (string, error) {
    // Construir el comando ffprobe
    cmd := exec.Command("ffprobe",
        "-v", "quiet",
        "-print_format", "json",
        "-show_format",
        videoPath)

    // Ejecutar el comando y obtener la salida
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("error ejecutando ffprobe: %v", err)
    }

    // Parsear la salida JSON
    var ffprobeOutput FFProbeOutput
    if err := json.Unmarshal(output, &ffprobeOutput); err != nil {
        return "", fmt.Errorf("error parseando la salida de ffprobe: %v", err)
    }

    // Convertir la duración a float64
    seconds, err := strconv.ParseFloat(ffprobeOutput.Format.Duration, 64)
    if err != nil {
        return "", fmt.Errorf("error convirtiendo la duración a número: %v", err)
    }

    return formatDuration(seconds), nil
}

func SaveThumbnail(videoPath string, folderPath string) (string, error) {
    thumbnailName := fmt.Sprintf("%s.webp", "thumbnail")
    thumbnailPath := filepath.Join(folderPath, thumbnailName)

    // El cambio principal está en mover -ss antes de -i
    cmd := exec.Command("ffmpeg",
        "-ss", "00:00:08",      // MOVIDO: colocar antes de -i
        "-i", videoPath,         // archivo de entrada
        "-frames:v", "1",        // usar frames:v en lugar de vframes
        "-vf", "scale=480:-1",   
        "-y",                   
        thumbnailPath,          
    )

    if output, err := cmd.CombinedOutput(); err != nil {
        return "", fmt.Errorf("error generando miniatura: %w, output: %s", err, string(output))
    }

    return thumbnailPath, nil
}
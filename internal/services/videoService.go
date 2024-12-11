package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

type videoServiceImp struct{
	S3configuration S3Configuration
}

type VideoService interface {
	GetVideos()
	SaveVideo(c *gin.Context) (*models.Video, error)
	FormatVideo(videoName string) (string, error) 
	ensureDir(dirName string) error
	UploadFilesFromFolderToS3(folder string) ([]string, error)

}

func NewVideoService(S3Configuration S3Configuration) VideoService {
	return &videoServiceImp{S3configuration: S3Configuration}
}

func (vs *videoServiceImp) GetVideos() {
	fmt.Println("GetVideos")
}

func (vs *videoServiceImp) SaveVideo(c *gin.Context) (*models.Video, error) {
	if err := vs.ensureDir("static/videos"); err != nil {
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
	uniqueName := fmt.Sprintf("%s_%s", uuid.New().String(), header.Filename)

	// Guardar el archivo directamente con Gin
	savePath := filepath.Join(storagePath, uniqueName)
	if err := c.SaveUploadedFile(header, savePath); err != nil {
		return nil, fmt.Errorf("error al guardar el archivo: %w", err)
	}

	videoData := &models.Video{
		Title:    		title,
		Description:	description,
		Video: 	 		header.Filename,
		LocalPath: 	 	savePath,
		UniqueName: 	uniqueName,
	}

	return videoData, nil

}

func (vs *videoServiceImp) FormatVideo(VideoName string) (string, error) {

	//obtener el nombre del video sin la extensión
	stringName := strings.Split(VideoName, ".")

	//crear la carpeta donde se guardará el video formateado
	err := createFolder("static/temp/" + stringName[0])

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


func (vs *videoServiceImp) ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error al crear directorio: %w", err)
	}

	return nil
}

func createFolder(path string) error {
	// Crea la carpeta y sus carpetas padres si no existen
	err := os.MkdirAll(path, os.ModePerm) // os.ModePerm otorga permisos de lectura, escritura y ejecución
	if err != nil {
		return fmt.Errorf("error al crear la carpeta: %w", err)
	}
	return nil
}
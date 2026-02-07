package services

import (
	"fmt"
	"log/slog"
	"os"
)

type filesService struct{}

type FilesService interface {
	EnsureDir(dirName string) error
	CreateFolder(path string) error
	RemoveFolder(folder string) error
	RemoveFile(filePath string) error
}

func NewFilesService() FilesService {
	return &filesService{}
}

func (fs *filesService) EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error al crear directorio: %w", err)
	}

	return nil
}

func (fs *filesService) RemoveFile(filePath string) error {
	err := os.Remove(filePath)

	if err != nil {
		return fmt.Errorf("error al borrar archivo: %w", err)
	}

	return nil
}


func (fs *filesService) CreateFolder(path string) error {
	// Crea la carpeta y sus carpetas padres si no existen
	err := os.MkdirAll(path, os.ModePerm) // os.ModePerm otorga permisos de lectura, escritura y ejecución
	if err != nil {
		return fmt.Errorf("error al crear la carpeta: %w", err)
	}
	return nil
}

func (fs *filesService) RemoveFolder(folder string) error {
	err := os.RemoveAll(folder)
	if err != nil {
		slog.Error("error removing folder", slog.Any("error", err))
		return err
	}
	return nil
}
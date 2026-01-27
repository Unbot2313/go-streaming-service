package storage

import (
	"context"

	"github.com/unbot2313/go-streaming-service/config"
)

// UploadResult contiene las URLs de los archivos importantes después de subir
type UploadResult struct {
	M3u8FileURL  string
	ThumbnailURL string
	BaseFolder   string
}

// ObjectInfo representa información básica de un objeto en storage
type ObjectInfo struct {
	Key  string
	Size int64
}

// StorageService define la interfaz genérica para operaciones de almacenamiento
type StorageService interface {
	// UploadFolder sube todos los archivos de una carpeta local al storage
	UploadFolder(ctx context.Context, localFolder string) (UploadResult, error)

	// DeleteFolder elimina todos los objetos dentro de una carpeta en el storage
	DeleteFolder(ctx context.Context, folderName string) error

	// ListObjects lista los objetos dentro de una carpeta
	ListObjects(ctx context.Context, folder string) ([]ObjectInfo, error)
}

// NewStorageService crea una instancia del servicio de storage según la configuración
func NewStorageService() StorageService {
	cfg := config.GetConfig()

	if cfg.StorageType == "minio" {
		return NewMinIOStorage()
	}

	return NewS3Storage()
}

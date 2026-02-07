package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/unbot2313/go-streaming-service/config"
)

// MinIOStorage implementa StorageService para MinIO
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
}

// NewMinIOStorage crea una nueva instancia de MinIOStorage
func NewMinIOStorage() StorageService {
	cfg := config.GetConfig()

	return &MinIOStorage{
		client:     config.GetMinIOClient(),
		bucketName: cfg.MinIOBucketName,
		endpoint:   cfg.MinIOEndpoint,
	}
}

// UploadFolder sube todos los archivos de una carpeta a MinIO
func (m *MinIOStorage) UploadFolder(ctx context.Context, localFolder string) (UploadResult, error) {
	baseFolder := filepath.Base(localFolder)

	var m3u8FileURL string
	var thumbnailURL string

	files, err := os.ReadDir(localFolder)
	if err != nil {
		return UploadResult{BaseFolder: baseFolder}, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(localFolder, file.Name())
		objectName := filepath.Join(baseFolder, file.Name())

		// Detectar content type
		contentType := "application/octet-stream"
		if strings.HasSuffix(file.Name(), ".m3u8") {
			contentType = "application/x-mpegURL"
		} else if strings.HasSuffix(file.Name(), ".ts") {
			contentType = "video/MP2T"
		} else if strings.HasSuffix(file.Name(), ".webp") {
			contentType = "image/webp"
		}

		_, err := m.client.FPutObject(
			ctx,
			m.bucketName,
			objectName,
			filePath,
			minio.PutObjectOptions{ContentType: contentType},
		)

		if err != nil {
			return UploadResult{BaseFolder: baseFolder}, fmt.Errorf("error subiendo %s a MinIO: %w", file.Name(), err)
		}

		// Construir URL del archivo
		fileURL := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucketName, objectName)

		if strings.HasSuffix(file.Name(), ".m3u8") {
			m3u8FileURL = fileURL
		}

		if strings.HasSuffix(file.Name(), ".webp") {
			thumbnailURL = fileURL
		}

		slog.Info("MinIO uploaded", slog.String("object", objectName))
	}

	if m3u8FileURL == "" {
		return UploadResult{BaseFolder: baseFolder}, errors.New("no se encontró el archivo .m3u8")
	}

	return UploadResult{
		M3u8FileURL:  m3u8FileURL,
		ThumbnailURL: thumbnailURL,
		BaseFolder:   baseFolder,
	}, nil
}

// DeleteFolder elimina todos los objetos dentro de una carpeta en MinIO
func (m *MinIOStorage) DeleteFolder(ctx context.Context, folderName string) error {
	slog.Info("MinIO deleting objects", slog.String("folder", folderName))

	objectsCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    folderName,
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			return fmt.Errorf("error listando objetos: %w", object.Err)
		}

		err := m.client.RemoveObject(ctx, m.bucketName, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("error eliminando %s: %w", object.Key, err)
		}

		slog.Info("MinIO deleted", slog.String("object", object.Key))
	}

	return nil
}

// ListObjects lista los objetos dentro de una carpeta en MinIO
func (m *MinIOStorage) ListObjects(ctx context.Context, folder string) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	objectsCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    folder,
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listando objetos: %w", object.Err)
		}

		objects = append(objects, ObjectInfo{
			Key:  object.Key,
			Size: object.Size,
		})
	}

	return objects, nil
}

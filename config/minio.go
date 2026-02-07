package config

import (
	"log/slog"
	"os"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	minioClient     *minio.Client
	minioClientOnce sync.Once
)

// GetMinIOClient retorna una instancia singleton del cliente MinIO
func GetMinIOClient() *minio.Client {
	minioClientOnce.Do(func() {
		cfg := GetConfig()

		client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
			Secure: false, // true para HTTPS, false para HTTP (desarrollo local)
		})

		if err != nil {
			slog.Error("error creating MinIO client", slog.Any("error", err))
			os.Exit(1)
		}

		minioClient = client
		slog.Info("MinIO client connected", slog.String("endpoint", cfg.MinIOEndpoint))
	})

	return minioClient
}

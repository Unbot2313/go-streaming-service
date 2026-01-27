package config

import (
	"log"
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
			log.Fatalf("Error creando cliente MinIO: %v", err)
		}

		minioClient = client
		log.Printf("Cliente MinIO conectado a %s", cfg.MinIOEndpoint)
	})

	return minioClient
}

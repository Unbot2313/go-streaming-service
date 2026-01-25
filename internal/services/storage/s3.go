package storage

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/unbot2313/go-streaming-service/config"
)

// S3Storage implementa StorageService para AWS S3
type S3Storage struct {
	client     *s3.Client
	uploader   *manager.Uploader
	bucketName string
	region     string
}

// NewS3Storage crea una nueva instancia de S3Storage
func NewS3Storage() StorageService {
	cfg := config.GetConfig()

	return &S3Storage{
		client:     config.GetS3Client(),
		uploader:   config.GetS3Uploader(),
		bucketName: cfg.AWSBucketName,
		region:     cfg.AWSRegion,
	}
}

// UploadFolder sube todos los archivos de una carpeta a S3
func (s *S3Storage) UploadFolder(localFolder string) (UploadResult, error) {
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
		f, err := os.Open(filePath)
		if err != nil {
			return UploadResult{BaseFolder: baseFolder}, err
		}
		defer f.Close()

		key := filepath.Join(baseFolder, file.Name())

		result, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(key),
			Body:   f,
		})

		if err != nil {
			return UploadResult{BaseFolder: baseFolder}, err
		}

		if strings.HasSuffix(file.Name(), ".m3u8") {
			m3u8FileURL = result.Location
		}

		if strings.HasSuffix(file.Name(), ".webp") {
			thumbnailURL = result.Location
		}
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

// DeleteFolder elimina todos los objetos dentro de una carpeta en S3
func (s *S3Storage) DeleteFolder(folderName string) error {
	ctx := context.Background()

	log.Println("Eliminando objetos en la carpeta:", folderName)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(folderName),
	}

	objectPaginator := s3.NewListObjectsV2Paginator(s.client, input)
	var objectsToDelete []types.ObjectIdentifier

	for objectPaginator.HasMorePages() {
		output, err := objectPaginator.NextPage(ctx)
		if err != nil {
			log.Printf("Error al listar objetos: %v\n", err)
			return err
		}

		for _, object := range output.Contents {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
				Key: object.Key,
			})
		}
	}

	if len(objectsToDelete) == 0 {
		log.Println("No se encontraron objetos para eliminar.")
		return nil
	}

	deleteInput := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucketName),
		Delete: &types.Delete{
			Objects: objectsToDelete,
		},
	}

	_, err := s.client.DeleteObjects(ctx, deleteInput)
	if err != nil {
		log.Printf("Error al eliminar objetos: %v\n", err)
		return err
	}

	log.Printf("Se han eliminado los objetos en la carpeta %v.\n", folderName)
	return nil
}

// ListObjects lista los objetos dentro de una carpeta en S3
func (s *S3Storage) ListObjects(ctx context.Context, folder string) ([]ObjectInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(folder),
	}

	var objects []ObjectInfo
	objectPaginator := s3.NewListObjectsV2Paginator(s.client, input)

	for objectPaginator.HasMorePages() {
		output, err := objectPaginator.NextPage(ctx)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", s.bucketName)
			}
			return nil, err
		}

		for _, obj := range output.Contents {
			objects = append(objects, ObjectInfo{
				Key:  *obj.Key,
				Size: *obj.Size,
			})
		}
	}

	return objects, nil
}

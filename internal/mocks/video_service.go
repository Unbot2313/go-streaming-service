package mocks

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
	"github.com/unbot2313/go-streaming-service/internal/services/storage"
)

type MockVideoService struct {
	SaveVideoFn             func(ctx context.Context, c *gin.Context) (*models.Video, error)
	FormatVideoFn           func(ctx context.Context, videoName string) (string, error)
	UploadFolderFn          func(ctx context.Context, folder string) (storage.UploadResult, error)
	DeleteFolderFn          func(ctx context.Context, folderName string) error
	GenerateThumbnailFn     func(ctx context.Context, videoPath, outputDir string) (string, error)
	GetFilesServiceFn       func() services.FilesService
	IsValidVideoExtensionFn func(c *gin.Context) bool
}

func (m *MockVideoService) SaveVideo(ctx context.Context, c *gin.Context) (*models.Video, error) {
	return m.SaveVideoFn(ctx, c)
}

func (m *MockVideoService) FormatVideo(ctx context.Context, videoName string) (string, error) {
	return m.FormatVideoFn(ctx, videoName)
}

func (m *MockVideoService) UploadFolder(ctx context.Context, folder string) (storage.UploadResult, error) {
	return m.UploadFolderFn(ctx, folder)
}

func (m *MockVideoService) DeleteFolder(ctx context.Context, folderName string) error {
	return m.DeleteFolderFn(ctx, folderName)
}

func (m *MockVideoService) GenerateThumbnail(ctx context.Context, videoPath, outputDir string) (string, error) {
	return m.GenerateThumbnailFn(ctx, videoPath, outputDir)
}

func (m *MockVideoService) GetFilesService() services.FilesService {
	return m.GetFilesServiceFn()
}

func (m *MockVideoService) IsValidVideoExtension(c *gin.Context) bool {
	return m.IsValidVideoExtensionFn(c)
}

package mocks

import (
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type MockDatabaseVideoService struct {
	FindLatestVideosFn func(page, pageSize int) (*services.PaginatedVideos, error)
	FindVideoByIDFn    func(videoId string) (*models.VideoModel, error)
	IncrementViewsFn   func(videoId string) (*models.VideoModel, error)
	FindUserVideosFn   func(userId string) ([]*models.VideoModel, error)
	CreateVideoFn      func(video *models.Video, userId string) (*models.VideoModel, error)
	UpdateVideoFn      func(video *models.VideoModel) (*models.VideoModel, error)
	DeleteVideoFn      func(videoId string) error
}

func (m *MockDatabaseVideoService) FindLatestVideos(page, pageSize int) (*services.PaginatedVideos, error) {
	return m.FindLatestVideosFn(page, pageSize)
}

func (m *MockDatabaseVideoService) FindVideoByID(videoId string) (*models.VideoModel, error) {
	return m.FindVideoByIDFn(videoId)
}

func (m *MockDatabaseVideoService) IncrementViews(videoId string) (*models.VideoModel, error) {
	return m.IncrementViewsFn(videoId)
}

func (m *MockDatabaseVideoService) FindUserVideos(userId string) ([]*models.VideoModel, error) {
	return m.FindUserVideosFn(userId)
}

func (m *MockDatabaseVideoService) CreateVideo(video *models.Video, userId string) (*models.VideoModel, error) {
	return m.CreateVideoFn(video, userId)
}

func (m *MockDatabaseVideoService) UpdateVideo(video *models.VideoModel) (*models.VideoModel, error) {
	return m.UpdateVideoFn(video)
}

func (m *MockDatabaseVideoService) DeleteVideo(videoId string) error {
	return m.DeleteVideoFn(videoId)
}

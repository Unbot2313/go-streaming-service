package mocks

import (
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type MockTagService struct {
	GetAllTagsFn         func() ([]models.Tag, error)
	FindVideosByTagFn    func(tagName string, page, pageSize int) (*services.PaginatedVideos, error)
	AddTagsToVideoFn     func(videoId string, tagNames []string) error
	RemoveTagFromVideoFn func(videoId string, tagName string) error
}

func (m *MockTagService) GetAllTags() ([]models.Tag, error) {
	return m.GetAllTagsFn()
}

func (m *MockTagService) FindVideosByTag(tagName string, page, pageSize int) (*services.PaginatedVideos, error) {
	return m.FindVideosByTagFn(tagName, page, pageSize)
}

func (m *MockTagService) AddTagsToVideo(videoId string, tagNames []string) error {
	return m.AddTagsToVideoFn(videoId, tagNames)
}

func (m *MockTagService) RemoveTagFromVideo(videoId string, tagName string) error {
	return m.RemoveTagFromVideoFn(videoId, tagName)
}

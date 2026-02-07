package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"gorm.io/gorm"
)

type tagServiceImp struct{}

type TagService interface {
	GetAllTags() ([]models.Tag, error)
	FindVideosByTag(tagName string, page, pageSize int) (*PaginatedVideos, error)
	AddTagsToVideo(videoId string, tagNames []string) error
	RemoveTagFromVideo(videoId string, tagName string) error
}

func NewTagService() TagService {
	return &tagServiceImp{}
}

func (s *tagServiceImp) GetAllTags() ([]models.Tag, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var tags []models.Tag
	if err := db.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (s *tagServiceImp) FindVideosByTag(tagName string, page, pageSize int) (*PaginatedVideos, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var tag models.Tag
	if err := db.Where("name = ?", tagName).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tag '%s' not found", tagName)
		}
		return nil, err
	}

	var total int64
	if err := db.Model(&models.VideoModel{}).
		Joins("JOIN video_tags ON video_tags.video_model_id = videos.id").
		Where("video_tags.tag_id = ?", tag.Id).
		Count(&total).Error; err != nil {
		return nil, err
	}

	var videos []*models.VideoModel
	offset := (page - 1) * pageSize

	if err := db.Preload("Tags").
		Joins("JOIN video_tags ON video_tags.video_model_id = videos.id").
		Where("video_tags.tag_id = ?", tag.Id).
		Order("videos.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&videos).Error; err != nil {
		return nil, err
	}

	return &PaginatedVideos{
		Data:     videos,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *tagServiceImp) AddTagsToVideo(videoId string, tagNames []string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	var video models.VideoModel
	if err := db.Where("id = ?", videoId).First(&video).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("video with id %s not found", videoId)
		}
		return err
	}

	var tags []models.Tag
	for _, name := range tagNames {
		var tag models.Tag
		result := db.Where("name = ?", name).First(&tag)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			tag = models.Tag{Id: uuid.New().String(), Name: name}
			if err := db.Create(&tag).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		}
		tags = append(tags, tag)
	}

	return db.Model(&video).Association("Tags").Append(tags)
}

func (s *tagServiceImp) RemoveTagFromVideo(videoId string, tagName string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	var video models.VideoModel
	if err := db.Where("id = ?", videoId).First(&video).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("video with id %s not found", videoId)
		}
		return err
	}

	var tag models.Tag
	if err := db.Where("name = ?", tagName).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("tag '%s' not found", tagName)
		}
		return err
	}

	return db.Model(&video).Association("Tags").Delete(&tag)
}

package services

import (
	"errors"
	"fmt"

	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"gorm.io/gorm"
)

type databaseVideoService struct{}

type PaginatedVideos struct {
	Data     []*models.VideoModel `json:"data"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	Total    int64                `json:"total"`
}

type DatabaseVideoService interface {
	FindLatestVideos(page, pageSize int) (*PaginatedVideos, error)
	FindVideoByID(videoId string) (*models.VideoModel, error)
	IncrementViews(videoId string) (*models.VideoModel, error)
	FindUserVideos(userId string) ([]*models.VideoModel, error)
	CreateVideo(video *models.Video, userId string) (*models.VideoModel, error)
	UpdateVideo(video *models.VideoModel) (*models.VideoModel, error)
	DeleteVideo(videoId string) error
	SearchVideos(query string, page, pageSize int) (*PaginatedVideos, error)
}

func NewDatabaseVideoService() DatabaseVideoService {
	return &databaseVideoService{}
}

func (service *databaseVideoService) FindLatestVideos(page, pageSize int) (*PaginatedVideos, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var total int64
	if err := db.Model(&models.VideoModel{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var videos []*models.VideoModel
	offset := (page - 1) * pageSize

	dbCtx := db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&videos)

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return &PaginatedVideos{
		Data:     videos,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (service *databaseVideoService) FindVideoByID(videoId string) (*models.VideoModel, error) {
	db, err := config.GetDB()

	if err != nil {
		return nil, err
	}

	var video models.VideoModel

	dbCtx := db.Where("id = ?", videoId).First(&video)

	if errors.Is(dbCtx.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("video with id %s not found", videoId)
	}

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return &video, nil
}

func (service *databaseVideoService) IncrementViews(videoId string) (*models.VideoModel, error) {
	db, err := config.GetDB()

	if err != nil {
		return nil, err
	}

	var video models.VideoModel

	dbCtx := db.Where("id = ?", videoId).First(&video)

	if errors.Is(dbCtx.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("video with id %s not found", videoId)
	}

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	video.Views++

	dbCtx = db.Save(&video)

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return &video, nil
}

func (service *databaseVideoService) FindUserVideos(userId string) ([]*models.VideoModel, error) {
	db, err := config.GetDB()

	if err != nil {
		return nil, err
	}

	var videos []*models.VideoModel

	dbCtx := db.Where(&models.VideoModel{UserID: userId}).Find(&videos)

	if errors.Is(dbCtx.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no videos found for user with id %s", userId)
	}

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return videos, nil
}

func (service *databaseVideoService) CreateVideo(videoData *models.Video, userId string) (*models.VideoModel, error) {

	Video := models.VideoModel{
		Id:           videoData.Id,
		Title:        videoData.Title,
		Description:  videoData.Description,
		UserID:       userId,
		VideoUrl:     videoData.M3u8FileURL,
		Duration:     videoData.Duration,
		ThumbnailURL: videoData.ThumbnailURL,
	}

	db, err := config.GetDB()

	if err != nil {
		return nil, err
	}

	dbCtx := db.Create(&Video)

	if errors.Is(dbCtx.Error, gorm.ErrDuplicatedKey) {
		return nil, fmt.Errorf("ya hay un video con el id %s", videoData.Id)
	}

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return &Video, nil
}

func (service *databaseVideoService) UpdateVideo(video *models.VideoModel) (*models.VideoModel, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var existing models.VideoModel
	if err := db.Where("id = ?", video.Id).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("video with id %s not found", video.Id)
		}
		return nil, err
	}

	if err := db.Model(&existing).Updates(map[string]interface{}{
		"title":       video.Title,
		"description": video.Description,
	}).Error; err != nil {
		return nil, err
	}

	return &existing, nil
}

func (service *databaseVideoService) DeleteVideo(videoId string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	result := db.Where("id = ?", videoId).Delete(&models.VideoModel{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("video with id %s not found", videoId)
	}

	return nil
}

func (service *databaseVideoService) SearchVideos(query string, page, pageSize int) (*PaginatedVideos, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	search := "%" + query + "%"

	var total int64
	if err := db.Model(&models.VideoModel{}).
		Where("title ILIKE ? OR description ILIKE ?", search, search).
		Count(&total).Error; err != nil {
		return nil, err
	}

	var videos []*models.VideoModel
	offset := (page - 1) * pageSize

	if err := db.Where("title ILIKE ? OR description ILIKE ?", search, search).
		Order("created_at DESC").
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

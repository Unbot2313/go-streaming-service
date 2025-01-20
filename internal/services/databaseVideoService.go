package services

import (
	"errors"
	"fmt"

	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"gorm.io/gorm"
)

func (service *databaseVideoService) FindLatestVideos() (*[]*models.VideoModel, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var videos []*models.VideoModel

	// Ordenar por CreatedAt en orden descendente
	dbCtx := db.Order("created_at DESC").Find(&videos)

	if errors.Is(dbCtx.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no videos found")
	}

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return &videos, nil
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
		Id: videoData.Id,
		Title: videoData.Title,
		Description: videoData.Description,
		UserID: userId,
		VideoUrl: videoData.M3u8FileURL,
		Duration: videoData.Duration,
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
	return &models.VideoModel{}, nil
}

func (service *databaseVideoService) DeleteVideo(videoId string) error {
	return nil
}


type databaseVideoService struct {}

type DatabaseVideoService interface {
	FindLatestVideos() (*[]*models.VideoModel, error)
	FindVideoByID(videoId string) (*models.VideoModel, error) 
	IncrementViews(videoId string) (*models.VideoModel, error)
	FindUserVideos(userId string) ([]*models.VideoModel, error)
	CreateVideo(video *models.Video, userId string) (*models.VideoModel, error)
	UpdateVideo(video *models.VideoModel) (*models.VideoModel, error)
	DeleteVideo(videoId string) error
}

func NewDatabaseVideoService() DatabaseVideoService {
	return &databaseVideoService{}
}

package services

import (
	"errors"
	"fmt"

	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"gorm.io/gorm"
)

type JobService interface {
	CreateJob(job *models.Job) (*models.JobModel, error)
	GetJobByID(jobId string) (*models.JobModel, error)
	UpdateJobStatus(jobId, status, errorMsg string) error
	UpdateJobCompleted(jobId, videoID string) error
}

type jobServiceImp struct{}

func NewJobService() JobService {
	return &jobServiceImp{}
}

func (service *jobServiceImp) CreateJob(job *models.Job) (*models.JobModel, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	jobModel := models.JobModel{
		Job: *job,
	}

	dbCtx := db.Create(&jobModel)

	if errors.Is(dbCtx.Error, gorm.ErrDuplicatedKey) {
		return nil, fmt.Errorf("ya existe un job con el id %s", job.Id)
	}

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return &jobModel, nil
}

func (service *jobServiceImp) GetJobByID(jobId string) (*models.JobModel, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var job models.JobModel

	dbCtx := db.Where("id = ?", jobId).First(&job)

	if errors.Is(dbCtx.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("job con id %s no encontrado", jobId)
	}

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return &job, nil
}

func (service *jobServiceImp) UpdateJobStatus(jobId, status, errorMsg string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	dbCtx := db.Model(&models.JobModel{}).Where("id = ?", jobId).Updates(updates)

	if dbCtx.Error != nil {
		return dbCtx.Error
	}

	if dbCtx.RowsAffected == 0 {
		return fmt.Errorf("job con id %s no encontrado", jobId)
	}

	return nil
}

func (service *jobServiceImp) UpdateJobCompleted(jobId, videoID string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	dbCtx := db.Model(&models.JobModel{}).Where("id = ?", jobId).Updates(map[string]interface{}{
		"status":   "completed",
		"video_id": videoID,
	})

	if dbCtx.Error != nil {
		return dbCtx.Error
	}

	if dbCtx.RowsAffected == 0 {
		return fmt.Errorf("job con id %s no encontrado", jobId)
	}

	return nil
}

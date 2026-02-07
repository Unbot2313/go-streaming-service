package mocks

import (
	"github.com/unbot2313/go-streaming-service/internal/models"
)

type MockJobService struct {
	CreateJobFn         func(job *models.Job) (*models.JobModel, error)
	GetJobByIDFn        func(jobId string) (*models.JobModel, error)
	UpdateJobStatusFn   func(jobId, status, errorMsg string) error
	UpdateJobCompletedFn func(jobId, videoID string) error
}

func (m *MockJobService) CreateJob(job *models.Job) (*models.JobModel, error) {
	return m.CreateJobFn(job)
}

func (m *MockJobService) GetJobByID(jobId string) (*models.JobModel, error) {
	return m.GetJobByIDFn(jobId)
}

func (m *MockJobService) UpdateJobStatus(jobId, status, errorMsg string) error {
	return m.UpdateJobStatusFn(jobId, status, errorMsg)
}

func (m *MockJobService) UpdateJobCompleted(jobId, videoID string) error {
	return m.UpdateJobCompletedFn(jobId, videoID)
}

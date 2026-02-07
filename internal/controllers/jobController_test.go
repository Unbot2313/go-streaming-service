package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/mocks"
)

func setupJobRouter(controller JobController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/jobs/:jobid", func(c *gin.Context) {
		// Simular usuario autenticado en el contexto
		c.Set("user", &models.User{Id: "user-123", Username: "testuser"})
		controller.GetJobByID(c)
	})
	return r
}

func TestGetJobByID_Success(t *testing.T) {
	mockJob := &mocks.MockJobService{
		GetJobByIDFn: func(jobId string) (*models.JobModel, error) {
			return &models.JobModel{
				Job: models.Job{Id: "job-123", UserID: "user-123", Status: "completed"},
			}, nil
		},
	}

	controller := NewJobController(mockJob)
	router := setupJobRouter(controller)

	req, _ := http.NewRequest("GET", "/jobs/job-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("expected success to be true")
	}
}

func TestGetJobByID_Forbidden(t *testing.T) {
	mockJob := &mocks.MockJobService{
		GetJobByIDFn: func(jobId string) (*models.JobModel, error) {
			return &models.JobModel{
				Job: models.Job{Id: "job-123", UserID: "other-user-456", Status: "completed"},
			}, nil
		},
	}

	controller := NewJobController(mockJob)
	router := setupJobRouter(controller)

	req, _ := http.NewRequest("GET", "/jobs/job-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestGetJobByID_NotFound(t *testing.T) {
	mockJob := &mocks.MockJobService{
		GetJobByIDFn: func(jobId string) (*models.JobModel, error) {
			return nil, errors.New("job not found")
		},
	}

	controller := NewJobController(mockJob)
	router := setupJobRouter(controller)

	req, _ := http.NewRequest("GET", "/jobs/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

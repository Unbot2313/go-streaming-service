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
	"github.com/unbot2313/go-streaming-service/internal/services"
)

func setupVideoRouter(controller VideoController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/streaming/latest", controller.GetLatestVideos)
	r.GET("/streaming/id/:videoid", controller.GetVideoByID)
	r.PATCH("/streaming/views/:videoid", controller.IncrementViews)
	return r
}

func TestGetLatestVideos_Success(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindLatestVideosFn: func(page, pageSize int) (*services.PaginatedVideos, error) {
			return &services.PaginatedVideos{
				Data:     []*models.VideoModel{{Title: "Test Video"}},
				Page:     1,
				PageSize: 10,
				Total:    1,
			}, nil
		},
	}

	controller := NewVideoController(nil, mockDBVideo, nil, nil)
	router := setupVideoRouter(controller)

	req, _ := http.NewRequest("GET", "/streaming/latest", nil)
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

func TestGetLatestVideos_WithPagination(t *testing.T) {
	var receivedPage, receivedPageSize int

	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindLatestVideosFn: func(page, pageSize int) (*services.PaginatedVideos, error) {
			receivedPage = page
			receivedPageSize = pageSize
			return &services.PaginatedVideos{
				Data:     []*models.VideoModel{},
				Page:     page,
				PageSize: pageSize,
				Total:    0,
			}, nil
		},
	}

	controller := NewVideoController(nil, mockDBVideo, nil, nil)
	router := setupVideoRouter(controller)

	req, _ := http.NewRequest("GET", "/streaming/latest?page=2&page_size=25", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if receivedPage != 2 {
		t.Errorf("expected page 2, got %d", receivedPage)
	}

	if receivedPageSize != 25 {
		t.Errorf("expected pageSize 25, got %d", receivedPageSize)
	}
}

func TestGetLatestVideos_PageSizeMax(t *testing.T) {
	var receivedPageSize int

	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindLatestVideosFn: func(page, pageSize int) (*services.PaginatedVideos, error) {
			receivedPageSize = pageSize
			return &services.PaginatedVideos{Data: []*models.VideoModel{}, Page: 1, PageSize: pageSize, Total: 0}, nil
		},
	}

	controller := NewVideoController(nil, mockDBVideo, nil, nil)
	router := setupVideoRouter(controller)

	req, _ := http.NewRequest("GET", "/streaming/latest?page_size=999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if receivedPageSize != 50 {
		t.Errorf("expected pageSize capped at 50, got %d", receivedPageSize)
	}
}

func TestGetVideoByID_Success(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindVideoByIDFn: func(videoId string) (*models.VideoModel, error) {
			return &models.VideoModel{Title: "Found Video"}, nil
		},
	}

	controller := NewVideoController(nil, mockDBVideo, nil, nil)
	router := setupVideoRouter(controller)

	req, _ := http.NewRequest("GET", "/streaming/id/video-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetVideoByID_NotFound(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindVideoByIDFn: func(videoId string) (*models.VideoModel, error) {
			return nil, errors.New("video not found")
		},
	}

	controller := NewVideoController(nil, mockDBVideo, nil, nil)
	router := setupVideoRouter(controller)

	req, _ := http.NewRequest("GET", "/streaming/id/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestIncrementViews_Success(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		IncrementViewsFn: func(videoId string) (*models.VideoModel, error) {
			return &models.VideoModel{Title: "Video", Views: 1}, nil
		},
	}

	controller := NewVideoController(nil, mockDBVideo, nil, nil)
	router := setupVideoRouter(controller)

	req, _ := http.NewRequest("PATCH", "/streaming/views/video-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestIncrementViews_Error(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		IncrementViewsFn: func(videoId string) (*models.VideoModel, error) {
			return nil, errors.New("database error")
		},
	}

	controller := NewVideoController(nil, mockDBVideo, nil, nil)
	router := setupVideoRouter(controller)

	req, _ := http.NewRequest("PATCH", "/streaming/views/video-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

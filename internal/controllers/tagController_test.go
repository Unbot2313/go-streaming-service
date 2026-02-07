package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/mocks"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

func setupTagRouter(controller TagController, mockDBVideo *mocks.MockDatabaseVideoService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Rutas públicas
	r.GET("/tags", controller.GetAllTags)
	r.GET("/tags/:tag/videos", controller.GetVideosByTag)

	// Rutas protegidas con usuario simulado
	protected := r.Group("")
	protected.Use(func(c *gin.Context) {
		c.Set("user", &models.User{Id: "user-123", Username: "testuser"})
		c.Next()
	})
	protected.POST("/tags/:videoid", controller.AddTagsToVideo)
	protected.DELETE("/tags/:videoid", controller.RemoveTagFromVideo)

	return r
}

func TestGetAllTags_Success(t *testing.T) {
	mockTag := &mocks.MockTagService{
		GetAllTagsFn: func() ([]models.Tag, error) {
			return []models.Tag{
				{Id: "1", Name: "golang"},
				{Id: "2", Name: "tutorial"},
			}, nil
		},
	}

	controller := NewTagController(mockTag, nil)
	router := setupTagRouter(controller, nil)

	req, _ := http.NewRequest("GET", "/tags", nil)
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

func TestGetVideosByTag_Success(t *testing.T) {
	mockTag := &mocks.MockTagService{
		FindVideosByTagFn: func(tagName string, page, pageSize int) (*services.PaginatedVideos, error) {
			return &services.PaginatedVideos{
				Data:     []*models.VideoModel{{Title: "Go Tutorial"}},
				Page:     1,
				PageSize: 10,
				Total:    1,
			}, nil
		},
	}

	controller := NewTagController(mockTag, nil)
	router := setupTagRouter(controller, nil)

	req, _ := http.NewRequest("GET", "/tags/golang/videos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetVideosByTag_NotFound(t *testing.T) {
	mockTag := &mocks.MockTagService{
		FindVideosByTagFn: func(tagName string, page, pageSize int) (*services.PaginatedVideos, error) {
			return nil, errors.New("tag 'nonexistent' not found")
		},
	}

	controller := NewTagController(mockTag, nil)
	router := setupTagRouter(controller, nil)

	req, _ := http.NewRequest("GET", "/tags/nonexistent/videos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestAddTagsToVideo_Success(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindVideoByIDFn: func(videoId string) (*models.VideoModel, error) {
			return &models.VideoModel{Id: "video-1", UserID: "user-123", Title: "Test"}, nil
		},
	}

	mockTag := &mocks.MockTagService{
		AddTagsToVideoFn: func(videoId string, tagNames []string) error {
			return nil
		},
	}

	controller := NewTagController(mockTag, mockDBVideo)
	router := setupTagRouter(controller, mockDBVideo)

	body := strings.NewReader(`{"tags": ["golang", "tutorial"]}`)
	req, _ := http.NewRequest("POST", "/tags/video-1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAddTagsToVideo_Forbidden(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindVideoByIDFn: func(videoId string) (*models.VideoModel, error) {
			return &models.VideoModel{Id: "video-1", UserID: "other-user", Title: "Test"}, nil
		},
	}

	controller := NewTagController(&mocks.MockTagService{}, mockDBVideo)
	router := setupTagRouter(controller, mockDBVideo)

	body := strings.NewReader(`{"tags": ["golang"]}`)
	req, _ := http.NewRequest("POST", "/tags/video-1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRemoveTagFromVideo_Success(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindVideoByIDFn: func(videoId string) (*models.VideoModel, error) {
			return &models.VideoModel{Id: "video-1", UserID: "user-123", Title: "Test"}, nil
		},
	}

	mockTag := &mocks.MockTagService{
		RemoveTagFromVideoFn: func(videoId string, tagName string) error {
			return nil
		},
	}

	controller := NewTagController(mockTag, mockDBVideo)
	router := setupTagRouter(controller, mockDBVideo)

	body := strings.NewReader(`{"tag": "golang"}`)
	req, _ := http.NewRequest("DELETE", "/tags/video-1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRemoveTagFromVideo_TagNotFound(t *testing.T) {
	mockDBVideo := &mocks.MockDatabaseVideoService{
		FindVideoByIDFn: func(videoId string) (*models.VideoModel, error) {
			return &models.VideoModel{Id: "video-1", UserID: "user-123", Title: "Test"}, nil
		},
	}

	mockTag := &mocks.MockTagService{
		RemoveTagFromVideoFn: func(videoId string, tagName string) error {
			return errors.New("tag 'unknown' not found")
		},
	}

	controller := NewTagController(mockTag, mockDBVideo)
	router := setupTagRouter(controller, mockDBVideo)

	body := strings.NewReader(`{"tag": "unknown"}`)
	req, _ := http.NewRequest("DELETE", "/tags/video-1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

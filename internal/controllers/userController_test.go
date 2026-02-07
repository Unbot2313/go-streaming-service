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

func setupUserRouter(controller UserController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/id/:id", controller.GetUserByID)
	r.GET("/users/username/:username", controller.GetUserByUserName)
	return r
}

func TestGetUserByID_Success(t *testing.T) {
	mockUser := &mocks.MockUserService{
		GetUserByIDFn: func(id string) (*models.User, error) {
			return &models.User{Id: "123", Username: "testuser", Email: "test@test.com"}, nil
		},
	}

	controller := NewUserController(mockUser)
	router := setupUserRouter(controller)

	req, _ := http.NewRequest("GET", "/users/id/123", nil)
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

func TestGetUserByID_NotFound(t *testing.T) {
	mockUser := &mocks.MockUserService{
		GetUserByIDFn: func(id string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
	}

	controller := NewUserController(mockUser)
	router := setupUserRouter(controller)

	req, _ := http.NewRequest("GET", "/users/id/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetUserByUserName_Success(t *testing.T) {
	mockUser := &mocks.MockUserService{
		GetUserByUserNameFn: func(userName string) (*models.User, error) {
			return &models.User{Id: "123", Username: userName, Email: "test@test.com"}, nil
		},
	}

	controller := NewUserController(mockUser)
	router := setupUserRouter(controller)

	req, _ := http.NewRequest("GET", "/users/username/testuser", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetUserByUserName_NotFound(t *testing.T) {
	mockUser := &mocks.MockUserService{
		GetUserByUserNameFn: func(userName string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
	}

	controller := NewUserController(mockUser)
	router := setupUserRouter(controller)

	req, _ := http.NewRequest("GET", "/users/username/unknown", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

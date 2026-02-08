package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

	// Rutas protegidas con usuario simulado
	protected := r.Group("")
	protected.Use(func(c *gin.Context) {
		c.Set("user", &models.User{Id: "user-123", Username: "testuser", Email: "old@test.com"})
		c.Next()
	})
	protected.PATCH("/users/email", controller.UpdateEmail)
	protected.PATCH("/users/password", controller.UpdatePassword)
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

func TestUpdateEmail_Success(t *testing.T) {
	mockUser := &mocks.MockUserService{
		UpdateEmailFn: func(userId, newEmail string) error {
			return nil
		},
	}

	controller := NewUserController(mockUser)
	router := setupUserRouter(controller)

	body := strings.NewReader(`{"email": "new@test.com"}`)
	req, _ := http.NewRequest("PATCH", "/users/email", body)
	req.Header.Set("Content-Type", "application/json")
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

func TestUpdateEmail_InvalidFormat(t *testing.T) {
	controller := NewUserController(&mocks.MockUserService{})
	router := setupUserRouter(controller)

	body := strings.NewReader(`{"email": "not-an-email"}`)
	req, _ := http.NewRequest("PATCH", "/users/email", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateEmail_AlreadyInUse(t *testing.T) {
	mockUser := &mocks.MockUserService{
		UpdateEmailFn: func(userId, newEmail string) error {
			return errors.New("email already in use")
		},
	}

	controller := NewUserController(mockUser)
	router := setupUserRouter(controller)

	body := strings.NewReader(`{"email": "taken@test.com"}`)
	req, _ := http.NewRequest("PATCH", "/users/email", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestUpdatePassword_Success(t *testing.T) {
	mockUser := &mocks.MockUserService{
		UpdatePasswordFn: func(userId, currentPassword, newPassword string) error {
			return nil
		},
	}

	controller := NewUserController(mockUser)
	router := setupUserRouter(controller)

	body := strings.NewReader(`{"current_password": "oldpass123", "new_password": "newpass123"}`)
	req, _ := http.NewRequest("PATCH", "/users/password", body)
	req.Header.Set("Content-Type", "application/json")
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

func TestUpdatePassword_BadRequest(t *testing.T) {
	controller := NewUserController(&mocks.MockUserService{})
	router := setupUserRouter(controller)

	body := strings.NewReader(`{"current_password": "old"}`)
	req, _ := http.NewRequest("PATCH", "/users/password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdatePassword_WrongCurrent(t *testing.T) {
	mockUser := &mocks.MockUserService{
		UpdatePasswordFn: func(userId, currentPassword, newPassword string) error {
			return errors.New("current password is incorrect")
		},
	}

	controller := NewUserController(mockUser)
	router := setupUserRouter(controller)

	body := strings.NewReader(`{"current_password": "wrongpass", "new_password": "newpass123"}`)
	req, _ := http.NewRequest("PATCH", "/users/password", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

package controllers

import (
	"bytes"
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

func setupAuthRouter(controller AuthController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", controller.Login)
	r.POST("/auth/register", controller.Register)
	r.POST("/auth/refresh", controller.RefreshToken)
	return r
}

func TestLogin_Success(t *testing.T) {
	mockAuth := &mocks.MockAuthService{
		LoginFn: func(username, password string) (*services.TokenPair, error) {
			return &services.TokenPair{
				AccessToken:  "access-token-123",
				RefreshToken: "refresh-token-456",
			}, nil
		},
	}

	controller := NewAuthController(mockAuth, nil)
	router := setupAuthRouter(controller)

	body, _ := json.Marshal(models.UserLogin{Username: "testuser", Password: "password123"})
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
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

	data := response["data"].(map[string]interface{})
	if data["access_token"] != "access-token-123" {
		t.Errorf("expected access_token 'access-token-123', got '%s'", data["access_token"])
	}
}

func TestLogin_BadRequest(t *testing.T) {
	controller := NewAuthController(&mocks.MockAuthService{}, nil)
	router := setupAuthRouter(controller)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mockAuth := &mocks.MockAuthService{
		LoginFn: func(username, password string) (*services.TokenPair, error) {
			return nil, errors.New("invalid credentials")
		},
	}

	controller := NewAuthController(mockAuth, nil)
	router := setupAuthRouter(controller)

	body, _ := json.Marshal(models.UserLogin{Username: "testuser", Password: "wrongpass"})
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRegister_Success(t *testing.T) {
	mockUser := &mocks.MockUserService{
		CreateUserFn: func(user *models.User) (*models.User, error) {
			user.Id = "user-123"
			return user, nil
		},
	}

	mockAuth := &mocks.MockAuthService{
		GenerateTokenFn: func(user *models.User) (string, error) {
			return "access-token", nil
		},
		GenerateRefreshTokenFn: func(user *models.User) (string, error) {
			return "refresh-token", nil
		},
		SaveRefreshTokenFn: func(userId, refreshToken string) error {
			return nil
		},
	}

	controller := NewAuthController(mockAuth, mockUser)
	router := setupAuthRouter(controller)

	body, _ := json.Marshal(map[string]string{
		"username": "newuser",
		"password": "password123",
		"email":    "test@example.com",
	})
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestRegister_Conflict(t *testing.T) {
	mockUser := &mocks.MockUserService{
		CreateUserFn: func(user *models.User) (*models.User, error) {
			return nil, errors.New("username already exists")
		},
	}

	controller := NewAuthController(&mocks.MockAuthService{}, mockUser)
	router := setupAuthRouter(controller)

	body, _ := json.Marshal(map[string]string{
		"username": "existinguser",
		"password": "password123",
		"email":    "test@example.com",
	})
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	mockAuth := &mocks.MockAuthService{
		RefreshTokensFn: func(refreshToken string) (*services.TokenPair, error) {
			return &services.TokenPair{
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
			}, nil
		},
	}

	controller := NewAuthController(mockAuth, nil)
	router := setupAuthRouter(controller)

	body, _ := json.Marshal(map[string]string{"refresh_token": "old-refresh"})
	req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRefreshToken_Invalid(t *testing.T) {
	mockAuth := &mocks.MockAuthService{
		RefreshTokensFn: func(refreshToken string) (*services.TokenPair, error) {
			return nil, errors.New("invalid token")
		},
	}

	controller := NewAuthController(mockAuth, nil)
	router := setupAuthRouter(controller)

	body, _ := json.Marshal(map[string]string{"refresh_token": "bad-token"})
	req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

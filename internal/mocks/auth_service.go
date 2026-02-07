package mocks

import (
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type MockAuthService struct {
	GenerateTokenFn        func(user *models.User) (string, error)
	ValidateTokenFn        func(token string) (*models.User, error)
	LoginFn                func(username, password string) (*services.TokenPair, error)
	GenerateRefreshTokenFn func(user *models.User) (string, error)
	ValidateRefreshTokenFn func(tokenString string) (*models.User, error)
	SaveRefreshTokenFn     func(userId, refreshToken string) error
	ClearRefreshTokenFn    func(userId string) error
	RefreshTokensFn        func(refreshToken string) (*services.TokenPair, error)
}

func (m *MockAuthService) GenerateToken(user *models.User) (string, error) {
	return m.GenerateTokenFn(user)
}

func (m *MockAuthService) ValidateToken(token string) (*models.User, error) {
	return m.ValidateTokenFn(token)
}

func (m *MockAuthService) Login(username, password string) (*services.TokenPair, error) {
	return m.LoginFn(username, password)
}

func (m *MockAuthService) GenerateRefreshToken(user *models.User) (string, error) {
	return m.GenerateRefreshTokenFn(user)
}

func (m *MockAuthService) ValidateRefreshToken(tokenString string) (*models.User, error) {
	return m.ValidateRefreshTokenFn(tokenString)
}

func (m *MockAuthService) SaveRefreshToken(userId, refreshToken string) error {
	return m.SaveRefreshTokenFn(userId, refreshToken)
}

func (m *MockAuthService) ClearRefreshToken(userId string) error {
	return m.ClearRefreshTokenFn(userId)
}

func (m *MockAuthService) RefreshTokens(refreshToken string) (*services.TokenPair, error) {
	return m.RefreshTokensFn(refreshToken)
}

package mocks

import (
	"github.com/unbot2313/go-streaming-service/internal/models"
)

type MockUserService struct {
	GetUserByIDFn       func(id string) (*models.User, error)
	GetUserByUserNameFn func(userName string) (*models.User, error)
	CreateUserFn        func(user *models.User) (*models.User, error)
	DeleteUserByIDFn    func(id string) error
	UpdateUserByIDFn    func(id string, user *models.User) (*models.User, error)
	UpdateEmailFn       func(userId, newEmail string) error
	UpdatePasswordFn    func(userId, currentPassword, newPassword string) error
}

func (m *MockUserService) GetUserByID(id string) (*models.User, error) {
	return m.GetUserByIDFn(id)
}

func (m *MockUserService) GetUserByUserName(userName string) (*models.User, error) {
	return m.GetUserByUserNameFn(userName)
}

func (m *MockUserService) CreateUser(user *models.User) (*models.User, error) {
	return m.CreateUserFn(user)
}

func (m *MockUserService) DeleteUserByID(id string) error {
	return m.DeleteUserByIDFn(id)
}

func (m *MockUserService) UpdateUserByID(id string, user *models.User) (*models.User, error) {
	return m.UpdateUserByIDFn(id, user)
}

func (m *MockUserService) UpdateEmail(userId, newEmail string) error {
	return m.UpdateEmailFn(userId, newEmail)
}

func (m *MockUserService) UpdatePassword(userId, currentPassword, newPassword string) error {
	return m.UpdatePasswordFn(userId, currentPassword, newPassword)
}

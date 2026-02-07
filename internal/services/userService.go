package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/unbot2313/go-streaming-service/config"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"gorm.io/gorm"
)

type UserServiceImp struct{}

type UserService interface {
	GetUserByID(Id string) (*models.User, error)
	GetUserByUserName(userName string) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	DeleteUserByID(Id string) error
	UpdateUserByID(Id string, user *models.User) (*models.User, error)
	UpdateEmail(userId, newEmail string) error
	UpdatePassword(userId, currentPassword, newPassword string) error
}

func (service *UserServiceImp) GetUserByID(Id string) (*models.User, error) {
	var user models.User

	// Obtén la conexión a la base de datos
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	// Busca el usuario por ID e incluye los videos asociados
	err = db.Preload("Videos").First(&user, "id = ?", Id).Error

	// Maneja el caso de usuario no encontrado
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with ID %s not found", Id)
	}

	// Maneja cualquier otro error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (service *UserServiceImp) GetUserByUserName(userName string) (*models.User, error) {
	var user models.User

	// Obtén la conexión a la base de datos
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	// Busca el usuario por username e incluye los videos asociados
	err = db.Preload("Videos").First(&user, "username = ?", userName).Error

	// Maneja el caso de usuario no encontrado
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with username %s not found", userName)
	}

	// Maneja cualquier otro error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (service *UserServiceImp) CreateUser(user *models.User) (*models.User, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	user.Id = uuid.New().String()

	hashedPassword, err := HashPassword(user.Password)

	if err != nil {
		return nil, err
	}

	user.Password = hashedPassword

	dbCtx := db.Create(user)

	if errors.Is(dbCtx.Error, gorm.ErrDuplicatedKey) {
		return nil, fmt.Errorf("user with username %s already exists", user.Username)
	}

	if dbCtx.Error != nil {
		return nil, dbCtx.Error
	}

	return user, nil
}

func (service *UserServiceImp) DeleteUserByID(Id string) error {

	db, err := config.GetDB()
	if err != nil {
		return err
	}

	 // Elimina el usuario usando el campo personalizado `Id`
    dbCtx := db.Where("id = ?", Id).Delete(&models.User{})

	if errors.Is(dbCtx.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("user with ID %s not found", Id)
	}

    if dbCtx.Error != nil {
        return dbCtx.Error
    }

	return nil
}

func (service *UserServiceImp) UpdateUserByID(Id string, user *models.User) (*models.User, error) {
	return nil, nil
}

func (service *UserServiceImp) UpdateEmail(userId, newEmail string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	result := db.Model(&models.User{}).Where("id = ?", userId).Update("email", newEmail)

	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return fmt.Errorf("email already in use")
	}

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userId)
	}

	return nil
}

func (service *UserServiceImp) UpdatePassword(userId, currentPassword, newPassword string) error {
	user, err := service.GetUserByID(userId)
	if err != nil {
		return err
	}

	if !CheckPasswordHash(currentPassword, user.Password) {
		return fmt.Errorf("current password is incorrect")
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	db, err := config.GetDB()
	if err != nil {
		return err
	}

	result := db.Model(&models.User{}).Where("id = ?", userId).Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func NewUserService() UserService {
	return &UserServiceImp{}
}
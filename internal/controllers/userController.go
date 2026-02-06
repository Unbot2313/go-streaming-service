package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/helpers"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type UserControllerImp struct {
	service	 services.UserService
}

type UserController interface {
	CreateUser(c *gin.Context)
	GetUserByID(c *gin.Context)
	GetUserByUserName(c *gin.Context)
	DeleteUserByID(c *gin.Context)
}


// GetUserByID		godoc
// @Summary 		Get user by ID
// @Description 	Search user by ID in Db
// @Tags 			users
// @Param 			Id path string true "User ID"
// @Produce 		json
// @Success 		200 {object} models.UserSwagger{}
// @Failure 		404 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/users/id/{UserId} [get]
func (controller *UserControllerImp) GetUserByID(c *gin.Context) {
	Id := c.Param("id")
	users, err := controller.service.GetUserByID(Id)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "User not found", err)
		return
	}
	helpers.Success(c, http.StatusOK, users)
}

// GetUserByUserName		godoc
// @Summary 				Get user by userName
// @Description 			Search user by userName in Db
// @Tags 					users
// @Param 					userName path string true "User Name"
// @Produce 				json
// @Success 				200 {object} models.UserSwagger{}
// @Failure 				404 {object} map[string]string
// @Failure 				500 {object} map[string]string
// @Router 					/users/username/{userName} [get]
func (controller *UserControllerImp) GetUserByUserName(c *gin.Context) {
	userName := c.Param("username")
	users, err := controller.service.GetUserByUserName(userName)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "User not found", err)
		return
	}
	helpers.Success(c, http.StatusOK, users)
}

// CreateUser		godoc
// @Summary 		Create a new user
// @Description 	Save user in Db
// @Tags 			users
// @Accept 			json
// @Param 			user body models.UserLogin{} true "User object containing all user details"
// @Produce 		json
// @Success 		200 {object} models.UserSwagger{}
// @Failure 		400 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/users/ [post]
func (controller *UserControllerImp) CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	newUser, err := controller.service.CreateUser(&user)
	if err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "Could not create user", err)
		return
	}

	helpers.Success(c, http.StatusCreated, newUser)
}

// GetUserByID		godoc
// @Summary 		Delete user by ID
// @Description 	Delete user by ID ni Db
// @Tags 			users
// @Param 			Id path string true "User ID"
// @Produce 		json
// @Success 		200 {object} models.UserSwagger{}
// @Failure 		404 {object} map[string]string
// @Failure 		500 {object} map[string]string
// @Router 			/users/{UserId} [delete]
func (controller *UserControllerImp) DeleteUserByID(c *gin.Context) {
	id := c.Param("id")

	// Verificar que el usuario autenticado es el dueno de la cuenta
	user, exists := c.Get("user")
	if !exists {
		helpers.HandleError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	authenticatedUser, ok := user.(*models.User)
	if !ok || authenticatedUser.Id != id {
		helpers.HandleError(c, http.StatusForbidden, "You can only delete your own account", nil)
		return
	}

	err := controller.service.DeleteUserByID(id)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "User not found", err)
		return
	}
	helpers.Success(c, http.StatusOK, gin.H{"message": "User deleted"})
}




func NewUserController(service services.UserService) *UserControllerImp {
	return &UserControllerImp{service: service}
}
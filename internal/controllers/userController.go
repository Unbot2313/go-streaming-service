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
	GetUserByID(c *gin.Context)
	GetUserByUserName(c *gin.Context)
	DeleteUserByID(c *gin.Context)
}


// GetUserByID godoc
// @Summary		Get user by ID
// @Description	Search user by ID in Db
// @Tags		users
// @Produce		json
// @Param		UserId path string true "User ID"
// @Success		200 {object} helpers.APIResponse{data=models.UserSwagger}
// @Failure		404 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/users/id/{UserId} [get]
func (controller *UserControllerImp) GetUserByID(c *gin.Context) {
	Id := c.Param("id")
	users, err := controller.service.GetUserByID(Id)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "User not found", err)
		return
	}
	helpers.Success(c, http.StatusOK, users)
}

// GetUserByUserName godoc
// @Summary		Get user by userName
// @Description	Search user by userName in Db
// @Tags		users
// @Produce		json
// @Param		userName path string true "User Name"
// @Success		200 {object} helpers.APIResponse{data=models.UserSwagger}
// @Failure		404 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/users/username/{userName} [get]
func (controller *UserControllerImp) GetUserByUserName(c *gin.Context) {
	userName := c.Param("username")
	users, err := controller.service.GetUserByUserName(userName)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "User not found", err)
		return
	}
	helpers.Success(c, http.StatusOK, users)
}

// DeleteUserByID godoc
// @Summary		Delete user by ID
// @Description	Delete user by ID. Only the owner can delete their own account.
// @Tags		users
// @Produce		json
// @Security	BearerAuth
// @Param		UserId path string true "User ID"
// @Success		200 {object} helpers.APIResponse{data=object{message=string}}
// @Failure		401 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		403 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		404 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/users/{UserId} [delete]
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




func NewUserController(service services.UserService) UserController {
	return &UserControllerImp{service: service}
}
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/helpers"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type AuthController interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
}

// GetUserByUserName		godoc
// @Summary 				Log in user
// @Tags 					Auth
// @Produce 				json
// @Accept 					json
// @Param 					user body models.UserLogin{} true "User object containing all user details"
// @Success 				200 {object} map[string]string
// @Failure 				404 {object} map[string]string
// @Failure 				500 {object} map[string]string
// @Router 					/auth/login [post]
func (controller *AuthControllerImp) Login(c *gin.Context) {
	
	var userLogin models.UserLogin

	if err := c.ShouldBindJSON(&userLogin); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "Username and password are required", err)
		return
	}

	token, err := controller.authService.Login(userLogin.Username, userLogin.Password)

	if err != nil {
		helpers.HandleError(c, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user": userLogin.Username,
	})
}

// Register godoc
// @Summary		Register a new user
// @Description	Create a new user account and return a JWT token
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Param		user body models.UserRegister{} true "User registration data"
// @Success		201 {object} map[string]string
// @Failure		400 {object} map[string]string
// @Failure		409 {object} map[string]string
// @Router		/auth/register [post]
func (controller *AuthControllerImp) Register(c *gin.Context) {
	var req models.UserRegister

	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "Invalid input: username (3-100 chars), password (min 8 chars) and valid email are required", err)
		return
	}

	user := &models.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	createdUser, err := controller.userService.CreateUser(user)
	if err != nil {
		helpers.HandleError(c, http.StatusConflict, "Username already exists", err)
		return
	}

	token, err := controller.authService.GenerateToken(createdUser)
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not generate token", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  createdUser,
	})
}


type AuthControllerImp struct {
	authService services.AuthService
	userService services.UserService
}

// NewAuthController crea una nueva instancia del controlador de autenticación
func NewAuthController(authService services.AuthService, userService services.UserService) AuthController {
	return &AuthControllerImp{authService: authService, userService: userService}
}

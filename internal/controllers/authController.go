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
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
}

// Login godoc
// @Summary		Log in user
// @Description	Authenticate with username and password to get access and refresh tokens
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Param		user body models.UserLogin true "User credentials"
// @Success		200 {object} helpers.APIResponse{data=object{access_token=string,refresh_token=string,user=string}}
// @Failure		400 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		401 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/auth/login [post]
func (controller *AuthControllerImp) Login(c *gin.Context) {
	
	var userLogin models.UserLogin

	if err := c.ShouldBindJSON(&userLogin); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "Username and password are required", err)
		return
	}

	tokens, err := controller.authService.Login(userLogin.Username, userLogin.Password)

	if err != nil {
		helpers.HandleError(c, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	helpers.Success(c, http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":          userLogin.Username,
	})
}

// Register godoc
// @Summary		Register a new user
// @Description	Create a new user account and return access and refresh tokens
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Param		user body models.UserRegister true "User registration data"
// @Success		201 {object} helpers.APIResponse{data=object{access_token=string,refresh_token=string,user=models.UserSwagger}}
// @Failure		400 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		409 {object} helpers.APIResponse{error=helpers.APIError}
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

	accessToken, err := controller.authService.GenerateToken(createdUser)
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not generate token", err)
		return
	}

	refreshToken, err := controller.authService.GenerateRefreshToken(createdUser)
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not generate refresh token", err)
		return
	}

	if err := controller.authService.SaveRefreshToken(createdUser.Id, refreshToken); err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not save refresh token", err)
		return
	}

	helpers.Success(c, http.StatusCreated, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          createdUser,
	})
}


// RefreshToken godoc
// @Summary		Refresh access token
// @Description	Exchange a valid refresh token for a new access/refresh token pair
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Param		body body object{refresh_token=string} true "Refresh token"
// @Success		200 {object} helpers.APIResponse{data=object{access_token=string,refresh_token=string}}
// @Failure		400 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		401 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/auth/refresh [post]
func (controller *AuthControllerImp) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "refresh_token is required", err)
		return
	}

	tokens, err := controller.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		helpers.HandleError(c, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		return
	}

	helpers.Success(c, http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

// Logout godoc
// @Summary		Logout user
// @Description	Revoke the refresh token for the authenticated user
// @Tags		Auth
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} helpers.APIResponse{data=object{message=string}}
// @Failure		401 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		500 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/auth/logout [post]
func (controller *AuthControllerImp) Logout(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		helpers.HandleError(c, http.StatusInternalServerError, "User not found in context", nil)
		return
	}

	authenticatedUser := user.(*models.User)

	if err := controller.authService.ClearRefreshToken(authenticatedUser.Id); err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not logout", err)
		return
	}

	helpers.Success(c, http.StatusOK, gin.H{"message": "Logged out successfully"})
}

type AuthControllerImp struct {
	authService services.AuthService
	userService services.UserService
}

// NewAuthController crea una nueva instancia del controlador de autenticación
func NewAuthController(authService services.AuthService, userService services.UserService) AuthController {
	return &AuthControllerImp{authService: authService, userService: userService}
}

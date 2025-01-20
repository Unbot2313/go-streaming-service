package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/controllers"
	"github.com/unbot2313/go-streaming-service/internal/middlewares"
)

// SetupRoutes configura todas las rutas
func SetupRoutes(router *gin.RouterGroup, userController controllers.UserController, authController controllers.AuthController, videoController controllers.VideoController) {
	// Rutas de usuarios
	userRoutes := router.Group("/users")
	{
		userRoutes.GET("/id/:id", userController.GetUserByID)
		userRoutes.GET("/username/:username", userController.GetUserByUserName)
		userRoutes.POST("/", userController.CreateUser)
		userRoutes.DELETE("/:id", userController.DeleteUserByID)
	}

	// Rutas de autenticación
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/register", authController.Register)
	}

    VideoRoutes := router.Group("/streaming")
    {
		ProtectedRoute := VideoRoutes.Group("")
		ProtectedRoute.Use(middlewares.AuthMiddleware)

		// Rutas públicas
        VideoRoutes.GET("/latest", videoController.GetLatestVideos)
		VideoRoutes.GET("/id/:videoid", videoController.GetVideoByID)
		VideoRoutes.PATCH("/views/:videoid", videoController.IncrementViews)

		// Ruta protegida
        ProtectedRoute.POST("/upload", videoController.CreateVideo)
    }
	
}

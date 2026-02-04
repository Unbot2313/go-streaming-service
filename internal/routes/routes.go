package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/controllers"
	"github.com/unbot2313/go-streaming-service/internal/middlewares"
	"golang.org/x/time/rate"
)

// SetupRoutes configura todas las rutas
func SetupRoutes(router *gin.RouterGroup, userController controllers.UserController, authController controllers.AuthController, videoController controllers.VideoController, jobController controllers.JobController) {
	// Rutas de usuarios
	userRoutes := router.Group("/users")
	{
		userRoutes.GET("/id/:id", userController.GetUserByID)
		userRoutes.GET("/username/:username", userController.GetUserByUserName)
		userRoutes.POST("/", userController.CreateUser)
		userRoutes.DELETE("/:id", userController.DeleteUserByID)
	}

	// Rutas de autenticación
	// Rate limiter estricto: 1 token cada 20s, burst 3 (anti fuerza bruta)
	authLimiter := middlewares.NewRateLimiter(rate.Every(20*time.Second), 3)
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", authLimiter.Middleware(), authController.Login)
		authRoutes.POST("/register", authLimiter.Middleware(), authController.Register)
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

	// Rutas de jobs (públicas)
	jobRoutes := router.Group("/jobs")
	{
		jobRoutes.GET("/:jobid", jobController.GetJobByID)
	}
}

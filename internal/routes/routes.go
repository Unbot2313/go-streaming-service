package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/controllers"
	"github.com/unbot2313/go-streaming-service/internal/middlewares"
	"github.com/unbot2313/go-streaming-service/internal/services"
	"golang.org/x/time/rate"
)

// SetupRoutes configura todas las rutas
func SetupRoutes(router *gin.RouterGroup, userController controllers.UserController, authController controllers.AuthController, videoController controllers.VideoController, jobController controllers.JobController, authService services.AuthService) {
	// Middleware de autenticación (una sola instancia reutilizada)
	authMiddleware := middlewares.AuthMiddleware(authService)

	// Rutas de usuarios
	userRoutes := router.Group("/users")
	{
		// Rutas publicas (lectura)
		userRoutes.GET("/id/:id", userController.GetUserByID)
		userRoutes.GET("/username/:username", userController.GetUserByUserName)

		// Rutas protegidas
		protectedUserRoutes := userRoutes.Group("")
		protectedUserRoutes.Use(authMiddleware)
		protectedUserRoutes.DELETE("/:id", userController.DeleteUserByID)
	}

	// Rutas de autenticación
	// Rate limiter estricto: 1 token cada 20s, burst 3 (anti fuerza bruta)
	authLimiter := middlewares.NewRateLimiter(rate.Every(20*time.Second), 3)
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", authLimiter.Middleware(), authController.Login)
		authRoutes.POST("/register", authLimiter.Middleware(), authController.Register)
		authRoutes.POST("/refresh", authLimiter.Middleware(), authController.RefreshToken)

		// Logout requiere estar autenticado
		protectedAuthRoutes := authRoutes.Group("")
		protectedAuthRoutes.Use(authMiddleware)
		protectedAuthRoutes.POST("/logout", authController.Logout)
	}

    VideoRoutes := router.Group("/streaming")
    {
		ProtectedRoute := VideoRoutes.Group("")
		ProtectedRoute.Use(authMiddleware)

		// Rutas públicas
        VideoRoutes.GET("/latest", videoController.GetLatestVideos)
		VideoRoutes.GET("/id/:videoid", videoController.GetVideoByID)
		VideoRoutes.PATCH("/views/:videoid", videoController.IncrementViews)

		// Rutas protegidas
        ProtectedRoute.POST("/upload", videoController.CreateVideo)
		ProtectedRoute.PUT("/:videoid", videoController.UpdateVideo)
		ProtectedRoute.DELETE("/:videoid", videoController.DeleteVideo)
    }

	// Rutas de jobs (protegidas)
	jobRoutes := router.Group("/jobs")
	jobRoutes.Use(authMiddleware)
	{
		jobRoutes.GET("/:jobid", jobController.GetJobByID)
	}
}

package app

import (
	"github.com/unbot2313/go-streaming-service/internal/controllers"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

// InitializeComponents crea las instancias de los servicios y controladores
func InitializeComponents() (controllers.UserController, controllers.AuthController, controllers.VideoController, controllers.JobController) {
	// Inicializa los servicios base
	userService := services.NewUserService()
	authService := services.NewAuthService()

	// Inicializa los controladores de usuario y auth
	userController := controllers.NewUserController(userService)
	authController := controllers.NewAuthController(authService)

	// Inicializa servicios de video
	S3configuration := services.GetS3Configuration()
	filesService := services.NewFilesService()
	videoService := services.NewVideoService(S3configuration, filesService)
	databaseVideoService := services.NewDatabaseVideoService()

	// Inicializa servicios de jobs y RabbitMQ
	jobService := services.NewJobService()
	rabbitMQService := services.NewRabbitMQService()

	// Inicializa controladores
	videoController := controllers.NewVideoController(videoService, databaseVideoService, jobService, rabbitMQService)
	jobController := controllers.NewJobController(jobService)

	return userController, authController, videoController, jobController
}

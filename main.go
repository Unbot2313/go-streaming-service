package main

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/unbot2313/go-streaming-service/config"
	_ "github.com/unbot2313/go-streaming-service/docs"
	"github.com/unbot2313/go-streaming-service/internal/app"
	"github.com/unbot2313/go-streaming-service/internal/routes"
)

// @title Go Streaming Service API
// @version 1.0
// @description A streaming service API using Go and Gin framework, with Swagger documentation and ffmpeg integration.

// @host	localhost:3003
// @BasePath /api/v1

func main() {

	cfg := config.GetConfig()

	r := gin.Default()

	allowedOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	apiGroup := r.Group("/api")

	v1Group := apiGroup.Group("/v1")

	// conect to database
	_, err := config.GetDB()
	if err != nil {
		panic(err)
	}

	// Servir archivos staticos (STREAMING)
	// cuando se accede a la ruta /static, se sirven los archivos que estan en la carpeta public,
	// ejm: http://localhost:3003/static/index.html, se sirve /public/index.html
	v1Group.Static("/static", "./static/temp")

	// Inicializar los componentes de la aplicación
	userController, authController, videoController, jobController := app.InitializeComponents()

	// Configurar las rutas
	routes.SetupRoutes(v1Group, userController, authController, videoController, jobController)
	// Configurar la documentación de Swagger
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))


	r.Run(":3003")

}
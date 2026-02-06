package main

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/unbot2313/go-streaming-service/config"
	_ "github.com/unbot2313/go-streaming-service/docs"
	"github.com/unbot2313/go-streaming-service/internal/app"
	"github.com/unbot2313/go-streaming-service/internal/middlewares"
	"github.com/unbot2313/go-streaming-service/internal/routes"
	"github.com/unbot2313/go-streaming-service/internal/services"
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

	// Rate limiter general: 10 req/s, burst 20
	generalLimiter := middlewares.NewRateLimiter(10, 20)
	r.Use(generalLimiter.Middleware())

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
	userController, authController, videoController, jobController, authService := app.InitializeComponents()

	// Configurar las rutas
	routes.SetupRoutes(v1Group, userController, authController, videoController, jobController, authService)
	// Configurar la documentación de Swagger
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))


	// Health check endpoints (fuera de /api/v1, sin auth ni rate limit)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/ready", func(c *gin.Context) {
		checks := gin.H{}

		// Check DB
		db, err := config.GetDB()
		if err != nil {
			checks["database"] = "error: " + err.Error()
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "checks": checks})
			return
		}
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			checks["database"] = "unreachable"
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "checks": checks})
			return
		}
		checks["database"] = "ok"

		// Check RabbitMQ
		rabbitService := services.NewRabbitMQService()
		if err := rabbitService.Connect(); err != nil {
			checks["rabbitmq"] = "unreachable"
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "checks": checks})
			return
		}
		rabbitService.Close()
		checks["rabbitmq"] = "ok"

		c.JSON(http.StatusOK, gin.H{"status": "ready", "checks": checks})
	})

	r.Run(":3003")

}
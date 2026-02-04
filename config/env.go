package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecretKey string
	AWSRegion	 string
	AWSBucketName string
	AWSAccessKey string
	AWSSecretKey string
	LocalStoragePath string

	PostgresHost	 string
	PostgresPort	 string
	PostgresUser	 string
	PostgresPassword string
	PostgresDBName	 string

	RabbitMQHost           string
	RabbitMQPort           string
	RabbitMQUser           string
	RabbitMQPassword       string
	RabbitMQVideoQueue     string
	RabbitMQThumbnailQueue string

	CORSAllowedOrigins string

	StorageType     string
	MinIOEndpoint   string
	MinIOBucketName string
	MinIOAccessKey  string
	MinIOSecretKey  string
}


// el singleton de configuracion 
var (
	config     *Config
	configOnce sync.Once
)


func GetConfig() *Config {

	// usar Sync.Once para garantizar que la configuración se cargue solo una vez y evitar problemas de rendimiento
	// y usa singleton para garantizar que solo haya una instancia de la configuración en toda la aplicación.

	configOnce.Do(func() {
		err := loadEnv()
		if err != nil {
			panic(fmt.Sprintf("Error al cargar el archivo .env: %v", err))
		}

		config = &Config{
			Port:         getEnv("PORT", "8080"),
			JWTSecretKey: getEnv("JWT_SECRET_KEY", ""),
			LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "videos"),
			AWSRegion:    getEnv("AWS_REGION", ""),
			AWSBucketName: getEnv("AWS_BUCKET_NAME", ""),
			AWSAccessKey: getEnv("AWS_ACCESS_KEY_ID", ""),
			AWSSecretKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),

			PostgresHost: getEnv("POSTGRES_HOST", "localhost"),
			PostgresPort: getEnv("POSTGRES_PORT", "5432"),
			PostgresUser: getEnv("POSTGRES_USER", "postgres"),
			PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
			PostgresDBName: getEnv("POSTGRES_DB", "golang"),

			RabbitMQHost:           getEnv("RABBITMQ_HOST", "localhost"),
			RabbitMQPort:           getEnv("RABBITMQ_PORT", "5672"),
			RabbitMQUser:           getEnv("RABBITMQ_USER", "guest"),
			RabbitMQPassword:       getEnv("RABBITMQ_PASSWORD", "guest"),
			RabbitMQVideoQueue:     getEnv("RABBITMQ_VIDEO_QUEUE", "video_processing"),
			RabbitMQThumbnailQueue: getEnv("RABBITMQ_THUMBNAIL_QUEUE", "thumbnail_generation"),

			CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),

			StorageType:     getEnv("STORAGE_TYPE", "minio"),
			MinIOEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
			MinIOBucketName: getEnv("MINIO_BUCKET_NAME", "streaming-videos"),
			MinIOAccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			MinIOSecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		}

		validateConfig(config)
	})

	return config
}

func validateConfig(cfg *Config) {
	if cfg.JWTSecretKey == "" {
		panic("JWT_SECRET_KEY environment variable is required")
	}
	if len(cfg.JWTSecretKey) < 32 {
		panic("JWT_SECRET_KEY must be at least 32 characters long")
	}

	if cfg.PostgresPassword == "postgres" {
		log.Println("WARNING: using default PostgreSQL password, set POSTGRES_PASSWORD in .env")
	}
	if cfg.RabbitMQPassword == "guest" {
		log.Println("WARNING: using default RabbitMQ password, set RABBITMQ_PASSWORD in .env")
	}
}

func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env: %v", err)
	}

	return nil

}

// getEnvAsBool obtiene una variable de entorno como booleano o retorna un valor por defecto.
func getEnvAsBool(key string, defaultValue bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}


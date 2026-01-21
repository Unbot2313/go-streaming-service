package config

import (
	"fmt"
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

	DOCKER_MODE 	bool

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
			JWTSecretKey: getEnv("JWT_SECRET_KEY", "secretJwtKey"),
			LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "videos"),
			AWSRegion:    getEnv("AWS_REGION", ""),
			AWSBucketName: getEnv("AWS_BUCKET_NAME", ""),
			AWSAccessKey: getEnv("AWS_ACCESS_KEY_ID", ""),
			AWSSecretKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),

			DOCKER_MODE: getEnvAsBool("DOCKER_MODE", false),

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
		}
	})

	return config
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


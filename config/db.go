package config

import (
	"fmt"
	"sync"

	"github.com/unbot2313/go-streaming-service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbInstance *gorm.DB
	once       sync.Once
)

// GetDsn genera la cadena de conexión para la base de datos.
func getDsn() string {
	config := GetConfig()

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.PostgresHost,
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDBName,
		config.PostgresPort,
	)
}

// GetDB devuelve una instancia única de la conexión a la base de datos.
func GetDB() (*gorm.DB, error) {
	var err error
	once.Do(func() {
		dsn := getDsn()
		dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	})
	if err != nil {
		return nil, err
	}

	// Migra las tablas a la base de datos.
	err = migrations(dbInstance)
	if err != nil {
		return nil, err
	}
	return dbInstance, nil
}

func migrations(db *gorm.DB) error {
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&models.VideoModel{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&models.JobModel{})
	if err != nil {
		return err
	}

	return nil
}

package config

import (
	"fmt"
	"sync"

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

	return dbInstance, nil
}

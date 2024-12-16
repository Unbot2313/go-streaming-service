package models

import (
	"time"

	"gorm.io/gorm"
)

// Esto es lo que deberia recibir el controlador al crear
// un nuevo usuario
type UserCreate struct {
	Id         string `json:"id"`
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Email      string `json:"email"`
}

type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserSwagger se usa en la documentacion ya que Swaggo no reconoce
// los tags de gorms
type UserSwagger struct {
	Id           string    `json:"id" gorm:"primaryKey;not null;uniqueIndex"`
	Username     string    `json:"username" gorm:"type:varchar(100);not null;uniqueIndex"`
	Password     string    `json:"password" gorm:"not null"`
	Email        string    `json:"email" gorm:"type:varchar(100)"`
	RefreshToken string    `json:"refresh_token"`
	Videos []VideoSwagger 	`json:"videos" gorm:"foreignKey:UserID"`
}

type User struct {
	Id           string    `json:"id" gorm:"primaryKey;not null;uniqueIndex"`
	Username     string    `json:"username" gorm:"type:varchar(100);not null;uniqueIndex"`
	Password     string    `json:"password" gorm:"not null"`
	Email        string    `json:"email" gorm:"type:varchar(100)"`
	RefreshToken string    `json:"refresh_token"`
	Videos []VideoModel 	`json:"videos" gorm:"foreignKey:UserID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
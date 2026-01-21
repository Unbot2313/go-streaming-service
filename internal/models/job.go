package models

import (
	"time"

	"gorm.io/gorm"
)

// Job es la estructura base para tareas de procesamiento de video
type Job struct {
	Id           string `json:"id" gorm:"primaryKey;not null;uniqueIndex"`
	VideoID      string `json:"video_id"`
	UserID       string `json:"user_id" gorm:"not null"`
	Status       string `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	LocalPath    string `json:"-" gorm:"not null"`
	UniqueName   string `json:"-"`
	Title        string `json:"title" gorm:"type:varchar(100)"`
	Description  string `json:"description"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// JobModel embebe Job y agrega campos de GORM para la base de datos
type JobModel struct {
	Job
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName especifica el nombre de la tabla
func (JobModel) TableName() string {
	return "jobs"
}

// VideoTask es la estructura del mensaje enviado a RabbitMQ
type VideoTask struct {
	JobID       string `json:"job_id"`
	UserID      string `json:"user_id"`
	LocalPath   string `json:"local_path"`
	UniqueName  string `json:"unique_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

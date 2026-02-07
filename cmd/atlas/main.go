package main

import (
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/unbot2313/go-streaming-service/internal/models"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&models.User{},
		&models.Tag{},
		&models.VideoModel{},
		&models.JobModel{},
	)
	if err != nil {
		io.WriteString(os.Stderr, err.Error())
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}

package main

import (
	"github.com/sirupsen/logrus"
	"mount-service/internal/api"
	"mount-service/internal/log"
	"mount-service/internal/models"
	"os"
)

func setupLogger() {
	log.Logger.SetLevel(logrus.DebugLevel)
	log.Logger.SetFormatter(&logrus.TextFormatter{})
	log.Logger.SetOutput(os.Stdout)
}

func main() {
	setupLogger()
	config, _ := models.NewConfig()

	log.Logger.Infoln("Creating server...")

	server := api.CreateNewServer(config)

	log.Logger.WithFields(logrus.Fields{
		"port": 8080,
	}).Infoln("Start listen port 8080...")

	server.Run()
}

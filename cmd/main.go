package main

import (
	"github.com/sirupsen/logrus"
	"mount-service/internal/api"
	"mount-service/internal/model"
	"os"
)

func createLogger() *logrus.Logger {
	log := logrus.New()

	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{})
	log.SetOutput(os.Stdout)

	return log
}

func main() {
	log := createLogger()

	config, _ := model.NewConfig()

	log.Infoln("Creating server...")

	server := api.CreateNewServer(config)

	log.WithFields(logrus.Fields{
		"port": 8080,
	}).Infoln("Start listen port 8080...")

	server.Run()
}

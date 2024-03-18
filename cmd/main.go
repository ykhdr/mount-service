package main

import (
	log "github.com/sirupsen/logrus"
	"mount-service/internal/api"
	"mount-service/internal/model"
	"os"
)

func setupLogger() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
}

func main() {

	config, _ := model.NewConfig()

	log.Infoln("Creating server...")

	server := api.CreateNewServer(config)

	log.WithFields(log.Fields{
		"port": 8080,
	}).Infoln("Start listen port 8080...")

	server.Run()
}

package main

import (
	"mount-service/internal/api"
	"mount-service/internal/model"
)

func main() {
	config := model.NewConfig()

	server := api.NewMountServer(config)

	api.FillRoute(server)
}

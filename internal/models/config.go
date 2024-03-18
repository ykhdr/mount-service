package models

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	DBHost       string
	DBName       string
	DBPort       int
	DBUser       string
	DBPassword   string
	HostUser     string
	HostPassword string
}

func NewConfig() (*Config, error) {
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, errors.Join(errors.New("DB_PORT isn't passed or not integer"), err)
	}

	hostUser := os.Getenv("HOST_USER")
	hostPassword := os.Getenv("HOST_PASSWORD")

	return &Config{
		DBHost:       dbHost,
		DBName:       dbName,
		DBPort:       dbPort,
		DBUser:       dbUser,
		DBPassword:   dbPassword,
		HostPassword: hostPassword,
		HostUser:     hostUser,
	}, nil

}

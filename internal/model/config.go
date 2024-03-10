package model

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBName     string
	DBPort     int
	DBUser     string
	DBPassword string
}

func NewConfig() *Config {
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		panic(err)
	}

	return &Config{
		DBHost:     dbHost,
		DBName:     dbName,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
	}

}

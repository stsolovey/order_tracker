package config

import (
	"fmt"
	"net"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	AppPort     string
	AppHost     string
	LogLevel    string
	NATSURL     string
}

func New(path string) *Config {
	err := godotenv.Load(path)
	if err != nil {
		panic("Error loading .env file")
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	appHost := os.Getenv("APP_HOST")
	appPort := os.Getenv("APP_PORT")
	logLevel := os.Getenv("LOG_LEVEL")
	natsURL := os.Getenv("NATS_URL")

	var dsn string

	switch {
	case postgresHost == "":
		panic("postgresHost environment variable is missing")
	case postgresPort == "":
		panic("postgresPort environment variable is missing")
	case postgresUser == "":
		panic("postgresUser environment variable is missing")
	case postgresPassword == "":
		panic("postgresPassword environment variable is missing")
	case postgresDB == "":
		panic("postgresDB environment variable is missing")
	case appPort == "":
		panic("appPort environment variable is missing")
	case natsURL == "":
		panic("natsURL environment variable is missing")
	default:
		hostPort := net.JoinHostPort(postgresHost, postgresPort)
		dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			postgresUser, postgresPassword, hostPort, postgresDB)

		return &Config{
			DatabaseURL: dsn,
			AppPort:     appPort,
			AppHost:     appHost,
			LogLevel:    logLevel,
			NATSURL:     natsURL,
		}
	}
}

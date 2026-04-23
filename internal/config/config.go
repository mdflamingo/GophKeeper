package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Postgres struct {
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
}

type Config struct {
	RunAddr     string
	LogLevel    string
	DataBaseDSN Postgres
}

func GetConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Missing required .env file: ", err)

	}

	pgConfig := Postgres{
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSW0RD"),
	}
	mainConfig := Config{RunAddr: os.Getenv("RUN_ADDR"), LogLevel: os.Getenv("LOG_LEVEL"), DataBaseDSN: pgConfig}

	return &mainConfig
}

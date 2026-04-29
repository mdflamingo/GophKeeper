package config

import (
	"fmt"
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
	Debug       string
	SecretKey   string
}

func GetConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Missing required .env file: ", err)

	}
	debug := os.Getenv("DEBUG")
	fmt.Println(debug)
	fmt.Println(os.Getenv("LOG_LEVEL"))
	pgConfig := Postgres{
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
	}
	mainConfig := Config{RunAddr: os.Getenv("RUN_ADDR"), LogLevel: os.Getenv("LOG_LEVEL"), DataBaseDSN: pgConfig, Debug: debug, SecretKey: os.Getenv("SECRET_KEY")}

	return &mainConfig
}

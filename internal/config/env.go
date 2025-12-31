package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func AppPort() string {
    if port := os.Getenv("APP_PORT"); port != "" {
        return port
    }
    if port := os.Getenv("PORT"); port != "" {
        return port
    }
    return "8080"
}

func DatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}

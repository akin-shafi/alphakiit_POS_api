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
    port := os.Getenv("PORT")
    if port == "" {
        return "8080"
    }
    return port
}

func DatabaseURL() string {
    return os.Getenv("DATABASE_URL")
}

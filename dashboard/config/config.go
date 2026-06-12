package config

import (
    "os"

    "github.com/joho/godotenv"
)


type Config struct {
	ServerURL string
}

func Load() Config {
	godotenv.Load()
	
	return Config{
		ServerURL: getOrDefault("SERVER_URL", "ws://localhost:8080/ws/dashboard"),
	}
}

func getOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
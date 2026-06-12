package config

import (
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
	Addr string
}

func Load() Config {
	godotenv.Load()
	
	return Config{
		Addr: getOrDefault("ADDR", ":8080"),
	}
}

func getOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
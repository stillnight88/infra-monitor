package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerURL string
	AgentID   string
	Hostname  string
}

func Load() Config {
	godotenv.Load()
	hostname := mustHostname()

	return Config{
		ServerURL: getOrDefault("SERVER_URL", "ws://localhost:8080/ws/agent"),
		AgentID:   getOrDefault("AGENT_ID", hostname),
		Hostname:  hostname,
	}
}

func mustHostname() string {
	h, err := os.Hostname()
	if err != nil {
		slog.Error("hostname lookup failed", "err", err)
		os.Exit(1)
	}
	return h
}

func getOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
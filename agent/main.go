package main

import (
	"log"
	"os"

	"github.com/stillnight88/infra-monitor/agent/ws"
)

func main() {
	serverURL := envOrDefault("SERVER_URL", "ws://localhost:8080/ws/agent")
	agentID := envOrDefault("AGENT_ID", mustHostname())
	hostname := mustHostname()

	log.Printf("starting agent — id: %s  host: %s  server: %s", agentID, hostname, serverURL)

	client, err := ws.New(serverURL, agentID, hostname)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	if err := client.Run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func mustHostname() string {
	h, err := os.Hostname()
	if err != nil {
		log.Fatalf("hostname: %v", err)
	}
	return h
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

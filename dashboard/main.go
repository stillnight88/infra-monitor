package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stillnight88/infra-monitor/dashboard/ui"
	"github.com/stillnight88/infra-monitor/dashboard/ws"
)

func main()  {
	serverURL := envOrDefault("SERVER_URL", "ws://localhost:8080/ws/dashboard")

	client, err := ws.New(serverURL)
	if err != nil {
		log.Fatalf("connect to server: %v", err)
	}

	ch := make(chan ws.SnapshotMsg, 1)

	go client.Listen(ch)

	model := ui.New(ch)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ui error: %v\n", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string  {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
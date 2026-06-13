package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stillnight88/infra-monitor/dashboard/config"
	"github.com/stillnight88/infra-monitor/dashboard/ui"
	"github.com/stillnight88/infra-monitor/dashboard/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := ws.New(cfg.ServerURL)

	ch := make(chan ws.SnapshotMsg, 1)

	go client.Listen(ctx, ch)

	model := ui.New(ch)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ui error: %v\n", err)
		os.Exit(1)
	}

	slog.Info("dashboard stopped cleanly")
}
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/stillnight88/infra-monitor/agent/config"
	"github.com/stillnight88/infra-monitor/agent/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg := config.Load()

	slog.Info("agent starting",
		"agent_id", cfg.AgentID,
		"hostname", cfg.Hostname,
		"server", cfg.ServerURL,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := ws.New(cfg.ServerURL, cfg.AgentID, cfg.Hostname)
	client.Run(ctx)

	slog.Info("agent stopped cleanly")
}

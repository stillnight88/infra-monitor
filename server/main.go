package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stillnight88/infra-monitor/server/config"
	"github.com/stillnight88/infra-monitor/server/handler"
	"github.com/stillnight88/infra-monitor/server/hub"
	"github.com/stillnight88/infra-monitor/server/metrics"
)

const (
	offlineThreshold  = 10 * time.Second
	heartbeatInterval = 5 * time.Second
	shutdownTimeout   = 5 * time.Second
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil))) // Structured logger — JSON

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store := metrics.New()
	h := hub.New()
	agentHandler := handler.New(store, h)
	dashboardHandler := handler.NewDashboard(h)

	// Hub and heartbeat run in their own goroutines.
	go h.Run(ctx)
	go runHeartbeat(ctx, store, h)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/ws/agent", agentHandler.AgentWS(ctx))
	r.GET("/ws/dashboard", dashboardHandler.DashboardWS(ctx))
	r.GET("/state", func(c *gin.Context) {
		c.JSON(http.StatusOK, store.All())
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	go func() {
		slog.Info("server listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "err", err)
	}

	slog.Info("server stopped cleanly")
}

func runHeartbeat(ctx context.Context, store *metrics.Store, h *hub.Hub) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("heartbeat shutting down")
			return

		case <-ticker.C:
			snapshot := store.All()
			changed := false

			for agentID, state := range snapshot {
				if !state.Online {
					continue
				}
				if time.Since(time.Unix(state.Payload.Timestamp, 0)) > offlineThreshold {
					slog.Info("agent offline", "agent", agentID)
					store.MarkOffline(agentID)
					changed = true
				}
			}

			if changed {
				select {
				case h.Broadcast <- store.All():
				default:
					slog.Warn("heartbeat: broadcast channel full — skipping")
				}
			}
		}
	}
}

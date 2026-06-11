package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stillnight88/infra-monitor/server/hub"
	"github.com/stillnight88/infra-monitor/server/metrics"
	"github.com/stillnight88/infra-monitor/shared"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler holds dependencies for agent WebSocket handling.
type Handler struct {
	store *metrics.Store
	hub   *hub.Hub
}

// New returns a Handler wired to the store and hub.
func New(store *metrics.Store, h *hub.Hub) *Handler {
	return &Handler{store: store, hub: h}
}

// AgentWS handles one agent connection, Each agent gets its own goroutine running this function.
func (h *Handler) AgentWS(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			slog.Error("agent upgrade", "err", err)
			return
		}
		defer conn.Close()

		slog.Info("agent connected", "addr", c.Request.RemoteAddr)

		for {
			select {
			case <-ctx.Done():
				slog.Info("agent handler shutting down", "addr", c.Request.RemoteAddr)
				return
			default:
			}

			_, data, err := conn.ReadMessage()
			if err != nil {
				slog.Info("agent disconnected", "addr", c.Request.RemoteAddr, "err", err)
				return
			}

			var payload shared.MetricsPayload
			if err := json.Unmarshal(data, &payload); err != nil {
				slog.Warn("agent unmarshal", "err", err)
				continue
			}

			h.store.Set(payload)

			select {
			case h.hub.Broadcast <- h.store.All():
			default:
				slog.Warn("broadcast channel full — skipping tick", "agent", payload.AgentID)
			}

			slog.Info("metrics received",
				"agent", payload.AgentID,
				"cpu", payload.CPU,
				"ram", payload.RAM,
				"disk", payload.Disk,
			)
		}
	}
}

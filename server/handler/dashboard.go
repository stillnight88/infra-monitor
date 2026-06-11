package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stillnight88/infra-monitor/server/hub"
)

// DashboardHandler handles dashboard WebSocket connections.
type DashboardHandler struct {
	hub *hub.Hub
}

// NewDashboard returns a DashboardHandler wired to the given hub.
func NewDashboard(h *hub.Hub) *DashboardHandler {
	return &DashboardHandler{hub: h}
}

func (d *DashboardHandler) DashboardWS(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			slog.Error("dashboard upgrade", "err", err)
			return
		}

		d.hub.Register(conn)
		defer d.hub.Unregister(conn)
		defer conn.Close()

		slog.Info("dashboard connected", "addr", c.Request.RemoteAddr)

		// server knows when dashboard disconnects when (Browser closes)err != nil becomes true.
		for {
			select {
			case <-ctx.Done():
				slog.Info("dashboard handler shutting down", "addr", c.Request.RemoteAddr)
				return
			default:
			}
			_, _, err := conn.ReadMessage()
			if err != nil {
				slog.Info("dashboard disconnected", "addr", c.Request.RemoteAddr, "err", err)
				return
			}
		}
	}
}

// StateHandler returns a snapshot of all agent states as JSON.
func StateHandler(store interface {
	All() map[string]interface{}
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, store.All())
	}
}

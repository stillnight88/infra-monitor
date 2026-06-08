package handler

import (
	"log"
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

func (d *DashboardHandler) DashboardWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("dashboard upgrade: %v", err)
		return
	}

	d.hub.Register(conn)
	defer d.hub.Unregister(conn)
	defer conn.Close()

	log.Printf("dashboard connected: %s", c.Request.RemoteAddr)

	// server knows when dashboard disconnects when (Browser closes)err != nil becomes true.
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("dashboard disconnected: %v", err)
			return
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

package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stillnight88/infra-monitor/server/metrics"
	"github.com/stillnight88/infra-monitor/shared"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	store *metrics.Store
}

func New(store *metrics.Store) *Handler {
	return &Handler{store: store}
}

// AgentWS handles one agent connection, Each agent gets its own goroutine running this function.
func (h *Handler) AgentWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}
	defer conn.Close()
	log.Printf("agent connected: %s", c.Request.RemoteAddr)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("agent disconnected: %v", err)
			return
		}

		var payload shared.MetricsPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Printf("unmarshal: %v", err)
			continue
		}

		h.store.Set(payload)

		log.Printf("[%s] CPU: %.1f%%  RAM: %.1f%%  Disk: %.1f%%",
			payload.AgentID, payload.CPU, payload.RAM, payload.Disk)
	}
}

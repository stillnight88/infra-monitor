package hub

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/stillnight88/infra-monitor/server/metrics"
)

type Hub struct {
	mu         sync.RWMutex
	dashboards map[*websocket.Conn]struct{}
	Broadcast  chan map[string]metrics.AgentState
}

// New returns an initialised Hub with a buffered broadcast channel.
func New() *Hub {
	return &Hub{
		dashboards: make(map[*websocket.Conn]struct{}),
		Broadcast:  make(chan map[string]metrics.AgentState, 16),
	}
}

// Register adds a dashboard connection to the hub.
func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dashboards[conn] = struct{}{}
	log.Printf("dashboard registered — total: %d", len(h.dashboards))
}

// Unregister removes a dashboard connection from the hub.
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.dashboards, conn)
	log.Printf("dashboard unregistered — total: %d", len(h.dashboards))
}

// Run starts the broadcast loop.
func (h *Hub) Run() {
	for snapshot := range h.Broadcast {
		h.broadcast(snapshot)
	}
}

// broadcast sends the snapshot to every registered dashboard.
func (h *Hub) broadcast(snapshot map[string]metrics.AgentState) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("hub marshal: %v", err)
		return
	}

	h.mu.RLock()

	conns := make([]*websocket.Conn, 0, len(h.dashboards))

	// Copy connections under read lock, don't hold the lock while writing to each conn
	for conn := range h.dashboards {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage,data); err != nil {
			log.Printf("hub write error — removing dashboard: %v", err)
			h.Unregister(conn)
			conn.Close()
		}
	}
}

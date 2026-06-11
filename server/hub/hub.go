package hub

import (
	"context"
	"encoding/json"
	"log/slog"
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
	slog.Info("dashboard registered", "total", len(h.dashboards))
}

// Unregister removes a dashboard connection from the hub.
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.dashboards, conn)
	slog.Info("dashboard unregistered", "total", len(h.dashboards))
}

// Run starts the broadcast loop.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("hub shutting down")
			h.closeAll()
			return

		case snapshot, ok := <-h.Broadcast:
			if !ok {
				return
			}
			h.broadcast(snapshot)
		}
	}
}

// broadcast sends the snapshot to every registered dashboard.
func (h *Hub) broadcast(snapshot map[string]metrics.AgentState) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		slog.Error("hub marshal", "err", err)
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
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			slog.Warn("hub write failed — removing dashboard", "err", err)
			h.Unregister(conn)
			conn.Close()
		}
	}
}

func (h *Hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.dashboards {
		conn.Close()
		delete(h.dashboards, conn)
	}
}

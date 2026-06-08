package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stillnight88/infra-monitor/server/handler"
	"github.com/stillnight88/infra-monitor/server/hub"
	"github.com/stillnight88/infra-monitor/server/metrics"
)

const offlineThreshold = 10 * time.Second
const heartbeatInterval = 5 * time.Second

func main() {
	store := metrics.New()
	h := hub.New()
	agentHandler := handler.New(store, h)
	dashboardHandler := handler.NewDashboard(h)

	// Hub broadcast loop — must run before any connections arrive.
	go h.Run()

	// Heartbeat — checks for stale agents every 5 seconds.
	go runHeartbeat(store, h)

	r := gin.Default()

	r.GET("/ws/agent", agentHandler.AgentWS)
	r.GET("/ws/dashboard", dashboardHandler.DashboardWS)
	r.GET("/state", func(c *gin.Context) {
		c.JSON(http.StatusOK, store.All())
	})

	log.Println("server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func runHeartbeat(store *metrics.Store, h *hub.Hub) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		snapshot := store.All()
		changed := false

		for agentID, state := range snapshot {
			if !state.Online {
				continue
			}
			if time.Since(time.Unix(state.Payload.Timestamp, 0)) > offlineThreshold {
				log.Printf("agent offline: %s", agentID)
				store.MarkOffline(agentID)
				changed = true
			}
		}

		if changed {
			select {
			case h.Broadcast <- store.All():
			default:
				log.Printf("heartbeat: broadcast channel full — skipping")
			}
		}
	}
}

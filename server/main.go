package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stillnight88/infra-monitor/server/handler"
	"github.com/stillnight88/infra-monitor/server/metrics"
)

func main() {
	store := metrics.New()
	h := handler.New(store)

	r := gin.Default()

	r.GET("/ws/agent", h.AgentWS)

	r.GET("/state", func(c *gin.Context) {
		c.JSON(http.StatusOK, store.All())
	})

	log.Println("server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server: %v", err)
	}
}

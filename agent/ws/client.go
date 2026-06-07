package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stillnight88/infra-monitor/agent/collector"
	"github.com/stillnight88/infra-monitor/shared"
)

const sendInterval = 2 * time.Second

type Client struct {
	agentID string
	hostname string
	conn *websocket.Conn
}

// New dials the server and returns a ready Client.
func New(serverURL, agentID, hostname string) (*Client, error) {
	con,_,err := websocket.DefaultDialer.Dial(serverURL,nil)
	if err != nil {
		return  nil, fmt.Errorf("dial %s: %w", serverURL, err)
	}

	return &Client{
		agentID: agentID,
		hostname: hostname,
		conn: con,
	}, nil
}

// Run starts the send loop.
// Collects metrics every 2 seconds and sends them to the server.
// Blocks until an error occurs.
func (c *Client) Run() error {
	defer c.conn.Close()

	ticker := time.NewTicker(sendInterval)
	defer ticker.Stop()

	for range ticker.C {
		payload, err := collector.Collect(c.agentID, c.hostname)
		if err != nil {
			log.Printf("collect error: %v", err)
			continue 
		}

		if err := c.send(payload); err != nil {
			return fmt.Errorf("send: %w", err)
		}

		log.Printf("sent — CPU: %.1f%%  RAM: %.1f%%  Disk: %.1f%%", payload.CPU, payload.RAM, payload.Disk)
	}
	return  nil
}

// send encodes the payload as JSON and writes it to the connection.
func (c *Client) send(payload shared.MetricsPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}
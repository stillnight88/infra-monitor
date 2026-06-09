package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stillnight88/infra-monitor/server/metrics"
)

// SnapshotMsg is what the dashboard receives on every server broadcast.
type SnapshotMsg map[string]metrics.AgentState

// Client holds the WebSocket connection to the server.
type Client struct {
	conn *websocket.Conn
}

func New(serverURL string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", serverURL, err)
	}

	return &Client{conn: conn}, nil
}

// Listen reads incoming snapshots and pushes them into the provided channel.
func (c *Client) Listen(ch chan<- SnapshotMsg) {
	defer c.conn.Close()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("dashboard ws read error: %v", err)
			close(ch)
			return
		}

		var snapshot SnapshotMsg
		if err := json.Unmarshal(data, &snapshot); err != nil {
			log.Printf("dashboard unmarshal: %v", err)
			continue
		}

		ch <- snapshot
	}
}

func LastSeen(ts int64) string {
	if ts == 0 {
		return "never"
	}
	d := time.Since(time.Unix(ts, 0)).Round(time.Second)
	return fmt.Sprintf("%v ago", d)
}

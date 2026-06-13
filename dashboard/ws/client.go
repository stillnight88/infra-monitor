package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stillnight88/infra-monitor/server/metrics"
)

const (
	initialBackoff = 2 * time.Second
	maxBackoff     = 30 * time.Second
)

// SnapshotMsg is what the dashboard receives on every server broadcast.
type SnapshotMsg map[string]metrics.AgentState

// Client holds the server address and manages the WebSocket connection.
type Client struct {
	serverURL string
}

func New(serverURL string) *Client {
	return &Client{serverURL: serverURL}
}

// Listen connects to the server and pushes snapshots into ch.
func (c *Client) Listen(ctx context.Context, ch chan<- SnapshotMsg) {
	defer close(ch)

	backoff := initialBackoff

	for {
		select {
		case <-ctx.Done():
			slog.Info("dashboard ws shutting down")
			return
		default:
		}

		slog.Info("dashboard connecting to server", "url", c.serverURL)

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.serverURL, nil)
		if err != nil {
			slog.Warn("dashboard connection failed — retrying",
				"err", err,
				"backoff", backoff,
			)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		slog.Info("dashboard connected to server")
		backoff = initialBackoff

		if err := c.readLoop(ctx, conn, ch); err != nil {
			slog.Warn("dashboard read loop exited", "err", err)
		}
	}
}

// readLoop reads snapshots from an established connection.
func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn, ch chan<- SnapshotMsg) error {
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"),
			)
			return nil
		default:
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		var snapshot SnapshotMsg
		if err := json.Unmarshal(data, &snapshot); err != nil {
			slog.Warn("dashboard unmarshal", "err", err)
			continue
		}

		select {
		case ch <- snapshot:
		default:
		}
	}
}

func LastSeen(ts int64) string {
	if ts == 0 {
		return "never"
	}
	d := time.Since(time.Unix(ts, 0)).Round(time.Second)
	return fmt.Sprintf("%v ago", d)
}

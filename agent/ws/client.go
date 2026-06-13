package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stillnight88/infra-monitor/agent/collector"
	"github.com/stillnight88/infra-monitor/shared"
)

const (
	sendInterval   = 2 * time.Second
	initialBackoff = 2 * time.Second
	maxBackoff     = 30 * time.Second
)

// Client holds agent identity and server address.
type Client struct {
	agentID   string
	hostname  string
	serverURL string
}

func New(serverURL, agentID, hostname string) *Client {
	return &Client{
		agentID:   agentID,
		hostname:  hostname,
		serverURL: serverURL,
	}
}

// Run connects to the server and starts the send loop.
// If the connection drops, it reconnects with exponential backoff.
func (c *Client) Run(ctx context.Context) {
	backoff := initialBackoff

	for {
		select {
		case <-ctx.Done():
			slog.Info("agent shutting down")
			return
		default:
		}

		slog.Info("connecting to server", "url", c.serverURL)

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.serverURL, nil)
		if err != nil {
			slog.Warn("connection failed — retrying",
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

		slog.Info("connected to server", "url", c.serverURL)

		backoff = initialBackoff // Reset backoff on successful connection.
		if err := c.runLoop(ctx, conn); err != nil {
			slog.Warn("send loop exited", "err", err)
		}
	}
}

// runLoop runs the send ticker for an established connection.
func (c *Client) runLoop(ctx context.Context, conn *websocket.Conn) error {
	defer conn.Close()

	ticker := time.NewTicker(sendInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"),
			)
			return nil

		case <-ticker.C:
			payload, err := collector.Collect(c.agentID, c.hostname)
			if err != nil {
				slog.Warn("collect error", "err", err)
				continue
			}

			if err := send(conn, payload); err != nil {
				return fmt.Errorf("send: %w", err)
			}

			slog.Info("metrics sent",
				"agent", payload.AgentID,
				"cpu", payload.CPU,
				"ram", payload.RAM,
				"disk", payload.Disk,
			)
		}
	}
}

// send encodes the payload as JSON and writes it to the connection.
func send(conn *websocket.Conn, payload shared.MetricsPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}
package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/coder/websocket"
	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

// Connection manages the WebSocket connection to Discord
type Connection struct {
	conn   *websocket.Conn
	logger logger.Logger
	mu     sync.Mutex
}

// NewConnection creates a new connection manager
func NewConnection(logger logger.Logger) *Connection {
	return &Connection{
		logger: logger,
	}
}

// Connect establishes a WebSocket connection
func (c *Connection) Connect(ctx context.Context, url string) (*websocket.Conn, error) {
	conn, resp, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	if resp.StatusCode != 101 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	c.logger.Info("WebSocket connection established")
	return conn, nil
}

// ReadMessages continuously reads messages from the WebSocket
func (c *Connection) ReadMessages(ctx context.Context, eventChan chan<- *Event) {
	defer c.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		messageType, reader, err := c.conn.Reader(ctx)
		if err != nil {
			c.logger.Error("Failed to read message", logger.Field{Key: "error", Value: err})
			return
		}

		if messageType != websocket.MessageText {
			continue
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			c.logger.Error("Failed to read message data", logger.Field{Key: "error", Value: err})
			return
		}

		event, err := c.parseEvent(data)
		if err != nil {
			c.logger.Error("Failed to parse event", logger.Field{Key: "error", Value: err})
			continue
		}

		select {
		case eventChan <- event:
		case <-ctx.Done():
			return
		}
	}
}

// WriteMessage sends a message over the WebSocket
func (c *Connection) WriteMessage(ctx context.Context, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection not established")
	}

	if err := c.conn.Write(ctx, websocket.MessageText, data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// Close closes the WebSocket connection
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.CloseNow()
		c.conn = nil
	}
}

// parseEvent parses raw message data into an Event
func (c *Connection) parseEvent(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &event, nil
}

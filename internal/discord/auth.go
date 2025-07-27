package discord

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

// authenticate sends an IDENTIFY payload to Discord
func (c *Client) authenticate(ctx context.Context) error {
	identifyPayload := struct {
		Op   int          `json:"op"`
		Data IdentifyData `json:"d"`
	}{
		Op: OpIdentify,
		Data: IdentifyData{
			Token: c.token,
			Properties: ConnectionProperties{
				OS:      "linux",
				Browser: "discord-bot",
				Device:  "discord-bot",
			},
			Intents: 67584, // MESSAGE_CONTENT + GUILD_MESSAGES + DIRECT_MESSAGES
		},
	}

	data, err := json.Marshal(identifyPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal identify payload: %w", err)
	}

	if err := c.connection.WriteMessage(ctx, data); err != nil {
		return fmt.Errorf("failed to send identify: %w", err)
	}

	c.logger.Info("Sent IDENTIFY payload")
	return nil
}

// handleHello processes the HELLO event from Discord
func (c *Client) handleHello(ctx context.Context, event *Event) error {
	var helloData HelloData
	if err := json.Unmarshal(event.Data, &helloData); err != nil {
		return fmt.Errorf("failed to unmarshal HELLO data: %w", err)
	}

	c.logger.Info("Received HELLO",
		logger.Field{Key: "heartbeat_interval", Value: helloData.HeartbeatInterval})

	// Set heartbeat interval
	c.heartbeat.SetInterval(helloData.HeartbeatInterval)

	return nil
}

// handleDispatch processes DISPATCH events from Discord
func (c *Client) handleDispatch(ctx context.Context, event *Event) error {
	// Update sequence number
	if event.Sequence != nil {
		c.state.UpdateSequence(*event.Sequence)
		c.heartbeat.UpdateSequence(*event.Sequence)
	}

	switch event.Type {
	case "READY":
		return c.handleReady(ctx, event)
	case "MESSAGE_CREATE":
		return c.handleMessage(ctx, event)
	default:
		c.logger.Debug("Unhandled dispatch event",
			logger.Field{Key: "type", Value: event.Type})
	}

	return nil
}

// handleReady processes the READY event
func (c *Client) handleReady(ctx context.Context, event *Event) error {
	var readyData ReadyData
	if err := json.Unmarshal(event.Data, &readyData); err != nil {
		return fmt.Errorf("failed to unmarshal READY data: %w", err)
	}

	// Update client state
	c.state.SetSessionInfo(readyData.SessionID, readyData.ResumeURL)
	c.state.SetAuthenticated(true)

	c.logger.Info("Bot is ready",
		logger.Field{Key: "user", Value: readyData.User.Username},
		logger.Field{Key: "session_id", Value: readyData.SessionID})

	return nil
}

// handleMessage processes MESSAGE_CREATE events
func (c *Client) handleMessage(ctx context.Context, event *Event) error {
	// This is where you'd handle incoming messages
	// For now, just log that we received a message
	c.logger.Debug("Received message event")
	return nil
}

// handleInvalidSession handles INVALID_SESSION events
func (c *Client) handleInvalidSession(ctx context.Context, event *Event) error {
	var canResume bool
	if err := json.Unmarshal(event.Data, &canResume); err != nil {
		return fmt.Errorf("failed to unmarshal INVALID_SESSION data: %w", err)
	}

	if canResume {
		c.logger.Info("Session invalid but resumable, attempting resume")
		return fmt.Errorf("session invalid but resumable")
	}

	c.logger.Info("Session invalid and not resumable, need fresh connection")
	c.state.SetAuthenticated(false)
	return fmt.Errorf("session invalid and not resumable")
}

// ConnectionProperties represents Discord connection properties
type ConnectionProperties struct {
	OS      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

// IdentifyData represents the data sent in an IDENTIFY payload
type IdentifyData struct {
	Token      string               `json:"token"`
	Properties ConnectionProperties `json:"properties"`
	Intents    int                  `json:"intents"`
}

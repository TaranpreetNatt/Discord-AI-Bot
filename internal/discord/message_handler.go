package discord

import (
	"context"
	"encoding/json"

	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

// MessageHandler handles different types of Discord messages
type MessageHandler struct {
	logger logger.Logger
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(logger logger.Logger) *MessageHandler {
	return &MessageHandler{
		logger: logger,
	}
}

// HandleMessage processes a MESSAGE_CREATE event
func (m *MessageHandler) HandleMessage(ctx context.Context, event *Event) error {
	var message Message
	if err := json.Unmarshal(event.Data, &message); err != nil {
		return err
	}

	m.logger.Info("Received message",
		logger.Field{Key: "author", Value: message.Author.Username},
		logger.Field{Key: "content", Value: message.Content},
		logger.Field{Key: "channel_id", Value: message.ChannelID})

	// Here you would add your message processing logic
	// For example, responding to commands, AI processing, etc.

	return nil
}

// Message represents a Discord message
type Message struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id,omitempty"`
	Author    User   `json:"author"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Type      int    `json:"type"`
}

// Channel represents a Discord channel
type Channel struct {
	ID      string `json:"id"`
	Type    int    `json:"type"`
	GuildID string `json:"guild_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Topic   string `json:"topic,omitempty"`
}

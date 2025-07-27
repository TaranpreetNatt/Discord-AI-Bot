package discord

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

// ResumeManager handles session resumption
type ResumeManager struct {
	logger logger.Logger
}

// NewResumeManager creates a new resume manager
func NewResumeManager(logger logger.Logger) *ResumeManager {
	return &ResumeManager{
		logger: logger,
	}
}

// Resume attempts to resume a Discord session
func (r *ResumeManager) Resume(ctx context.Context, conn *Connection, token, sessionID string, sequence int) error {
	resumePayload := struct {
		Op   int        `json:"op"`
		Data ResumeData `json:"d"`
	}{
		Op: OpResume,
		Data: ResumeData{
			Token:     token,
			SessionID: sessionID,
			Sequence:  sequence,
		},
	}
	
	data, err := json.Marshal(resumePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal resume payload: %w", err)
	}
	
	if err := conn.WriteMessage(ctx, data); err != nil {
		return fmt.Errorf("failed to send resume: %w", err)
	}
	
	r.logger.Info("Sent RESUME payload", 
		logger.Field{Key: "session_id", Value: sessionID},
		logger.Field{Key: "sequence", Value: sequence})
	
	return nil
}

// ResumeData represents the data sent in a RESUME payload
type ResumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Sequence  int    `json:"seq"`
}

// internal/discord/errors.go
package discord

import (
	"fmt"
)

// DiscordError represents a Discord-specific error
type DiscordError struct {
	Code    int
	Message string
	Retry   bool
}

func (e *DiscordError) Error() string {
	return fmt.Sprintf("discord error %d: %s", e.Code, e.Message)
}

// CanRetry returns whether this error allows for retry
func (e *DiscordError) CanRetry() bool {
	return e.Retry
}

// Gateway close codes from Discord documentation
var GatewayCloseErrors = map[int]*DiscordError{
	4000: {Code: 4000, Message: "Unknown error", Retry: true},
	4001: {Code: 4001, Message: "Unknown opcode", Retry: true},
	4002: {Code: 4002, Message: "Decode error", Retry: true},
	4003: {Code: 4003, Message: "Not authenticated", Retry: true},
	4004: {Code: 4004, Message: "Authentication failed", Retry: false},
	4005: {Code: 4005, Message: "Already authenticated", Retry: true},
	4007: {Code: 4007, Message: "Invalid seq", Retry: true},
	4008: {Code: 4008, Message: "Rate limited", Retry: true},
	4009: {Code: 4009, Message: "Session timed out", Retry: true},
	4010: {Code: 4010, Message: "Invalid shard", Retry: false},
	4011: {Code: 4011, Message: "Sharding required", Retry: false},
	4012: {Code: 4012, Message: "Invalid API version", Retry: false},
	4013: {Code: 4013, Message: "Invalid intent(s)", Retry: false},
	4014: {Code: 4014, Message: "Disallowed intent(s)", Retry: false},
}

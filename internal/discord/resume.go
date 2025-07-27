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

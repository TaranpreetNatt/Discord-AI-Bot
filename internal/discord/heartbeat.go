package discord

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

// HeartbeatManager handles Discord heartbeat protocol
type HeartbeatManager struct {
	logger   logger.Logger
	interval time.Duration
	ticker   *time.Ticker
	ackTimer *time.Timer
	sequence *int
	mu       sync.RWMutex

	// State
	started       bool
	ackReceived   bool
	missedAcks    int
	maxMissedAcks int
}

// NewHeartbeatManager creates a new heartbeat manager
func NewHeartbeatManager(logger logger.Logger) *HeartbeatManager {
	return &HeartbeatManager{
		logger:        logger,
		maxMissedAcks: 3,
	}
}

// Start begins the heartbeat process
func (h *HeartbeatManager) Start(ctx context.Context, conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return
	}

	h.started = true
	go h.run(ctx, conn)
}

// SetInterval sets the heartbeat interval from Discord's HELLO event
func (h *HeartbeatManager) SetInterval(intervalMs int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add jitter to prevent thundering herd
	jitter := rand.Float64()
	h.interval = time.Duration(float64(intervalMs)*jitter) * time.Millisecond

	h.logger.Info("Heartbeat interval set",
		logger.Field{Key: "interval", Value: h.interval})
}

// HandleAck handles heartbeat acknowledgment from Discord
func (h *HeartbeatManager) HandleAck() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.ackReceived = true
	h.missedAcks = 0

	if h.ackTimer != nil {
		h.ackTimer.Stop()
	}
}

// UpdateSequence updates the sequence number for heartbeats
func (h *HeartbeatManager) UpdateSequence(seq int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sequence = &seq
}

// run is the main heartbeat loop
func (h *HeartbeatManager) run(ctx context.Context, conn *Connection) {
	defer h.cleanup()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.ticker.C:
			if err := h.sendHeartbeat(ctx, conn); err != nil {
				h.logger.Error("Failed to send heartbeat", logger.Field{Key: "error", Value: err})
				return
			}
		}
	}
}

// sendHeartbeat sends a heartbeat message to Discord
func (h *HeartbeatManager) sendHeartbeat(ctx context.Context, conn *Connection) error {
	h.mu.RLock()
	sequence := h.sequence
	h.mu.RUnlock()

	heartbeat := struct {
		Op   int  `json:"op"`
		Data *int `json:"d"`
	}{
		Op:   OpHeartbeat,
		Data: sequence,
	}

	data, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	if err := conn.WriteMessage(ctx, data); err != nil {
		return err
	}

	// Start ack timeout
	h.startAckTimeout(ctx)

	h.logger.Debug("Heartbeat sent")
	return nil
}

// startAckTimeout starts a timeout for heartbeat acknowledgment
func (h *HeartbeatManager) startAckTimeout(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ackTimer != nil {
		h.ackTimer.Stop()
	}

	h.ackReceived = false
	h.ackTimer = time.AfterFunc(30*time.Second, func() {
		h.handleMissedAck(ctx)
	})
}

// handleMissedAck handles a missed heartbeat acknowledgment
func (h *HeartbeatManager) handleMissedAck(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ackReceived {
		return // Ack was received, ignore timeout
	}

	h.missedAcks++
	h.logger.Warn("Missed heartbeat ack",
		logger.Field{Key: "count", Value: h.missedAcks})

	if h.missedAcks >= h.maxMissedAcks {
		h.logger.Error("Too many missed heartbeat acks, connection likely dead")
		// Signal that we need to reconnect
		// This would be handled by the main client loop
	}
}

// cleanup cleans up resources
func (h *HeartbeatManager) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ticker != nil {
		h.ticker.Stop()
	}

	if h.ackTimer != nil {
		h.ackTimer.Stop()
	}

	h.started = false
}

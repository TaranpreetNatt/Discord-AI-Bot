package discord

import (
	"encoding/json"
)

// Discord Gateway opcodes
const (
	OpDispatch = iota
	OpHeartbeat
	OpIdentify
	OpPresenceUpdate
	OpVoiceStateUpdate
	_
	OpResume
	OpReconnect
	OpRequestGuildMembers
	OpInvalidSession
	OpHello
	OpHeartbeatAck
)

// Event represents a Discord Gateway event
type Event struct {
	Op       int             `json:"op"`
	Data     json.RawMessage `json:"d"`
	Sequence *int            `json:"s,omitempty"`
	Type     string          `json:"t,omitempty"`
}

// HelloData represents the HELLO event data
type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// ReadyData represents the READY event data
type ReadyData struct {
	Version     int    `json:"v"`
	User        User   `json:"user"`
	SessionID   string `json:"session_id"`
	ResumeURL   string `json:"resume_gateway_url"`
	Application struct {
		ID    string `json:"id"`
		Flags int    `json:"flags"`
	} `json:"application"`
}

// User represents a Discord user
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
}


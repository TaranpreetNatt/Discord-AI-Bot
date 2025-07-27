package discord

import (
	"sync"
)

// ClientState holds the current state of the Discord client
type ClientState struct {
	mu sync.RWMutex

	// Connection state
	SessionID     string
	ResumeURL     string
	Sequence      int
	Authenticated bool

	// User state
	User *User

	// Gateway state
	HeartbeatInterval int
}

// NewClientState creates a new client state
func NewClientState() *ClientState {
	return &ClientState{}
}

// SetSessionInfo updates session information
func (s *ClientState) SetSessionInfo(sessionID, resumeURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.SessionID = sessionID
	s.ResumeURL = resumeURL
}

// UpdateSequence updates the sequence number
func (s *ClientState) UpdateSequence(seq int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Sequence = seq
}

// GetSequence returns the current sequence number
func (s *ClientState) GetSequence() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Sequence
}

// SetAuthenticated sets the authentication status
func (s *ClientState) SetAuthenticated(auth bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Authenticated = auth
}

// IsAuthenticated returns the authentication status
func (s *ClientState) IsAuthenticated() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Authenticated
}

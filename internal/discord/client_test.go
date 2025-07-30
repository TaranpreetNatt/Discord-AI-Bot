package discord

import (
	"testing"
)

// Test NewClient
/*
	- Happy path
	- What happens if the token is empty? Should not be empty
	- What happens if the url is empty? Should not be empty
	- What happens if the logger is nil?
	- Is the gateway being initialized correctly?
	- Is the connection being initialized correctly?
	- is the heartbreat being initialized correctly?
*/

func TestNewClient_Success(t *testing.T) {
	validToken := "valid_token"
	validApiBase := "valid_apiBase"
	l := setupLogger(t)

	client := NewClient(validToken, validApiBase, l)

	if client.token != validToken {
		t.Fatalf("Expected token %s, got %s", validToken, client.token)
	}

	if client.apiBase != validApiBase {
		t.Fatalf("Expected apiBase %s, got %s", validApiBase, validApiBase)
	}

	if client.gateway == nil {
		t.Fatal("Gateway should be initialized")
	}
	if client.connection == nil {
		t.Fatal("Connection should be initialized")
	}
	if client.heartbeat == nil {
		t.Fatal("Heartbeat should be initialized")
	}

	if client.state == nil {
		t.Fatal("State should be initialized")
	}

	if client.eventChan == nil {
		t.Fatal("Event channel should be initialized")
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

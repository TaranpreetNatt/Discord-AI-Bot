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

func TestNewClient_Inputs(t *testing.T) {
	validToken := "valid_token"
	validApiBase := "valid_apiBase"
	l := setupLogger(t)

	client, err := NewClient(validToken, validApiBase, l)
	if err != nil {
		t.Fatalf("Expected error to be nil, %s", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if client.token != validToken {
		t.Fatalf("Expected token %s, got %s", validToken, client.token)
	}

	if client.apiBase != validApiBase {
		t.Fatalf("Expected apiBase %s, got %s", validApiBase, client.apiBase)
	}
}

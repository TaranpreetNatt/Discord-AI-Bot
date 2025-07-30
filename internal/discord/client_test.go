package discord

import (
	"context"
	"testing"
)

func TestNewClient_Success(t *testing.T) {
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

func TestClient_Start_Success(t *testing.T) {
	ctx := context.Background()

	token := "valid_token"
	config := MockConfig{
		ValidToken:      token,
		StatusCode:      200,
		ReturnValidJson: true,
	}

	apiBase := setupMockServer(t, config).URL
	client, _ := NewClient(token, apiBase, setupLogger(t))

	err := client.Start(ctx)
	if err != nil {
		t.Fatalf("Expected client to start without error, %s", err)
	}
}

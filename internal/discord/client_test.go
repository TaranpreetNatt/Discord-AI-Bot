package discord

import (
	"context"
	"testing"
	"time"
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

func TestClient_Integration_Start(t *testing.T) {
	ctx := context.Background()

	mockWsServer := NewWebSocketServer(t)
	defer mockWsServer.Close()

	config := MockConfig{
		ValidToken:      "valid_token",
		StatusCode:      200,
		ReturnValidJson: true,
		WebSocketURL:    mockWsServer.URL(),
	}

	server := setupMockServer(t, config)

	client, err := NewClient("valid_token", server.URL, setupLogger(t))
	if err != nil {
		t.Fatalf("Expected client to be created: %v", err)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Start(ctx)
	}()

	time.Sleep(1 * time.Second)
	select {
	case err := <-errChan:
		t.Fatalf("Error in client startup, err: %v", err)
	default:
		t.Log("SUCCESS: client.Start() completed startup and is running")
	}
	mockWsServer.Close()
}

func TestClient_Integration_GatewayUrlFailure(t *testing.T) {
	ctx := context.Background()

	config := MockConfig{
		ValidToken:      "valid_token",
		StatusCode:      500,
		ReturnValidJson: true,
		WebSocketURL:    "wss://test.gg",
	}

	server := setupMockServer(t, config)

	client, err := NewClient("valid_token", server.URL, setupLogger(t))
	if err != nil {
		t.Fatalf("Expected client to be created: %v", err)
	}

	if err := client.Start(ctx); err == nil {
		t.Fatalf("Expected an error when server returns 500, %v", err)
	}
}

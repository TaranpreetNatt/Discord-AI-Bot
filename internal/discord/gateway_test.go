package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

/*
	- Test the happy path, what happens if everything works
	- Test what happens if the token is not the correct token?
	- What happens if the server returns an invalid json response?
	- What happens if the server doesn't return anything and times out?
	- What happens if the server returns 500?
	- What happens if the context is cancelled?
	- What happens if the context times out?
	- What happens if the apiBase is wrong?
*/

type MockConfig struct {
	ValidToken      string
	StatusCode      int
	ReturnValidJson bool
	ResponseBody    string // for custom invalid JSON
	Delay           time.Duration
}

func MockServer(config MockConfig) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

		if config.StatusCode == 500 {
			writer.WriteHeader(config.StatusCode)
			return
		}

		if config.Delay > 0 {
			time.Sleep(config.Delay)
		}

		token := req.Header.Get("Authorization")
		expectedToken := "Bot " + config.ValidToken

		if token != expectedToken {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		if config.ReturnValidJson {
			response := GatewayResponse{
				URL:    "wss://gateway.discord.gg",
				Shards: 9,
				SessionStartLimit: SessionStartLimit{
					Total:          1000,
					Remaining:      999,
					ResetAfter:     14400000,
					MaxConcurrency: 1,
				},
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(config.StatusCode)
			json.NewEncoder(writer).Encode(response)

		} else if config.ResponseBody != "" {
			writer.Write([]byte(config.ResponseBody))
		}
	}))
	return server
}

func TestGateway_GetURL_Success(t *testing.T) {
	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      200,
		ReturnValidJson: true,
	}

	server := MockServer(config)
	server.Start()
	defer server.Close()

	l, _ := logger.NewZapAdapter()
	correctURL := "wss://gateway.discord.gg/?v=10&encoding=json"
	ctx := context.Background()
	gateway := NewGateway(server.URL, l)

	url, err := gateway.GetURL(ctx, config.ValidToken)
	if err != nil {
		t.Fatalf("Got error %v\n", err)
	}
	if url != correctURL {
		t.Fatalf("wanted %s, got %s\n", correctURL, url)
	}
}

func TestGateway_Invalid_Token(t *testing.T) {
	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      401,
		ReturnValidJson: true,
	}
	server := MockServer(config)
	server.Start()
	defer server.Close()

	l, _ := logger.NewZapAdapter()
	ctx := context.Background()
	token := "incorrect_token"
	gateway := NewGateway(server.URL, l)

	url, err := gateway.GetURL(ctx, token)
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}

	if url != "" {
		t.Fatalf("Expected URL to be empty for invalid token, %s", url)
	}
}

func TestGateway_Invalid_Json(t *testing.T) {
	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      200,
		ReturnValidJson: false,
		ResponseBody:    `{ "Broken Json"`,
	}
	server := MockServer(config)
	server.Start()
	defer server.Close()

	l, _ := logger.NewZapAdapter()
	ctx := context.Background()
	token := "correct_token"
	gateway := NewGateway(server.URL, l)

	url, err := gateway.GetURL(ctx, token)
	if err == nil {
		t.Fatal("Expected error with invalid returned Json")
	}

	if url != "" {
		t.Fatal("Expected url to be empty with invalid JSON")
	}
}

func TestGateway_Timeout(t *testing.T) {
	timeout := 100 * time.Millisecond

	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      200,
		ReturnValidJson: true,
		Delay:           2 * timeout,
	}
	server := MockServer(config)
	server.Start()
	defer server.Close()

	l, _ := logger.NewZapAdapter()
	ctx := context.Background()
	token := "correct_token"
	gateway := NewGatewayWithTimeout(server.URL, l, timeout)

	url, err := gateway.GetURL(ctx, token)
	if err == nil {
		t.Fatal("Expected an error when client times out")
	}

	if url != "" {
		t.Fatal("Expected url to be empty when client times out")
	}
}

func TestGateway_ServerError(t *testing.T) {
	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      500,
		ReturnValidJson: true,
	}
	server := MockServer(config)
	server.Start()
	defer server.Close()

	l, _ := logger.NewZapAdapter()
	ctx := context.Background()
	token := "correct_token"
	gateway := NewGateway(server.URL, l)

	url, err := gateway.GetURL(ctx, token)
	if err == nil {
		t.Fatal("Expected an error when server returns 500")
	}

	if url != "" {
		t.Fatal("Expected url to be empty when server returns 500")
	}
}

func TestGateway_Context(t *testing.T) {
	l, _ := logger.NewZapAdapter()
	timeout := 100 * time.Millisecond

	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      200,
		ReturnValidJson: true,
		Delay:           2 * timeout,
	}
	server := MockServer(config)
	server.Start()
	defer server.Close()

	ctx := context.Background()
	ctxTimeout, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	token := "correct_token"
	gateway := NewGateway(server.URL, l)
	url, err := gateway.GetURL(ctxTimeout, token)

	if err == nil {
		t.Fatal("Expected an error with context timeout")
	}

	if url != "" {
		t.Fatal("Expected url to be empty when context timeouts before server returns data")
	}
}

func TestGateway_Wrong_ApiBase(t *testing.T) {
	l, _ := logger.NewZapAdapter()

	ctx := context.Background()
	token := "correct_token"
	wrongUrl := "https://nonexistent-domain-12345.com"
	gateway := NewGateway(wrongUrl, l)

	url, err := gateway.GetURL(ctx, token)
	if err == nil {
		t.Fatal("Expected an error with wrong apiBase specifed")
	}

	if url != "" {
		t.Fatal("Expected url to be empty when wrong apiBase is supplied")
	}
}

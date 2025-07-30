package discord

import (
	"context"
	"testing"
	"time"
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

func assertError(t *testing.T, url string, err error, scenario string) {
	if err == nil {
		t.Fatalf("Expected an error for %s, ", scenario)
	}
	if url != "" {
		t.Fatalf("Expected url to be empty for %s, got %s", scenario, url)
	}
}

func TestGateway_GetURL_Success(t *testing.T) {
	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      200,
		ReturnValidJson: true,
	}

	server := setupMockServer(t, config)
	l := setupLogger(t)

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

	server := setupMockServer(t, config)
	l := setupLogger(t)
	ctx := context.Background()
	token := "incorrect_token"
	gateway := NewGateway(server.URL, l)

	url, err := gateway.GetURL(ctx, token)
	assertError(t, url, err, "invalid token")
}

func TestGateway_Invalid_Json(t *testing.T) {
	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      200,
		ReturnValidJson: false,
		ResponseBody:    `{ "Broken Json"`,
	}
	server := setupMockServer(t, config)

	l := setupLogger(t)
	ctx := context.Background()
	token := "correct_token"
	gateway := NewGateway(server.URL, l)

	url, err := gateway.GetURL(ctx, token)
	assertError(t, url, err, "Invalid Json")
}

func TestGateway_Timeout(t *testing.T) {
	timeout := 100 * time.Millisecond

	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      200,
		ReturnValidJson: true,
		Delay:           2 * timeout,
	}
	server := setupMockServer(t, config)

	l := setupLogger(t)
	ctx := context.Background()
	token := "correct_token"
	gateway := NewGatewayWithTimeout(server.URL, l, timeout)

	url, err := gateway.GetURL(ctx, token)
	assertError(t, url, err, "Gateway timeout")
}

func TestGateway_ServerError(t *testing.T) {
	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      500,
		ReturnValidJson: true,
	}
	server := setupMockServer(t, config)

	l := setupLogger(t)
	ctx := context.Background()
	token := "correct_token"
	gateway := NewGateway(server.URL, l)

	url, err := gateway.GetURL(ctx, token)
	assertError(t, url, err, "server error")
}

func TestGateway_Context_Cancel(t *testing.T) {
	l := setupLogger(t)
	timeout := 100 * time.Millisecond

	config := MockConfig{
		ValidToken:      "correct_token",
		StatusCode:      200,
		ReturnValidJson: true,
		Delay:           2 * timeout,
	}
	server := setupMockServer(t, config)

	ctx := context.Background()
	ctxTimeout, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	token := "correct_token"
	gateway := NewGateway(server.URL, l)
	url, err := gateway.GetURL(ctxTimeout, token)

	assertError(t, url, err, "context cancellation")
}

func TestGateway_Wrong_ApiBase(t *testing.T) {
	l := setupLogger(t)

	ctx := context.Background()
	token := "correct_token"
	wrongUrl := "https://nonexistent-domain-12345.com"
	gateway := NewGateway(wrongUrl, l)

	url, err := gateway.GetURL(ctx, token)
	assertError(t, url, err, "wrong apiBase")
}

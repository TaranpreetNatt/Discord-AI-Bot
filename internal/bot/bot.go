package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	// "github.com/coder/websocket"
)

const (
	discordApiVersion = "/?v=10&encoding=json"
)

type SessionStartLimit struct {
	Total          int `json:"total"`
	Remaining      int `json:"remaining"`
	ResetAfter     int `json:"reset_after"`
	MaxConcurrency int `json:"max_concurrency"`
}

type GatewayResponse struct {
	URL               string            `json:"url"`
	Shards            int               `json:"shards"`
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

func GetGatewayUrl(ctx context.Context, botToken, apiBase string, client *http.Client) (string, error) {

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiBase+"/gateway/bot", nil)
	if reqErr != nil {
		return "", reqErr
	}
	req.Header.Set("Authorization", "Bot "+botToken)

	if client == nil {
		client = &http.Client{}
	}

	resp, respErr := client.Do(req)
	if respErr != nil {
		return "", fmt.Errorf("Request failed: %v", respErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unexpected return status code: %s", resp.Status)
	}

	var gatewayResponse GatewayResponse
	unmarshalErr := json.NewDecoder(resp.Body).Decode(&gatewayResponse)
	if unmarshalErr != nil {
		return "", unmarshalErr
	}

	if gatewayResponse.URL == "" {
		return "", fmt.Errorf("Missing field url, after receiving data from server with 200")
	}

	return (gatewayResponse.URL + discordApiVersion), nil
}

func StartBot(ctx context.Context, botToken, apiBase string) error {
	fmt.Println("Starting bot")
	url, gatewayErr := GetGatewayUrl(ctx, botToken, apiBase, nil)
	if gatewayErr != nil {
		return gatewayErr
	}
	fmt.Println("URL: " + url)
	return nil
}

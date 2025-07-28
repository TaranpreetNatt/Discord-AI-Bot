package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

// Gateway handles Discord Gateway API operations
type Gateway struct {
	apiBase string
	logger  logger.Logger
	client  *http.Client
}

type SessionStartLimit struct {
	Total          int `json:"total"`
	Remaining      int `json:"remaining"`
	ResetAfter     int `json:"reset_after"`
	MaxConcurrency int `json:"max_concurrency"`
}

// GatewayResponse represents the response from Discord's gateway endpoint
type GatewayResponse struct {
	URL               string            `json:"url"`
	Shards            int               `json:"shards"`
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

// NewGateway creates a new gateway client
func NewGateway(apiBase string, logger logger.Logger) *Gateway {
	return &Gateway{
		apiBase: apiBase,
		logger:  logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewGatewayWithTimeout(apiBase string, logger logger.Logger, timeout time.Duration) *Gateway {
	return &Gateway{
		apiBase: apiBase,
		logger:  logger,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetURL retrieves the Discord Gateway URL
func (g *Gateway) GetURL(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiBase+"/gateway/bot", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("User-Agent", "DiscordBot (https://github.com/taranpreetnatt/Discord-AI-Bot, 1.0)")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var gatewayResp GatewayResponse
	if err := json.NewDecoder(resp.Body).Decode(&gatewayResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if gatewayResp.URL == "" {
		return "", fmt.Errorf("gateway URL is empty")
	}

	// Add Discord gateway version and encoding
	gatewayURL := gatewayResp.URL + "/?v=10&encoding=json"

	g.logger.Info("Retrieved gateway URL",
		logger.Field{Key: "url", Value: gatewayURL},
		logger.Field{Key: "shards", Value: gatewayResp.Shards})

	return gatewayURL, nil
}

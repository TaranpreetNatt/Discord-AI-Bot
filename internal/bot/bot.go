package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	// "github.com/coder/websocket"
)

const (
	//TODO: Remove bot token and put it as a env
	botToken = "MTM1MzU3NTY3NzEzNTQyNTU0Nw.GHSxaY.7KV0iuIOEcCZTbCKwISLFRIpmeyUMCp9NYC988"
	apiBase  = "https://discord.com/api/v10"
)

type GatewayResponse struct {
	URL               string `json:"url"`
	Shards            int    `json:"shards"`
	SessionStartLimit struct {
		Total          int `json:"total"`
		Remaining      int `json:"remaining"`
		ResetAfter     int `json:"reset_after"`
		MaxConcurrency int `json:"max_concurrency"`
	} `json:"session_start_limit"`
}

func getGatewayUrl() (string, error) {
	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		return "", fmt.Errorf("Bot token does not exist in the environment variables")
	}
	apiBase, ok := os.LookupEnv("API_BASE")
	if !ok {
		return "", fmt.Errorf("API_BASE url does not exist in the environment variables")
	}

	req, reqErr := http.NewRequest("GET", apiBase+"/gateway/bot", nil)
	if reqErr != nil {
		return "", reqErr
	}
	req.Header.Set("Authorization", "Bot "+botToken)
	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		return "", respErr
	}
	defer resp.Body.Close()

	var gatewayResponse GatewayResponse
	unmarshalErr := json.NewDecoder(resp.Body).Decode(&gatewayResponse)
	if unmarshalErr != nil {
		return "", unmarshalErr
	}

	return gatewayResponse.URL + "/?v=10&encoding=json", nil
}

// func ConnectToDiscord() {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
// 	defer cancel()
//
// }

func StartBot() error {
	fmt.Println("Starting bot")
	url, gatewayErr := getGatewayUrl()
	if gatewayErr != nil {
		return gatewayErr
	}
	fmt.Println("URL: " + url)
	return nil
}

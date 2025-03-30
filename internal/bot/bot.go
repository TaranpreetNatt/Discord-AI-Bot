package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	// "github.com/coder/websocket"
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

func GetGatewayUrl(botToken, apiBase string) (string, error) {
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

func StartBot(botToken, apiBase string) error {
	fmt.Println("Starting bot")
	url, gatewayErr := GetGatewayUrl(botToken, apiBase)
	if gatewayErr != nil {
		return gatewayErr
	}
	fmt.Println("URL: " + url)
	return nil
}

package bot_test

import (
	"testing"

	bot "github.com/taranpreetnatt/Discord-AI-Bot/internal/bot"
	"github.com/taranpreetnatt/Discord-AI-Bot/testutils"
)

const (
	test_token = "test_token"
	apiBase    = "https://discord.com/api/v10"
	discordURL = "wss://gateway.discord.gg/?v=10&encoding=json"
)

func TestGetGatewayUrl(t *testing.T) {
	server, client := testutils.NewMockServer(t, test_token)
	defer server.Close()

	tests := []struct {
		name    string
		token   string
		apiBase string
		wantURL string
		wantErr bool
	}{
		{
			name:    "Happy path, everything is correct",
			token:   test_token,
			apiBase: apiBase,
			wantURL: discordURL,
			wantErr: false,
		},
		{
			name:    "Incorrect token",
			token:   "wrong_token",
			apiBase: apiBase,
			wantURL: "",
			wantErr: true,
		},
		{
			name:    "Invalid json response",
			token:   "invalid_json",
			apiBase: apiBase,
			wantURL: "",
			wantErr: true,
		},
		{
			name:    "500 internal server error",
			token:   "internal_server_error",
			apiBase: apiBase,
			wantURL: "",
			wantErr: true,
		},
		{
			name:    "Missing discord URL field",
			token:   "missing_url_field",
			apiBase: apiBase,
			wantURL: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bot.GetGatewayUrl(tt.token, apiBase, client)
			if (err != nil) != tt.wantErr {
				t.Errorf("Got error %v", err)
				return
			}
			if got != tt.wantURL {
				t.Errorf("wanted %s, got %s", tt.wantURL, got)
			}

			if tt.wantErr && err != nil {
				t.Logf("Expected error received: %v", err)
			}
		})
	}
}

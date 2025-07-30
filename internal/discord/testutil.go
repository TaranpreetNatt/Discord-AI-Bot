package discord

import (
	"encoding/json"
	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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

func setupLogger(t *testing.T) logger.Logger {
	l, err := logger.NewZapAdapter()
	if err != nil {
		t.Fatalf("Failed to create logger in gateway_test, %v", err)
	}
	return l
}

func setupMockServer(t *testing.T, config MockConfig) *httptest.Server {
	server := MockServer(config)
	server.Start()
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

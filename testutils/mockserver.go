package testutils

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	bot "github.com/taranpreetnatt/Discord-AI-Bot/internal/bot"
	"golang.org/x/net/http2"
)

func NewMockServer(t *testing.T, validToken string) (*httptest.Server, *http.Client) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		switch auth {

		case "Bot " + validToken:
			response := bot.GatewayResponse{
				URL:    "wss://gateway.discord.gg",
				Shards: 9,
				SessionStartLimit: bot.SessionStartLimit{
					Total:          999,
					Remaining:      999,
					ResetAfter:     14400000,
					MaxConcurrency: 1,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			encodeErr := json.NewEncoder(w).Encode(response)
			if encodeErr != nil {
				t.Errorf("Error with encoding response in mock server: %v", encodeErr)
			}

		case "Bot invalid_json":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("not json"))

		case "Bot internal_server_error":
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})

		case "Bot missing_url_field":
			response := bot.GatewayResponse{
				Shards: 9,
				SessionStartLimit: bot.SessionStartLimit{
					Total:          999,
					Remaining:      999,
					ResetAfter:     14400000,
					MaxConcurrency: 1,
				},
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Error encondign response in missing_url_field mock server. Err: %v", err)
			}

		default:
			w.WriteHeader(http.StatusUnauthorized) // 401
			json.NewEncoder(w).Encode(map[string]string{"message": "Unathorized"})
		}
	}))

	// Configures HTTP/2 Server
	if err := http2.ConfigureServer(server.Config, &http2.Server{}); err != nil {
		t.Fatalf("Failted to configure HTTP/2 sever: %v", err)
	}

	server.EnableHTTP2 = true
	server.TLS = &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2"},
	}
	server.StartTLS()

	// Configure client for HTTP/2
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	if err := http2.ConfigureTransport(tr); err != nil {
		t.Fatalf("Failed to configure HTTP/2 transport: %v", err)
	}

	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}

	client := &http.Client{Transport: tr}
	return server, client
}

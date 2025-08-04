package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

type MockConfig struct {
	ValidToken      string
	StatusCode      int
	ReturnValidJson bool
	ResponseBody    string // for custom invalid JSON
	Delay           time.Duration
	WebSocketURL    string
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

		webSocketURL := "wss://gateway.discord.gg"
		if config.WebSocketURL != "" {
			webSocketURL = config.WebSocketURL
		}

		if config.ReturnValidJson {
			response := GatewayResponse{
				URL:    webSocketURL,
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

type MockWebSocketServer struct {
	server *httptest.Server
}

func NewWebSocketServer(t *testing.T) *MockWebSocketServer {
	onPingReceived := func(ctx context.Context, payload []byte) bool {
		return true
	}
	onPongReceived := func(ctx context.Context, payload []byte) {
		return
	}

	opts := websocket.AcceptOptions{
		OnPingReceived: onPingReceived,
		OnPongReceived: onPongReceived,
	}

	handler := http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		wsConn, err := websocket.Accept(writer, req, &opts)
		if err != nil {
			t.Fatalf("Error creating MockWebSocketServer with error %s", err)
		}
		defer wsConn.Close(websocket.StatusNormalClosure, "Closing Mock Server")

		helloEvent := Event{
			Op:   10,
			Data: json.RawMessage(`{"heartbeat_interval": 4500}`),
		}
		helloEventByte, err := json.Marshal(helloEvent)
		if err != nil {
			t.Fatalf("Failed to marshal hello event in MockWebSocketServer, err: %v", err)
		}

		ctx := context.Background()
		if err := wsConn.Write(ctx, websocket.MessageText, helloEventByte); err != nil {
			t.Fatalf("Error sending hello event in MockWebSocketServer, err: %v", err)
		}

		for {
			messageType, reader, err := wsConn.Reader(ctx)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
					t.Logf("Websocket closed normaly")
					return
				}
				t.Logf("Failed to read from MockWebServer connection, err: %v", err)
				return
			}

			if messageType == websocket.MessageBinary {
				continue
			}

			var event Event
			decoder := json.NewDecoder(reader)
			if err := decoder.Decode(&event); err != nil {
				t.Logf("Failed to decode JSON,  err: %v", err)
				continue
			}

			t.Logf("Received event: Op=%d", event.Op)

			if event.Op == OpHeartbeat {
				heartBeat := Event{
					Op: 11,
				}
				heartBeatByte, err := json.Marshal(heartBeat)
				if err != nil {
					t.Logf("Failed to Marshal heartbeat struct, err: %v", err)
					return
				}
				if err := wsConn.Write(ctx, websocket.MessageText, heartBeatByte); err != nil {
					t.Logf("Error writing heartbeat data to MockWebSocket connection, %v", err)
					return
				} else {
					t.Logf("Sent HEARTBEAT_ACK")
				}
			}

			if event.Op == OpIdentify {
				readyData := ReadyData{
					Version: 10,
					User: User{
						ID:       "123456",
						Username: "test_bot",
					},
					SessionID: "test-session-123",
					ResumeURL: "wss://gateway.discord.gg",
					Application: struct {
						ID    string `json:"id"`
						Flags int    `json:"flags"`
					}{
						ID:    "app123",
						Flags: 0,
					},
				}

				readyDataByte, err := json.Marshal(readyData)
				if err != nil {
					t.Logf("Error Marshling readyData in MockWebServer, %v", err)
				}

				identify := Event{
					Op:   OpDispatch,
					Type: "READY",
					Data: json.RawMessage(readyDataByte),
				}

				identifyByte, err := json.Marshal(identify)
				if err != nil {
					t.Logf("Failed to Marshal identify struct, err %v	", err)
				}
				if err := wsConn.Write(ctx, websocket.MessageText, identifyByte); err != nil {
					t.Logf("Failed to write identify struct to MockWebSocket connection, err: %v", err)
				} else {
					t.Logf("Sent IDENTIFY")
				}
			}
		}

	})
	server := httptest.NewServer(handler)

	wsMockServer := MockWebSocketServer{
		server: server,
	}
	return &wsMockServer
}

func (mockWs MockWebSocketServer) URL() string {
	return strings.Replace(mockWs.server.URL, "http://", "ws://", 1)
}

func (mockWs MockWebSocketServer) Close() {
	mockWs.server.Close()
}

package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type WebSocketConn struct {
	Conn         *websocket.Conn
	mu           sync.Mutex
	lastSequence int
	msgChan      chan *Payload
	pingChan     chan *Payload
}

type HeartbeatInterval struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type Payload struct {
	Op        int             `json:"op"`
	Data      json.RawMessage `json:"d"`
	Sequence  int             `json:"s"`
	EventName string          `json:"t"`
}

type Heartbeat struct {
	Op   int `json:"op"`
	Data int `json:"d"`
}

func NewWebsocketConnection(ctx context.Context, url string) (*WebSocketConn, error) {
	conn, resp, err := websocket.Dial(ctx, url, nil)
	if err != nil || resp.StatusCode != 101 {
		return nil, fmt.Errorf("Error creating websocket connection, %v", err)
	}

	discordWebSocket := &WebSocketConn{
		Conn:     conn,
		mu:       sync.Mutex{},
		msgChan:  make(chan *Payload, 100),
		pingChan: make(chan *Payload, 10),
	}

	return discordWebSocket, nil
}

func ParseMessage(data []byte) (*Payload, error) {
	if data == nil {
		return nil, fmt.Errorf("Cannot parse nil data slice")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("Cannot parse an empty data slice")
	}

	var p Payload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("Error parsing data, %v", err)
	}

	return &p, nil
}

func ParseHeartbeatInterval(data []byte) (int, error) {
	//#TODO: Create a helper function to validate []byte
	if data == nil {
		return 0, fmt.Errorf("Cannot parse nil data slice")
	}

	if len(data) == 0 {
		return 0, fmt.Errorf("Cannot parse an empty data slice")
	}

	var heartbeat HeartbeatInterval
	if err := json.Unmarshal(data, &heartbeat); err != nil {
		return 0, fmt.Errorf("Could not parse heartbeat payload: %v\n", err)
	}
	return heartbeat.HeartbeatInterval, nil
}

func (w *WebSocketConn) GetMessages(ctx context.Context) {
	defer w.Conn.CloseNow()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			messageType, reader, err := w.Conn.Reader(ctx)
			if err != nil {
				fmt.Printf("Error reading from websocket connection: %v", err)
				return
			}

			if messageType == websocket.MessageText {
				//TODO: Change from ReadAll to using the reader.read and implement a max message size
				data, err := io.ReadAll(reader)
				if err != nil {
					fmt.Printf("Error reading all of discord websocket message: %v", err)
					return
				}

				payload, err := ParseMessage(data)
				if err != nil {
					fmt.Printf("Error parsing payload: %v", err)
					return
				}
				fmt.Printf("Payload: %+v\n", payload)
				w.lastSequence = payload.Sequence

				if payload.Op == 10 || payload.Op == 1 || payload.Op == 11 {
					w.pingChan <- payload
					continue
				}
				w.msgChan <- payload
			}
		}

	}
}

func (w *WebSocketConn) WriteConn(ctx context.Context, data []byte) error {
	w.mu.Lock()
	fmt.Println("Mutex locked to Write")

	defer func() {
		w.mu.Unlock()
		fmt.Println("Mutex unlocked in Write")
	}()

	if err := w.Conn.Write(ctx, websocket.MessageText, data); err != nil {
		fmt.Printf("Error writing to discord during heartbeat: %v\n", err)
		return err
	}
	fmt.Println("Sent message successfully")
	return nil
}

func setHeartbeatInterval(heartbeatinterval int, rand float64) time.Duration {
	interval := float64(heartbeatinterval) * rand
	return time.Duration(interval) * time.Millisecond
}

// TODO:If there is no heartbeat response from discord, add reconnect
func (w *WebSocketConn) PingDiscord(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(1) * time.Millisecond)
	heartbeatAckTimer := time.NewTicker(time.Duration(1) * time.Millisecond)
	var heartbeatinterval int

	heartbeat := Heartbeat{Op: 1, Data: w.lastSequence}
	heartbeatmessage, err := json.Marshal(heartbeat)
	if err != nil {
		fmt.Printf("Error marshaling heartbeatmessage: %v\n", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-w.pingChan:
			if payload.Op == 10 {
				var err error
				heartbeatinterval, err = ParseHeartbeatInterval(payload.Data)
				if err != nil {
					fmt.Printf("Error parsing heartbeat interval: %v\n", err)
					return
				}
			}

			if payload.Op == 11 {
				heartbeatAckTimer.Reset(time.Duration(120) * time.Second)
				continue
			}
			if err := w.WriteConn(ctx, heartbeatmessage); err != nil {
				return
			}
			heartbeatAckTimer.Reset(time.Duration(3) * time.Second)
			ticker.Reset(setHeartbeatInterval(heartbeatinterval, rand.Float64()))
		case <-ticker.C:
			if err := w.WriteConn(ctx, heartbeatmessage); err != nil {
				return
			}
			heartbeatAckTimer.Reset(time.Duration(3) * time.Second)
			ticker.Reset(setHeartbeatInterval(heartbeatinterval, rand.Float64()))
		case <-heartbeatAckTimer.C:
			fmt.Println("No heartbeat response from discord")
			return
		}
	}
}

func (w *WebSocketConn) Coordinator(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-w.msgChan:
			fmt.Printf("Message is %+v\n", msg)
		}
	}
}

package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"io"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type GatewayCloseEventCodes struct {
	Code        int
	Description string
	Explanation string
	Reconnect   bool
}

var gatewayCloseEventCodes = map[int]GatewayCloseEventCodes{
	4000: {Code: 4000, Description: "Unknown error", Explanation: "We're not sure what went wrong. Try reconnecting?", Reconnect: true},
	4001: {Code: 4001, Description: "Unknown opcode", Explanation: "You sent an invalid Gateway opcode or an invalid payload for an opcode. Don't do that!", Reconnect: true},
	4002: {Code: 4002, Description: "Decode error", Explanation: "You sent an invalid payload to Discord. Don't do that!", Reconnect: true},
	4003: {Code: 4003, Description: "Not authenticated", Explanation: "You sent us a payload prior to identifying, or this session has been invalidated.", Reconnect: true},
	4004: {Code: 4004, Description: "Authentication failed", Explanation: "The account token sent with your identify payload is incorrect.", Reconnect: false},
	4005: {Code: 4005, Description: "Already authenticated", Explanation: "You sent more than one identify payload. Don't do that!", Reconnect: true},
	4007: {Code: 4007, Description: "Invalid seq", Explanation: "The sequence sent when resuming the session was invalid. Reconnect and start a new session.", Reconnect: true},
	4008: {Code: 4008, Description: "Rate limited", Explanation: "Woah nelly! You're sending payloads to us too quickly. Slow it down! You will be disconnected on receiving this.", Reconnect: true},
	4009: {Code: 4009, Description: "Session timed out", Explanation: "Your session timed out. Reconnect and start a new one.", Reconnect: true},
	4010: {Code: 4010, Description: "Invalid shard", Explanation: "You sent us an invalid shard when identifying.", Reconnect: false},
	4011: {Code: 4011, Description: "Sharding required", Explanation: "The session would have handled too many guilds - you are required to shard your connection in order to connect.", Reconnect: false},
	4012: {Code: 4012, Description: "Invalid API version", Explanation: "You sent an invalid version for the gateway.", Reconnect: false},
	4013: {Code: 4013, Description: "Invalid intent(s)", Explanation: "You sent an invalid intent for a Gateway Intent. You may have incorrectly calculated the bitwise value.", Reconnect: false},
	4014: {Code: 4014, Description: "Disallowed intent(s)", Explanation: "You sent a disallowed intent for a Gateway Intent. You may have tried to specify an intent that you have not enabled or are not approved for.", Reconnect: false},
	7:    {Code: 7, Description: "Reconnect", Explanation: "Discord requested a reconnect, likely due to server changes.", Reconnect: true},
	9:    {Code: 9, Description: "Invalid Session", Explanation: "The session is invalid. Reconnect and resume if possible, or start a new session.", Reconnect: true},
}

type SnowflakeID snowflake.ID

type WebSocketConn struct {
	Conn                        *websocket.Conn
	mu                          sync.Mutex
	lastSequence                int
	payloadChan                 chan *Payload
	pingChan                    chan *Payload
	initialHeartbeatAckReceived bool
	Token                       string
	resumeUrl                   string
	sessionId                   string
}

type HeartbeatInterval struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type Payload struct {
	Op        int             `json:"op"`
	Data      json.RawMessage `json:"d"`
	Sequence  *int            `json:"s,omitempty"`
	EventName string          `json:"t"`
}

type Heartbeat struct {
	Op   int  `json:"op"`
	Data *int `json:"d"`
}

type Resume struct {
	Op   int         `json:"op"`
	Data *ResumeData `json:"d"`
}

type ResumeData struct {
	Token     string `json:"token"`
	SessionId string `json:"session_id"`
	Seq       int    `json:"seq"`
}

type ConnectionProperties struct {
	Os      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

type IdentifyData struct {
	//#TODO: Hide token from log output
	Token           string               `json:"token"`
	Properties      ConnectionProperties `json:"properties"`
	Compress        *bool                `json:"compress,omitempty"`
	Large_threshold *int                 `json:"large_threshold,omitempty"`
	Shard           *[2]int              `json:"shard,omitempty"`
	Intents         int                  `json:"intents"`
}

type Identify struct {
	Op int `json:"op"`
	//TODO: Make this a pointer
	//TODO: Refactor the Identify and Heartbeat to be one general struct
	Data IdentifyData `json:"d"`
}

type User struct {
	Id            SnowflakeID `json:"id"`
	Username      string      `json:"username"`
	Discriminator string      `json:"discriminator"`
}

type UnavailableGuild struct {
	Id          SnowflakeID `json:"id"`
	Unavailable bool        `json:"unavailable"`
}

type PartialApplication struct {
	Id    SnowflakeID `json:"id"`
	Flags int         `json:"flags"`
}

type Ready struct {
	V           int                `json:"v"`
	User        User               `json:"user"`
	Guilds      []UnavailableGuild `json:"guilds"`
	SessionId   string             `json:"session_id"`
	ResumeUrl   string             `json:"resume_gateway_url"`
	Shard       *[2]int            `json:"shard,omitempty"`
	Application PartialApplication `json:"application"`
}

func (id *SnowflakeID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Error unmarshalJSON for snowflake data: %v", err)
	}
	ID, err := snowflake.ParseString(str)
	if err != nil {
		return fmt.Errorf("Error unmarshalJSON for snowflake id: %v", err)
	}
	*id = SnowflakeID(ID)
	return nil
}

func NewWebsocketConnection(ctx context.Context, url string, token string) (*WebSocketConn, error) {
	conn, resp, err := websocket.Dial(ctx, url, nil)
	if err != nil || resp.StatusCode != 101 {
		return nil, fmt.Errorf("Error creating websocket connection, %v", err)
	}

	discordWebSocket := &WebSocketConn{
		Conn:        conn,
		mu:          sync.Mutex{},
		payloadChan: make(chan *Payload, 100),
		pingChan:    make(chan *Payload, 10),
		Token:       token,
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
		return 0, fmt.Errorf("Cannot parse nil data slice\n")
	}

	if len(data) == 0 {
		return 0, fmt.Errorf("Cannot parse an empty data slice\n")
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
				fmt.Printf("Error reading from websocket connection: %v\n", err)
				return
			}

			if messageType == websocket.MessageText {
				//TODO: Change from ReadAll to using the reader.read and implement a max message size
				data, err := io.ReadAll(reader)
				if err != nil {
					fmt.Printf("Error reading all of discord websocket message: %v\n", err)
					return
				}

				payload, err := ParseMessage(data)
				if err != nil {
					fmt.Printf("Error parsing payload: %v\n", err)
					return
				}
				fmt.Printf("Payload: %+v\n", payload)
				if payload.Op == 0 && payload.Sequence != nil && *payload.Sequence > 0 {
					w.lastSequence = *payload.Sequence
				}

				w.payloadChan <- payload
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
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer ticker.Stop()

	heartbeatAckTimer := time.NewTicker(time.Duration(5) * time.Second)
	defer heartbeatAckTimer.Stop()

	var heartbeatinterval int

	var sequence *int
	if w.lastSequence > 0 {
		sequence = &w.lastSequence
	}
	heartbeat := Heartbeat{Op: 1, Data: sequence}
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
			heartbeatAckTimer.Reset(time.Duration(5) * time.Second)
			ticker.Reset(setHeartbeatInterval(heartbeatinterval, rand.Float64()))
		case <-ticker.C:
			if err := w.WriteConn(ctx, heartbeatmessage); err != nil {
				return
			}
			heartbeatAckTimer.Reset(time.Duration(5) * time.Second)
			ticker.Reset(setHeartbeatInterval(heartbeatinterval, rand.Float64()))
		case <-heartbeatAckTimer.C:
			fmt.Println("No heartbeat response from discord")
			return
		}
	}
}

func (w *WebSocketConn) discordIdentify(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		fmt.Println("Context cancelled:", err)
		return
	}

	connectionProperties := ConnectionProperties{Os: "linux_bot", Browser: "bot", Device: "bot"}
	identifyData := IdentifyData{
		Token:      w.Token,
		Properties: connectionProperties,
		Intents:    67584,
	}
	identify := Identify{Op: 2, Data: identifyData}

	identifyMessage, err := json.Marshal(identify)
	if err != nil {
		fmt.Printf("Error marshaling identify payload: %v\n", err)
		return
	}

	fmt.Printf("Sending identify to discord: %+v\n", identify)
	if err := w.WriteConn(ctx, identifyMessage); err != nil {
		fmt.Printf("Error writing identify payload to websocker: %v\n", err)
		return
	}
}

func (w *WebSocketConn) parseReadyData(data []byte) (*Ready, error) {
	var payload Ready

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("Error in parseReadyData: %v", err)
	}

	return &payload, nil
}

func (w *WebSocketConn) Reconnect(ctx context.Context) (*WebSocketConn, error) {

	resumeData := ResumeData{
		Token:     w.Token,
		SessionId: w.sessionId,
		Seq:       w.lastSequence,
	}

	resume := Resume{
		Op:   6,
		Data: &resumeData,
	}

	conn, resp, err := websocket.Dial(ctx, w.resumeUrl, nil)
	if err != nil || resp.StatusCode != 101 {
		return nil, fmt.Errorf("Error creating websocket connection, %v\n", err)
	}

	discordWebSocket := &WebSocketConn{
		Conn:        conn,
		mu:          sync.Mutex{},
		payloadChan: make(chan *Payload, 100),
		pingChan:    make(chan *Payload, 10),
		Token:       w.Token,
	}

	resumeByte, err := json.Marshal(resume)
	if err != nil {
		return nil, fmt.Errorf("Error marshiling resume payload during Reconnect: %v\n", err)
	}
	if err := w.WriteConn(ctx, resumeByte); err != nil {
		return nil, fmt.Errorf("Error writing to conn during Reconnect: %v\n", err)
	}
	return discordWebSocket, nil
}

func (w *WebSocketConn) Coordinator(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-w.payloadChan:
			fmt.Printf("Payload is %+v\n", payload)

			if payload.Op == 10 || payload.Op == 1 || payload.Op == 11 {
				w.pingChan <- payload
			}

			if payload.Op == 11 {
				if !w.initialHeartbeatAckReceived {
					w.initialHeartbeatAckReceived = true
					w.discordIdentify(ctx)
				}
			}

			if payload.Op == 0 && payload.EventName == "READY" {
				readyPayload, err := w.parseReadyData(payload.Data)
				if err != nil {
					return
				}
				fmt.Printf("\nReady Payload: %+v\n", readyPayload)
				w.resumeUrl = readyPayload.ResumeUrl
				w.sessionId = readyPayload.SessionId
			}

			if gatewayCloseEventCodes[payload.Op].Reconnect {
				fmt.Println("Attempting reconnect")
			}
		}
	}
}

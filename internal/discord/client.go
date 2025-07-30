package discord

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

// Client represents the main Discord client
type Client struct {
	token   string
	apiBase string
	logger  logger.Logger

	// Components
	gateway    *Gateway
	connection *Connection
	heartbeat  *HeartbeatManager

	// State
	state   *ClientState
	stateMu sync.RWMutex

	// Channels for coordination
	eventChan    chan *Event
	shutdownChan chan struct{}

	// Context management
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient creates a new Discord client
func NewClient(token, apiBase string, logger logger.Logger) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	if apiBase == "" {
		return nil, fmt.Errorf("apiBase cannot be empty")
	}

	client := &Client{
		token:        token,
		apiBase:      apiBase,
		logger:       logger,
		state:        NewClientState(),
		eventChan:    make(chan *Event, 100),
		shutdownChan: make(chan struct{}),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Initialize components
	client.gateway = NewGateway(apiBase, logger)
	client.connection = NewConnection(logger)
	client.heartbeat = NewHeartbeatManager(logger)

	return client, nil
}

// Start starts the Discord client
func (c *Client) Start(ctx context.Context) error {
	c.logger.Info("Starting Discord client")

	// Get gateway URL
	gatewayURL, err := c.gateway.GetURL(ctx, c.token)
	if err != nil {
		return fmt.Errorf("failed to get gateway URL: %w", err)
	}

	// Start connection with reconnection handling
	if err := c.startWithReconnect(ctx, gatewayURL); err != nil {
		return fmt.Errorf("failed to start connection: %w", err)
	}

	return nil
}

// startWithReconnect handles the connection lifecycle with automatic reconnection
func (c *Client) startWithReconnect(ctx context.Context, gatewayURL string) error {
	reconnectAttempts := 0
	maxReconnectAttempts := 5

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Connect to Discord
		if err := c.connect(ctx, gatewayURL); err != nil {
			c.logger.Error("Connection failed", logger.Field{Key: "error", Value: err})

			reconnectAttempts++
			if reconnectAttempts >= maxReconnectAttempts {
				return fmt.Errorf("max reconnection attempts reached: %w", err)
			}

			// Exponential backoff
			delay := time.Duration(1<<reconnectAttempts) * time.Second
			c.logger.Info("Reconnecting", logger.Field{Key: "delay", Value: delay})

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		// Reset reconnect attempts on successful connection
		reconnectAttempts = 0

		// Run the main event loop
		if err := c.run(ctx); err != nil {
			c.logger.Error("Event loop error", logger.Field{Key: "error", Value: err})

			// Check if we should reconnect
			if c.shouldReconnect(err) {
				c.logger.Info("Attempting to reconnect")
				continue
			}

			return err
		}
	}
}

// connect establishes a connection to Discord
func (c *Client) connect(ctx context.Context, gatewayURL string) error {
	// Create WebSocket connection
	if _, err := c.connection.Connect(ctx, gatewayURL); err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	// Start message reader
	go c.connection.ReadMessages(ctx, c.eventChan)

	// Start heartbeat manager
	go c.heartbeat.Start(ctx, c.connection)

	return nil
}

// run handles the main event processing loop
func (c *Client) run(ctx context.Context) error {
	authenticated := false

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event := <-c.eventChan:
			if err := c.handleEvent(ctx, event, &authenticated); err != nil {
				return err
			}
		}
	}
}

// handleEvent processes Discord events
func (c *Client) handleEvent(ctx context.Context, event *Event, authenticated *bool) error {
	c.logger.Debug("Handling event",
		logger.Field{Key: "op", Value: event.Op},
		logger.Field{Key: "type", Value: event.Type})

	switch event.Op {
	case OpHello:
		return c.handleHello(ctx, event)

	case OpHeartbeatAck:
		c.heartbeat.HandleAck()
		if !*authenticated {
			*authenticated = true
			return c.authenticate(ctx)
		}

	case OpDispatch:
		return c.handleDispatch(ctx, event)

	case OpReconnect:
		return fmt.Errorf("reconnect requested by Discord")

	case OpInvalidSession:
		return c.handleInvalidSession(ctx, event)

	default:
		c.logger.Debug("Unhandled event", logger.Field{Key: "op", Value: event.Op})
	}

	return nil
}

// shouldReconnect determines if we should attempt to reconnect
func (c *Client) shouldReconnect(err error) bool {
	// Add logic to determine if error is recoverable
	// For now, always attempt reconnect for non-context errors
	return err != context.Canceled && err != context.DeadlineExceeded
}

// Stop gracefully shuts down the client
func (c *Client) Stop() {
	c.logger.Info("Stopping Discord client")
	c.cancel()

	if c.connection != nil {
		c.connection.Close()
	}

	close(c.shutdownChan)
}

package discord

import (
	"fmt"
)

// DiscordError represents a Discord-specific error
type DiscordError struct {
	Code    int
	Message string
	Retry   bool
}

func (e *DiscordError) Error() string {
	return fmt.Sprintf("discord error %d: %s", e.Code, e.Message)
}

// CanRetry returns whether this error allows for retry
func (e *DiscordError) CanRetry() bool {
	return e.Retry
}

// Gateway close codes from Discord documentation
var GatewayCloseErrors = map[int]*DiscordError{
	4000: {Code: 4000, Message: "Unknown error", Retry: true},
	4001: {Code: 4001, Message: "Unknown opcode", Retry: true},
	4002: {Code: 4002, Message: "Decode error", Retry: true},
	4003: {Code: 4003, Message: "Not authenticated", Retry: true},
	4004: {Code: 4004, Message: "Authentication failed", Retry: false},
	4005: {Code: 4005, Message: "Already authenticated", Retry: true},
	4007: {Code: 4007, Message: "Invalid seq", Retry: true},
	4008: {Code: 4008, Message: "Rate limited", Retry: true},
	4009: {Code: 4009, Message: "Session timed out", Retry: true},
	4010: {Code: 4010, Message: "Invalid shard", Retry: false},
	4011: {Code: 4011, Message: "Sharding required", Retry: false},
	4012: {Code: 4012, Message: "Invalid API version", Retry: false},
	4013: {Code: 4013, Message: "Invalid intent(s)", Retry: false},
	4014: {Code: 4014, Message: "Disallowed intent(s)", Retry: false},
}

package discord

import (
	"context"
	"sync"
	"time"

	"github.com/taranpreetnatt/Discord-AI-Bot/internal/logger"
)

// RateLimiter handles Discord API rate limiting
type RateLimiter struct {
	logger logger.Logger
	mu     sync.Mutex

	// Global rate limit
	globalReset time.Time
	globalLimit int

	// Per-route rate limits
	routes map[string]*RouteLimit
}

// RouteLimit represents rate limit info for a specific route
type RouteLimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(logger logger.Logger) *RateLimiter {
	return &RateLimiter{
		logger: logger,
		routes: make(map[string]*RouteLimit),
	}
}

// Wait waits if necessary to respect rate limits
func (r *RateLimiter) Wait(ctx context.Context, route string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Check global rate limit
	if now.Before(r.globalReset) {
		waitTime := r.globalReset.Sub(now)
		r.logger.Warn("Global rate limit hit, waiting",
			logger.Field{Key: "wait_time", Value: waitTime})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	// Check route-specific rate limit
	if limit, exists := r.routes[route]; exists && limit.Remaining <= 0 && now.Before(limit.Reset) {
		waitTime := limit.Reset.Sub(now)
		r.logger.Warn("Route rate limit hit, waiting",
			logger.Field{Key: "route", Value: route},
			logger.Field{Key: "wait_time", Value: waitTime})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	return nil
}

// UpdateLimits updates rate limit information from response headers
func (r *RateLimiter) UpdateLimits(route string, limit, remaining int, resetAfter time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes[route] = &RouteLimit{
		Limit:     limit,
		Remaining: remaining,
		Reset:     time.Now().Add(resetAfter),
	}
}

// SetGlobalLimit sets the global rate limit
func (r *RateLimiter) SetGlobalLimit(resetAfter time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.globalReset = time.Now().Add(resetAfter)
}

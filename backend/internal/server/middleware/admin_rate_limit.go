package middleware

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// adminRateLimitEntry tracks request counts per IP.
type adminRateLimitEntry struct {
	count    int
	windowAt time.Time
}

// AdminRateLimiterConfig configures admin rate limiting.
type AdminRateLimiterConfig struct {
	// MaxRequests is the maximum number of requests allowed per window per IP.
	MaxRequests int
	// Window is the time window for counting requests.
	Window time.Duration
	// Enabled controls whether rate limiting is active.
	Enabled bool
}

// AdminRateLimitConfigProvider is a callback that returns current rate limit config.
// Used to dynamically load settings from the database.
type AdminRateLimitConfigProvider func() *AdminRateLimiterConfig

// AdminRateLimiter provides IP-based rate limiting for admin endpoints.
// Uses in-memory counters (no Redis dependency) — suitable since admin traffic is low volume.
// Supports dynamic configuration via SetConfigProvider.
type AdminRateLimiter struct {
	mu             sync.Mutex
	entries        map[string]*adminRateLimitEntry
	cfg            AdminRateLimiterConfig
	configProvider atomic.Pointer[AdminRateLimitConfigProvider]
}

// NewAdminRateLimiter creates a new admin rate limiter with default config.
// Default: enabled, 60 requests per minute per IP (generous for normal admin usage).
func NewAdminRateLimiter(cfg AdminRateLimiterConfig) *AdminRateLimiter {
	if cfg.MaxRequests <= 0 {
		cfg.MaxRequests = 60
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	rl := &AdminRateLimiter{
		entries: make(map[string]*adminRateLimitEntry),
		cfg:     cfg,
	}
	// Background cleanup every 5 minutes to prevent memory leak
	go rl.cleanupLoop()
	return rl
}

// SetConfigProvider sets a dynamic config provider that overrides the static config.
// The provider is called on each request to get the current configuration.
func (rl *AdminRateLimiter) SetConfigProvider(provider AdminRateLimitConfigProvider) {
	rl.configProvider.Store(&provider)
}

// getEffectiveConfig returns the current effective configuration.
func (rl *AdminRateLimiter) getEffectiveConfig() AdminRateLimiterConfig {
	if p := rl.configProvider.Load(); p != nil && *p != nil {
		if cfg := (*p)(); cfg != nil {
			return *cfg
		}
	}
	return rl.cfg
}

// Middleware returns a gin middleware that rate limits based on client IP.
func (rl *AdminRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := rl.getEffectiveConfig()

		// If disabled, pass through
		if !cfg.Enabled {
			c.Next()
			return
		}

		ip := c.ClientIP()
		window := cfg.Window
		if window <= 0 {
			window = time.Minute
		}
		maxReqs := cfg.MaxRequests
		if maxReqs <= 0 {
			maxReqs = 60
		}

		rl.mu.Lock()
		entry, exists := rl.entries[ip]
		now := time.Now()

		if !exists || now.Sub(entry.windowAt) > window {
			// New window
			rl.entries[ip] = &adminRateLimitEntry{count: 1, windowAt: now}
			rl.mu.Unlock()
			c.Next()
			return
		}

		entry.count++
		count := entry.count
		rl.mu.Unlock()

		if count > maxReqs {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "RATE_LIMIT_EXCEEDED",
				"message": "Too many requests to admin API, please try again later",
			})
			return
		}

		c.Next()
	}
}

func (rl *AdminRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.entries {
			if now.Sub(entry.windowAt) > 2*time.Minute {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

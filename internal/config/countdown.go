package config

import (
	"os"
	"strconv"
	"time"
)

// Countdown configuration constants with default values
const (
	// DefaultUpdateInterval is the default interval between countdown updates (in seconds)
	DefaultUpdateInterval = 60

	// DefaultMaxErrors is the default maximum number of consecutive errors before cleanup
	DefaultMaxErrors = 5

	// DefaultCleanupInterval is the default interval for cleanup operations (in minutes)
	DefaultCleanupInterval = 30

	// DefaultMaxActiveCountdowns is the default maximum number of simultaneous countdowns
	DefaultMaxActiveCountdowns = 100

	// DefaultMessageEditTimeout is the default timeout for editMessageText API calls (in seconds)
	DefaultMessageEditTimeout = 10

	// DefaultRateLimitDelay is the default delay between message edits to respect Telegram limits (in milliseconds)
	DefaultRateLimitDelay = 1000
)

// CountdownConfig holds all countdown-related configuration
type CountdownConfig struct {
	// UpdateInterval defines how often countdown messages should be updated (seconds)
	UpdateInterval int

	// MaxErrors is the maximum number of consecutive errors before removing a countdown
	MaxErrors int

	// CleanupInterval defines how often expired/errored countdowns are cleaned up (minutes)
	CleanupInterval int

	// MaxActiveCountdowns is the maximum number of simultaneous active countdowns
	MaxActiveCountdowns int

	// MessageEditTimeout is the timeout for Telegram editMessageText API calls (seconds)
	MessageEditTimeout time.Duration

	// RateLimitDelay is the delay between message edits to respect Telegram rate limits (milliseconds)
	RateLimitDelay time.Duration

	// EnableAutoCleanup determines if automatic cleanup of expired countdowns is enabled
	EnableAutoCleanup bool

	// EnableErrorRecovery determines if automatic error recovery is enabled
	EnableErrorRecovery bool
}

// LoadCountdownConfig loads countdown configuration from environment variables with fallback to defaults
func LoadCountdownConfig() *CountdownConfig {
	config := &CountdownConfig{
		UpdateInterval:      getEnvInt("COUNTDOWN_UPDATE_INTERVAL", DefaultUpdateInterval),
		MaxErrors:           getEnvInt("COUNTDOWN_MAX_ERRORS", DefaultMaxErrors),
		CleanupInterval:     getEnvInt("COUNTDOWN_CLEANUP_INTERVAL", DefaultCleanupInterval),
		MaxActiveCountdowns: getEnvInt("COUNTDOWN_MAX_ACTIVE", DefaultMaxActiveCountdowns),
		MessageEditTimeout:  time.Duration(getEnvInt("COUNTDOWN_EDIT_TIMEOUT", DefaultMessageEditTimeout)) * time.Second,
		RateLimitDelay:      time.Duration(getEnvInt("COUNTDOWN_RATE_LIMIT_DELAY", DefaultRateLimitDelay)) * time.Millisecond,
		EnableAutoCleanup:   getEnvBool("COUNTDOWN_ENABLE_AUTO_CLEANUP", true),
		EnableErrorRecovery: getEnvBool("COUNTDOWN_ENABLE_ERROR_RECOVERY", true),
	}

	return config
}

// GetUpdateInterval returns the update interval for countdown messages
func (c *CountdownConfig) GetUpdateInterval() time.Duration {
	return time.Duration(c.UpdateInterval) * time.Second
}

// GetCleanupInterval returns the cleanup interval as a time.Duration
func (c *CountdownConfig) GetCleanupInterval() time.Duration {
	return time.Duration(c.CleanupInterval) * time.Minute
}

// IsWithinLimits checks if the number of active countdowns is within limits
func (c *CountdownConfig) IsWithinLimits(activeCount int) bool {
	return activeCount < c.MaxActiveCountdowns
}

// ShouldAttemptRecovery determines if error recovery should be attempted based on error count
func (c *CountdownConfig) ShouldAttemptRecovery(errorCount int) bool {
	return c.EnableErrorRecovery && errorCount < c.MaxErrors
}

// getEnvInt retrieves an integer value from environment variable with fallback to default
func getEnvInt(key string, defaultValue int) int {
	if envValue := os.Getenv(key); envValue != "" {
		if parsed, err := strconv.Atoi(envValue); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvBool retrieves a boolean value from environment variable with fallback to default
func getEnvBool(key string, defaultValue bool) bool {
	if envValue := os.Getenv(key); envValue != "" {
		if parsed, err := strconv.ParseBool(envValue); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Global countdown configuration instance
var countdownConfig *CountdownConfig

// GetCountdownConfig returns the global countdown configuration instance
// Initializes it on first call
func GetCountdownConfig() *CountdownConfig {
	if countdownConfig == nil {
		countdownConfig = LoadCountdownConfig()
	}
	return countdownConfig
}

// ReloadCountdownConfig forces a reload of the countdown configuration
// Useful for testing or runtime configuration changes
func ReloadCountdownConfig() *CountdownConfig {
	countdownConfig = LoadCountdownConfig()
	return countdownConfig
}
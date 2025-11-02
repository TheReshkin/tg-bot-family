package services

import (
	"github.com/TheReshkin/tg-bot-family/internal/models"
)

// MessageTracker defines the interface for managing countdown messages
// This service is responsible for tracking active countdown messages,
// their state, and coordinating with the Telegram Bot API for live updates.
type MessageTracker interface {
	// AddCountdownMessage registers a new countdown message for tracking
	// Returns error if the countdown already exists for this chat/event combination
	AddCountdownMessage(countdown *models.CountdownMessage) error
	
	// GetCountdownMessage retrieves a countdown message by chat ID and event name
	// Returns nil if not found
	GetCountdownMessage(chatID int64, eventName string) (*models.CountdownMessage, error)
	
	// GetActiveCountdowns returns all countdown messages with status "active"
	GetActiveCountdowns() ([]*models.CountdownMessage, error)
	
	// GetCountdownsByChatID returns all countdown messages for a specific chat
	GetCountdownsByChatID(chatID int64) ([]*models.CountdownMessage, error)
	
	// UpdateCountdownMessage updates an existing countdown message
	// This includes updating the last_updated_at timestamp and any status changes
	UpdateCountdownMessage(countdown *models.CountdownMessage) error
	
	// RemoveCountdownMessage removes a countdown message from tracking
	// Used when countdown is stopped or encounters too many errors
	RemoveCountdownMessage(chatID int64, eventName string) error
	
	// GetCountdownsNeedingUpdate returns countdown messages that should be updated
	// based on their update interval and current status
	GetCountdownsNeedingUpdate() ([]*models.CountdownMessage, error)
	
	// SetCountdownStatus updates the status of a countdown message
	SetCountdownStatus(chatID int64, eventName string, status models.CountdownStatus) error
	
	// IncrementErrorCount increments the error counter for a countdown message
	// Returns the new error count, or error if countdown not found
	IncrementErrorCount(chatID int64, eventName string) (int, error)
	
	// ResetErrorCount resets the error counter for a countdown message to zero
	ResetErrorCount(chatID int64, eventName string) error
	
	// CountdownExists checks if a countdown message exists for the given chat/event
	CountdownExists(chatID int64, eventName string) bool
	
	// GetTotalActiveCountdowns returns the count of currently active countdowns
	GetTotalActiveCountdowns() (int, error)
	
	// CleanupExpiredCountdowns removes countdown messages that have expired
	// and updates their status to "expired" in storage
	CleanupExpiredCountdowns() error
	
	// CleanupErroredCountdowns removes countdown messages with too many consecutive errors
	// Configurable error threshold to prevent infinite retry loops
	CleanupErroredCountdowns(maxErrors int) error
	
	// GetCountdownByMessageID retrieves a countdown by Telegram message ID
	// Useful for handling callback queries or message-specific operations
	GetCountdownByMessageID(chatID int64, messageID int) (*models.CountdownMessage, error)
	
	// UpdateMessageID updates the Telegram message ID for a countdown
	// Used when messages are recreated or edited
	UpdateMessageID(chatID int64, eventName string, newMessageID int) error
}
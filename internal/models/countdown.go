package models

import (
	"fmt"
	"time"
)

// CountdownStatus represents the current state of a countdown
type CountdownStatus string

const (
	CountdownActive   CountdownStatus = "active"
	CountdownStopped  CountdownStatus = "stopped"
	CountdownExpired  CountdownStatus = "expired"
	CountdownError    CountdownStatus = "error"
)

// CountdownMessage represents a live countdown message that gets updated periodically
type CountdownMessage struct {
	// MessageID is the Telegram message ID for editMessageText
	MessageID int `json:"message_id"`
	
	// ChatID is the Telegram chat ID where the countdown is displayed
	ChatID int64 `json:"chat_id"`
	
	// EventName is the name of the event being counted down to
	EventName string `json:"event_name"`
	
	// EventDate is the target date/time for the countdown (YYYY-MM-DD HH:MM format)
	EventDate string `json:"event_date"`
	
	// Description is optional description text for the event
	Description string `json:"description"`
	
	// Status indicates the current state of the countdown
	Status CountdownStatus `json:"status"`
	
	// CreatedAt tracks when the countdown was started
	CreatedAt time.Time `json:"created_at"`
	
	// LastUpdatedAt tracks the last time the message was updated
	LastUpdatedAt time.Time `json:"last_updated_at"`
	
	// UpdateInterval defines how often the countdown should be updated (in seconds)
	UpdateInterval int `json:"update_interval"`
	
	// ErrorCount tracks consecutive errors for this countdown
	ErrorCount int `json:"error_count"`
}

// NewCountdownMessage creates a new CountdownMessage instance
func NewCountdownMessage(chatID int64, messageID int, eventName, eventDate, description string) *CountdownMessage {
	now := time.Now()
	return &CountdownMessage{
		MessageID:      messageID,
		ChatID:         chatID,
		EventName:      eventName,
		EventDate:      eventDate,
		Description:    description,
		Status:         CountdownActive,
		CreatedAt:      now,
		LastUpdatedAt:  now,
		UpdateInterval: 60, // Default: update every 60 seconds
		ErrorCount:     0,
	}
}

// IsValid validates the CountdownMessage fields
func (cm *CountdownMessage) IsValid() error {
	if cm.MessageID <= 0 {
		return fmt.Errorf("invalid message_id: must be positive")
	}
	
	if cm.ChatID == 0 {
		return fmt.Errorf("invalid chat_id: cannot be zero")
	}
	
	if !IsValidEventName(cm.EventName) {
		return fmt.Errorf("invalid event_name: must contain only letters, numbers, and underscores")
	}
	
	if !IsValidDate(cm.EventDate) {
		return fmt.Errorf("invalid event_date: must be in YYYY-MM-DD or YYYY-MM-DD HH:MM format")
	}
	
	if cm.UpdateInterval <= 0 {
		return fmt.Errorf("invalid update_interval: must be positive")
	}
	
	return nil
}

// IsExpired checks if the countdown target date has passed
func (cm *CountdownMessage) IsExpired() bool {
	targetDate, err := ParseEventDate(cm.EventDate)
	if err != nil {
		return false // If we can't parse the date, don't consider it expired
	}
	
	return time.Now().After(targetDate)
}

// ShouldUpdate determines if the countdown message should be updated based on the interval
func (cm *CountdownMessage) ShouldUpdate() bool {
	if cm.Status != CountdownActive {
		return false
	}
	
	if cm.IsExpired() {
		return true // Update one final time to show "expired"
	}
	
	timeSinceUpdate := time.Since(cm.LastUpdatedAt)
	return timeSinceUpdate.Seconds() >= float64(cm.UpdateInterval)
}

// MarkUpdated updates the LastUpdatedAt timestamp
func (cm *CountdownMessage) MarkUpdated() {
	cm.LastUpdatedAt = time.Now()
}

// IncrementErrorCount increments the error counter
func (cm *CountdownMessage) IncrementErrorCount() {
	cm.ErrorCount++
}

// ResetErrorCount resets the error counter to zero
func (cm *CountdownMessage) ResetErrorCount() {
	cm.ErrorCount = 0
}

// SetStatus updates the countdown status
func (cm *CountdownMessage) SetStatus(status CountdownStatus) {
	cm.Status = status
}

// GetTimeRemaining calculates the time remaining until the target date
func (cm *CountdownMessage) GetTimeRemaining() (time.Duration, error) {
	targetDate, err := ParseEventDate(cm.EventDate)
	if err != nil {
		return 0, fmt.Errorf("failed to parse event date: %w", err)
	}
	
	duration := time.Until(targetDate)
	return duration, nil
}

// FormatCountdownMessage generates the formatted countdown message text
func (cm *CountdownMessage) FormatCountdownMessage() string {
	duration, err := cm.GetTimeRemaining()
	if err != nil {
		return fmt.Sprintf("🚫 Ошибка: %s", err.Error())
	}
	
	message := fmt.Sprintf("🕒 Событие: %s\n📅 Дата: %s\n", cm.EventName, cm.EventDate)
	
	if cm.Description != "" {
		message += fmt.Sprintf("📝 Описание: %s\n", cm.Description)
	}
	
	if duration > 0 {
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24
		minutes := int(duration.Minutes()) % 60
		
		message += fmt.Sprintf("⏰ Осталось: %d дней, %d часов, %d минут\n", days, hours, minutes)
	} else {
		message += "✅ Событие наступило!\n"
	}
	
	// Add last update timestamp
	updateTime := cm.LastUpdatedAt.Format("15:04")
	message += fmt.Sprintf("\n🔄 Последнее обновление: %s", updateTime)
	
	return message
}

// GetUniqueKey returns a unique identifier for this countdown message
func (cm *CountdownMessage) GetUniqueKey() string {
	return fmt.Sprintf("%d_%s", cm.ChatID, cm.EventName)
}
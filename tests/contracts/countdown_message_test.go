package contracts

import (
	"testing"
	"time"

	"github.com/TheReshkin/tg-bot-family/internal/models"
)

// TestCountdownMessageCreationContract tests the contract for creating countdown messages
// This test MUST FAIL initially before implementation (TDD requirement)
func TestCountdownMessageCreationContract(t *testing.T) {
	// Contract: CountdownMessage can be created with valid parameters
	chatID := int64(123456789)
	messageID := 42
	eventName := "test_event"
	eventDate := "2025-12-31 23:59"
	description := "Test countdown event"

	// Test: NewCountdownMessage constructor should create valid instance
	countdown := models.NewCountdownMessage(chatID, messageID, eventName, eventDate, description)
	
	if countdown == nil {
		t.Fatal("NewCountdownMessage должна возвращать валидный объект")
	}

	// Contract: All required fields should be set correctly
	if countdown.ChatID != chatID {
		t.Errorf("ChatID: ожидалось %d, получено %d", chatID, countdown.ChatID)
	}
	
	if countdown.MessageID != messageID {
		t.Errorf("MessageID: ожидалось %d, получено %d", messageID, countdown.MessageID)
	}
	
	if countdown.EventName != eventName {
		t.Errorf("EventName: ожидалось %s, получено %s", eventName, countdown.EventName)
	}
	
	if countdown.EventDate != eventDate {
		t.Errorf("EventDate: ожидалось %s, получено %s", eventDate, countdown.EventDate)
	}
	
	if countdown.Description != description {
		t.Errorf("Description: ожидалось %s, получено %s", description, countdown.Description)
	}

	// Contract: Default status should be active
	if countdown.Status != models.CountdownActive {
		t.Errorf("Status: ожидался %s, получен %s", models.CountdownActive, countdown.Status)
	}

	// Contract: Default update interval should be 60 seconds
	if countdown.UpdateInterval != 60 {
		t.Errorf("UpdateInterval: ожидалось 60, получено %d", countdown.UpdateInterval)
	}

	// Contract: Error count should start at zero
	if countdown.ErrorCount != 0 {
		t.Errorf("ErrorCount: ожидалось 0, получено %d", countdown.ErrorCount)
	}

	// Contract: Timestamps should be set to current time (within 1 second tolerance)
	now := time.Now()
	if time.Since(countdown.CreatedAt) > time.Second {
		t.Error("CreatedAt должен быть установлен на текущее время")
	}
	
	if time.Since(countdown.LastUpdatedAt) > time.Second {
		t.Error("LastUpdatedAt должен быть установлен на текущее время")
	}
}

// TestCountdownMessageValidationContract tests the validation contract
func TestCountdownMessageValidationContract(t *testing.T) {
	// Contract: Valid countdown message should pass validation
	validCountdown := models.NewCountdownMessage(123456789, 42, "valid_event", "2025-12-31 23:59", "Test")
	
	if err := validCountdown.IsValid(); err != nil {
		t.Errorf("Валидный countdown должен проходить валидацию: %v", err)
	}

	// Contract: Invalid message ID should fail validation
	invalidMessageID := models.NewCountdownMessage(123456789, 0, "valid_event", "2025-12-31 23:59", "Test")
	if err := invalidMessageID.IsValid(); err == nil {
		t.Error("Countdown с messageID=0 должен не проходить валидацию")
	}

	// Contract: Invalid chat ID should fail validation
	invalidChatID := models.NewCountdownMessage(0, 42, "valid_event", "2025-12-31 23:59", "Test")
	if err := invalidChatID.IsValid(); err == nil {
		t.Error("Countdown с chatID=0 должен не проходить валидацию")
	}

	// Contract: Invalid event name should fail validation
	invalidEventName := models.NewCountdownMessage(123456789, 42, "invalid event!", "2025-12-31 23:59", "Test")
	if err := invalidEventName.IsValid(); err == nil {
		t.Error("Countdown с некорректным именем события должен не проходить валидацию")
	}

	// Contract: Invalid date should fail validation
	invalidDate := models.NewCountdownMessage(123456789, 42, "valid_event", "invalid-date", "Test")
	if err := invalidDate.IsValid(); err == nil {
		t.Error("Countdown с некорректной датой должен не проходить валидацию")
	}

	// Contract: Invalid update interval should fail validation
	invalidInterval := models.NewCountdownMessage(123456789, 42, "valid_event", "2025-12-31 23:59", "Test")
	invalidInterval.UpdateInterval = 0
	if err := invalidInterval.IsValid(); err == nil {
		t.Error("Countdown с updateInterval=0 должен не проходить валидацию")
	}
}

// TestCountdownMessageStateContract tests state management contract
func TestCountdownMessageStateContract(t *testing.T) {
	countdown := models.NewCountdownMessage(123456789, 42, "test_event", "2025-12-31 23:59", "Test")

	// Contract: ShouldUpdate should return false for non-active countdowns
	countdown.SetStatus(models.CountdownStopped)
	if countdown.ShouldUpdate() {
		t.Error("Остановленный countdown не должен обновляться")
	}

	// Contract: Active countdown should update based on interval
	countdown.SetStatus(models.CountdownActive)
	countdown.LastUpdatedAt = time.Now().Add(-2 * time.Minute) // 2 minutes ago
	if !countdown.ShouldUpdate() {
		t.Error("Активный countdown должен обновляться через 60 секунд")
	}

	// Contract: Fresh countdown should not need update
	countdown.MarkUpdated()
	if countdown.ShouldUpdate() {
		t.Error("Свежий countdown не должен требовать обновления")
	}

	// Contract: Error count should increment and reset correctly
	initialErrors := countdown.ErrorCount
	countdown.IncrementErrorCount()
	if countdown.ErrorCount != initialErrors+1 {
		t.Error("ErrorCount должен увеличиваться на 1")
	}

	countdown.ResetErrorCount()
	if countdown.ErrorCount != 0 {
		t.Error("ResetErrorCount должен обнулять счётчик ошибок")
	}
}

// TestCountdownMessageFormattingContract tests message formatting contract
func TestCountdownMessageFormattingContract(t *testing.T) {
	// Contract: Message formatting should include all required elements
	futureDate := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04")
	countdown := models.NewCountdownMessage(123456789, 42, "test_event", futureDate, "Test description")
	
	message := countdown.FormatCountdownMessage()
	
	// Contract: Message should contain event name
	if !contains(message, "test_event") {
		t.Error("Форматированное сообщение должно содержать имя события")
	}
	
	// Contract: Message should contain date
	if !contains(message, futureDate) {
		t.Error("Форматированное сообщение должно содержать дату")
	}
	
	// Contract: Message should contain description
	if !contains(message, "Test description") {
		t.Error("Форматированное сообщение должно содержать описание")
	}
	
	// Contract: Message should contain countdown (for future events)
	if !contains(message, "Осталось:") {
		t.Error("Форматированное сообщение должно содержать обратный отсчёт")
	}
	
	// Contract: Message should contain update timestamp
	if !contains(message, "Последнее обновление:") {
		t.Error("Форматированное сообщение должно содержать время обновления")
	}
}

// TestCountdownMessageExpirationContract tests expiration logic contract
func TestCountdownMessageExpirationContract(t *testing.T) {
	// Contract: Future event should not be expired
	futureDate := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04")
	futureCountdown := models.NewCountdownMessage(123456789, 42, "future_event", futureDate, "Test")
	
	if futureCountdown.IsExpired() {
		t.Error("Будущее событие не должно быть просроченным")
	}

	// Contract: Past event should be expired
	pastDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04")
	pastCountdown := models.NewCountdownMessage(123456789, 42, "past_event", pastDate, "Test")
	
	if !pastCountdown.IsExpired() {
		t.Error("Прошедшее событие должно быть просроченным")
	}

	// Contract: Expired countdown should update one final time
	pastCountdown.SetStatus(models.CountdownActive)
	if !pastCountdown.ShouldUpdate() {
		t.Error("Просроченный countdown должен обновиться финальный раз")
	}
}

// TestCountdownMessageUniqueKeyContract tests unique key generation contract
func TestCountdownMessageUniqueKeyContract(t *testing.T) {
	// Contract: Unique key should combine chat ID and event name
	countdown1 := models.NewCountdownMessage(123456789, 42, "event1", "2025-12-31 23:59", "Test")
	countdown2 := models.NewCountdownMessage(123456789, 43, "event2", "2025-12-31 23:59", "Test")
	countdown3 := models.NewCountdownMessage(987654321, 44, "event1", "2025-12-31 23:59", "Test")

	key1 := countdown1.GetUniqueKey()
	key2 := countdown2.GetUniqueKey()
	key3 := countdown3.GetUniqueKey()

	// Contract: Different events in same chat should have different keys
	if key1 == key2 {
		t.Error("Разные события в одном чате должны иметь разные ключи")
	}

	// Contract: Same event in different chats should have different keys
	if key1 == key3 {
		t.Error("Одинаковые события в разных чатах должны иметь разные ключи")
	}

	// Contract: Key should contain both chat ID and event name
	expectedKey1 := "123456789_event1"
	if key1 != expectedKey1 {
		t.Errorf("Ключ должен быть %s, получен %s", expectedKey1, key1)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 indexOf(s, substr) >= 0))
}

// Simple indexOf implementation
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
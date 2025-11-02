package integration

import (
	"context"
	"testing"
	"time"

	"github.com/TheReshkin/tg-bot-family/internal/config"
	"github.com/TheReshkin/tg-bot-family/internal/models"
	"github.com/TheReshkin/tg-bot-family/internal/services"
	"github.com/TheReshkin/tg-bot-family/internal/storage"
)

// TestCountdownIntegrationFlow tests the complete live countdown flow
// This test MUST FAIL initially before implementation (TDD requirement)
func TestCountdownIntegrationFlow(t *testing.T) {
	// Setup: Create test dependencies
	store := storage.NewJSONStorage()
	eventService := services.NewEventService(store)
	
	// This will fail until MessageTracker and CountdownService are implemented
	// messageTracker := services.NewMessageTracker(store)
	// countdownService := services.NewCountdownService(messageTracker, eventService)

	// Test data
	chatID := int64(123456789)
	eventName := "integration_test_event"
	eventDate := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04")
	description := "Integration test countdown event"
	messageID := 42

	// Step 1: Create an event first
	err := eventService.CreateEvent(chatID, eventName, eventDate, description)
	if err != nil {
		t.Fatalf("Не удалось создать событие для интеграционного теста: %v", err)
	}

	// Step 2: Create countdown message (will fail until implemented)
	countdown := models.NewCountdownMessage(chatID, messageID, eventName, eventDate, description)
	if countdown == nil {
		t.Fatal("Не удалось создать CountdownMessage")
	}

	// Step 3: Validate countdown message
	if err := countdown.IsValid(); err != nil {
		t.Fatalf("CountdownMessage не прошёл валидацию: %v", err)
	}

	// Step 4: Test countdown formatting
	formattedMessage := countdown.FormatCountdownMessage()
	if formattedMessage == "" {
		t.Error("Форматированное сообщение не должно быть пустым")
	}

	// Step 5: Test that countdown should update (future event)
	if !countdown.ShouldUpdate() {
		t.Error("Новый countdown должен требовать обновления")
	}

	// Step 6: Test countdown state transitions
	countdown.MarkUpdated()
	if countdown.ShouldUpdate() {
		t.Error("Обновлённый countdown не должен сразу требовать нового обновления")
	}

	// Step 7: Test error handling
	countdown.IncrementErrorCount()
	if countdown.ErrorCount != 1 {
		t.Error("Счётчик ошибок должен увеличиться")
	}

	// Step 8: Test status changes
	countdown.SetStatus(models.CountdownStopped)
	if countdown.Status != models.CountdownStopped {
		t.Error("Статус countdown должен измениться на stopped")
	}

	// This test will fail here until MessageTracker is implemented
	t.Log("ОЖИДАЕМАЯ ОШИБКА: MessageTracker сервис не реализован")
}

// TestCountdownServiceIntegration tests countdown service integration
// This test MUST FAIL initially before implementation (TDD requirement)
func TestCountdownServiceIntegration(t *testing.T) {
	// Setup test environment
	store := storage.NewJSONStorage()
	eventService := services.NewEventService(store)
	
	// Create test event
	chatID := int64(123456789)
	eventName := "service_test_event"
	eventDate := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04")
	
	err := eventService.CreateEvent(chatID, eventName, eventDate, "Service test")
	if err != nil {
		t.Fatalf("Не удалось создать тестовое событие: %v", err)
	}

	// Test integration between Event and Countdown
	event, err := eventService.GetEvent(chatID, eventName)
	if err != nil {
		t.Fatalf("Не удалось получить созданное событие: %v", err)
	}

	// Create countdown from event
	messageID := 100
	countdown := models.NewCountdownMessage(event.ChatID, messageID, event.Name, event.Date, event.Description)
	
	// Validate integration
	if countdown.ChatID != event.ChatID {
		t.Error("ChatID должен совпадать между Event и CountdownMessage")
	}
	
	if countdown.EventName != event.Name {
		t.Error("EventName должен совпадать между Event и CountdownMessage")
	}
	
	if countdown.EventDate != event.Date {
		t.Error("EventDate должен совпадать между Event и CountdownMessage")
	}

	// Test countdown calculation
	timeRemaining, err := countdown.GetTimeRemaining()
	if err != nil {
		t.Fatalf("Не удалось рассчитать оставшееся время: %v", err)
	}
	
	if timeRemaining <= 0 {
		t.Error("Оставшееся время должно быть положительным для будущего события")
	}

	// This test will fail here until CountdownService is implemented
	t.Log("ОЖИДАЕМАЯ ОШИБКА: CountdownService не реализован")
}

// TestCountdownConfigurationIntegration tests configuration integration
func TestCountdownConfigurationIntegration(t *testing.T) {
	// Test config loading
	countdownConfig := config.GetCountdownConfig()
	if countdownConfig == nil {
		t.Fatal("Конфигурация countdown не должна быть nil")
	}

	// Test default values
	if countdownConfig.UpdateInterval <= 0 {
		t.Error("UpdateInterval должен быть положительным")
	}

	if countdownConfig.MaxErrors <= 0 {
		t.Error("MaxErrors должен быть положительным")
	}

	if countdownConfig.MaxActiveCountdowns <= 0 {
		t.Error("MaxActiveCountdowns должен быть положительным")
	}

	// Test configuration methods
	updateDuration := countdownConfig.GetUpdateInterval()
	if updateDuration <= 0 {
		t.Error("GetUpdateInterval должен возвращать положительную длительность")
	}

	cleanupDuration := countdownConfig.GetCleanupInterval()
	if cleanupDuration <= 0 {
		t.Error("GetCleanupInterval должен возвращать положительную длительность")
	}

	// Test limits checking
	if !countdownConfig.IsWithinLimits(0) {
		t.Error("0 активных countdown должно быть в пределах лимитов")
	}

	if countdownConfig.IsWithinLimits(countdownConfig.MaxActiveCountdowns + 1) {
		t.Error("Превышение лимита не должно быть разрешено")
	}

	// Test error recovery logic
	if !countdownConfig.ShouldAttemptRecovery(1) {
		t.Error("Должна разрешаться попытка восстановления с 1 ошибкой")
	}

	if countdownConfig.ShouldAttemptRecovery(countdownConfig.MaxErrors + 1) {
		t.Error("Не должна разрешаться попытка восстановления при превышении лимита ошибок")
	}
}

// TestCountdownMessageLifecycleIntegration tests complete message lifecycle
func TestCountdownMessageLifecycleIntegration(t *testing.T) {
	// Test complete lifecycle: creation -> active -> updating -> stopped/expired
	
	// Phase 1: Creation
	chatID := int64(123456789)
	messageID := 200
	eventName := "lifecycle_test"
	futureDate := time.Now().Add(30 * time.Minute).Format("2006-01-02 15:04")
	
	countdown := models.NewCountdownMessage(chatID, messageID, eventName, futureDate, "Lifecycle test")
	
	// Validate initial state
	if countdown.Status != models.CountdownActive {
		t.Error("Новый countdown должен иметь статус active")
	}
	
	if countdown.ErrorCount != 0 {
		t.Error("Новый countdown должен иметь ErrorCount = 0")
	}

	// Phase 2: Active state
	if !countdown.ShouldUpdate() {
		t.Error("Новый countdown должен требовать обновления")
	}
	
	formattedMsg1 := countdown.FormatCountdownMessage()
	if !contains(formattedMsg1, "Осталось:") {
		t.Error("Сообщение активного countdown должно содержать 'Осталось:'")
	}

	// Phase 3: Update cycle
	countdown.MarkUpdated()
	lastUpdate := countdown.LastUpdatedAt
	
	time.Sleep(10 * time.Millisecond) // Small delay to ensure time difference
	
	countdown.MarkUpdated()
	if !countdown.LastUpdatedAt.After(lastUpdate) {
		t.Error("LastUpdatedAt должен обновляться при вызове MarkUpdated")
	}

	// Phase 4: Error handling
	for i := 0; i < 3; i++ {
		countdown.IncrementErrorCount()
	}
	
	if countdown.ErrorCount != 3 {
		t.Errorf("ErrorCount должен быть 3, получен %d", countdown.ErrorCount)
	}
	
	countdown.ResetErrorCount()
	if countdown.ErrorCount != 0 {
		t.Error("ErrorCount должен сброситься в 0")
	}

	// Phase 5: Status transitions
	countdown.SetStatus(models.CountdownStopped)
	if countdown.ShouldUpdate() {
		t.Error("Остановленный countdown не должен требовать обновления")
	}
	
	countdown.SetStatus(models.CountdownExpired)
	if countdown.Status != models.CountdownExpired {
		t.Error("Статус должен измениться на expired")
	}

	// Phase 6: Expiration test with past date
	pastDate := time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04")
	expiredCountdown := models.NewCountdownMessage(chatID, 201, "expired_test", pastDate, "Expired test")
	
	if !expiredCountdown.IsExpired() {
		t.Error("Countdown с прошедшей датой должен быть expired")
	}
	
	formattedMsg2 := expiredCountdown.FormatCountdownMessage()
	if !contains(formattedMsg2, "Событие наступило!") {
		t.Error("Сообщение просроченного countdown должно содержать 'Событие наступило!'")
	}
}

// TestCountdownStorageIntegration tests storage integration
// This test MUST FAIL initially before storage implementation
func TestCountdownStorageIntegration(t *testing.T) {
	// Test storage operations for countdown messages
	store := storage.NewJSONStorage()
	
	// Create test countdown data
	chatID := int64(123456789)
	messageID := 300
	eventName := "storage_test"
	eventDate := time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04")
	
	countdown := models.NewCountdownMessage(chatID, messageID, eventName, eventDate, "Storage test")
	
	// This will fail until storage methods for countdown are implemented
	t.Log("ОЖИДАЕМАЯ ОШИБКА: Методы хранения для countdown не реализованы")
	
	// Expected storage operations (will fail):
	// err := store.SaveCountdownMessage(countdown)
	// retrievedCountdown, err := store.GetCountdownMessage(chatID, eventName)
	// countdowns, err := store.GetActiveCountdowns()
	// err = store.RemoveCountdownMessage(chatID, eventName)
	
	t.Log("Требуется реализация методов: SaveCountdownMessage, GetCountdownMessage, GetActiveCountdowns, RemoveCountdownMessage")
}

// Helper function (reused from other tests)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
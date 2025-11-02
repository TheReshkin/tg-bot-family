package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/TheReshkin/tg-bot-family/internal/config"
	"github.com/TheReshkin/tg-bot-family/internal/models"
)

const eventsFile = "./data/events.json"

type Storage interface {
	SaveEvent(chatID int64, event models.Event) error
	GetEvents(chatID int64) ([]models.Event, error)
	GetAllEvents() ([]models.Event, error)
	GetEvent(chatID int64, name string) (*models.Event, error)
	FindEventAcrossChats(name string, excludeChatID int64) (*models.Event, int64, error)
	EventExists(chatID int64, name string) bool
	GetUser(chatID, userID int64) (*models.User, error)
	AddEventToUser(chatID, userID int64, event models.Event) error

	// Countdown message storage methods
	SaveCountdownMessage(countdown *models.CountdownMessage) error
	GetCountdownMessage(chatID int64, eventName string) (*models.CountdownMessage, error)
	GetCountdownsByChatID(chatID int64) ([]*models.CountdownMessage, error)
	GetActiveCountdowns() ([]*models.CountdownMessage, error)
	UpdateCountdownMessage(countdown *models.CountdownMessage) error
	RemoveCountdownMessage(chatID int64, eventName string) error
	GetCountdownByMessageID(chatID int64, messageID int) (*models.CountdownMessage, error)
	CountdownExists(chatID int64, eventName string) bool
}

type JSONStorage struct {
	mu sync.RWMutex
}

func NewJSONStorage() *JSONStorage {
	return &JSONStorage{}
}

type ChatData struct {
	ChatID     int64                      `json:"chat_id"`
	Events     []models.Event             `json:"events"`
	Users      []models.User              `json:"users"`
	Countdowns []models.CountdownMessage  `json:"countdowns"`
}

func (s *JSONStorage) loadData() ([]ChatData, error) {
	file, err := os.Open(eventsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []ChatData{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var data []ChatData
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *JSONStorage) saveData(data []ChatData) error {
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		err := os.Mkdir("./data", os.ModePerm)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(eventsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(data)
}

func (s *JSONStorage) SaveEvent(chatID int64, event models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadData()
	if err != nil {
		return err
	}

	found := false
	for i, chat := range data {
		if chat.ChatID == chatID {
			data[i].Events = append(data[i].Events, event)
			found = true
			break
		}
	}
	if !found {
		data = append(data, ChatData{
			ChatID:     chatID,
			Events:     []models.Event{event},
			Users:      []models.User{},
			Countdowns: []models.CountdownMessage{},
		})
	}

	return s.saveData(data)
}

func (s *JSONStorage) GetEvents(chatID int64) ([]models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	for _, chat := range data {
		if chat.ChatID == chatID {
			return chat.Events, nil
		}
	}
	return []models.Event{}, nil
}

func (s *JSONStorage) GetAllEvents() ([]models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	var allEvents []models.Event
	for _, chat := range data {
		allEvents = append(allEvents, chat.Events...)
	}
	return allEvents, nil
}

func (s *JSONStorage) GetEvent(chatID int64, name string) (*models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	for _, chat := range data {
		if chat.ChatID == chatID {
			for _, event := range chat.Events {
				if event.Name == name {
					return &event, nil
				}
			}
		}
	}
	return nil, errors.New("event not found")
}

func (s *JSONStorage) FindEventAcrossChats(name string, excludeChatID int64) (*models.Event, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, 0, err
	}

	// First, try to find in the test chat (prioritize it for cross-chat searches)
	testChatID := config.LoadTestChatID()
	if excludeChatID != testChatID {
		for _, chat := range data {
			if chat.ChatID == testChatID {
				for _, event := range chat.Events {
					if event.Name == name {
						return &event, testChatID, nil
					}
				}
			}
		}
	}

	// If not found in test chat, search other chats
	for _, chat := range data {
		if chat.ChatID != excludeChatID && chat.ChatID != testChatID {
			for _, event := range chat.Events {
				if event.Name == name {
					return &event, chat.ChatID, nil
				}
			}
		}
	}

	return nil, 0, errors.New("event not found")
}

func (s *JSONStorage) EventExists(chatID int64, name string) bool {
	_, err := s.GetEvent(chatID, name)
	return err == nil
}

func (s *JSONStorage) GetUser(chatID, userID int64) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	for _, chat := range data {
		if chat.ChatID == chatID {
			for _, user := range chat.Users {
				if user.UserID == userID {
					return &user, nil
				}
			}
		}
	}
	return nil, errors.New("user not found")
}

func (s *JSONStorage) AddEventToUser(chatID, userID int64, event models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadData()
	if err != nil {
		return err
	}

	found := false
	for i, chat := range data {
		if chat.ChatID == chatID {
			for j, user := range chat.Users {
				if user.UserID == userID {
					data[i].Users[j].Events = append(data[i].Users[j].Events, event)
					found = true
					break
				}
			}
			if !found {
				newUser := models.User{
					UserID: userID,
					ChatID: chatID,
					Events: []models.Event{event},
				}
				data[i].Users = append(data[i].Users, newUser)
				found = true
			}
			break
		}
	}
	if !found {
		newUser := models.User{
			UserID: userID,
			ChatID: chatID,
			Events: []models.Event{event},
		}
		data = append(data, ChatData{
			ChatID:     chatID,
			Events:     []models.Event{},
			Users:      []models.User{newUser},
			Countdowns: []models.CountdownMessage{},
		})
	}

	return s.saveData(data)
}

// Countdown message storage methods

func (s *JSONStorage) SaveCountdownMessage(countdown *models.CountdownMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadData()
	if err != nil {
		return err
	}

	found := false
	for i, chat := range data {
		if chat.ChatID == countdown.ChatID {
			// Check if countdown already exists and update it
			for j, existingCountdown := range chat.Countdowns {
				if existingCountdown.EventName == countdown.EventName {
					data[i].Countdowns[j] = *countdown
					found = true
					break
				}
			}
			if !found {
				data[i].Countdowns = append(data[i].Countdowns, *countdown)
				found = true
			}
			break
		}
	}
	if !found {
		data = append(data, ChatData{
			ChatID:     countdown.ChatID,
			Events:     []models.Event{},
			Users:      []models.User{},
			Countdowns: []models.CountdownMessage{*countdown},
		})
	}

	return s.saveData(data)
}

func (s *JSONStorage) GetCountdownMessage(chatID int64, eventName string) (*models.CountdownMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	for _, chat := range data {
		if chat.ChatID == chatID {
			for _, countdown := range chat.Countdowns {
				if countdown.EventName == eventName {
					return &countdown, nil
				}
			}
		}
	}
	return nil, errors.New("countdown message not found")
}

func (s *JSONStorage) GetCountdownsByChatID(chatID int64) ([]*models.CountdownMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	for _, chat := range data {
		if chat.ChatID == chatID {
			countdowns := make([]*models.CountdownMessage, len(chat.Countdowns))
			for i := range chat.Countdowns {
				countdowns[i] = &chat.Countdowns[i]
			}
			return countdowns, nil
		}
	}
	return []*models.CountdownMessage{}, nil
}

func (s *JSONStorage) GetActiveCountdowns() ([]*models.CountdownMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	var activeCountdowns []*models.CountdownMessage
	for _, chat := range data {
		for i, countdown := range chat.Countdowns {
			if countdown.Status == models.CountdownActive {
				activeCountdowns = append(activeCountdowns, &chat.Countdowns[i])
			}
		}
	}
	return activeCountdowns, nil
}

func (s *JSONStorage) UpdateCountdownMessage(countdown *models.CountdownMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadData()
	if err != nil {
		return err
	}

	for i, chat := range data {
		if chat.ChatID == countdown.ChatID {
			for j, existingCountdown := range chat.Countdowns {
				if existingCountdown.EventName == countdown.EventName {
					data[i].Countdowns[j] = *countdown
					return s.saveData(data)
				}
			}
		}
	}
	return errors.New("countdown message not found for update")
}

func (s *JSONStorage) RemoveCountdownMessage(chatID int64, eventName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadData()
	if err != nil {
		return err
	}

	for i, chat := range data {
		if chat.ChatID == chatID {
			for j, countdown := range chat.Countdowns {
				if countdown.EventName == eventName {
					// Remove countdown from slice
					data[i].Countdowns = append(data[i].Countdowns[:j], data[i].Countdowns[j+1:]...)
					return s.saveData(data)
				}
			}
		}
	}
	return errors.New("countdown message not found for removal")
}

func (s *JSONStorage) GetCountdownByMessageID(chatID int64, messageID int) (*models.CountdownMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.loadData()
	if err != nil {
		return nil, err
	}

	for _, chat := range data {
		if chat.ChatID == chatID {
			for _, countdown := range chat.Countdowns {
				if countdown.MessageID == messageID {
					return &countdown, nil
				}
			}
		}
	}
	return nil, errors.New("countdown message not found by message ID")
}

func (s *JSONStorage) CountdownExists(chatID int64, eventName string) bool {
	_, err := s.GetCountdownMessage(chatID, eventName)
	return err == nil
}

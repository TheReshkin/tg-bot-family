package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const datesFile = "dates.json"

type DateEntry struct {
	Date string `json:"date"`
	Name string `json:"name"`
}

type ChatDates struct {
	ChatID int64       `json:"chat_id"`
	Dates  []DateEntry `json:"dates"`
}

var baseCommands []models.BotCommand

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка при загрузке .env файла")
	}

	telegram_token := os.Getenv("TELEGRAM_TOKEN")
	// Инициализация бота с токеном
	b, err := bot.New(telegram_token)
	if err != nil {
		log.Fatal(err)
	}

	// Регистрируем обработчики команд
	b.RegisterHandler(bot.HandlerTypeMessageText, "/setdate", bot.MatchTypePrefix, setDateHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/dates", bot.MatchTypeExact, listDatesHandler)

	// Устанавливаем базовые команды
	baseCommands := []models.BotCommand{
		{Command: "setdate", Description: "Добавить новую дату (/setdate YYYY-MM-DD [название])"},
		{Command: "dates", Description: "Показать список всех запланированных дат"},
	}
	updateBotCommands(b, baseCommands)

	// Запускаем бота
	b.Start(context.Background())
}

func setDateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Получаем дату и название из сообщения
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) < 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Используйте формат: /setdate YYYY-MM-DD [HH:MM] [название]",
		})
		return
	}

	dateTime := parts[1]
	if len(parts) > 2 && strings.Contains(parts[2], ":") {
		dateTime += " " + parts[2]
		parts = append(parts[:2], parts[3:]...)
	}

	parsedDate, err := parseDateWithTimezone(dateTime)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Неверный формат даты. Используйте формат: YYYY-MM-DD [HH:MM]",
		})
		return
	}

	name := ""
	if len(parts) > 2 {
		name = parts[2]
	}

	// Загружаем существующие даты
	chatDates := loadDates()

	// Добавляем дату для текущего чата
	found := false
	for i, chat := range chatDates {
		if chat.ChatID == update.Message.Chat.ID {
			chatDates[i].Dates = append(chatDates[i].Dates, DateEntry{Date: parsedDate.Format("2006-01-02 15:04"), Name: name})
			found = true
			break
		}
	}

	if !found {
		chatDates = append(chatDates, ChatDates{
			ChatID: update.Message.Chat.ID,
			Dates:  []DateEntry{{Date: parsedDate.Format("2006-01-02 15:04"), Name: name}},
		})
	}

	// Сохраняем даты обратно в файл
	saveDates(chatDates)

	// Регистрируем новую команду, если указано название
	if name != "" {
		command := "/" + name
		b.RegisterHandler(bot.HandlerTypeMessageText, command, bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
			handleDynamicCommand(ctx, b, update, name)
		})

		// Добавляем описание для новой команды
		newCommand := models.BotCommand{
			Command:     name,
			Description: fmt.Sprintf("Показать время до события '%s'", name),
		}
		baseCommands = append(baseCommands, newCommand)
		err := updateBotCommands(b, baseCommands)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Дата добавлена, но команда '%s' не была зарегистрирована: %v", command, err),
			})
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Дата успешно добавлена! Используйте команду %s для просмотра.", command),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Дата успешно добавлена!",
	})
}

func listDatesHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Загружаем даты из файла
	chatDates := loadDates()

	// Загружаем часовой пояс Europe/Moscow
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Ошибка загрузки часового пояса: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка загрузки часового пояса.",
		})
		return
	}

	// Ищем даты для текущего чата
	for _, chat := range chatDates {
		if chat.ChatID == update.Message.Chat.ID {
			if len(chat.Dates) == 0 {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Нет запланированных дат.",
				})
				return
			}

			message := "Запланированные даты:\n"
			now := time.Now().In(location) // Преобразуем текущее время в Europe/Moscow

			for i, entry := range chat.Dates {
				parsedDate, err := time.ParseInLocation("2006-01-02 15:04", entry.Date, location)
				if err != nil {
					message += fmt.Sprintf("%d. %s (неверный формат даты)\n", i+1, entry.Date)
					continue
				}

				// Рассчитываем оставшееся время
				duration := parsedDate.Sub(now)
				if duration < 0 {
					message += fmt.Sprintf("%d. %s (%s) (уже прошло)\n", i+1, entry.Date, entry.Name)
				} else {
					days := int(duration.Hours()) / 24
					hours := int(duration.Hours()) % 24
					minutes := int(duration.Minutes()) % 60
					message += fmt.Sprintf("%d. %s (%s) (осталось: %d дней, %d часов, %d минут)\n", i+1, entry.Date, entry.Name, days, hours, minutes)
				}
			}

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   message,
			})
			return
		}
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Нет запланированных дат.",
	})
}

func loadDates() []ChatDates {
	file, err := os.Open(datesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []ChatDates{}
		}
		log.Fatal(err)
	}
	defer file.Close()

	var chatDates []ChatDates
	err = json.NewDecoder(file).Decode(&chatDates)
	if err != nil {
		log.Fatal(err)
	}

	return chatDates
}

func handleDynamicCommand(ctx context.Context, b *bot.Bot, update *models.Update, name string) {
	// Загружаем даты из файла
	chatDates := loadDates()

	// Загружаем часовой пояс Europe/Moscow
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Ошибка загрузки часового пояса: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка загрузки часового пояса.",
		})
		return
	}

	// Ищем дату с указанным названием
	for _, chat := range chatDates {
		if chat.ChatID == update.Message.Chat.ID {
			for _, entry := range chat.Dates {
				if entry.Name == name {
					parsedDate, err := time.ParseInLocation("2006-01-02 15:04", entry.Date, location)
					if err != nil {
						b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID: update.Message.Chat.ID,
							Text:   "Ошибка: неверный формат даты.",
						})
						return
					}

					// Рассчитываем оставшееся время
					now := time.Now().In(location) // Преобразуем текущее время в Europe/Moscow
					duration := parsedDate.Sub(now)
					if duration < 0 {
						b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID: update.Message.Chat.ID,
							Text:   fmt.Sprintf("Дата %s (%s) уже прошла.", entry.Name, entry.Date),
						})
					} else {
						days := int(duration.Hours()) / 24
						hours := int(duration.Hours()) % 24
						minutes := int(duration.Minutes()) % 60
						b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID: update.Message.Chat.ID,
							Text:   fmt.Sprintf("До события %s (%s) осталось: %d дней, %d часов, %d минут.", entry.Name, entry.Date, days, hours, minutes),
						})
					}
					return
				}
			}
		}
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Событие с названием '%s' не найдено.", name),
	})
}

func saveDates(chatDates []ChatDates) {
	file, err := os.Create(datesFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(chatDates)
	if err != nil {
		log.Fatal(err)
	}
}

func updateBotCommands(b *bot.Bot, commands []models.BotCommand) error {
	_, err := b.SetMyCommands(context.Background(), &bot.SetMyCommandsParams{
		Commands: commands,
	})
	if err != nil {
		log.Printf("Ошибка при обновлении команд: %v", err)
		return err
	}
	return nil
}

func parseDateWithTimezone(dateTime string) (time.Time, error) {
	// Загружаем часовой пояс Europe/Moscow (UTC+3)
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.Time{}, fmt.Errorf("не удалось загрузить часовой пояс: %v", err)
	}

	// Парсим дату с учётом указанного часового пояса
	parsedDate, err := time.ParseInLocation("2006-01-02 15:04", dateTime, location)
	if err != nil {
		// Пробуем парсить только дату без времени
		parsedDate, err = time.ParseInLocation("2006-01-02", dateTime, location)
		if err != nil {
			return time.Time{}, err
		}
	}

	return parsedDate, nil
}

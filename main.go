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

// ChatDates хранит даты для конкретного чата
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
	b.RegisterHandler(bot.HandlerTypeMessageText, "*", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || !isCommand(update.Message.Text) {
			return // Игнорируем сообщения, которые не являются командами
		}

		// Нормализуем команду
		command := normalizeCommand(update.Message.Text)

		// Обрабатываем команды
		switch command {
		case "/setdate":
			setDateHandler(ctx, b, update)
		case "/dates":
			listDatesHandler(ctx, b, update)
		default:
			// Для динамических команд
			for _, cmd := range baseCommands {
				if command == "/"+cmd.Command {
					handleDynamicCommand(ctx, b, update, cmd.Command)
					return
				}
			}
			log.Printf("Неизвестная команда: %s", command)
			fmt.Printf("Неизвестная команда: %s\n", command)
		}
	})

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
	// Проверяем, является ли сообщение командой
	if update.Message == nil || !isCommand(update.Message.Text) {
		return
	}

	// Логируем и выводим команду
	log.Printf("Обработка команды: %s", update.Message.Text)
	fmt.Printf("Обработка команды: %s\n", update.Message.Text)

	// Нормализуем команду
	command := normalizeCommand(update.Message.Text)
	if !strings.HasPrefix(command, "/setdate") {
		return // Игнорируем, если это не команда /setdate
	}

	// Получаем дату и название из сообщения
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) < 2 {
		sendMessage(ctx, b, &bot.SendMessageParams{
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
		sendMessage(ctx, b, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Неверный формат даты. Используйте формат: YYYY-MM-DD [HH:MM]",
		})
		return
	}

	name := ""
	if len(parts) > 2 {
		name = parts[2]
	}

	// Сохраняем дату для текущего чата
	saveChatDate(update.Message.Chat.ID, DateEntry{Date: parsedDate.Format("2006-01-02 15:04"), Name: name})

	// Регистрируем новую команду, если указано название
	if name != "" {
		command := "/" + name
		log.Printf("Регистрируется команда: %s", command)
		fmt.Printf("Регистрируется команда: %s\n", command)

		b.RegisterHandler(bot.HandlerTypeMessageText, normalizeCommand(command), bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
			handleDynamicCommand(ctx, b, update, name)
		})

		// Добавляем описание для новой команды
		newCommand := models.BotCommand{
			Command:     name,
			Description: fmt.Sprintf("Показать время до события '%s'", name),
		}
		baseCommands = append(baseCommands, newCommand)

		// Обновляем команды в Telegram
		err := updateBotCommands(b, baseCommands)
		if err != nil {
			sendMessage(ctx, b, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Дата добавлена, но команда '%s' не была зарегистрирована: %v", command, err),
			})
			return
		}

		sendMessage(ctx, b, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Дата успешно добавлена! Используйте команду %s для просмотра.", command),
		})
		return
	}

	sendMessage(ctx, b, &bot.SendMessageParams{
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
	// Устанавливаем команды только для личных чатов
	_, err := b.SetMyCommands(context.Background(), &bot.SetMyCommandsParams{
		Commands: commands,
		Scope:    &models.BotCommandScopeDefault{}, // Применяем команды только для личных чатов
	})
	if err != nil {
		log.Printf("Ошибка при обновлении команд: %v", err)
		fmt.Printf("Ошибка при обновлении команд: %v\n", err) // Дублируем вывод
		return err
	}

	log.Printf("Команды успешно обновлены для личных чатов.")
	fmt.Printf("Команды успешно обновлены для личных чатов.\n") // Дублируем вывод
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

func addCommandIfNotExists(name string, description string) {
	for _, cmd := range baseCommands {
		if cmd.Command == name {
			return // Команда уже существует
		}
	}
	baseCommands = append(baseCommands, models.BotCommand{
		Command:     name,
		Description: description,
	})
}

func normalizeCommand(command string) string {
	fmt.Printf("Нормализация команды: %s\n", command) // Логируем входящую команду
	if strings.Contains(command, "@") {
		parts := strings.Split(command, "@")
		fmt.Printf("Команда после нормализации: %s\n", parts[0]) // Логируем результат нормализации
		return parts[0]                                          // Возвращаем только команду без имени бота
	}
	fmt.Printf("Команда не требует нормализации: %s\n", command) // Логируем, если нормализация не требуется
	return command
}

func saveChatDate(chatID int64, date DateEntry) {
	chatDates := loadDates()
	found := false
	for i, chat := range chatDates {
		if chat.ChatID == chatID {
			chatDates[i].Dates = append(chatDates[i].Dates, date)
			found = true
			break
		}
	}
	if !found {
		chatDates = append(chatDates, ChatDates{
			ChatID: chatID,
			Dates:  []DateEntry{date},
		})
	}
	saveDates(chatDates)
}

func isCommand(message string) bool {
	return strings.HasPrefix(message, "/")
}

func sendMessage(ctx context.Context, b *bot.Bot, params *bot.SendMessageParams) {
	_, err := b.SendMessage(ctx, params)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

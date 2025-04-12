package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
	b.RegisterHandler(bot.HandlerTypeMessageText, "/murmansk", bot.MatchTypeExact, murmanskHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/setdate", bot.MatchTypePrefix, setDateHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/dates", bot.MatchTypeExact, listDatesHandler)

	// Устанавливаем базовые команды
	baseCommands := []models.BotCommand{
		{Command: "murmansk", Description: "Показать время до следующей даты"},
		{Command: "setdate", Description: "Добавить новую дату (/setdate YYYY-MM-DD [название])"},
		{Command: "dates", Description: "Показать список всех запланированных дат"},
	}
	updateBotCommands(b, baseCommands)

	// Запускаем бота
	b.Start(context.Background())
}

func murmanskHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Загружаем даты из файла
	chatDates := loadDates()

	// Ищем даты для текущего чата
	var targetDate time.Time
	for _, chat := range chatDates {
		if chat.ChatID == update.Message.Chat.ID && len(chat.Dates) > 0 {
			// Берём первую дату из списка
			targetDate, _ = time.Parse("2006-01-02", chat.Dates[0].Date)
			break
		}
	}

	if targetDate.IsZero() {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Дата не установлена. Используйте команду /setdate YYYY-MM-DD для установки даты.",
		})
		return
	}

	// Текущее время
	now := time.Now()

	// Рассчитываем продолжительность до целевой даты
	duration := targetDate.Sub(now)

	// Получаем количество дней, часов и минут
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	// Случайные фразы
	funnyPhrases := []string{
		"Дамы и господа, на ваших экранах — космический рейс, и время до старта составляет... 🚀⏳",
		"Забудьте все, что вы знали о времени, вот оно — ваше будущее! 🔮✨",
		"Секунды тают, как снег на солнце, до события осталась совсем малость... ❄️☀️",
	}

	// Генерируем случайный индекс для фразы
	rand.Seed(time.Now().Unix())
	randomPhrase := funnyPhrases[rand.Intn(len(funnyPhrases))]

	// Формируем сообщение
	message := fmt.Sprintf("%s\n**%d дней, %d часов, %d минут.**", randomPhrase, days, hours, minutes)

	// Отправляем сообщение
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      message,
		ParseMode: "Markdown",
	})
}

func setDateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Получаем дату и название из сообщения
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) < 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Используйте формат: /setdate YYYY-MM-DD [название]",
		})
		return
	}

	date := parts[1]
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Неверный формат даты. Используйте формат: YYYY-MM-DD",
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
			chatDates[i].Dates = append(chatDates[i].Dates, DateEntry{Date: date, Name: name})
			found = true
			break
		}
	}

	if !found {
		chatDates = append(chatDates, ChatDates{
			ChatID: update.Message.Chat.ID,
			Dates:  []DateEntry{{Date: date, Name: name}},
		})
	}

	// Сохраняем даты обратно в файл
	saveDates(chatDates)

	// Регистрируем новую команду, если указано название
	if name != "" {
		b.RegisterHandler(bot.HandlerTypeMessageText, "/"+name, bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
			handleDynamicCommand(ctx, b, update, name)
		})

		// Добавляем описание для новой команды
		newCommand := models.BotCommand{
			Command:     name,
			Description: fmt.Sprintf("Показать время до события '%s'", name),
		}
		baseCommands = append(baseCommands, newCommand)
		updateBotCommands(b, baseCommands)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Дата успешно добавлена!",
	})
}

func listDatesHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Загружаем даты из файла
	chatDates := loadDates()

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
			now := time.Now()

			for i, entry := range chat.Dates {
				parsedDate, err := time.Parse("2006-01-02", entry.Date)
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

	// Ищем дату с указанным названием
	for _, chat := range chatDates {
		if chat.ChatID == update.Message.Chat.ID {
			for _, entry := range chat.Dates {
				if entry.Name == name {
					parsedDate, err := time.Parse("2006-01-02", entry.Date)
					if err != nil {
						b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID: update.Message.Chat.ID,
							Text:   "Ошибка: неверный формат даты.",
						})
						return
					}

					// Рассчитываем оставшееся время
					now := time.Now()
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

func updateBotCommands(b *bot.Bot, commands []models.BotCommand) {
	_, err := b.SetMyCommands(context.Background(), &bot.SetMyCommandsParams{
		Commands: commands,
	})
	if err != nil {
		log.Printf("Ошибка при обновлении команд: %v", err)
	}
}

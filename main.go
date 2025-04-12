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

type ChatDates struct {
	ChatID int64    `json:"chat_id"`
	Dates  []string `json:"dates"`
}

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
			targetDate, _ = time.Parse("2006-01-02", chat.Dates[0])
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
	// Получаем дату из сообщения
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) < 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Используйте формат: /setdate YYYY-MM-DD",
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

	// Загружаем существующие даты
	chatDates := loadDates()

	// Добавляем дату для текущего чата
	found := false
	for i, chat := range chatDates {
		if chat.ChatID == update.Message.Chat.ID {
			chatDates[i].Dates = append(chatDates[i].Dates, date)
			found = true
			break
		}
	}

	if !found {
		chatDates = append(chatDates, ChatDates{
			ChatID: update.Message.Chat.ID,
			Dates:  []string{date},
		})
	}

	// Сохраняем даты обратно в файл
	saveDates(chatDates)

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
			for _, date := range chat.Dates {
				message += fmt.Sprintf("- %s\n", date)
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

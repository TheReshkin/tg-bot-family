package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

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

	// Устанавливаем подсказки команд
	commands := []models.BotCommand{
		{Command: "murmansk", Description: "Узнать, сколько осталось до поездки в Мурманск"},
		// Добавьте другие команды, если нужно
	}
	_, err = b.SetMyCommands(context.Background(), &bot.SetMyCommandsParams{
		Commands: commands,
	})
	if err != nil {
		log.Fatalf("Ошибка при установке команд: %v", err)
	}

	// Регистрируем обработчик команды /murманск
	b.RegisterHandler(bot.HandlerTypeMessageText, "/murmansk", bot.MatchTypeExact, murmanskHandler)

	// Запускаем бота
	b.Start(context.Background())
}

func murmanskHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Задаём часовой пояс для Москвы (UTC+3)
	moscowLocation := time.FixedZone("MSK", 3*60*60)

	funnyPhrases := []string{
		"Дамы и господа, на ваших экранах — космический рейс, и время до старта составляет... 🚀⏳",
		"Забудьте все, что вы знали о времени, вот оно — ваше будущее! 🔮✨",
		"А тем временем на секундомере… до точки старта осталось всего... ⏱️🔥",
		"Дамы и господа, если вы хотели узнать, сколько осталось до этого момента — держитесь крепче! Осталось ⏳💥",
		"Присаживайтесь поудобнее, время до поездки... 🛋️🕒",
		"Пока мы тут болтаем, до важной даты осталось всего... 🗓️⏳",
		"Секунды тают, как снег на солнце, до события осталась совсем малость... ❄️☀️",
		"Представьте, что вы в гонке, и до старта остаётся всего... 🏁⏰",
	}

	// Инициализируем генератор случайных чисел один раз
	rand.Seed(time.Now().Unix())

	// Целевая дата с учётом московского времени
	targetDate := time.Date(2025, time.April, 26, 0, 0, 0, 0, moscowLocation)

	// Текущее время
	now := time.Now().In(moscowLocation)

	// Рассчитываем продолжительность до целевой даты
	duration := targetDate.Sub(now)

	// Получаем количество дней, часов и минут
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	// Генерируем случайный индекс для фразы
	randomPhrase := funnyPhrases[rand.Intn(len(funnyPhrases))]

	// Формируем сообщение с жирным текстом
	message := fmt.Sprintf("%s\n**%d дней, %d часов, %d минут.**", randomPhrase, days, hours, minutes)

	// Отправляем сообщение в чат
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      message,
		ParseMode: "Markdown",
	})
}

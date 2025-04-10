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

	// Регистрируем обработчик команды /murmansk
	b.RegisterHandler(bot.HandlerTypeMessageText, "/murmansk", bot.MatchTypeExact, murmanskHandler)

	// Запускаем бота
	b.Start(context.Background())
}

func murmanskHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

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

	// Целевая дата
	targetDate := time.Date(2025, time.April, 26, 4, 0, 0, 0, time.UTC)
	// Текущее время
	now := time.Now()

	// Рассчитываем время до целевой даты
	duration := targetDate.Sub(now)

	// Получаем количество дней, часов и минут
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	// Генерируем случайный индекс для фразы
	rand.Seed(time.Now().Unix())
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

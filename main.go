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

const datesFile = "./data/dates.json"

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

	botName := ""
	me, err := b.GetMe(context.Background())
	if err != nil {
		log.Fatalf("Не удалось получить информацию о боте: %v", err)
	} else {
		botName = me.Username
		log.Printf("Имя бота: %s", botName)
	}

	// Регистрируем обработчики команд
	b.RegisterHandler(bot.HandlerTypeMessageText, "/setdate", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		setDateHandler(ctx, b, update, botName)
	})
	b.RegisterHandler(bot.HandlerTypeMessageText, "/dates", bot.MatchTypeExact, listDatesHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/dates", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		listDatesHandler(ctx, b, update)
	})
	b.RegisterHandler(bot.HandlerTypeMessageText, "/dates@"+botName, bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		listDatesHandler(ctx, b, update)
	})
	b.RegisterHandler(bot.HandlerTypeMessageText, "*", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || !isCommand(update.Message.Text) {
			return // Игнорируем сообщения, которые не являются командами
		}

		// Нормализуем команду
		command := normalizeCommand(update.Message.Text)

		// Логируем нормализованную команду
		log.Printf("Получена команда: %s", update.Message.Text)
		log.Printf("Нормализованная команда: %s", command)

		// Обрабатываем стандартные команды
		switch command {
		case "/setdate":
			setDateHandler(ctx, b, update, botName)
			return
		case "/dates", "/dates@" + botName: // Учитываем оба варианта команды
			listDatesHandler(ctx, b, update)
			return
		}

		// Обрабатываем кастомные команды с учётом имени бота
		for _, cmd := range baseCommands {
			fullCommand := "/" + cmd.Command + "@" + botName
			if command == "/"+cmd.Command || command == fullCommand {
				handleDynamicCommand(ctx, b, update, cmd.Command)
				return
			}
		}

		log.Printf("Неизвестная команда: %s", command)
		fmt.Printf("Неизвестная команда: %s\n", command)
	})

	// Устанавливаем базовые команды
	baseCommands := []models.BotCommand{
		{Command: "setdate", Description: "Добавить новую дату (/setdate YYYY-MM-DD [название])"},
		{Command: "dates", Description: "Показать список всех запланированных дат"},
	}
	updateBotCommands(b, baseCommands)

	// Регистрируем динамические команды из файла dates.json
	registerDynamicCommandsFromFile(b, botName)

	// Запускаем бота
	b.Start(context.Background())
	updateBotCommands(b, baseCommands)

	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())
	updateBotCommands(b, baseCommands)
	updateBotCommands(b, baseCommands)

	// Регистрируем динамические команды
	registerDynamicCommands(b, botName)
}
func setDateHandler(ctx context.Context, b *bot.Bot, update *models.Update, botName string) {
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
		fullCommand := command + "@" + botName // Команда с именем бота
		log.Printf("Регистрируется команда: %s", fullCommand)
		fmt.Printf("Регистрируется команда: %s\n", fullCommand)

		b.RegisterHandler(bot.HandlerTypeMessageText, fullCommand, bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
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
				Text:   fmt.Sprintf("Дата добавлена, но команда '%s' не была зарегистрирована: %v", fullCommand, err),
			})
			return
		}

		sendMessage(ctx, b, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Дата успешно добавлена! Используйте команду %s для просмотра.", fullCommand),
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
	// Забавные фразочки
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
					now := time.Now().In(location)
					duration := parsedDate.Sub(now)
					if duration < 0 {
						b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID: update.Message.Chat.ID,
							Text:   fmt.Sprintf("Дата %s (%s) уже прошла.", entry.Name, entry.Date),
						})
					} else {
						totalHours := int(duration.Hours())
						days := totalHours / 24
						hours := totalHours % 24
						minutes := int(duration.Minutes()) % 60

						// Форматируем вывод
						var timeParts []string
						if days >= 4 {
							if days > 0 {
								timeParts = append(timeParts, fmt.Sprintf("%d дней", days))
							}
							if hours > 0 {
								timeParts = append(timeParts, fmt.Sprintf("%d часов", hours))
							}
							if minutes > 0 {
								timeParts = append(timeParts, fmt.Sprintf("%d минут", minutes))
							}
						} else {
							// Если осталось меньше 3 дней, преобразуем дни в часы
							remainingHours := totalHours
							if remainingHours > 0 {
								timeParts = append(timeParts, fmt.Sprintf("%d часов", remainingHours))
							}
							if minutes > 0 {
								timeParts = append(timeParts, fmt.Sprintf("%d минут", minutes))
							}
						}

						timeLeft := strings.Join(timeParts, ", ")

						// Выбираем случайную фразу
						randomPhrase := funnyPhrases[rand.Intn(len(funnyPhrases))]

						b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID: update.Message.Chat.ID,
							Text:   fmt.Sprintf("%s\nДо события %s (%s)\nосталось: %s.", randomPhrase, entry.Name, entry.Date, timeLeft),
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
	// Создаём папку ./data/, если её нет
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		err := os.Mkdir("./data", os.ModePerm)
		if err != nil {
			log.Fatalf("Не удалось создать папку ./data/: %v", err)
		}
	}

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
		Scope:    &models.BotCommandScopeAllGroupChats{}, // Применяем команды для всех групп
	})
	if err != nil {
		log.Printf("Ошибка при обновлении команд: %v", err)
		return err
	}

	log.Printf("Команды успешно обновлены для всех чатов.")
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

func registerDynamicCommands(b *bot.Bot, botName string) {
	chatDates := loadDates()

	for _, chat := range chatDates {
		for _, entry := range chat.Dates {
			if entry.Name != "" {
				command := "/" + entry.Name
				fullCommand := command + "@" + botName // Команда с именем бота

				// Регистрируем обработчик для команды
				b.RegisterHandler(bot.HandlerTypeMessageText, fullCommand, bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
					handleDynamicCommand(ctx, b, update, entry.Name)
				})

				// Добавляем описание для команды
				newCommand := models.BotCommand{
					Command:     entry.Name,
					Description: fmt.Sprintf("Показать время до события '%s'", entry.Name),
				}
				baseCommands = append(baseCommands, newCommand)
			}
		}
	}

	// Обновляем команды в Telegram
	err := updateBotCommands(b, baseCommands)
	if err != nil {
		log.Printf("Ошибка при обновлении динамических команд: %v", err)
	}
}

func registerDynamicCommandsFromFile(b *bot.Bot, botName string) {
	chatDates := loadDates()

	for _, chat := range chatDates {
		for _, entry := range chat.Dates {
			// Проверяем, что имя команды не пустое
			if entry.Name != "" {
				command := "/" + entry.Name
				fullCommand := command + "@" + botName // Команда с именем бота

				// Регистрируем обработчик для команды
				b.RegisterHandler(bot.HandlerTypeMessageText, fullCommand, bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
					handleDynamicCommand(ctx, b, update, entry.Name)
				})

				// Добавляем описание для команды
				newCommand := models.BotCommand{
					Command:     entry.Name,
					Description: fmt.Sprintf("Показать время до события '%s'", entry.Name),
				}
				baseCommands = append(baseCommands, newCommand)
			}
		}
	}

	// Обновляем команды в Telegram
	err := updateBotCommands(b, baseCommands)
	if err != nil {
		log.Printf("Ошибка при обновлении динамических команд: %v", err)
	}
}

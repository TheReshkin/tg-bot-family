# Tasks: murmansk-bot MVP

**Input**: Design documents from `/specs/001-project-murmansk-bot/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/

## Phase 3.1: Setup
- [x] T001 Создать структуру проекта: cmd/, internal/, pkg/, tests/
- [x] T002 Инициализировать Go-модуль и зависимости (go mod tidy)
- [ ] T003 [P] Настроить линтер и форматирование (golangci-lint)

## Phase 3.2: Tests First (TDD)
- [x] T004 [P] Реализовать контрактные тесты для /set_date и /list в specs/001-project-murmansk-bot/contracts/events_contract_test.go (должны падать)
- [x] T005 [P] Написать юнит-тесты для парсинга дат и расчёта времени до события в tests/unit/date_test.go
- [x] T006 [P] Написать интеграционный тест для команды /set_date в tests/integration/set_date_test.go
- [x] T007 [P] Написать интеграционный тест для команды /list в tests/integration/list_test.go

## Phase 3.3: Core Implementation
- [x] T008 [P] Реализовать модель Event в internal/models/event.go
- [x] T009 [P] Реализовать модель User в internal/models/user.go
- [x] T010 Реализовать интерфейс Storage (json, расширяемо до PostgreSQL) в internal/storage/storage.go
- [x] T011 Реализовать сервис управления событиями (EventService) в internal/services/event_service.go
- [x] T012 Реализовать сервис управления пользователями (UserService) в internal/services/user_service.go
- [x] T013 Реализовать обработчики команд /set_date, /list, /all, /active, /outdated, /help в cmd/main.go
- [x] T014 Реализовать регистрацию динамических команд в cmd/main.go
- [x] T015 Реализовать логирование событий и ошибок (log/logrus/zap) в cmd/main.go и сервисах

## Phase 3.4: Integration & Polish
- [x] T016 [P] Реализовать интеграцию с Docker (Dockerfile, docker-compose)
- [x] T017 [P] Добавить README и quickstart.md с инструкциями и списком команд
- [ ] T018 [P] Добавить поддержку нескольких чатов (разделение событий по chat_id)
- [ ] T019 [P] Добавить локализацию (структура для будущих языков)
- [ ] T020 [P] Провести рефакторинг и оптимизацию кода
- [ ] T021 [P] Провести тестирование производительности и устойчивости

## Phase 3.5: Refactoring & Enhancements
- [x] T022 Обновить парсинг даты и времени: добавить поддержку часов и минут (формат YYYY-MM-DD HH:MM), по умолчанию 00:00, часовой пояс Europe/Moscow
- [x] T023 Изменить формат даты в моделях и хранилище на YYYY-MM-DD HH:MM
- [x] T024 Обновить команду /list для отображения динамических команд событий
- [x] T025 Обновить все тесты для нового формата даты и времени
- [ ] T026 Обновить обработчики команд для работы с новым форматом

## Phase 3.6: Countdown Functionality (Live Message Updates)
**Requirement**: Добавить функционал "живого" обратного отсчёта через editMessageText
- [x] T027 [P] Создать модель CountdownMessage в `internal/models/countdown.go`
- [x] T028 [P] Создать интерфейс MessageTracker в `internal/services/message_tracker.go`
- [x] T029 [P] Добавить настройки countdown в `internal/config/countdown.go`

### Phase 3.6.1: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE IMPLEMENTATION
- [x] T030 [P] Контрактный тест создания countdown сообщения в `tests/contracts/countdown_message_test.go`
- [x] T031 [P] Контрактный тест editMessageText API в `tests/contracts/edit_message_test.go`
- [x] T032 [P] Интеграционный тест живого countdown в `tests/integration/countdown_integration_test.go`
- [x] T033 [P] Интеграционный тест отслеживания сообщений в `tests/integration/message_tracking_test.go`
- [x] T034 [P] Интеграционный тест countdown ticker поведения в `tests/integration/countdown_ticker_test.go`

### Phase 3.6.2: Core Implementation (ТОЛЬКО после падающих тестов)
- [x] T035 [P] Реализовать CountdownMessage struct и валидацию в `internal/models/countdown.go`
- [ ] T036 [P] Реализовать MessageTracker сервис в `internal/services/message_tracker.go`
- [ ] T037 [P] Реализовать CountdownService с editMessageText в `internal/services/countdown_service.go`
- [ ] T038 Управление countdown ticker (goroutine) в `internal/services/countdown_service.go`
- [ ] T039 Интегрировать countdown в handleDynamicCommand в `cmd/main.go`
- [ ] T040 Добавить команды /countdown и /stop_countdown в `cmd/main.go`
- [ ] T041 Хранение message_id и retrieval в `internal/storage/storage.go`
- [ ] T042 Обработка ошибок editMessageText (удалённые сообщения и т.д.)

### Phase 3.6.3: Integration & Error Handling
- [ ] T043 Подключить CountdownService к существующему EventService
- [ ] T044 Реализовать cleanup countdown при перезапуске бота
- [ ] T045 Добавить countdown статус в event storage
- [ ] T046 Graceful shutdown для активных countdown
- [ ] T047 Rate limiting для message edits (ограничения Telegram API)

### Phase 3.6.4: Polish & Testing
- [ ] T048 [P] Юнит-тесты для countdown расчётов в `tests/unit/countdown_calc_test.go`
- [ ] T049 [P] Юнит-тесты для форматирования сообщений в `tests/unit/countdown_format_test.go`
- [ ] T050 [P] Performance тесты (множественные simultaneous countdown)
- [ ] T051 [P] Обновить README.md с countdown командами
- [ ] T052 [P] Добавить countdown примеры в quickstart.md
- [ ] T053 Рефакторинг и удаление дублирования кода
- [ ] T054 Проверить memory usage и cleanup countdown

## Parallel Execution Guidance
- Все задачи, отмеченные [P], могут выполняться параллельно, если не зависят от одних и тех же файлов.
- Пример: тесты, модели, интеграция с Docker, документация, локализация, оптимизация.
- **Countdown tests (T030-T034)** могут выполняться параллельно - разные файлы
- **Countdown models (T027-T029, T035-T036)** могут выполняться параллельно - разные файлы
- **Polish tasks (T048-T052)** могут выполняться параллельно - разные файлы

## Dependency Notes
- T001, T002 — всегда первыми
- Тесты (T004-T007) — до реализации
- Модели (T008-T009) — до сервисов
- Сервисы (T011-T012) — до обработчиков команд
- Логирование, интеграция, документация — после основных функций
- Рефакторинг (T022-T026) — после завершения основных функций

### Countdown Dependencies
- **Setup (T027-T029)** — могут выполняться параллельно после завершения основного MVP
- **Countdown tests (T030-T034)** — должны быть написаны и падать перед реализацией
- **T035 блокирует T036, T037** — CountdownMessage до сервисов
- **T037 блокирует T038, T039** — CountdownService до ticker и integration
- **T041 блокирует T043, T044** — Storage до EventService integration
- **Implementation (T035-T047) до Polish (T048-T054)**

## Task Agent Commands
- Для параллельных задач используйте: `/run-tasks T004 T005 T006 T007`
- Для последовательных: `/run-task T001`, затем `/run-task T002`, затем `/run-task T008`
- Для рефакторинга: `/run-task T022`, затем `/run-task T023`, затем `/run-task T024`

### Countdown Implementation Commands
```bash
# Phase 3.6 Setup (parallel):
/run-tasks T027 T028 T029

# Phase 3.6.1 Tests First (parallel - MUST FAIL before implementation):
/run-tasks T030 T031 T032 T033 T034

# Phase 3.6.2 Core Implementation (sequential due to dependencies):
/run-tasks T035 T036  # Models (parallel)
/run-task T037        # CountdownService
/run-task T038        # Ticker management  
/run-task T039        # Integration
/run-task T040        # Commands
/run-task T041        # Storage
/run-task T042        # Error handling

# Phase 3.6.3 Integration (sequential):
/run-task T043        # EventService integration
/run-task T044        # Cleanup
/run-task T045        # Status storage
/run-task T046        # Shutdown
/run-task T047        # Rate limiting

# Phase 3.6.4 Polish (parallel):
/run-tasks T048 T049 T050 T051 T052
/run-task T053        # Refactor
/run-task T054        # Memory check
```

## Countdown Feature Specification

### Core Requirements
1. **Live Countdown**: Бот отправляет сообщение с countdown, затем периодически обновляет то же сообщение через Telegram `editMessageText` API
2. **Message Tracking**: Сохранять `chat_id` и `message_id` отправленных countdown сообщений 
3. **Ticker Management**: Использовать Go `time.Ticker` или goroutines для обновления countdown каждую минуту
4. **Multiple Countdowns**: Поддержка множественных одновременных countdown в разных чатах
5. **Persistence**: Состояние countdown переживает перезапуски бота

### Technical Implementation
- **editMessageText**: Использовать `bot.EditMessageText()` из go-telegram/bot library
- **Storage**: Расширить существующее JSON хранилище для включения message IDs и countdown состояний
- **Concurrency**: Одна goroutine на активный countdown с proper cleanup
- **Error Handling**: Gracefully обрабатывать ошибки message edit (удалённые сообщения, и т.д.)

### New Commands
- `/countdown <event_name>` - Запустить live countdown для события
- `/stop_countdown <event_name>` - Остановить live countdown для события

### Message Format
```
🕒 Событие: new_year
📅 Дата: 31.12.2025 00:00
⏰ Осталось: 45 дней, 12 часов, 30 минут

Последнее обновление: 15:45
```


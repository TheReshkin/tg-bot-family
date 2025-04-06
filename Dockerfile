# Этап сборки
FROM golang:1.22-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git

# Создание рабочей директории
WORKDIR /app

# Копируем исходный файл бота
COPY main.go .

# Инициализируем модуль, если его нет (опционально)
RUN go mod init murmansk-bot || true
# Подключаем нужные зависимости
RUN go get github.com/go-telegram/bot \
           github.com/go-telegram/bot/models \
           github.com/joho/godotenv
# Компиляция
RUN go build -o murmansk-bot main.go

# Финальный образ (без компилятора)
FROM alpine:latest

# Устанавливаем необходимые библиотеки
RUN apk --no-cache add ca-certificates

# Рабочая директория
WORKDIR /app

# Копируем бинарник из стадии сборки
COPY --from=builder /app/murmansk-bot .

# Точка входа
ENTRYPOINT ["./murmansk-bot"]

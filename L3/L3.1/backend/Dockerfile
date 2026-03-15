# Этап сборки
FROM golang:1.25-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git

# Рабочая директория - устанавливаем на уровень корня проекта
WORKDIR /app

# Копируем корневые go.mod и go.sum (они находятся уровнем выше)
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download && go mod verify

# Копируем весь проект (сохраняя структуру)
COPY . .

# Переходим в директорию backend для сборки
WORKDIR /app/L3/L3.1/backend

# Сборка приложения с правильным путем к модулю
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/delayed-notifier ./cmd/main.go

# Финальный образ
FROM alpine:latest

# Установка CA сертификатов
RUN apk --no-cache add ca-certificates tzdata

# Копируем бинарный файл
COPY --from=builder /app/bin/delayed-notifier /usr/local/bin/delayed-notifier

# Копируем веб-файлы (фронтенд)
COPY --from=builder /app/L3/L3.1/frontend /web

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["delayed-notifier"]
# Calendar Service

HTTP-сервис для управления событиями с in-memory хранилищем, асинхронным логгированием и фоновыми воркерами (напоминания, архивация).

## Структура

- `cmd/server/main.go` — точка входа
- `config/config.go` — конфигурация из environment-переменных
- `internal/domain/` — модели и ошибки
- `internal/delivery/http/` — HTTP-обработчики
- `internal/usecase/` — бизнес-логика
- `internal/repository/memory/` — in-memory хранилище
- `internal/middleware/` — middleware логирования
- `internal/workers/` — фоновые воркеры (reminder, cleaner)
- `pkg/logger/` — асинхронный логгер

## Запуск локально

```bash
go mod tidy
go run ./cmd/server/
```

Сервис запустится на `localhost:8080`.

## Запуск в Docker

```bash
docker build -t calendar-service .
docker run -p 8080:8080 calendar-service

# или через docker-compose
docker compose up --build
```

Сбросить сервис:
```bash
docker compose down
```

## Конфигурация

Переменные среды (по умолчанию в скобках):

| Переменная | По умолчанию | Описание |
|---|---|---|
| `PORT` | `:8080` | Порт сервера |
| `READ_TIMEOUT` | `15s` | Таймаут чтения |
| `WRITE_TIMEOUT` | `15s` | Таймаут записи |
| `IDLE_TIMEOUT` | `60s` | Idle-таймаут |
| `LOGGER_BUFFER_SIZE` | `1000` | Размер буфера лога |
| `REMINDER_CHECK_INTERVAL` | `1m` | Интервал проверки напоминаний |
| `CLEANER_INTERVAL` | `1h` | Интервал архивации |
| `ARCHIVED_DAYS_THRESHOLD` | `30` | Дней до архивации |

## API

> Базовый URL: `http://localhost:8080`

### Создание события

```bash
curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "date": "2025-06-15",
    "event": "Встреча с командой",
    "has_reminder": true,
    "reminder_at": "2025-06-15 10:50:00"
  }'
```

**Ответ:**
```json
{
  "result": {
    "message": "Event created successfully",
    "event": {
      "id": "20250624120000.000000000",
      "user_id": "user1",
      "date": "2025-06-15",
      "event": "Встреча с командой",
      "has_reminder": true,
      "reminder_at": "2025-06-15 10:50:00",
      "created_at": "2025-06-24T12:00:00Z",
      "updatedAt": "2025-06-24T12:00:00Z"
    }
  }
}
```

### Обновление события

```bash
curl -s -X PUT http://localhost:8080/update_event \
  -H "Content-Type: application/json" \
  -d '{
    "id": "20250624120000.000000000",
    "user_id": "user1",
    "event": "Встреча перенесена на 15:00"
  }'
```

### Удаление события

```bash
curl -s -X DELETE http://localhost:8080/delete_event \
  -H "Content-Type: application/json" \
  -d '{
    "id": "20250624120000.000000000",
    "user_id": "user1"
  }'
```

### События на день

```bash
curl -s 'http://localhost:8080/events_for_day?user_id=user1&date=2025-06-15'
```

### События на неделю

```bash
curl -s 'http://localhost:8080/events_for_week?user_id=user1&date=2025-06-15'
```

### События на месяц

```bash
curl -s 'http://localhost:8080/events_for_month?user_id=user1&date=2025-06-15'
```

## Ошибки

| HTTP Status | Описание |
|---|---|
| `400` | Неверный формат данных (data, user_id, event) |
| `400` | Invalid request body |
| `500` | Внутренняя ошибка сервера |
| `503` | Event not found |

Сценарии ошибок:

```bash
# Missing fields
curl -s -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1"}'

# No such event
curl -s -X DELETE http://localhost:8080/delete_event \
  -H "Content-Type: application/json" \
  -d '{"id": "nonexistent", "user_id": "user1"}'

# Wrong date format
curl -s 'http://localhost:8080/events_for_day?user_id=user1&date=15-06-2025'
```

## Воркеры

В сервисе запускаются два фоновых модуля:

- **ReminderWorker** — каждую минуту проверяет события с напоминаниями и логирует при приближении ReminderAt
- **CleanerWorker** — каждые 1 час(настраивается) архивирует события старше 30 дней

## Тесты

```bash
# Запуск всех тестов
go test ./... -v

# Только usecase
go test ./internal/usecase/ -v

# Только logger
go test ./pkg/logger/ -v

# С покрытием
go test ./... -cover
```

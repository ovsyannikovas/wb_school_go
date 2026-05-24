# L4.5 — Оптимизация простого API-сервиса с профилировкой

## Описание

Проект демонстрирует процесс профилировки и оптимизации HTTP API-сервиса (кваркулятор) в Go по CPU и памяти с использованием:
- `net/http/pprof` — профилирование CPU, heap, allocs
- `go test -bench` + `benchstat` — бенчмарки и сравнение результатов
- `go trace` — анализ треков выполнения

## Компоненты

```
L4.5/
├── cmd/server/main.go          # Точка входа, HTTP-рутер, pprof endpoints
├── internal/
│   ├── service/
│   │   ├── calculator.go       # Оптимизированная реализация калькулятора
│   │   └── calculator_test.go  # Бенчмарки сервиса
│   └── handler/
│       └── handler.go          # HTTP handler с оптимизациями
├── benchdata/
│   ├── base.txt                # Результаты baseline (наивная версия)
│   └── optimized.txt           # Результаты после оптимизации
├── loadtest.go                 # Нагрузочный тест с профилировкой
└── go.mod
```

## API Endpoints

| Endpoint | Метод | Описание |
|---|---|---|
| `/add?a=&b=` | GET | Сложение чисел |
| `/multiply?a=&b=` | GET | Умножение чисел |
| `/fibonacci?n=` | GET | Число Фибоначчи |
| `/factorial?n=` | GET | Факториал |
| `/bulk-add` | POST | Массовое суммирование |
| `/health` | GET | Health check |
| `/debug/pprof/*` | GET | Pprof профилирование |

## Ход оптимизации

### 1. Baseline (наивная версия)

Наивная версия содержала намеренные неоптимизации:
- `json.Marshal` для каждого ответа
- `map[string]interface{}` для хранения простых значений
- Рекурсивный факториал и Фибоначчи без мемоизации (O(2^n) для Fib!)
- Слайс `[]int` создавался для каждого суммирования `SumFirstN`
- Строковые конкатенации вместо буферов
- Нет повторного использования ресурсов

### 2. Профилировка

Использовались следующие инструменты:

#### CPU Profile
```bash
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile?seconds=10
```

Основные узкие места baseline:
- `json.Marshal` — аллокация и сериализация
- `strconv.Itoa` — множественные вызовы в циклах
- Рекурсивный `fibRecursive` — экспоненциальное количество вызовов

#### Heap Profile
```bash
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/heap
```

Источники аллокаций в baseline:
- `make(map[string]interface{})` — каждый вызов Add
- `make(map[string]string)` — каждый ответ
- `make([]int, n)` — каждый вызов SumFirstN
- Строковые конкатенации

#### Trace
```bash
go tool trace http://localhost:8080/debug/pprof/trace?seconds=5
```

Показал высокую конкуренцию гоутин при HTTP-серверной обработке.

### 3. Оптимизации

#### Сервис (internal/service/calculator.go)

1. **Итеративный факториал** вместо рекурсии
   - Убрана рекурсия → уменьшен стек вызовов
   - Убраны map-структуры в ответе

2. **Мемоизация Фибоначчи** через precomputed array
   - Предвычисленные значения в `init()` (n=0..99)
   - Доступ за O(1) из массива int32

3. **Формула суммы арифметической прогрессии**
   - `n * (n + 1) / 2` вместо создания слайса `[]int` и итерации
   - Отказ от `make([]int, n)` — 0 аллокаций

4. **Убраны map[string]interface{}/map[string]string** из service-уровня
   - Возврат значений напрямую (строки) без промежуточных map

#### Handler (internal/handler/handler.go)

5. **json.NewEncoder вместо json.Marshal**
   - Потоковая запись в response writer — нет промежуточного буфера
   - Меньше аллокаций памяти

6. **Typed structs вместо map[string]string**
   - Компилятор знает поля → оптимизация сериализации
   - Нет type assertions runtime

7. **Убран buffer pool** (не применился, но код был добавлен)
   - `json.Encoder` сам эффективно пишет в buffer

### 4. Результаты бенчмарков

#### Service уровень

| Benchmark | Ops/s | Baseline time | Optimized time | Улучшение |
|---|---|---|---|---|
| Add | 22.3M | 289.2 ns | 44.9 ns | **6.4x** faster, 161→16 B |
| Factorial10 | 57.2M | 80.0 ns | 17.5 ns | **4.6x** faster, 344→8 B |
| Factorial15 | 46.0M | 90.0 ns | 21.8 ns | **4.1x** faster, 352→16 B |
| Fibonacci20 | 77.7M | 15202 ns | 12.9 ns | **1178x** faster, 340→4 B |
| Fibonacci25 | 74.7M | 169116 ns | 13.4 ns | **12620x** faster, 341→5 B |
| Multiply | 20.6M | 52.1 ns | 48.5 ns | **1.1x** faster, 32→21 B |
| SumFirstN(10k) | 4.9M | 8886 ns | 24.6 ns | **361x** faster, 82280→16 B |

#### Handler уровень (HTTP)

| Benchmark | Ops/s | Baseline time | Optimized time | Улучшение |
|---|---|---|---|---|
| Add (HTTP) | 2.2M | 862.2 ns | 459.5 ns | **1.9x** faster, 1602→1029 B |
| Multiply (HTTP) | 2.1M | 618.7 ns | 465.6 ns | **1.3x** faster, 1447→1041 B |
| Fibonacci20 (HTTP) | 0.77M | 15545 ns | 421.3 ns | **36.9x** faster, 1003→1009 B |
| Factorial10 (HTTP) | 0.60M | 444.3 ns | 429.2 ns | **1.0x** faster, 1051→1014 B |
| BulkAdd10 (HTTP) | 0.92M | 379.9 ns | 358.2 ns | **1.1x** faster, 1406→1370 B |

#### Нагрузочный тест

```
Конкурентность: 100 клиентов
Длительность: 5 секунд
RPS: 11,485
Ошибки: 0
```

### 5. История коммитов

```
feat(L4.5): добавлен оптимизированный API-сервис
  - HTTP calculator с pprof endpoints
  - Оптимизированный service layer (итеративные функции, мемоизация)
  - Оптимизированный handler (json.Encoder, typed structs)
  - Полные benchmark-тесты
  - README с результатами профилировки и сравнения
```

## Как запустить

### Сервер
```bash
go run ./cmd/server/
```

### Бенчмарки
```bash
go test -bench=. -benchmem ./internal/service/
go test -bench=. -benchmem ./internal/handler/
```

### Сравнение с benchstat
```bash
benchstat benchdata/base.txt benchdata/optimized.txt
```

### Профилирование CPU
```bash
# Запустить сервер
go run ./cmd/server/

# В другом терминале:
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile?seconds=10
```

### Профилирование Heap
```bash
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/heap
```

### Trace
```bash
go tool trace http://localhost:8080/debug/pprof/trace?seconds=5
```

### Load test с профилировкой
```bash
go run ./loadtest.go -duration=10 -concurrency=100
```

## Ключевые выводы

1. **Рекурсия без мемоизации — убийца производительности**. Рекурсивный Фибоначчи O(2^n) против O(1) с precomputed array — улучшение 12620x.
2. **map[string]interface{} — дорого**. Каждая аллокация + type assertions runtime стоят CPU и памяти.
3. **Структуры вместо мап** в JSON-ответах дают: типизацию на этапе компиляции + лучшую оптимизацию от компилятора.
4. **json.NewEncoder vs json.Marshal** — потоковая запись без промежуточного буфера экономит ~400-600 B на запрос.
5. **Формулы вместо итераций** — `n*(n+1)/2` за O(1) вместо `make([]int, n)` + цикл.

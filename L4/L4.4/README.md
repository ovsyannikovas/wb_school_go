# GC Metrics Server

Сервер на Go, предоставляющий метрики памяти и сборщика мусора в формате Prometheus через HTTP-эндпоинт `/metrics`.

## Возможности

- 📊 **Метрики Prometheus**:
    - Количество аллокаций и объектов
    - Количество сборок мусора (GC)
    - Используемая и системная память (heap, stack)
    - Время и длительность последнего GC
    - Целевой размер heap для следующего GC (`NextGC`)
    - Текущий `GCPercent`
    - Количество горутин и CGO вызовов

- 🛠 **Управление GC**:
    - Принудительный запуск GC: `GET /gc`
    - Изменение `GCPercent`: `GET /debug/set_gc_percent?percent=<value>`

- 🔍 **Профилирование (pprof)**:
    - CPU профиль: `/debug/pprof/profile`
    - Heap профиль: `/debug/pprof/heap`
    - Goroutine профиль: `/debug/pprof/goroutine`
    - Трассировка: `/debug/pprof/trace`

## Запуск

```bash
go run main.go
```

## Пример запуска
```bash
➜  L4.4 git:(main) ✗ go run main.go
2026/05/18 22:03:03 GC Percent установлен в 100%
2026/05/18 22:03:03 Сервер запущен на http://localhost:8080
2026/05/18 22:03:03 Метрики Prometheus: http://localhost:8080/metrics
2026/05/18 22:03:03 Pprof профилирование: http://localhost:8080/debug/pprof/
```

http://localhost:8080/metrics:

```bash
# HELP go_alloc_total_bytes Всего выделено байт (включая freed)
# TYPE go_alloc_total_bytes counter
go_alloc_total_bytes 249200
# HELP go_alloc_objects Всего выделено объектов
# TYPE go_alloc_objects counter
go_alloc_objects 736
# HELP go_freed_objects Всего освобождено объектов
# TYPE go_freed_objects counter
go_freed_objects 41
# HELP go_alloc_current_bytes Текущая выделенная память (in use)
# TYPE go_alloc_current_bytes gauge
go_alloc_current_bytes 249200
# HELP go_heap_alloc_bytes Выделенная heap-память
# TYPE go_heap_alloc_bytes gauge
go_heap_alloc_bytes 249200
# HELP go_heap_sys_bytes Системная heap-память
# TYPE go_heap_sys_bytes gauge
go_heap_sys_bytes 3899392
# HELP go_heap_idle_bytes Неиспользуемая heap-память
# TYPE go_heap_idle_bytes gauge
go_heap_idle_bytes 3022848
# HELP go_heap_inuse_bytes Используемая heap-память
# TYPE go_heap_inuse_bytes gauge
go_heap_inuse_bytes 876544
# HELP go_heap_released_bytes Возвращённая ОС память
# TYPE go_heap_released_bytes gauge
go_heap_released_bytes 3022848
# HELP go_stack_inuse_bytes Используемая стек-память
# TYPE go_stack_inuse_bytes gauge
go_stack_inuse_bytes 294912
# HELP go_gc_count_total Количество завершённых GC
# TYPE go_gc_count_total counter
go_gc_count_total 0
# HELP go_gc_pause_total_ns Общее время пауз GC в наносекундах
# TYPE go_gc_pause_total_ns counter
go_gc_pause_total_ns 0
# HELP go_gc_last_pause_ns Время последней паузы GC в наносекундах
# TYPE go_gc_last_pause_ns gauge
go_gc_last_pause_ns 0
# HELP go_gc_seconds_since_last Время с последнего GC (секунды)
# TYPE go_gc_seconds_since_last gauge
go_gc_seconds_since_last 0.000000
# HELP go_gc_next_gc_bytes Целевой размер heap для следующего GC
# TYPE go_gc_next_gc_bytes gauge
go_gc_next_gc_bytes 4194304
# HELP go_gc_percent Текущий GCPercent (debug.SetGCPercent)
# TYPE go_gc_percent gauge
go_gc_percent 100
# HELP go_goroutines_count Количество горутин
# TYPE go_goroutines_count gauge
go_goroutines_count 4
# HELP go_cgo_calls_total Всего CGO вызовов
# TYPE go_cgo_calls_total counter
go_cgo_calls_total 0
```

http://localhost:8080/debug/pprof/:

```bash
/debug/pprof/
Set debug=1 as a query parameter to export in legacy text format


Types of profiles available:
Count	Profile
2	allocs
0	block
0	cmdline
4	goroutine
2	heap
0	mutex
0	profile
0	symbol
5	threadcreate
0	trace
full goroutine stack dump
Profile Descriptions:

allocs: A sampling of all past memory allocations
block: Stack traces that led to blocking on synchronization primitives
cmdline: The command line invocation of the current program
goroutine: Stack traces of all current goroutines. Use debug=2 as a query parameter to export in the same format as an unrecovered panic.
heap: A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.
mutex: Stack traces of holders of contended mutexes
profile: CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.
symbol: Maps given program counters to function names. Counters can be specified in a GET raw query or POST body, multiple counters are separated by '+'.
threadcreate: Stack traces that led to the creation of new OS threads
trace: A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.
```



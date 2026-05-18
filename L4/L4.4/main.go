package main

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

func main() {
	initialGCPercent := 100
	debug.SetGCPercent(initialGCPercent)
	log.Printf("GC Percent установлен в %d%%", initialGCPercent)

	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/gc", gcHandler)
	http.HandleFunc("/debug/set_gc_percent", setGCPercentHandler)

	// Экспорт метрик
	http.Handle("/debug/vars", expvar.Handler())

	port := ":8080"
	log.Printf("Сервер запущен на http://localhost%s", port)
	log.Printf("Метрики Prometheus: http://localhost%s/metrics", port)
	log.Printf("Pprof профилирование: http://localhost%s/debug/pprof/", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Fprintf(w, "# HELP go_alloc_total_bytes Всего выделено байт (включая freed)\n")
	fmt.Fprintf(w, "# TYPE go_alloc_total_bytes counter\n")
	fmt.Fprintf(w, "go_alloc_total_bytes %d\n", memStats.TotalAlloc)

	fmt.Fprintf(w, "# HELP go_alloc_objects Всего выделено объектов\n")
	fmt.Fprintf(w, "# TYPE go_alloc_objects counter\n")
	fmt.Fprintf(w, "go_alloc_objects %d\n", memStats.Mallocs)

	fmt.Fprintf(w, "# HELP go_freed_objects Всего освобождено объектов\n")
	fmt.Fprintf(w, "# TYPE go_freed_objects counter\n")
	fmt.Fprintf(w, "go_freed_objects %d\n", memStats.Frees)

	fmt.Fprintf(w, "# HELP go_alloc_current_bytes Текущая выделенная память (in use)\n")
	fmt.Fprintf(w, "# TYPE go_alloc_current_bytes gauge\n")
	fmt.Fprintf(w, "go_alloc_current_bytes %d\n", memStats.Alloc)

	fmt.Fprintf(w, "# HELP go_heap_alloc_bytes Выделенная heap-память\n")
	fmt.Fprintf(w, "# TYPE go_heap_alloc_bytes gauge\n")
	fmt.Fprintf(w, "go_heap_alloc_bytes %d\n", memStats.HeapAlloc)

	fmt.Fprintf(w, "# HELP go_heap_sys_bytes Системная heap-память\n")
	fmt.Fprintf(w, "# TYPE go_heap_sys_bytes gauge\n")
	fmt.Fprintf(w, "go_heap_sys_bytes %d\n", memStats.HeapSys)

	fmt.Fprintf(w, "# HELP go_heap_idle_bytes Неиспользуемая heap-память\n")
	fmt.Fprintf(w, "# TYPE go_heap_idle_bytes gauge\n")
	fmt.Fprintf(w, "go_heap_idle_bytes %d\n", memStats.HeapIdle)

	fmt.Fprintf(w, "# HELP go_heap_inuse_bytes Используемая heap-память\n")
	fmt.Fprintf(w, "# TYPE go_heap_inuse_bytes gauge\n")
	fmt.Fprintf(w, "go_heap_inuse_bytes %d\n", memStats.HeapInuse)

	fmt.Fprintf(w, "# HELP go_heap_released_bytes Возвращённая ОС память\n")
	fmt.Fprintf(w, "# TYPE go_heap_released_bytes gauge\n")
	fmt.Fprintf(w, "go_heap_released_bytes %d\n", memStats.HeapReleased)

	fmt.Fprintf(w, "# HELP go_stack_inuse_bytes Используемая стек-память\n")
	fmt.Fprintf(w, "# TYPE go_stack_inuse_bytes gauge\n")
	fmt.Fprintf(w, "go_stack_inuse_bytes %d\n", memStats.StackInuse)

	fmt.Fprintf(w, "# HELP go_gc_count_total Количество завершённых GC\n")
	fmt.Fprintf(w, "# TYPE go_gc_count_total counter\n")
	fmt.Fprintf(w, "go_gc_count_total %d\n", memStats.NumGC)

	fmt.Fprintf(w, "# HELP go_gc_pause_total_ns Общее время пауз GC в наносекундах\n")
	fmt.Fprintf(w, "# TYPE go_gc_pause_total_ns counter\n")
	fmt.Fprintf(w, "go_gc_pause_total_ns %d\n", memStats.PauseTotalNs)

	fmt.Fprintf(w, "# HELP go_gc_last_pause_ns Время последней паузы GC в наносекундах\n")
	fmt.Fprintf(w, "# TYPE go_gc_last_pause_ns gauge\n")
	fmt.Fprintf(w, "go_gc_last_pause_ns %d\n", memStats.PauseNs[(memStats.NumGC+255)%256])

	if memStats.LastGC != 0 {
		fmt.Fprintf(w, "# HELP go_gc_last_time_seconds Время последнего GC (Unix timestamp)\n")
		fmt.Fprintf(w, "# TYPE go_gc_last_time_seconds gauge\n")
		fmt.Fprintf(w, "go_gc_last_time_seconds %f\n", float64(memStats.LastGC)/1e9)
	}

	secondsSinceLastGC := 0.0
	if memStats.LastGC != 0 {
		secondsSinceLastGC = time.Since(time.Unix(0, int64(memStats.LastGC))).Seconds()
	}
	fmt.Fprintf(w, "# HELP go_gc_seconds_since_last Время с последнего GC (секунды)\n")
	fmt.Fprintf(w, "# TYPE go_gc_seconds_since_last gauge\n")
	fmt.Fprintf(w, "go_gc_seconds_since_last %f\n", secondsSinceLastGC)

	fmt.Fprintf(w, "# HELP go_gc_next_gc_bytes Целевой размер heap для следующего GC\n")
	fmt.Fprintf(w, "# TYPE go_gc_next_gc_bytes gauge\n")
	fmt.Fprintf(w, "go_gc_next_gc_bytes %d\n", memStats.NextGC)

	fmt.Fprintf(w, "# HELP go_gc_percent Текущий GCPercent (debug.SetGCPercent)\n")
	fmt.Fprintf(w, "# TYPE go_gc_percent gauge\n")
	fmt.Fprintf(w, "go_gc_percent %d\n", debug.SetGCPercent(-1))

	fmt.Fprintf(w, "# HELP go_goroutines_count Количество горутин\n")
	fmt.Fprintf(w, "# TYPE go_goroutines_count gauge\n")
	fmt.Fprintf(w, "go_goroutines_count %d\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "# HELP go_cgo_calls_total Всего CGO вызовов\n")
	fmt.Fprintf(w, "# TYPE go_cgo_calls_total counter\n")
	fmt.Fprintf(w, "go_cgo_calls_total %d\n", runtime.NumCgoCall())
}

func gcHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Принудительный вызов GC")
	runtime.GC()
	fmt.Fprintf(w, "GC выполнен принудительно\n")
}

func setGCPercentHandler(w http.ResponseWriter, r *http.Request) {
	percentStr := r.URL.Query().Get("percent")
	if percentStr == "" {
		http.Error(w, "Необходимо указать параметр percent (целое число, -1 для отключения)", http.StatusBadRequest)
		return
	}

	percent, err := strconv.Atoi(percentStr)
	if err != nil {
		http.Error(w, "Параметр percent должен быть целым числом", http.StatusBadRequest)
		return
	}

	oldPercent := debug.SetGCPercent(percent)
	log.Printf("GC Percent изменён с %d на %d", oldPercent, percent)
	fmt.Fprintf(w, "GC Percent изменён с %d на %d\n", oldPercent, percent)
}

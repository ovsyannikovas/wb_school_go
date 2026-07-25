package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task определяет интерфейс задачи, которую могут выполнять воркеры
type Task interface {
	Execute()
}

// WorkerPool представляет пул воркеров с динамическим управлением
type WorkerPool struct {
	jobChan       chan Task
	commands      chan command
	metricsTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	// Статистика
	totalTasks  atomic.Int64
	workerCount atomic.Int32

	// Для graceful stop воркеров
	workerStopCh chan struct{}
	workerWG     sync.WaitGroup

	mu        sync.RWMutex
	isRunning bool
}

// command представляет команду управления пулом
type command struct {
	cmdType string // "add" или "remove"
	count   int
}

// NewWorkerPool создает новый пул воркеров
func NewWorkerPool(queue int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool{
		jobChan:      make(chan Task, queue),
		commands:     make(chan command, 10),
		ctx:          ctx,
		cancel:       cancel,
		workerStopCh: make(chan struct{}),
		isRunning:    false,
	}

	return wp
}

// AddWorker увеличивает число воркеров на n
func (wp *WorkerPool) AddWorker(n int) {
	if n <= 0 {
		return
	}

	wp.commands <- command{cmdType: "add", count: n}
}

// RemoveWorker уменьшает число воркеров на n (graceful stop)
func (wp *WorkerPool) RemoveWorker(n int) {
	if n <= 0 {
		return
	}

	wp.commands <- command{cmdType: "remove", count: n}
}

// Start запускает пул и все начальные воркеры
func (wp *WorkerPool) Start() {
	wp.mu.Lock()
	if wp.isRunning {
		wp.mu.Unlock()
		return
	}
	wp.isRunning = true
	wp.mu.Unlock()

	// Запускаем 1 воркер по умолчанию
	wp.AddWorker(1)

	// Запускаем горутину для обработки команд
	go wp.commandProcessor()

	// Запускаем горутину для метрик
	go wp.metricsReporter()
}

// Stop останавливает пул воркеров
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	if !wp.isRunning {
		wp.mu.Unlock()
		return
	}
	wp.isRunning = false
	wp.mu.Unlock()

	// Отменяем контекст
	wp.cancel()

	// Останавливаем метрики
	if wp.metricsTicker != nil {
		wp.metricsTicker.Stop()
	}

	// Закрываем канал команд
	close(wp.commands)

	// Закрываем канал задач
	close(wp.jobChan)

	// Закрываем канал остановки воркеров
	close(wp.workerStopCh)

	// Ждем завершения всех воркеров
	wp.workerWG.Wait()

	// Ждем завершения commandProcessor
	wp.wg.Wait()
}

// commandProcessor обрабатывает команды управления пулом
func (wp *WorkerPool) commandProcessor() {
	wp.wg.Add(1)
	defer wp.wg.Done()

	for cmd := range wp.commands {
		select {
		case <-wp.ctx.Done():
			return
		default:
			switch cmd.cmdType {
			case "add":
				wp.addWorkers(cmd.count)
			case "remove":
				wp.removeWorkers(cmd.count)
			}
		}
	}
}

// addWorkers добавляет новых воркеров
func (wp *WorkerPool) addWorkers(n int) {
	for i := 0; i < n; i++ {
		wp.workerWG.Add(1)
		go wp.worker()
	}
}

// removeWorkers удаляет воркеров (graceful stop)
func (wp *WorkerPool) removeWorkers(n int) {
	for i := 0; i < n; i++ {
		select {
		case <-wp.ctx.Done():
			return
		default:
			wp.workerStopCh <- struct{}{}
		}
	}
}

// worker представляет горутину, выполняющую задачи
func (wp *WorkerPool) worker() {
	defer wp.workerWG.Done()

	// Увеличиваем счетчик активных воркеров
	wp.workerCount.Add(1)
	defer wp.workerCount.Add(-1)

	for {
		select {
		case <-wp.ctx.Done():
			return

		case task, ok := <-wp.jobChan:
			if !ok {
				return
			}

			// Выполняем задачу
			task.Execute()

			// Увеличиваем счетчик выполненных задач
			wp.totalTasks.Add(1)

		case <-wp.workerStopCh:
			// Получаем сигнал на graceful stop
			// Проверяем, есть ли еще задачи в канале
			select {
			case task, ok := <-wp.jobChan:
				if !ok {
					return
				}
				// Выполняем текущую задачу перед завершением
				task.Execute()
				wp.totalTasks.Add(1)
				return
			default:
				// Нет задач - завершаемся
				return
			}
		}
	}
}

// metricsReporter выводит метрики с заданным интервалом
func (wp *WorkerPool) metricsReporter() {
	wp.metricsTicker = time.NewTicker(5 * time.Second) // Настраиваемый интервал
	defer wp.metricsTicker.Stop()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case <-wp.metricsTicker.C:
			total := wp.totalTasks.Load()
			workers := wp.workerCount.Load()
			queueLen := len(wp.jobChan)

			fmt.Printf("[Metrics] Total tasks: %d, Active workers: %d, Queue length: %d\n",
				total, workers, queueLen)
		}
	}
}

// Submit добавляет задачу в пул
func (wp *WorkerPool) Submit(task Task) bool {
	select {
	case <-wp.ctx.Done():
		return false
	case wp.jobChan <- task:
		return true
	}
}

// GetStats возвращает текущую статистику
func (wp *WorkerPool) GetStats() (total int64, workers int32, queueLen int) {
	return wp.totalTasks.Load(), wp.workerCount.Load(), len(wp.jobChan)
}

// PrintTask - пример задачи, реализующей интерфейс Task
type PrintTask struct {
	ID int
}

func (t PrintTask) Execute() {
	fmt.Printf("Executing task %d\n", t.ID)
	time.Sleep(500 * time.Millisecond) // Имитируем работу
}

// Пример использования
func main() {
	// Создаем пул с буфером очереди 10
	pool := NewWorkerPool(10)

	// Запускаем пул
	pool.Start()

	// Отправляем задачи
	go func() {
		for i := 0; i < 20; i++ {
			pool.Submit(PrintTask{ID: i})
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Динамически управляем количеством воркеров
	go func() {
		time.Sleep(1 * time.Second)
		pool.AddWorker(3) // Добавляем 3 воркера

		time.Sleep(2 * time.Second)
		pool.RemoveWorker(2) // Убираем 2 воркера

		time.Sleep(2 * time.Second)
		pool.AddWorker(2) // Добавляем еще 2
	}()

	// Ждем завершения
	time.Sleep(8 * time.Second)

	// Останавливаем пул
	pool.Stop()
	fmt.Println("Pool stopped")
}

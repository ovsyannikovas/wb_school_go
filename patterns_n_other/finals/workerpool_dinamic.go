package main

import (
	"errors"
	"sync"
)

var (
	ErrPoolClosed = errors.New("worker pool is closed")
	ErrQueueFull  = errors.New("task queue is full")
	ErrNilTask    = errors.New("task cannot be nil")
)

type WorkerPool struct {
	taskChan chan func()

	wg   sync.WaitGroup
	once sync.Once

	mu     sync.RWMutex
	closed bool
}

// NewWorkerPool создаёт пул с указанным количеством воркеров
// и размером очереди.
func NewWorkerPool(workerCount int, queueSize int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 1
	}

	if queueSize < 0 {
		queueSize = 0
	}

	p := &WorkerPool{
		taskChan: make(chan func(), queueSize),
	}

	p.wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go p.worker()
	}

	return p
}

// worker выполняет задачи до тех пор,
// пока taskChan не будет закрыт и полностью прочитан.
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for task := range p.taskChan {
		task()
	}
}

// Submit добавляет задачу в очередь.
//
// Если пул уже закрывается или закрыт — возвращает ErrPoolClosed.
// Если очередь заполнена — возвращает ErrQueueFull.
func (p *WorkerPool) Submit(task func()) error {
	if task == nil {
		return ErrNilTask
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrPoolClosed
	}

	select {
	case p.taskChan <- task:
		return nil

	default:
		return ErrQueueFull
	}
}

// Shutdown выполняет graceful shutdown.
//
// 1. Запрещает новые Submit.
// 2. Закрывает taskChan.
// 3. Воркеры выполняют все задачи, которые уже находятся в очереди.
// 4. После опустошения очереди воркеры завершаются.
// 5. Ждём завершения всех воркеров.
func (p *WorkerPool) Shutdown() {
	p.once.Do(func() {
		p.mu.Lock()

		p.closed = true
		close(p.taskChan)

		p.mu.Unlock()

		p.wg.Wait()
	})
}

package main

import (
	"fmt"
	"sync"
)

// Подсказка:
// type Barrier interface {
//	Await() error
//	Reset()
//}
// 6 значений в структуре, 3 каунтера
// резет дублирует часть await

type Worker interface {
	Work()
}

type Barrier interface {
	Await() error
	Reset()
}

// Нужно реализовать циклический барьер, который позволяет N горутинам ждать друг друга
// Когда все N горутин вызвали метод Await(), они одновременно освобождаются,
// и барьер сбрасывается (становится готовым к следующему использованию)
type CyclicBarrier struct {
	mu       sync.Mutex
	cond     *sync.Cond
	parties  int
	count    int
	gen      int // поколение успешного сбора
	resetGen int // поколение сброса
}

func NewCyclicBarrier(parties int) *CyclicBarrier {
	b := &CyclicBarrier{parties: parties}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Await блокирует горутину до тех пор, пока все parties не вызовут Await
// Когда все собрались, барьер сбрасывается и все ожидающие возвращаются
// Возвращает nil при успешной синхронизации, иначе ошибку (например, если барьер сброшен через Reset)
func (b *CyclicBarrier) Await() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	myGen := b.gen
	myResetGen := b.resetGen

	b.count++
	if b.count == b.parties {
		b.count = 0
		b.gen++
		b.cond.Broadcast()
		return nil
	}

	for b.gen == myGen && b.resetGen == myResetGen {
		b.cond.Wait()
	}

	if b.resetGen != myResetGen {
		return fmt.Errorf("barrier reset")
	}
	return nil
}

// Reset принудительно сбрасывает барьер, пробуждая все ожидающие горутины с ошибкой
// После сброса барьер готов к новому использованию
func (b *CyclicBarrier) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.count = 0
	b.resetGen++
	b.cond.Broadcast()
}

// usage

func main() {
	barrier := NewCyclicBarrier(3)
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			fmt.Printf("Горутина %d: ждет у барьера\n", id)
			barrier.Await()
			fmt.Printf("Горутина %d: прошла барьер!\n", id)

			// Барьер можно использовать повторно
			fmt.Printf("Горутина %d: снова ждет\n", id)
			barrier.Await()
			fmt.Printf("Горутина %d: снова прошла!\n", id)
		}(i)
	}

	wg.Wait()
}

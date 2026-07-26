package main

import (
	"context"
	"sync"
)

// Подсказка:
// generic
// out, wg, проходимся по каналам, по каждому горутиним с селектом, каждый читаем ДО КОНЦА + горутина закрыватор

func FanIn[T any](ctx context.Context, chans ...<-chan T) <-chan T {
	out := make(chan T)

	var wg sync.WaitGroup
	for _, ch := range chans {
		wg.Add(1)
		go func(ch <-chan T) {
			defer wg.Done()
			for v := range ch {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

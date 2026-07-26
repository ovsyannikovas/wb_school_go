package main

import (
	"context"
	"fmt"
	"sync"
)

// Подсказка:
// как fanin, но применяем transform

func Transformer[T any, R any](
	ctx context.Context, input <-chan T, transform func(T) R, workers int,
) <-chan R {
	out := make(chan R)
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for v := range input {
				select {
				case out <- transform(v):
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	ctx := context.Background()
	in := make(chan int)

	out := Transformer(ctx, in, func(v int) string {
		return fmt.Sprintf("value=%d", v)
	}, 3)

	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	for result := range out {
		fmt.Println(result)
	}
}

package main

import (
	"context"
	"fmt"
)

func Transformer[T any, R any](ctx context.Context, in <-chan T, fn func(T) R) <-chan R {
	out := make(chan R)
	go func() {
		defer close(out)
		for v := range in {
			select {
			case out <- fn(v):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func main() {
	ctx := context.Background()
	in := make(chan int)

	out := Transformer(ctx, in, func(v int) string {
		return fmt.Sprintf("value=%d", v)
	})

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

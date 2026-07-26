package main

import "context"

// Подсказка:
// ctx, resultCh
// Get()
// все самое важное в создании, буферизированный канал, одна горутина

type Future[T any] struct {
	ctx      context.Context
	resultCh chan T
}

func NewFuture[T any](ctx context.Context, action func(ctx context.Context) T) *Future[T] {
	future := &Future[T]{
		ctx:      ctx,
		resultCh: make(chan T, 1),
	}

	go func() {
		defer close(future.resultCh)

		res := action(ctx)

		select {
		case <-ctx.Done():
			return
		case future.resultCh <- res:
		}
	}()

	return future
}

func (f *Future[T]) Get() (T, error) {
	select {
	case <-f.ctx.Done():
		var zero T
		return zero, f.ctx.Err()
	case res := <-f.resultCh:
		return res, nil
	}
}

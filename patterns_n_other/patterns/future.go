package main

import "context"

type Future[T any] struct {
	ctx      context.Context
	cancel   context.CancelFunc
	resultCh chan T
}

func NewFuture[T any](ctx context.Context, action func(ctx context.Context) T) *Future[T] {
	ctx, cancel := context.WithCancel(ctx)

	future := &Future[T]{
		ctx:      ctx,
		cancel:   cancel,
		resultCh: make(chan T, 1),
	}

	go func() {
		defer func() {
			close(future.resultCh)
		}()

		done := make(chan T, 1)
		go func() {
			done <- action(ctx)
		}()

		select {
		case <-ctx.Done():
			return
		case result := <-done:
			future.resultCh <- result
		}
	}()

	return future
}

func (f *Future[T]) Get() (T, error) {
	select {
	case <-f.ctx.Done():
		var zero T
		return zero, f.ctx.Err()
	case result := <-f.resultCh:
		return result, nil
	}
}

package main

type result[T any] struct {
	val T
	err error
}

type Promise[T any] struct {
	resultCh chan result[T]
}

// NewPromise создает новый Promise
func NewPromise[T any](asyncFn func() (T, error)) *Promise[T] {
	promise := &Promise[T]{
		resultCh: make(chan result[T]),
	}

	go func() {
		defer close(promise.resultCh)
		val, err := asyncFn()
		promise.resultCh <- result[T]{val: val, err: err}
	}()

	return promise
}

// Then - выполняет successFn при успехе или errorFn при ошибке
func (p *Promise[T]) Then(successFn func(T), errorFn func(error)) *Promise[T] {
	go func() {
		res := <-p.resultCh
		if res.err == nil {
			successFn(res.val)
		} else {
			if errorFn != nil {
				errorFn(res.err)
			}
		}
	}()
	return p
}

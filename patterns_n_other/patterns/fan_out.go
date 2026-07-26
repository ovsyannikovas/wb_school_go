package main

import "context"

// Подсказка:
// generic
// outputs, result, одна горутина на все проходы
// рассылаем ВСЕМ outputs

func FanOut[T any](ctx context.Context, in <-chan T, numOut int) []<-chan T {
	if numOut <= 0 {
		return []<-chan T{}
	}
	outputs := make([]chan T, numOut)
	result := make([]<-chan T, numOut)

	for i := 0; i < numOut; i++ {
		ch := make(chan T)
		outputs[i] = ch
		result[i] = ch
	}
	go func() {
		defer func() {
			// Close all output channels when done
			for _, ch := range outputs {
				close(ch)
			}
		}()
		for v := range in {
			// Send value to all outputs
			for _, ch := range outputs {
				select {
				case ch <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return result
}

package channel

func Filter[T any](in <-chan T, predicate func(T) bool) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for v := range in {
			if predicate(v) {
				out <- v
			}
		}
	}()
	return out
}

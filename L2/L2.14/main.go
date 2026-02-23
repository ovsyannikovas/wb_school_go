package main

import (
	"fmt"
	"time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {
	orDone := make(chan interface{})

	go func() {
		defer close(orDone)

		if len(channels) == 0 {
			return
		}

		active := make([]<-chan interface{}, len(channels))
		copy(active, channels)

		for len(active) > 0 {
			for _, ch := range active {
				select {
				case <-ch:
					return
				default:
					// Канал не готов, продолжаем проверку
				}
			}
			time.Sleep(time.Microsecond * 10)
		}
	}()

	return orDone
}

func orWithRecursion(channels ...<-chan interface{}) <-chan interface{} {
	orDone := make(chan interface{})

	go func() {
		defer close(orDone)

		switch len(channels) {
		case 0:
			return
		case 1:
			<-channels[0]
			return
		}

		select {
		case <-channels[0]:
		case <-channels[1]:
		case <-or(channels[2:]...):
		}
	}()

	return orDone
}

func main() {
	orFuncs := []func(channels ...<-chan interface{}) <-chan interface{}{
		or, orWithRecursion,
	}

	for _, orFunc := range orFuncs {
		sig := func(after time.Duration) <-chan interface{} {
			c := make(chan interface{})
			go func() {
				defer close(c)
				time.Sleep(after)
			}()
			return c
		}

		start := time.Now()
		<-orFunc(
			sig(2*time.Second),
			sig(5*time.Second),
			sig(1*time.Second),
			sig(3*time.Second),
			sig(4*time.Second),
		)
		fmt.Printf("done after %v\n", time.Since(start))
	}
}

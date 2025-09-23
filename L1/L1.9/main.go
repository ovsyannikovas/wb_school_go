package main

import (
	"fmt"
	"sync"
)

func main() {
	channelX := make(chan int)
	channelX2 := make(chan int)

	wg := sync.WaitGroup{}

	wg.Add(3)

	go func(ch chan<- int, wg *sync.WaitGroup) {
		defer wg.Done()
		defer close(ch)

		for i := range 10 {
			channelX <- i
		}

	}(channelX, &wg)

	go func(ch1 <-chan int, ch2 chan<- int, wg *sync.WaitGroup) {
		defer wg.Done()
		defer close(ch2)

		for val := range ch1 {
			ch2 <- val * 2
		}

	}(channelX, channelX2, &wg)

	go func(ch <-chan int, wg *sync.WaitGroup) {
		defer wg.Done()

		for i := range ch {
			fmt.Println(i)
		}

	}(channelX2, &wg)

	wg.Wait()
}

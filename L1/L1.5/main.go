package main

import (
	"fmt"
	"sync"
	"time"
)

func reader(ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		i, ok := <-ch
		if !ok {
			return
		}
		fmt.Println(i)
	}
}

func writer(timer <-chan time.Time, ch chan int, wg *sync.WaitGroup) {
	defer close(ch)
	defer wg.Done()

	i := 0
	for {
		select {
		case <-timer:
			return
		default:
			ch <- i
			i++
			time.Sleep(500 * time.Millisecond)
		}
	}
}
func main() {
	timer := time.After(3 * time.Second)

	ch := make(chan int)
	wg := sync.WaitGroup{}

	wg.Add(2)
	go writer(timer, ch, &wg)
	go reader(ch, &wg)

	wg.Wait()
}

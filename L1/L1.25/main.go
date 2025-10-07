package main

import (
	"fmt"
	"sync"
	"time"
)

func sleep(duration time.Duration) {
	<-time.After(duration)
}

func main() {
	fmt.Println("Start waiting")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Start goroutine")
		sleep(3 * time.Second)
		fmt.Println("End goroutine")
	}()
	wg.Wait()
	fmt.Println("Stop waiting")
}

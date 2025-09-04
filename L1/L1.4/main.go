package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func worker(ctx context.Context, ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case i := <-ch:
			fmt.Println(i)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <num_workers>")
		return
	}

	numWorkers, err := strconv.Atoi(os.Args[1])
	if err != nil || numWorkers <= 0 {
		fmt.Println("Invalid number of workers")
		return
	}

	ch := make(chan int)

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Главная пишущая горутина
	go func(ctx context.Context) {
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			case ch <- i:
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}(ctx)

	wg.Add(numWorkers)
	for range numWorkers {
		go worker(ctx, ch, &wg)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	cancel()
	close(ch)
	wg.Wait()
}

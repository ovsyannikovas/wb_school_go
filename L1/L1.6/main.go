package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func main() {
	wg := sync.WaitGroup{}

	fmt.Println("\n1. By condition")
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for i := 0; i < 5; i++ {
			if i == 3 {
				fmt.Println("	Goroutine 1: ended on step", i)
				return
			}
			fmt.Printf("	Goroutine 1: step %d\n", i)
		}
	}(&wg)
	wg.Wait()

	fmt.Println("\n2. By notification channel")
	wg.Add(1)
	quit := make(chan struct{})
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-quit:
				fmt.Println("	Goroutine 2: got from quit channel")
				return
			default:
				fmt.Println("	Goroutine 2: working...")
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(&wg)
	time.Sleep(600 * time.Millisecond)
	quit <- struct{}{}
	wg.Wait()

	fmt.Println("\n3. By context cancel")
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("	Goroutine 3: got context done")
				return
			default:
				fmt.Println("	Goroutine 3: working...")
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(ctx, &wg)
	time.Sleep(600 * time.Millisecond)
	cancel()
	wg.Wait()

	fmt.Println("\n4. By context timeout")
	wg.Add(1)
	ctx, cancel = context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("	Goroutine 4: got context done")
				return
			default:
				fmt.Println("	Goroutine 4: working...")
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(ctx, &wg)
	wg.Wait()

	fmt.Println("\n5. By context deadline")
	wg.Add(1)
	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(600*time.Millisecond))
	defer cancel()
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("	Goroutine 5: got context done")
				return
			default:
				fmt.Println("	Goroutine 5: working...")
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(ctx, &wg)
	wg.Wait()

	fmt.Println("\n6. By runtime.Goexit()")
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for i := 0; i < 5; i++ {
			if i == 3 {
				fmt.Println("	Goroutine 6: ended on step", i)
				runtime.Goexit()
			}
			fmt.Printf("	Goroutine 6: step %d\n", i)
		}
	}(&wg)
	wg.Wait()

	fmt.Println("\n7. By os signal")
	wg.Add(1)

	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("	Goroutine 7: got os signal")
				return
			default:
				fmt.Println("	Goroutine 7: working...")
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(ctx, &wg)
	time.Sleep(600 * time.Millisecond)
	err := syscall.Kill(os.Getpid(), syscall.SIGTERM)
	if err != nil {
		fmt.Println("Error after signal sent:", err)
	}
	wg.Wait()
}

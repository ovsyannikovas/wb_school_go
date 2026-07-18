package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	NumJobs    = 5
	NumWorkers = 3
)

func process(id int, jobs <-chan int, result chan<- int) {
	for j := range jobs {
		fmt.Printf("worker: %d started job: %d\n", id, j)
		time.Sleep(time.Second)
		fmt.Printf("worker: %d finished job: %d\n", id, j)
		result <- j * j
	}
}

func main() {
	jobs := make(chan int, NumJobs)
	result := make(chan int, NumJobs)

	var wg sync.WaitGroup
	for w := 1; w <= NumWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			process(id, jobs, result)
		}(w)
	}

	for j := 1; j <= NumJobs; j++ {
		jobs <- j
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(result)
	}()

	for r := range result {
		fmt.Printf("result: %d\n", r)
	}
}

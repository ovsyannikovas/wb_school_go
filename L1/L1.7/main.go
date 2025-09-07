package main

import (
	"fmt"
	"sync"
)

type concurrentMap struct {
	mx sync.Mutex
	m  map[string]int
}

func main() {
	cMap := concurrentMap{
		m:  make(map[string]int),
		mx: sync.Mutex{},
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			cMap.mx.Lock()
			defer cMap.mx.Unlock()

			cMap.m["key"] = i

			fmt.Printf("Map in goroutine %d: %v\n", i, cMap.m)
		}(&wg)
	}

	wg.Wait()
	fmt.Printf("Map in main goroutine: %v\n", cMap.m)
}

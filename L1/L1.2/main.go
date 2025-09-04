package main

import (
	"fmt"
	"sync"
)

func main() {
	nums := []int{2, 4, 6, 8, 10}
	wg := sync.WaitGroup{}

	wg.Add(len(nums))
	for _, i := range nums {
		go func(num int) {
			defer wg.Done()
			fmt.Println(num * num)
		}(i)
	}
	wg.Wait()
}

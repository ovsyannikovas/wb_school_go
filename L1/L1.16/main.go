package main

import (
	"fmt"
)

func quickSort(array []int) []int {
	if len(array) <= 1 {
		return array
	}

	middle := array[len(array)/2]
	left := make([]int, 0, len(array))
	right := make([]int, 0, len(array))
	equal := make([]int, 0, len(array))

	for _, num := range array {
		if num < middle {
			left = append(left, num)
		} else if num > middle {
			right = append(right, num)
		} else {
			equal = append(equal, num)
		}
	}

	left = quickSort(left)
	right = quickSort(right)

	result := append(left, equal...)
	result = append(result, right...)

	return result
}

func main() {
	fmt.Printf("%v\n", quickSort([]int{3, 8, 2, 4, 4, 1, 10}))
}

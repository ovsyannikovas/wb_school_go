package main

import "fmt"

func main() {
	slice1 := []int{1, 2, 3}
	slice2 := []int{2, 3, 4}

	fmt.Println(intersect(slice1, slice2))
}

func intersect(slice1, slice2 []int) []int {
	set := make(map[int]struct{})

	for _, v := range slice1 {
		set[v] = struct{}{}
	}

	var result []int

	for _, v := range slice2 {
		_, ok := set[v]

		if ok {
			result = append(result, v)
		}
	}

	return result
}

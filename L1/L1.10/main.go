package main

import (
	"fmt"
	"sort"
)

func main() {
	temperature := []float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5}
	resultMap := make(map[int][]float64)

	sort.Float64s(temperature) // -27 -25.4 -21 13 15.5 19 24.5 32.5

	for _, temp := range temperature {
		key := int(temp/10) * 10
		resultMap[key] = append(resultMap[key], temp)
	}

	fmt.Printf("%v", resultMap)
}

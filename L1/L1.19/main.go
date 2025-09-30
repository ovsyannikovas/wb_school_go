package main

import (
	"fmt"
	"os"
)

func reverseString(s string) string {
	runes := []rune(s)
	result := make([]rune, 0, len(s))

	for i := len(runes) - 1; i >= 0; i-- {
		result = append(result, runes[i])
	}

	return string(result)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <world>")
		return
	}

	fmt.Println(reverseString(os.Args[1]))
}

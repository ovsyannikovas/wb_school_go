package main

import (
	"fmt"
	"os"
	"strings"
)

func isUnique(s string) bool {
	uniqueSet := make(map[rune]struct{})
	for _, c := range strings.ToLower(s) {
		uniqueSet[c] = struct{}{}
	}

	return len(uniqueSet) == len(s)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage go run main.go <string>")
		return
	}

	fmt.Println(isUnique(os.Args[1]))
}

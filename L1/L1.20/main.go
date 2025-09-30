package main

import (
	"fmt"
)

func reverseWords(s string) string {
	runes := []rune(s)
	n := len(runes)

	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	start := 0 // Начало слова
	for i := range n {
		if runes[i] == ' ' || i == n {
			for l, r := start, i-1; l < r; l, r = l+1, r-1 {
				runes[l], runes[r] = runes[r], runes[l]
			}
			start = i + 1
		}
	}

	return string(runes)
}

func main() {
	fmt.Println(reverseWords("snow dog sun"))
}

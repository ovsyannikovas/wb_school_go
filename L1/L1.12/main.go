package main

import (
	"fmt"
)

func main() {
	words := []string{"cat", "cat", "dog", "cat", "tree"}

	wordMap := make(map[string]struct{})

	for _, word := range words {
		wordMap[word] = struct{}{}
	}

	uniqueWords := make([]string, 0, len(wordMap))
	for word := range wordMap {
		uniqueWords = append(uniqueWords, word)
	}

	fmt.Printf("%v\n", uniqueWords)
}

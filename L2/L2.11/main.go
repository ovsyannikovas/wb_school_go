package main

import (
	"fmt"
	"sort"
	"strings"
)

type anagramMapValueStruct struct {
	firstWord string
	words     []string
}

func findAnagram(list []string) map[string][]string {
	anagramMap := make(map[string]anagramMapValueStruct) // отсортированное слово: {первое слово, массив слов}

	for _, word := range list {
		word = strings.ToLower(word)
		runes := []rune(word)
		sort.Slice(runes, func(i, j int) bool {
			return runes[i] < runes[j]
		})
		value, exists := anagramMap[string(runes)]
		if !exists {
			anagramMap[string(runes)] = anagramMapValueStruct{
				firstWord: word,
				words:     []string{word},
			}
		} else {
			value.words = append(value.words, word)
			anagramMap[string(runes)] = value
		}

	}

	// Формируем результирующую мапу в нужном формате
	result := make(map[string][]string)
	for _, group := range anagramMap {
		if len(group.words) > 1 {
			sort.Strings(group.words)
			result[group.firstWord] = group.words
		}
	}

	return result
}

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	fmt.Println(findAnagram(words))
}

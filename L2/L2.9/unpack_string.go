package main

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

func UnpackString(str string) (string, error) {
	var result []rune
	runes := []rune(str)

	for i := 0; i < len(runes); i++ {
		s := runes[i]

		if s == '\\' {
			if i+1 >= len(runes) {
				return "", errors.New("некорректная escape-последовательность")
			}

			result = append(result, runes[i+1])
			i++
			continue
		}

		if !unicode.IsNumber(s) {
			result = append(result, s)
			continue
		}

		if unicode.IsNumber(s) {
			j := i
			for ; j < len(runes); j++ {
				if !unicode.IsNumber(runes[j]) {
					break
				}
			}

			num, err := strconv.Atoi(string(runes[i:j]))
			if err != nil {
				return "", fmt.Errorf("ошибка преобразования числа: %w", err)
			}

			if num == 0 {
				return "", errors.New("некорректное число: 0")
			}

			if len(result) == 0 {
				return "", errors.New("некорректная строка: цифра без предшествующего символа")
			}

			lastChar := result[len(result)-1]

			for k := 1; k < num; k++ {
				result = append(result, lastChar)
			}

			i = j - 1
		}
	}

	return string(result), nil
}

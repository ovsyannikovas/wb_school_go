package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <value>")
		return
	}

	// Проверим другие типы
	parseAndDetectType(os.Args[1])

	// Проверим канал
	ch := make(chan interface{})
	detectType(ch)
}

func parseAndDetectType(value string) {
	if value == "true" || value == "false" {
		if value == "true" {
			detectType(true)
		} else {
			detectType(false)
		}
		return
	}

	if i, err := strconv.Atoi(value); err == nil {
		detectType(i)
		return
	}

	detectType(value)
}

func detectType(value interface{}) {
	switch value.(type) {
	case int:
		fmt.Println("int")
	case string:
		fmt.Println("string")
	case bool:
		fmt.Println("bool")
	case chan interface{}:
		fmt.Println("chan")
	default:
		fmt.Println("unknown type")
	}
}

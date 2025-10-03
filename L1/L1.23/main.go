package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Println("go run main.go <index>")
		return
	}

	slice := []int{1, 2, 3, 4, 5, 6, 7}
	fmt.Println("Исходный слайс:", slice)

	i, err := strconv.Atoi(os.Args[1])
	if err != nil || i < 0 || i >= len(slice) {
		fmt.Println("Invalid index")
		return
	}

	num := copy(slice[i:], slice[i+1:])
	if num != len(slice)-i-1 {
		fmt.Println("Invalid copy")
		return
	}

	slice = slice[:len(slice)-1]

	fmt.Println("Слайс после удаления:", slice)
}

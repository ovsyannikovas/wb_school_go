package main

import "fmt"

func main() {
	a := 2
	b := 3

	fmt.Println(a, b) // a = 2, b = 3

	a = a + b // a = 5, b = 3
	b = a - b // a = 5, b = 2
	a = a - b // a = 3, b = 2

	fmt.Println(a, b) // a = 3, b = 2

	a = a ^ b // a = 1, b = 3 / 11 ^ 10 = 01
	b = b ^ a // a = 1, b = 2 / 11 ^ 01 = 10
	a = b ^ a // a = 3, b = 2 / 01 ^ 10 = 11

	fmt.Println(a, b) // a = 2, b = 3
}

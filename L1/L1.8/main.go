package main

import (
	"fmt"
	"os"
	"strconv"
)

func setNumBitToBin(num int64, bit int, bin int) int64 {
	mask := int64(1) << bit
	if bin == 1 {
		// 0 OR 1 = 1, 1 OR 1 = 1
		num |= mask
	} else {
		// 1 AND NOT 1 = 0, 0 AND NOT 1 = 0
		num &^= mask
	}
	return num
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <num> <bit> <bin (1/0)>")
		return
	}

	num, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		fmt.Println("Invalid number")
		return
	}
	bit, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid number")
		return
	}
	bit--
	bin, err := strconv.Atoi(os.Args[3])
	if err != nil || (bin != 0 && bin != 1) {
		fmt.Println("Invalid number")
		return
	}

	result := setNumBitToBin(num, bit, bin)

	fmt.Printf("Result: %d, binary: %08b\n", result, result)
}

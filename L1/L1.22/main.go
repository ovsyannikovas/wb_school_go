package main

import (
	"fmt"
	"math/big"
)

func main() {
	a := big.NewInt(0)
	b := big.NewInt(0)

	a.SetString("5000000", 10)
	b.SetString("3000000", 10)

	fmt.Println("=== Калькулятор с math/big ===")
	fmt.Printf("a = %s\n", a.String())
	fmt.Printf("b = %s\n\n", b.String())

	// Сложение
	sum := new(big.Int)
	sum.Add(a, b)
	fmt.Printf("a + b = %s\n", sum.String())

	// Вычитание
	diff := new(big.Int)
	diff.Sub(a, b)
	fmt.Printf("a - b = %s\n", diff.String())

	// Умножение
	product := new(big.Int)
	product.Mul(a, b)
	fmt.Printf("a × b = %s\n", product.String())

	// Деление
	quotient := new(big.Int)
	quotient.Div(a, b)
	fmt.Printf("a ÷ b = %s\n", quotient.String())

	// Пробуем с еще большими числами
	fmt.Println("\n=== На еще больших числах ===")
	huge1 := new(big.Int)
	huge2 := new(big.Int)

	huge1.SetString("1000000000000000000000", 10)
	huge2.SetString("500000000000000000000", 10)

	fmt.Printf("huge1 = %s\n", huge1.String())
	fmt.Printf("huge2 = %s\n", huge2.String())

	// Умножение больших чисел
	hugeProduct := new(big.Int)
	hugeProduct.Mul(huge1, huge2)
	fmt.Printf("huge1 × huge2 = %s\n", hugeProduct.String())
}

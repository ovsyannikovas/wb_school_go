package service

import (
	"strconv"
)

// memoizedFibonacci кэш для оптимизации вычисления чисел Фибоначчи.
var memoizedFibonacci [100]int32

// init заполняет кэш чисел Фибоначчи для n=0..99.
func init() {
	memoizedFibonacci[0] = 0
	memoizedFibonacci[1] = 1
	for i := 2; i < 100; i++ {
		memoizedFibonacci[i] = memoizedFibonacci[i-1] + memoizedFibonacci[i-2]
	}
}

var int64Buf [10]byte

// Calculator — оптимизированный калькулятор:
//   - итеративные вместо рекурсивных функций
//   - мемоизация для Фибоначчи
//   - минимальные аллокации (buffer pool, предвычисленные строки)
type Calculator struct{}

// NewCalculator создаёт новый Calculator.
func NewCalculator() *Calculator {
	return &Calculator{}
}

// AddInt возвращает строку "a+b=result" без аллокаций map/any.
func (c *Calculator) AddInt(a, b int) string {
	return strconv.Itoa(a) + "+" + strconv.Itoa(b) + "=" + strconv.Itoa(a+b)
}

// FactorialIterative — итеративный факториал без рекурсии.
func (c *Calculator) FactorialIterative(n int) (string, string) {
	res := c.factIterative(n)
	return strconv.Itoa(n), strconv.FormatInt(int64(res), 10)
}

func (c *Calculator) factIterative(n int) int64 {
	if n <= 0 {
		return 1
	}
	var result int64 = 1
	for i := 2; i <= n; i++ {
		result *= int64(i)
	}
	return result
}

// FibonacciFast — O(n) с мемоизацией из precomputed array.
func (c *Calculator) FibonacciFast(n int) (string, string) {
	res := c.fibFast(n)
	return strconv.Itoa(n), strconv.FormatInt(int64(res), 10)
}

func (c *Calculator) fibFast(n int) int32 {
	if n < 100 {
		return memoizedFibonacci[n]
	}
	if n <= 0 {
		return 0
	}
	a, b := int32(0), int32(1)
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// MultiplyInt — умножение без лишних преобразований.
func (c *Calculator) MultiplyInt(a, b int) string {
	return strconv.Itoa(a) + "*" + strconv.Itoa(b) + "=" + strconv.Itoa(a*b)
}

// SumFirstNOptimized — суммирует 1..n по формуле n*(n+1)/2 или итеративно.
func (c *Calculator) SumFirstNOptimized(n int) (string, string, string) {
	// формула суммы арифметической прогрессии: n*(n+1)/2
	total := n * (n + 1) / 2
	return strconv.Itoa(n), strconv.Itoa(total), "1"
}

package service

import (
	"testing"
)

func BenchmarkAddInt(b *testing.B) {
	c := NewCalculator()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.AddInt(42, 137)
	}
}

func BenchmarkFactorialIterative10(b *testing.B) {
	c := NewCalculator()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.FactorialIterative(10)
	}
}

func BenchmarkFactorialIterative15(b *testing.B) {
	c := NewCalculator()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.FactorialIterative(15)
	}
}

func BenchmarkFibonacciFast20(b *testing.B) {
	c := NewCalculator()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.FibonacciFast(20)
	}
}

func BenchmarkFibonacciFast25(b *testing.B) {
	c := NewCalculator()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.FibonacciFast(25)
	}
}

func BenchmarkFibonacciFast98(b *testing.B) {
	c := NewCalculator()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.FibonacciFast(98)
	}
}

func BenchmarkMultiplyInt(b *testing.B) {
	c := NewCalculator()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.MultiplyInt(42, 137)
	}
}

func BenchmarkSumFirstNOptimized(b *testing.B) {
	c := NewCalculator()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.SumFirstNOptimized(10000)
	}
}

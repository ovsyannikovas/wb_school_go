package handler

import (
	"encoding/json"
	"l4.5/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler() *Handler {
	calculator := service.NewCalculator()
	return NewHandler(calculator)
}

func BenchmarkAdd(b *testing.B) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/add?a=42&b=137", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Add(w, req)
	}
}

func BenchmarkMultiply(b *testing.B) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/multiply?a=42&b=137", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Multiply(w, req)
	}
}

func BenchmarkFibonacci20(b *testing.B) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/fibonacci?n=20", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Fibonacci(w, req)
	}
}

func BenchmarkFactorial10(b *testing.B) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/factorial?n=10", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Factorial(w, req)
	}
}

func BenchmarkBulkAdd10(b *testing.B) {
	h := newTestHandler()
	payload := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/bulk-add", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.BulkAdd(w, req)
	}
}

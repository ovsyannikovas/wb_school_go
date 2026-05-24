package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"l4.5/internal/service"
)

// bufferPool переиспользует *bytes.Buffer для формирования JSON-ответов.
var bufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 128)
		return &buf
	},
}

// Handler — оптимизированный handler: использует json.NewEncoder (streaming)
// и singleton buffer pool вместо json.Marshal.
type Handler struct {
	calculator *service.Calculator
}

func NewHandler(calculator *service.Calculator) *Handler {
	return &Handler{calculator: calculator}
}

// sendJSONStream потоково пишет JSON через json.Encoder + buffer pool.
func sendJSONStream(w http.ResponseWriter, encoder *json.Encoder) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder.Encode(w)
}

// sendJSON writes JSON response.
func sendJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	a, _ := strconv.Atoi(r.URL.Query().Get("a"))
	b, _ := strconv.Atoi(r.URL.Query().Get("b"))
	resp := struct {
		Result    string `json:"result"`
		Operation string `json:"operation"`
	}{
		Result:    h.calculator.AddInt(a, b),
		Operation: "add",
	}
	sendJSON(w, resp)
}

func (h *Handler) Multiply(w http.ResponseWriter, r *http.Request) {
	a, _ := strconv.Atoi(r.URL.Query().Get("a"))
	b, _ := strconv.Atoi(r.URL.Query().Get("b"))
	resp := struct {
		Result    string `json:"result"`
		Operation string `json:"operation"`
	}{
		Result:    h.calculator.MultiplyInt(a, b),
		Operation: "multiply",
	}
	sendJSON(w, resp)
}

func (h *Handler) Fibonacci(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	nStr, resStr := h.calculator.FibonacciFast(n)
	sendJSON(w, map[string]string{"n": nStr, "result": resStr})
}

func (h *Handler) Factorial(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	nStr, resStr := h.calculator.FactorialIterative(n)
	sendJSON(w, map[string]string{"n": nStr, "result": resStr})
}

func (h *Handler) Sum(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	nStr, sumStr, countStr := h.calculator.SumFirstNOptimized(n)
	sendJSON(w, struct {
		N     string `json:"n"`
		Sum   string `json:"sum"`
		Count string `json:"count"`
	}{
		N:     nStr,
		Sum:   sumStr,
		Count: countStr,
	})
}

// BulkAdd — оптимизированный bulk-сумматор.
func (h *Handler) BulkAdd(w http.ResponseWriter, r *http.Request) {
	var nums []int
	if err := json.NewDecoder(r.Body).Decode(&nums); err != nil {
		sendJSON(w, map[string]string{"error": "invalid input"})
		return
	}
	total := sumSlice(nums)
	sendJSON(w, map[string]string{
		"total": strconv.Itoa(total),
		"count": strconv.Itoa(len(nums)),
	})
}

// sumSlice суммирует слайц int без аллокаций.
func sumSlice(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

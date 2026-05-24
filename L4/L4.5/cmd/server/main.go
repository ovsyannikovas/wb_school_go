package main

import (
	"log"
	"net/http"
	"net/http/pprof"

	"l4.5/internal/handler"
	"l4.5/internal/service"
)

func main() {
	calculator := service.NewCalculator()
	h := handler.NewHandler(calculator)

	mux := http.NewServeMux()
	mux.HandleFunc("/add", h.Add)
	mux.HandleFunc("/multiply", h.Multiply)
	mux.HandleFunc("/fibonacci", h.Fibonacci)
	mux.HandleFunc("/factorial", h.Factorial)
	mux.HandleFunc("/bulk-add", h.BulkAdd)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/debug/pprof/allocs", pprof.Index)
	mux.HandleFunc("/debug/pprof/block", pprof.Index)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
	mux.HandleFunc("/debug/pprof/heap", pprof.Index)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Index)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)

	log.Println("Server starting at :8080")
	log.Println("Pprof: http://localhost:8080/debug/pprof/")
	log.Println("Test endpoints: /add, /multiply, /fibonacci, /factorial, /bulk-add")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

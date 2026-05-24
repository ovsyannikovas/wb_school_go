package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	addr := flag.String("addr", "http://localhost:8080", "server address")
	duration := flag.Int("duration", 5, "duration in seconds")
	concurrency := flag.Int("concurrency", 50, "number of concurrent clients")
	flag.Parse()

	fmt.Printf("Load test: %d seconds, %d concurrent workers\n", *duration, *concurrency)

	start := time.Now()
	var wg sync.WaitGroup
	var totalRequests int64
	var totalSuccess int64
	var totalErrors int64

	pprofURL := fmt.Sprintf("%s/debug/pprof/profile?seconds=%d", *addr, *duration+5)
	go func() {
		time.Sleep(500 * time.Millisecond)
		resp, err := http.Get(pprofURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to collect pprof: %v\n", err)
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		os.WriteFile("cpuprofile.prof", data, 0644)
		fmt.Printf("CPU Profile saved: cpuprofile.prof\n")
	}()

	// Heap profile
	go func() {
		time.Sleep(time.Duration(*duration) * time.Second)
		resp, err := http.Get(fmt.Sprintf("%s/debug/pprof/heap", *addr))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		os.WriteFile("heapprofile.prof", data, 0644)
		fmt.Println("Heap Profile saved: heapprofile.prof")
	}()

	// Allocations profile
	go func() {
		time.Sleep(time.Duration(*duration) * time.Second)
		resp, err := http.Get(fmt.Sprintf("%s/debug/pprof/allocs", *addr))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		os.WriteFile("allocprofile.prof", data, 0644)
		fmt.Println("Alloc Profile saved: allocprofile.prof")
	}()

	// Trace
	go func() {
		time.Sleep(time.Duration(*duration) * time.Second)
		resp, err := http.Get(fmt.Sprintf("%s/debug/pprof/trace?seconds=5", *addr))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		os.WriteFile("trace.prof", data, 0644)
		fmt.Println("Trace saved: trace.prof")
	}()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		endpoints := []string{
			"/add?a=42&b=137",
			"/multiply?a=42&b=137",
			"/fibonacci?n=20",
			"/factorial?n=10",
			"/bulk-add",
		}
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 10 * time.Second}
			for time.Since(start) < time.Duration(*duration)*time.Second {
				ep := endpoints[workerID%len(endpoints)]
				if ep == "/bulk-add" {
					ep = "/add?a=42&b=137"
				}
				req, _ := http.NewRequest(http.MethodGet, *addr+ep, nil)
				resp, err := client.Do(req)
				if err != nil {
					atomicAdd(&totalErrors, 1)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode == 200 {
					atomicAdd(&totalSuccess, 1)
				} else {
					atomicAdd(&totalErrors, 1)
				}
				atomicAdd(&totalRequests, 1)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Printf("Total requests: %d\n", totalRequests)
	fmt.Printf("Successes: %d\n", totalSuccess)
	fmt.Printf("Errors: %d\n", totalErrors)
	if elapsed > 0 {
		fmt.Printf("RPS: %.1f\n", float64(totalRequests)/elapsed.Seconds())
	}
	fmt.Println("Pprof data generated. Use: go tool pprof -http=:8081 cpuprofile.prof")
}

func atomicAdd(ptr *int64, val int64) {
	*ptr += val
}

package grepper

import (
	"L4_2/grepper/types"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
)

func Run(file, pattern, workersList string, concurrency int) {
	nodeAddrs := strings.Split(workersList, ",")

	if len(nodeAddrs) == 0 || (len(nodeAddrs) == 1 && nodeAddrs[0] == "") {
		runLocal(file, pattern)
		return
	}

	fmt.Printf("[Master] Distributed mode with %d workers\n", len(nodeAddrs))

	chunks, err := SplitFile(file, len(nodeAddrs))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	taskID := uuid.New().String()
	fmt.Printf("[Master] Task ID: %s\n", taskID)

	distributor := NewDistributor(len(nodeAddrs), taskID)
	if err := distributor.Start("9090"); err != nil {
		fmt.Printf("Failed to start distributor: %v\n", err)
		return
	}
	defer distributor.Stop()

	var wg sync.WaitGroup
	for i, addr := range nodeAddrs {
		wg.Add(1)
		go func(addr string, chunk Chunk) {
			defer wg.Done()
			sendTaskAndWaitResult(addr, chunk, pattern, taskID)
		}(addr, chunks[i])
	}

	wg.Wait()

	results := distributor.WaitForQuorum()

	if len(results) == 0 {
		fmt.Printf("[Master] No results received from workers\n")
		return
	}

	seen := make(map[string]bool)
	fmt.Printf("\n=== RESULTS ===\n")
	for workerID, lines := range results {
		fmt.Printf("[Worker %s] %d lines\n", workerID, len(lines))
		for _, line := range lines {
			if !seen[line] {
				fmt.Print(line)
				seen[line] = true
			}
		}
	}
	fmt.Printf("\n=== Total unique lines: %d ===\n", len(seen))
}

// Новая функция: отправляет задачу и ЖДЁТ результат, затем отправляет в distributor
func sendTaskAndWaitResult(addr string, chunk Chunk, pattern, taskID string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("[Master] Failed to connect to %s: %v\n", addr, err)
		return
	}
	defer conn.Close()

	task := types.Task{
		TaskID: taskID,
		Path:   chunk.FileName,
		Start:  chunk.Start,
		End:    chunk.End,
		Regex:  pattern,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(task); err != nil {
		fmt.Printf("[Master] Failed to send task to %s: %v\n", addr, err)
		return
	}

	fmt.Printf("[Master] Task sent to %s (bytes %d-%d)\n", addr, chunk.Start, chunk.End)

	var result types.TaskResult
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&result); err != nil {
		fmt.Printf("[Master] Failed to receive result from %s: %v\n", addr, err)
		return
	}

	fmt.Printf("[Master] Received result from %s: %d lines\n", addr, len(result.Lines))

	sendResultToDistributor(result)
}

func sendResultToDistributor(result types.TaskResult) {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		fmt.Printf("[Master] Failed to connect to distributor: %v\n", err)
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(result); err != nil {
		fmt.Printf("[Master] Failed to send result to distributor: %v\n", err)
		return
	}
}

func runLocal(file, pattern string) {
	fmt.Printf("[Master] Local fallback mode\n")
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, pattern) {
			fmt.Println(line)
		}
	}
}

package main

import (
	"L4_2/grepper/types"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"sync"
	"time"
)

type Worker struct {
	workerID        string
	distributorAddr string
	listenPort      string
	concurrency     int
}

func NewWorker(workerID, distributorAddr, listenPort string, concurrency int) *Worker {
	return &Worker{
		workerID:        workerID,
		distributorAddr: distributorAddr,
		listenPort:      listenPort,
		concurrency:     concurrency,
	}
}

func (w *Worker) Start() {
	fmt.Printf("[Worker %s] Starting on port %s\n", w.workerID, w.listenPort)

	listener, err := net.Listen("tcp", ":"+w.listenPort)
	if err != nil {
		fmt.Printf("[Worker %s] Failed to listen: %v\n", w.workerID, err)
		return
	}
	defer listener.Close()

	go w.registerWithDistributor()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go w.handleConnection(conn)
	}
}

func (w *Worker) registerWithDistributor() {
	for {
		conn, err := net.Dial("tcp", w.distributorAddr)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		encoder := json.NewEncoder(conn)

		// Регистрируем воркера
		registration := map[string]interface{}{
			"type":     "register",
			"workerID": w.workerID,
			"address":  w.listenPort,
		}
		encoder.Encode(registration)
		conn.Close()

		fmt.Printf("[Worker %s] Registered with distributor at %s\n", w.workerID, w.distributorAddr)
		return
	}
}

func (w *Worker) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var task types.Task
	if err := decoder.Decode(&task); err != nil {
		return
	}

	fmt.Printf("[Worker %s] Received task: %+v\n", w.workerID, task)
	fmt.Printf("[Worker %s] Path=%s, Pattern=%s, Start=%d, End=%d, TaskID=%s\n",
		w.workerID, task.Path, task.Regex, task.Start, task.End, task.TaskID)

	fmt.Printf("[Worker %s] Processing task: file=%s, pattern=%s\n",
		w.workerID, task.Path, task.Regex)

	result := w.processTask(task)

	encoder := json.NewEncoder(conn)
	encoder.Encode(result)

	fmt.Printf("[Worker %s] Completed, found %d lines\n",
		w.workerID, len(result.Lines))
}

func (w *Worker) processTask(task types.Task) types.TaskResult {
	f, err := os.Open(task.Path)
	if err != nil {
		return types.TaskResult{
			TaskID:   task.TaskID,
			WorkerID: w.workerID,
			Success:  false,
			Error:    err.Error(),
		}
	}
	defer f.Close()

	if task.Start > 0 {
		f.Seek(task.Start, 0)
	}

	reader := bufio.NewReader(f)
	re := regexp.MustCompile(task.Regex)

	jobs := make(chan string, 100)
	results := make(chan string, 100)
	var wg sync.WaitGroup

	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range jobs {
				if re.MatchString(line) {
					results <- line
				}
			}
		}()
	}

	go func() {
		var pos = task.Start
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if pos > task.End && task.End > 0 {
				break
			}
			jobs <- line
			pos += int64(len(line))
		}
		close(jobs)
	}()

	var matched []string
	doneCollecting := make(chan bool)
	go func() {
		for line := range results {
			matched = append(matched, line)
		}
		doneCollecting <- true
	}()

	wg.Wait()
	close(results)
	<-doneCollecting

	return types.TaskResult{
		TaskID:   task.TaskID,
		WorkerID: w.workerID,
		Lines:    matched,
		Success:  true,
	}
}

func main() {
	var (
		nodeID          string
		distributorAddr string
		port            string
		concurrency     int
	)

	flag.StringVar(&nodeID, "id", "", "Worker ID")
	flag.StringVar(&distributorAddr, "distributorAddr", "localhost:9090", "Distributor address")
	flag.StringVar(&port, "port", "8081", "Worker port for listening")
	flag.IntVar(&concurrency, "c", 4, "Concurrency level")
	flag.Parse()

	if nodeID == "" {
		hostname, _ := os.Hostname()
		nodeID = fmt.Sprintf("worker-%s", hostname)
	}

	worker := NewWorker(nodeID, distributorAddr, port, concurrency)
	worker.Start()
}

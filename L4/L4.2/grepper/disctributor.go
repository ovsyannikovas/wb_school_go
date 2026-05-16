package grepper

import (
	"L4_2/grepper/types"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type Distributor struct {
	quorum        int
	workersNum    int
	mu            sync.RWMutex
	results       map[string][]string
	workerResults map[string]bool
	done          chan bool
	taskID        string
	listener      net.Listener
}

func NewDistributor(workersNum int, taskID string) *Distributor {
	return &Distributor{
		quorum:        workersNum/2 + 1,
		workersNum:    workersNum,
		results:       make(map[string][]string),
		workerResults: make(map[string]bool),
		done:          make(chan bool),
		taskID:        taskID,
	}
}

func (d *Distributor) Start(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	d.listener = listener

	go d.acceptConnections()
	fmt.Printf("[Distributor] Listening on port %s\n", port)

	return nil
}

func (d *Distributor) acceptConnections() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			return
		}
		go d.handleConnection(conn)
	}
}

func (d *Distributor) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)

	var result types.TaskResult
	if err := decoder.Decode(&result); err != nil {
		fmt.Printf("[Distributor] Error decoding result: %v\n", err)
		return
	}

	if result.TaskID != d.taskID {
		fmt.Printf("[Distributor] Wrong task ID: %s (expected %s)\n", result.TaskID, d.taskID)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.workerResults[result.WorkerID]; !exists {
		d.workerResults[result.WorkerID] = true
		if result.Success {
			d.results[result.WorkerID] = result.Lines
			fmt.Printf("[Distributor] Got results from %s (%d lines)\n",
				result.WorkerID, len(result.Lines))
		} else {
			fmt.Printf("[Distributor] Worker %s failed: %s\n", result.WorkerID, result.Error)
		}

		if len(d.workerResults) >= d.quorum {
			select {
			case d.done <- true:
			default:
			}
		}
	}
}

func (d *Distributor) WaitForQuorum() map[string][]string {
	select {
	case <-d.done:
	case <-time.After(30 * time.Second):
		fmt.Printf("[Distributor] Timeout waiting for quorum\n")
	}

	fmt.Printf("[Distributor] Quorum reached (%d/%d workers)\n",
		len(d.workerResults), d.workersNum)

	d.mu.RLock()
	defer d.mu.RUnlock()

	results := make(map[string][]string)
	for k, v := range d.results {
		results[k] = v
	}
	return results
}

func (d *Distributor) Stop() {
	if d.listener != nil {
		d.listener.Close()
	}
}

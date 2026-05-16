package types

// Task — задача, которую мастер отправляет воркеру (кусок, который нужно обработать)
type Task struct {
	Path   string `json:"path"`
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Regex  string `json:"regex"`
	TaskID string `json:"task_id"`
}

// TaskResult — результат выполнения задачи одним воркером
type TaskResult struct {
	TaskID   string   `json:"task_id"`
	WorkerID string   `json:"worker_id"`
	Lines    []string `json:"lines"`
	Success  bool     `json:"success"`
	Error    string   `json:"error,omitempty"`
}

type ReadinessProbe struct {
	WorkerID string
}

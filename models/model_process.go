// model_process.go - Модели процессов для Venera
//
// Этот модуль содержит определения структур для управления процессами
// сбора и обработки данных в системе Venera.
//
// Основные структуры:
// - Process - процесс обработки
// - ProcessManager - менеджер процессов
// - ProcessConfig - конфигурация процесса
// - ProcessStatus - статус процесса

package models

// Process - процесс обработки
type Process struct {
	ID                 string         `json:"id"`
	Config             ProcessConfig  `json:"config"`
	Status             ProcessStatus  `json:"status"`
	Metrics            ProcessMetrics `json:"metrics"`
	Queue              *ProcessQueue  `json:"queue"`
	StartedAt          int64          `json:"started_at"`
	LastActivity       int64          `json:"last_activity"`
	Stopped            bool           `json:"stopped"`
	ErrorMessage       string         `json:"error_message"`
}

// ProcessConfig - конфигурация процесса
type ProcessConfig struct {
	ID                 string `json:"id"`
	Type               string `json:"type"` // "network", "folder", "file"
	Name               string `json:"name"`
	IP                 string `json:"ip"`
	Port               int    `json:"port"`
	Path               string `json:"path"`
	ScanSubfolders     bool   `json:"scan_subfolders"`
	MonitorNewFiles    bool   `json:"monitor_new_files"`
}

// ProcessStatus - статус процесса
type ProcessStatus struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"` // "running", "stopped", "error"
	InputRate    float64 `json:"input_rate"`
	ProcessedCount int64  `json:"processed_count"`
	StartTime    int64  `json:"start_time"`
	LastActivity int64  `json:"last_activity"`
}

// ProcessMetrics - метрики процесса
type ProcessMetrics struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	ProcessID       int     `json:"process_id"`
	InputRate       float64 `json:"input_rate"`
	MemoryUsage     int64   `json:"memory_usage"`
	CPUUsage        float64 `json:"cpu_usage"`
	ProcessedCount  int64   `json:"processed_count"`
	FilteredCount   int64   `json:"filtered_count"`
	QueueSize       int64   `json:"queue_size"`
	StartTime       int64   `json:"start_time"`
	LastActivity    int64   `json:"last_activity"`
}

// ProcessQueue - очередь процесса
type ProcessQueue struct {
	ListKey       string `json:"list_key"`
	SortedSetKey  string `json:"sorted_set_key"`
	CurrentSize   int64  `json:"current_size"`
	MaxSize       int64  `json:"max_size"`
	BatchSize     int    `json:"batch_size"`
	Timeout       int    `json:"timeout"`
}

// ProcessManager - менеджер процессов
type ProcessManager struct {
	Processes    map[string]*Process `json:"processes"`
	MaxProcesses int                 `json:"max_processes"`
}

// NewProcessManager - создание нового менеджера процессов
func NewProcessManager(maxProcesses int) *ProcessManager {
	return &ProcessManager{
		Processes:    make(map[string]*Process),
		MaxProcesses: maxProcesses,
	}
}

// AddProcess - добавление процесса
func (m *ProcessManager) AddProcess(process *Process) error {
	if len(m.Processes) >= m.MaxProcesses {
		return fmt.Errorf("достигнуто максимальное количество процессов: %d", m.MaxProcesses)
	}

	if _, exists := m.Processes[process.ID]; exists {
		return fmt.Errorf("процесс с ID %s уже существует", process.ID)
	}

	m.Processes[process.ID] = process
	return nil
}

// RemoveProcess - удаление процесса
func (m *ProcessManager) RemoveProcess(id string) error {
	if _, exists := m.Processes[id]; !exists {
		return fmt.Errorf("процесс с ID %s не найден", id)
	}

	delete(m.Processes, id)
	return nil
}

// GetProcess - получение процесса по ID
func (m *ProcessManager) GetProcess(id string) (*Process, error) {
	process, exists := m.Processes[id]
	if !exists {
		return nil, fmt.Errorf("процесс с ID %s не найден", id)
	}
	return process, nil
}

// GetAllProcesses - получение всех процессов
func (m *ProcessManager) GetAllProcesses() []*Process {
	processes := make([]*Process, 0, len(m.Processes))
	for _, process := range m.Processes {
		processes = append(processes, process)
	}
	return processes
}

// GetRunningProcesses - получение запущенных процессов
func (m *ProcessManager) GetRunningProcesses() []*Process {
	var running []*Process
	for _, process := range m.Processes {
		if process.Status.Status == "running" {
			running = append(running, process)
		}
	}
	return running
}

// GetStoppedProcesses - получение остановленных процессов
func (m *ProcessManager) GetStoppedProcesses() []*Process {
	var stopped []*Process
	for _, process := range m.Processes {
		if process.Status.Status == "stopped" {
			stopped = append(stopped, process)
		}
	}
	return stopped
}

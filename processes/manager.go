// manager.go - Менеджер процессов обработки данных
//
// Этот модуль обеспечивает управление жизненным циклом процессов обработки данных,
// включая запуск, остановку и удаление процессов.
//
// Основные функции:
// - Управление жизненным циклом процессов
// - Запуск процессов обработки
// - Остановка процессов
// - Удаление процессов
// - Проверка максимального количества процессов (20)
// - Отслеживание состояния процессов
// - Каналы для коммуникации между процессами
//
// Использование:
// import "venera/processes"
// manager := NewProcessManager()
// manager.AddProcess(process)
// manager.StartProcess(processID)
// manager.StopProcess(processID)

package processes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"venera/config"
	"venera/data"
	"venera/models"
	"venera/logging"
)

// ProcessManager - менеджер процессов
type ProcessManager struct {
	processes    map[string]*ProcessWrapper
	maxProcesses int
	config       *config.Config
	dragonflyDB  *data.DragonflyDB
	postgresDB   *data.PostgreSQL
	dataFilter   *data.DataFilter
	controls     map[string]string
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	log          *logrus.Logger
}

// ProcessWrapper - обертка процесса
type ProcessWrapper struct {
	ID           string
	Config       config.ProcessConfig
	Tshark       *Tshark
	DataSelector *data.DataSelector
	Queue        *ProcessQueue
	Cancel       context.CancelFunc
	StartTime    int64
	Stopped      bool
}

// ProcessQueue - очередь процесса
type ProcessQueue struct {
	ListKey      string
	SortedSetKey string
	BatchSize    int
	Timeout      time.Duration
	Processor    *DataProcessor
}

// DataProcessor - обработчик данных
type DataProcessor struct {
	ListKey        string
	SortedSetKey   string
	DragonflyDB    *data.DragonflyDB
	PostgreSQL     *data.PostgreSQL
	DataFilter     *data.DataFilter
	Controls       map[string]string
	BatchSize      int
	Timeout        time.Duration
}

// NewProcessManager - создание нового менеджера процессов
func NewProcessManager(cfg *config.Config, dragonflyDB *data.DragonflyDB, postgresDB *data.PostgreSQL, dataFilter *data.DataFilter, controls map[string]string) *ProcessManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ProcessManager{
		processes:    make(map[string]*ProcessWrapper),
		maxProcesses: cfg.Generic.MaxProcesses,
		config:       cfg,
		dragonflyDB:  dragonflyDB,
		postgresDB:   postgresDB,
		dataFilter:   dataFilter,
		controls:     controls,
		ctx:          ctx,
		cancel:       cancel,
		log:          logrus.WithField("module", "process_manager"),
	}
}

// AddProcess - добавление процесса
func (m *ProcessManager) AddProcess(id string, cfg config.ProcessConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Проверка максимального количества процессов
	if len(m.processes) >= m.maxProcesses {
		return fmt.Errorf("достигнуто максимальное количество процессов: %d", m.maxProcesses)
	}

	// Проверка существования процесса
	if _, exists := m.processes[id]; exists {
		return fmt.Errorf("процесс с ID %s уже существует", id)
	}

	// Создание очереди
	queue := &ProcessQueue{
		ListKey:      fmt.Sprintf("queue:%s", id),
		SortedSetKey: fmt.Sprintf("sorted:%s", id),
		BatchSize:    m.config.DragonflyDB.BatchSize,
		Timeout:      time.Duration(m.config.DragonflyDB.Timeout) * time.Second,
	}

	// Создание обработчика данных
	processor := &DataProcessor{
		ListKey:      queue.ListKey,
		SortedSetKey: queue.SortedSetKey,
		DragonflyDB:  m.dragonflyDB,
		PostgreSQL:   m.postgresDB,
		DataFilter:   m.dataFilter,
		Controls:     m.controls,
		BatchSize:    m.config.DragonflyDB.BatchSize,
		Timeout:      queue.Timeout,
	}

	queue.Processor = processor

	// Создание Tshark
	tshark := NewTshark(TsharkConfig{
		Path:      m.config.Paths.TsharkPath,
		Type:      cfg.Type,
		IP:        cfg.IP,
		Port:      cfg.Port,
		PathInput: cfg.Path,
	})

	// Создание обертки процесса
	wrapper := &ProcessWrapper{
		ID:           id,
		Config:       cfg,
		Tshark:       tshark,
		DataSelector: data.NewDataSelector(m.dragonflyDB, m.postgresDB, m.dataFilter, m.controls),
		Queue:        queue,
		StartTime:    time.Now().Unix(),
		Stopped:      false,
	}

	m.processes[id] = wrapper
	m.log.Infof("Процесс %s добавлен", id)

	return nil
}

// StartProcess - запуск процесса
func (m *ProcessManager) StartProcess(id string) error {
	m.mu.Lock()
	wrapper, exists := m.processes[id]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("процесс с ID %s не найден", id)
	}

	if !wrapper.Stopped {
		return fmt.Errorf("процесс %s уже запущен", id)
	}

	// Запуск Tshark
	if err := wrapper.Tshark.Start(); err != nil {
		return fmt.Errorf("ошибка запуска Tshark: %w", err)
	}

	// Запуск обработчика данных
	wrapper.Stopped = false
	m.log.Infof("Процесс %s запущен", id)

	return nil
}

// StopProcess - остановка процесса
func (m *ProcessManager) StopProcess(id string) error {
	m.mu.Lock()
	wrapper, exists := m.processes[id]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("процесс с ID %s не найден", id)
	}

	if wrapper.Stopped {
		return fmt.Errorf("процесс %s уже остановлен", id)
	}

	// Остановка Tshark
	if err := wrapper.Tshark.Stop(); err != nil {
		m.log.Warnf("Ошибка остановки Tshark для процесса %s: %v", id, err)
	}

	// Остановка обработчика данных
	wrapper.Stopped = true
	m.log.Infof("Процесс %s остановлен", id)

	return nil
}

// DeleteProcess - удаление процесса
func (m *ProcessManager) DeleteProcess(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.processes[id]; !exists {
		return fmt.Errorf("процесс с ID %s не найден", id)
	}

	// Остановка процесса, если он запущен
	if !m.processes[id].Stopped {
		if err := m.StopProcess(id); err != nil {
			m.log.Warnf("Ошибка остановки процесса %s при удалении: %v", id, err)
		}
	}

	// Удаление очереди из DragonflyDB
	listKey := fmt.Sprintf("queue:%s", id)
	sortedSetKey := fmt.Sprintf("sorted:%s", id)

	m.dragonflyDB.Delete(listKey)
	m.dragonflyDB.Delete(sortedSetKey)

	// Удаление из карты
	delete(m.processes, id)
	m.log.Infof("Процесс %s удален", id)

	return nil
}

// StartAll - запуск всех процессов
func (m *ProcessManager) StartAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, wrapper := range m.processes {
		if wrapper.Stopped {
			if err := m.StartProcess(id); err != nil {
				m.log.Errorf("Ошибка запуска процесса %s: %v", id, err)
			}
		}
	}

	return nil
}

// StopAll - остановка всех процессов
func (m *ProcessManager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, wrapper := range m.processes {
		if !wrapper.Stopped {
			if err := m.StopProcess(id); err != nil {
				m.log.Errorf("Ошибка остановки процесса %s: %v", id, err)
			}
		}
	}

	return nil
}

// GetProcess - получение процесса по ID
func (m *ProcessManager) GetProcess(id string) (*ProcessWrapper, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wrapper, exists := m.processes[id]
	if !exists {
		return nil, fmt.Errorf("процесс с ID %s не найден", id)
	}

	return wrapper, nil
}

// GetAllProcesses - получение всех процессов
func (m *ProcessManager) GetAllProcesses() []*ProcessWrapper {
	m.mu.Lock()
	defer m.mu.Unlock()

	processes := make([]*ProcessWrapper, 0, len(m.processes))
	for _, wrapper := range m.processes {
		processes = append(processes, wrapper)
	}

	return processes
}

// GetRunningProcesses - получение запущенных процессов
func (m *ProcessManager) GetRunningProcesses() []*ProcessWrapper {
	m.mu.Lock()
	defer m.mu.Unlock()

	var running []*ProcessWrapper
	for _, wrapper := range m.processes {
		if !wrapper.Stopped {
			running = append(running, wrapper)
		}
	}

	return running
}

// GetStoppedProcesses - получение остановленных процессов
func (m *ProcessManager) GetStoppedProcesses() []*ProcessWrapper {
	m.mu.Lock()
	defer m.mu.Unlock()

	var stopped []*ProcessWrapper
	for _, wrapper := range m.processes {
		if wrapper.Stopped {
			stopped = append(stopped, wrapper)
		}
	}

	return stopped
}

// GetProcessCount - получение количества процессов
func (m *ProcessManager) GetProcessCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.processes)
}

// GetProcessMetrics - получение метрик всех процессов
func (m *ProcessManager) GetProcessMetrics() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := make(map[string]interface{})
	for id, wrapper := range m.processes {
		metrics[id] = map[string]interface{}{
			"status":      "stopped",
			"startTime":   wrapper.StartTime,
			"config":      wrapper.Config,
			"queueList":   wrapper.Queue.ListKey,
			"queueSorted": wrapper.Queue.SortedSetKey,
		}
		if !wrapper.Stopped {
			metrics[id].(map[string]interface{})["status"] = "running"
		}
	}
	return metrics
}

// GetProcessesList - получение списка конфигураций процессов
func (m *ProcessManager) GetProcessesList() []config.ProcessConfig {
	m.mu.Lock()
	defer m.mu.Unlock()

	configs := make([]config.ProcessConfig, 0, len(m.processes))
	for _, wrapper := range m.processes {
		configs = append(configs, wrapper.Config)
	}
	return configs
}

// GetProcessStatus - получение статуса процесса по ID
func (m *ProcessManager) GetProcessStatus(processID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	wrapper, exists := m.processes[processID]
	if !exists {
		return "error"
	}
	if wrapper.Stopped {
		return "stopped"
	}
	return "running"
}

// Shutdown - корректное завершение работы менеджера
func (m *ProcessManager) Shutdown() error {
	m.log.Info("Остановка всех процессов")

	// Остановка всех процессов
	m.StopAll()

	// Отмена контекста
	if m.cancel != nil {
		m.cancel()
	}

	m.log.Info("Менеджер процессов остановлен")
	return nil
}

// monitor.go - Мониторинг процессов для системы Venera
//
// Этот модуль обеспечивает мониторинг процессов обработки данных
// для приложения Venera, включая скорость входного потока,
// потребление RAM и загрузку CPU.
//
// Основные функции:
// - Мониторинг скорости входного потока данных
// - Мониторинг потребления RAM процессами
// - Мониторинг загрузки CPU процессами
// - Мониторинг количества обработанных записей
// - Планировщик мониторинга
//
// Использование:
// import "venera/metrics"
// monitor := metrics.NewMonitor(cfg, logger)
// monitor.Start()
// monitor.Stop()

package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"venera/config"
)

// ProcessMonitor - монитор процесса
type ProcessMonitor struct {
	ID                string
	Name              string
	StartTime         time.Time
	BytesProcessed    int64
	PacketsCount      int64
	LastBytesProcessed int64
	LastTime          time.Time
	CPUUsage          float64
	MemoryUsage       uint64
	Status            string
}

// Monitor - структура монитора процессов
type Monitor struct {
	cfg        *config.Config
	logger     *logrus.Logger
	processes  map[string]*ProcessMonitor
	mu         sync.Mutex
	ticker     *time.Ticker
	stopChan   chan struct{}
}

// NewMonitor - создание нового монитора процессов
func NewMonitor(cfg *config.Config, logger *logrus.Logger) *Monitor {
	return &Monitor{
		cfg:      cfg,
		logger:   logger,
		processes: make(map[string]*ProcessMonitor),
		stopChan: make(chan struct{}),
	}
}

// Start - запуск мониторинга
func (m *Monitor) Start() {
	m.logger.Info("Запуск мониторинга процессов")

	// Запуск планировщика
	m.ticker = time.NewTicker(time.Duration(m.cfg.Generic.ProcessingInterval) * time.Second)

	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.collectProcessMetrics()
			case <-m.stopChan:
				m.ticker.Stop()
				m.logger.Info("Мониторинг процессов остановлен")
				return
			}
		}
	}()
}

// Stop - остановка мониторинга
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// RegisterProcess - регистрация процесса для мониторинга
func (m *Monitor) RegisterProcess(id, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processes[id] = &ProcessMonitor{
		ID:        id,
		Name:      name,
		StartTime: time.Now(),
		Status:    "running",
	}
}

// UnregisterProcess - удаление процесса из мониторинга
func (m *Monitor) UnregisterProcess(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.processes, id)
}

// UpdateBytesProcessed - обновление количества обработанных байт
func (m *Monitor) UpdateBytesProcessed(id string, bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if process, ok := m.processes[id]; ok {
		process.BytesProcessed += bytes
	}
}

// UpdatePacketsCount - обновление количества обработанных пакетов
func (m *Monitor) UpdatePacketsCount(id string, count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if process, ok := m.processes[id]; ok {
		process.PacketsCount += count
	}
}

// UpdateStatus - обновление статуса процесса
func (m *Monitor) UpdateStatus(id, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if process, ok := m.processes[id]; ok {
		process.Status = status
	}
}

// collectProcessMetrics - сбор метрик процессов
func (m *Monitor) collectProcessMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Debug("Сбор метрик процессов")

	for id, process := range m.processes {
		// Получение использования памяти
		handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ,
			false, uint32(0)) // Получить реальный PID процесса
		if err == nil {
			defer windows.CloseHandle(handle)

			var memCounters windows.ProcessMemoryCounters
			err = windows.GetProcessMemoryInfo(handle, &memCounters, uint32(unsafe.Sizeof(memCounters)))
			if err == nil {
				process.MemoryUsage = memCounters.WorkingSetSize
			}
		}

		// Вычисление скорости входного потока
		now := time.Now()
		elapsed := now.Sub(process.LastTime).Seconds()
		if elapsed > 0 {
			bytesDiff := process.BytesProcessed - process.LastBytesProcessed
			process.CPUUsage = float64(bytesDiff) / elapsed / 1024 // KB/s
		}

		process.LastTime = now
		process.LastBytesProcessed = process.BytesProcessed
	}
}

// GetProcessMetrics - получение метрик процесса
func (m *Monitor) GetProcessMetrics(id string) (*ProcessMonitor, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, ok := m.processes[id]
	if !ok {
		return nil, fmt.Errorf("процесс с ID %s не найден", id)
	}

	return process, nil
}

// GetAllProcessMetrics - получение всех метрик процессов
func (m *Monitor) GetAllProcessMetrics() map[string]*ProcessMonitor {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]*ProcessMonitor)
	for id, process := range m.processes {
		result[id] = process
	}

	return result
}

// GetInputRate - получение скорости входного потока
func (m *Monitor) GetInputRate(id string) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, ok := m.processes[id]
	if !ok {
		return 0, fmt.Errorf("процесс с ID %s не найден", id)
	}

	return process.CPUUsage, nil
}

// GetMemoryUsage - получение потребления памяти процессом
func (m *Monitor) GetMemoryUsage(id string) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, ok := m.processes[id]
	if !ok {
		return 0, fmt.Errorf("процесс с ID %s не найден", id)
	}

	return process.MemoryUsage, nil
}

// GetPacketsCount - получение количества обработанных пакетов
func (m *Monitor) GetPacketsCount(id string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, ok := m.processes[id]
	if !ok {
		return 0, fmt.Errorf("процесс с ID %s не найден", id)
	}

	return process.PacketsCount, nil
}

// GetBytesProcessed - получение количества обработанных байт
func (m *Monitor) GetBytesProcessed(id string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, ok := m.processes[id]
	if !ok {
		return 0, fmt.Errorf("процесс с ID %s не найден", id)
	}

	return process.BytesProcessed, nil
}

// GetProcessStatus - получение статуса процесса
func (m *Monitor) GetProcessStatus(id string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, ok := m.processes[id]
	if !ok {
		return "", fmt.Errorf("процесс с ID %s не найден", id)
	}

	return process.Status, nil
}

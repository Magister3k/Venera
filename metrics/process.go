// process.go - Метрики процессов для системы Venera
//
// Этот модуль обеспечивает сбор метрик процессов для приложения Venera,
// включая скорость обработки JSON, количество отобранных пар ключ-значение
// и количество уникальных пар ключ-значение в PostgreSQL.
//
// Основные функции:
// - Метрики парсинга JSON (общее количество пар ключ-значение)
// - Метрики помещения в list DragonflyDB (количество отобранных пар)
// - Метрики добавления в PostgreSQL (количество уникальных пар)
// - Метрики процессов обработки данных
//
// Использование:
// import "venera/metrics"
// processMetrics := NewProcessMetrics(logger)
// processMetrics.IncrementJSONPairs(100)
// processMetrics.IncrementSelectedPairs(80)
// processMetrics.IncrementPostgresPairs(75)

package metrics

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ProcessMetrics - метрики процесса обработки
type ProcessMetrics struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	StartTime              time.Time `json:"start_time"`
	TotalJSONPairs         int64 `json:"total_json_pairs"`
	SelectedPairs          int64 `json:"selected_pairs"`
	PostgresPairs          int64 `json:"postgres_pairs"`
	FilteringTime          time.Duration `json:"filtering_time"`
	ParsingTime            time.Duration `json:"parsing_time"`
	InsertionTime          time.Duration `json:"insertion_time"`
	CurrentQueueSize       int64 `json:"current_queue_size"`
	MaxQueueSize           int64 `json:"max_queue_size"`
	ProcessingRate         float64 `json:"processing_rate"`
	PacketRate             float64 `json:"packet_rate"`
}

// ProcessMetricsCollector - сборщик метрик процессов
type ProcessMetricsCollector struct {
	metrics    map[string]*ProcessMetrics
	mu         sync.Mutex
	logger     *logrus.Logger
}

// NewProcessMetricsCollector - создание нового сборщика метрик процессов
func NewProcessMetricsCollector(logger *logrus.Logger) *ProcessMetricsCollector {
	return &ProcessMetricsCollector{
		metrics: make(map[string]*ProcessMetrics),
		logger:  logger,
	}
}

// StartProcess - запуск сбора метрик для процесса
func (c *ProcessMetricsCollector) StartProcess(id, name string) *ProcessMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics := &ProcessMetrics{
		ID:          id,
		Name:        name,
		StartTime:   time.Now(),
		MaxQueueSize: 1000, // Установить из конфигурации
	}

	c.metrics[id] = metrics

	c.logger.WithFields(logrus.Fields{
		"process_id":   id,
		"process_name": name,
	}).Info("Запуск сбора метрик процесса")

	return metrics
}

// StopProcess - остановка сбора метрик для процесса
func (c *ProcessMetricsCollector) StopProcess(id string) *ProcessMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return nil
	}

	elapsed := time.Since(metrics.StartTime).Seconds()
	if elapsed > 0 {
		metrics.ProcessingRate = float64(metrics.PostgresPairs) / elapsed
		metrics.PacketRate = float64(metrics.TotalJSONPairs) / elapsed
	}

	delete(c.metrics, id)

	c.logger.WithFields(logrus.Fields{
		"process_id": id,
		"total_json_pairs": metrics.TotalJSONPairs,
		"selected_pairs":   metrics.SelectedPairs,
		"postgres_pairs":   metrics.PostgresPairs,
		"processing_rate":  metrics.ProcessingRate,
	}).Info("Остановка сбора метрик процес��а")

	return metrics
}

// IncrementJSONPairs - увеличение количества пар ключ-значение из JSON
func (c *ProcessMetricsCollector) IncrementJSONPairs(id string, count int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metrics, ok := c.metrics[id]; ok {
		metrics.TotalJSONPairs += count
	}
}

// IncrementSelectedPairs - увеличение количества отобранных пар ключ-значение
func (c *ProcessMetricsCollector) IncrementSelectedPairs(id string, count int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metrics, ok := c.metrics[id]; ok {
		metrics.SelectedPairs += count
	}
}

// IncrementPostgresPairs - увеличение количества уникальных пар ключ-значение в PostgreSQL
func (c *ProcessMetricsCollector) IncrementPostgresPairs(id string, count int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metrics, ok := c.metrics[id]; ok {
		metrics.PostgresPairs += count
	}
}

// UpdateQueueSize - обновление размера очереди
func (c *ProcessMetricsCollector) UpdateQueueSize(id string, size int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metrics, ok := c.metrics[id]; ok {
		metrics.CurrentQueueSize = size
	}
}

// UpdateFilteringTime - обновление времени фильтрации
func (c *ProcessMetricsCollector) UpdateFilteringTime(id string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metrics, ok := c.metrics[id]; ok {
		metrics.FilteringTime += duration
	}
}

// UpdateParsingTime - обновление времени парсинга
func (c *ProcessMetricsCollector) UpdateParsingTime(id string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metrics, ok := c.metrics[id]; ok {
		metrics.ParsingTime += duration
	}
}

// UpdateInsertionTime - обновление времени вставки в PostgreSQL
func (c *ProcessMetricsCollector) UpdateInsertionTime(id string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if metrics, ok := c.metrics[id]; ok {
		metrics.InsertionTime += duration
	}
}

// GetProcessMetrics - получение метрик процесса
func (c *ProcessMetricsCollector) GetProcessMetrics(id string) (*ProcessMetrics, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return nil, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics, nil
}

// GetAllProcessMetrics - получение всех метрик процессов
func (c *ProcessMetricsCollector) GetAllProcessMetrics() map[string]*ProcessMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make(map[string]*ProcessMetrics)
	for id, metrics := range c.metrics {
		result[id] = metrics
	}

	return result
}

// GetTotalJSONPairs - получение общего количества пар ключ-значение из JSON
func (c *ProcessMetricsCollector) GetTotalJSONPairs(id string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return 0, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics.TotalJSONPairs, nil
}

// GetSelectedPairs - получение количества отобранных пар ключ-значение
func (c *ProcessMetricsCollector) GetSelectedPairs(id string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return 0, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics.SelectedPairs, nil
}

// GetPostgresPairs - получение количества уникальных пар ключ-значение в PostgreSQL
func (c *ProcessMetricsCollector) GetPostgresPairs(id string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return 0, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics.PostgresPairs, nil
}

// GetProcessingRate - получение скорости обработки
func (c *ProcessMetricsCollector) GetProcessingRate(id string) (float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return 0, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics.ProcessingRate, nil
}

// GetPacketRate - получение скорости пакетов
func (c *ProcessMetricsCollector) GetPacketRate(id string) (float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return 0, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics.PacketRate, nil
}

// GetFilteringTime - получение времени фильтрации
func (c *ProcessMetricsCollector) GetFilteringTime(id string) (time.Duration, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return 0, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics.FilteringTime, nil
}

// GetParsingTime - получение времени парсинга
func (c *ProcessMetricsCollector) GetParsingTime(id string) (time.Duration, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return 0, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics.ParsingTime, nil
}

// GetInsertionTime - получение времени вставки в PostgreSQL
func (c *ProcessMetricsCollector) GetInsertionTime(id string) (time.Duration, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metrics, ok := c.metrics[id]
	if !ok {
		return 0, fmt.Errorf("метрики процесса с ID %s не найдены", id)
	}

	return metrics.InsertionTime, nil
}

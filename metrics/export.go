// export.go - Экспорт метрик для системы Venera
//
// Этот модуль обеспечивает экспорт метрик для внешних систем мониторинга
// в различных форматах (JSON, CSV, текст).
//
// Основные функции:
// - Экспорт метрик в различные форматы
// - Экспорт статистики по процессам
// - Экспорт метрик баз данных
// - Планировщик экспорта метрик
//
// Использование:
// import "venera/metrics"
// exporter := metrics.NewExporter(cfg, logger)
// exporter.ExportToJSON("metrics.json")
// exporter.ExportToCSV("metrics.csv")

package metrics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Exporter - структура экспортера метрик
type Exporter struct {
	cfg    *config.Config
	logger *logrus.Logger
	collector *Collector
}

// NewExporter - создание нового экспортера метрик
func NewExporter(cfg *config.Config, logger *logrus.Logger, collector *Collector) *Exporter {
	return &Exporter{
		cfg:       cfg,
		logger:    logger,
		collector: collector,
	}
}

// ExportToJSON - экспорт метрик в JSON файл
func (e *Exporter) ExportToJSON(filePath string) error {
	// Получение метрик
	metrics, err := e.collector.GetAllMetrics()
	if err != nil {
		return fmt.Errorf("ошибка получения метрик: %w", err)
	}

	// Формирование данных для экспорта
	exportData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics":   metrics,
	}

	// Маршалинг в JSON
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	// Запись в файл
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	e.logger.WithField("file", filePath).Info("Метрики экспортированы в JSON")
	return nil
}

// ExportToCSV - экспорт метрик в CSV файл
func (e *Exporter) ExportToCSV(filePath string) error {
	// Создание файла
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	// Создание CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Получение метрик
	metrics, err := e.collector.GetAllMetrics()
	if err != nil {
		return fmt.Errorf("ошибка получения метрик: %w", err)
	}

	// Запись заголовков
	headers := []string{"timestamp", "metric_name", "metric_value", "metric_type"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("ошибка записи заголовков: %w", err)
	}

	// Запись данных
	timestamp := time.Now().Format(time.RFC3339)
	e.writeMetricsRecursive(writer, "system", metrics, timestamp)

	e.logger.WithField("file", filePath).Info("Метрики экспортированы в CSV")
	return nil
}

// writeMetricsRecursive - рекурсивная запись метрик в CSV
func (e *Exporter) writeMetricsRecursive(writer *csv.Writer, prefix string, metrics interface{}, timestamp string) {
	switch v := metrics.(type) {
	case map[string]interface{}:
		for key, value := range v {
			e.writeMetricsRecursive(writer, fmt.Sprintf("%s.%s", prefix, key), value, timestamp)
		}
	case float64:
		_ = writer.Write([]string{timestamp, prefix, fmt.Sprintf("%.2f", v), "float"})
	case int64:
		_ = writer.Write([]string{timestamp, prefix, fmt.Sprintf("%d", v), "int"})
	case string:
		_ = writer.Write([]string{timestamp, prefix, v, "string"})
	case bool:
		_ = writer.Write([]string{timestamp, prefix, fmt.Sprintf("%t", v), "bool"})
	}
}

// ExportToText - экспорт метрик в текстовый файл
func (e *Exporter) ExportToText(filePath string) error {
	// Получение метрик
	metrics, err := e.collector.GetAllMetrics()
	if err != nil {
		return fmt.Errorf("ошибка получения метрик: %w", err)
	}

	// Формирование текста
	var text string
	text += fmt.Sprintf("Venera Metrics Export\n")
	text += fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339))
	text += fmt.Sprintf("=====================\n\n")

	// Запись метрик
	text += e.formatMetrics(metrics, 0)

	// Запись в файл
	err = os.WriteFile(filePath, []byte(text), 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	e.logger.WithField("file", filePath).Info("Метрики экспортированы в текст")
	return nil
}

// formatMetrics - форматирование метрик в текст
func (e *Exporter) formatMetrics(metrics interface{}, indent int) string {
	var text string
	prefix := ""

	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	switch v := metrics.(type) {
	case map[string]interface{}:
		for key, value := range v {
			switch value.(type) {
			case map[string]interface{}:
				text += fmt.Sprintf("%s%s:\n", prefix, key)
				text += e.formatMetrics(value, indent+1)
			default:
				text += fmt.Sprintf("%s%s: %v\n", prefix, key, value)
			}
		}
	}

	return text
}

// ExportProcessMetrics - экспорт метрик процессов
func (e *Exporter) ExportProcessMetrics(filePath string, processMetrics map[string]*ProcessMetrics) error {
	// Маршалинг в JSON
	jsonData, err := json.MarshalIndent(processMetrics, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	// Запись в файл
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	e.logger.WithField("file", filePath).Info("Метрики процессов экспортированы")
	return nil
}

// ExportDatabaseMetrics - экспорт метрик баз данных
func (e *Exporter) ExportDatabaseMetrics(filePath string, dbMetrics *DatabaseMetrics) error {
	// Маршалинг в JSON
	jsonData, err := json.MarshalIndent(dbMetrics, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	// Запись в файл
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	e.logger.WithField("file", filePath).Info("Метрики баз данных экспортированы")
	return nil
}

// ExportSystemMetrics - экспорт системных метрик
func (e *Exporter) ExportSystemMetrics(filePath string, systemMetrics *SystemMetrics) error {
	// Маршалинг в JSON
	jsonData, err := json.MarshalIndent(systemMetrics, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	// Запись в файл
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	e.logger.WithField("file", filePath).Info("Системные метрики экспортированы")
	return nil
}

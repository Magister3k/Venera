// id.go - Генерация ID для процессов и объектов
//
// Этот модуль обеспечивает генерацию уникальных ID для процессов
// и других объектов системы Venera.
//
// Основные функции:
// - Генерация уникальных ID для процессов
// - Формат ID: process_YYYYMMDD_HHMMSS_N
// - Проверка уникальности ID
//
// Использование:
// import "venera/utils"
// id := utils.GenerateProcessID()
// fmt.Println(id) // "process_20260705_150405_123"

package utils

import (
	"fmt"
	"sync"
	"time"
)

// IDGenerator - генератор ID
type IDGenerator struct {
	mu       sync.Mutex
	counters map[string]int
}

// GlobalIDGenerator - глобальный генератор ID
var GlobalIDGenerator *IDGenerator

func init() {
	GlobalIDGenerator = &IDGenerator{
		counters: make(map[string]int),
	}
}

// GenerateProcessID - генерация ID процесса
func GenerateProcessID() string {
	return GenerateID("process")
}

// GenerateQueueID - генерация ID очереди
func GenerateQueueID() string {
	return GenerateID("queue")
}

// GenerateSetID - генерация ID множества
func GenerateSetID() string {
	return GenerateID("set")
}

// GenerateID - генерация ID с префиксом
func GenerateID(prefix string) string {
	GlobalIDGenerator.mu.Lock()
	defer GlobalIDGenerator.mu.Unlock()

	// Генерация времени
	now := time.Now()
	timestamp := now.Format("20060102_150405")

	// Генерация счетчика
	counter := GlobalIDGenerator.counters[prefix]
	GlobalIDGenerator.counters[prefix] = counter + 1

	// Формирование ID
	id := fmt.Sprintf("%s_%s_%d", prefix, timestamp, counter)

	return id
}

// GenerateProcessConfigID - генерация ID конфигурации процесса
func GenerateProcessConfigID() string {
	return GenerateID("config")
}

// GenerateAlertID - генерация ID алерта
func GenerateAlertID() string {
	return GenerateID("alert")
}

// GenerateLogID - генерация ID лога
func GenerateLogID() string {
	return GenerateID("log")
}

// IsProcessID - проверка, является ли ID ID процесса
func IsProcessID(id string) bool {
	return len(id) > 8 && id[:8] == "process_"
}

// IsQueueID - проверка, является ли ID ID очереди
func IsQueueID(id string) bool {
	return len(id) > 6 && id[:6] == "queue_"
}

// IsSetID - проверка, является ли ID ID множества
func IsSetID(id string) bool {
	return len(id) > 4 && id[:4] == "set_"
}

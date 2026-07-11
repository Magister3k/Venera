// id.go - Генерация уникальных идентификаторов для процессов
//
// Этот модуль обеспечивает генерацию уникальных ID для процессов обработки данных.
//
// Основные функции:
// - Генерация UUID для процессов
// - Генерация уникальных имен для очередей
// - Уникализация идентификаторов
//
// Использование:
// import "venera/utils"
// id := GenerateProcessID()
// queueName := GenerateQueueName(id)

package utils

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// IDGenerator - генератор уникальных ID
type IDGenerator struct {
	mu       sync.Mutex
	sequence uint64
}

// Global generator - глобальный генератор ID
var (
	globalGenerator *IDGenerator
	once            sync.Once
)

// GetIDGenerator - получить глобальный генератор ID
func GetIDGenerator() *IDGenerator {
	once.Do(func() {
		globalGenerator = &IDGenerator{
			sequence: uint64(time.Now().UnixNano()),
		}
	})
	return globalGenerator
}

// GenerateProcessID - генерация уникального ID для процесса
func GenerateProcessID() string {
	g := GetIDGenerator()
	g.mu.Lock()
	defer g.mu.Unlock()

	g.sequence++
	return uuid.New().String()[:8] // Первые 8 символов UUID
}

// GenerateQueueName - генерация имени очереди для процесса
func GenerateQueueName(processID string) string {
	return "queue_" + processID
}

// GenerateSortedSetName - генерация имени sorted set для процесса
func GenerateSortedSetName(processID string) string {
	return "sorted_" + processID
}

// GenerateCounterName - генерация имени счетчика для процесса
func GenerateCounterName(processID string) string {
	return "counter_" + processID
}

// GenerateTimerName - генерация имени таймера для процесса
func GenerateTimerName(processID string) string {
	return "timer_" + processID
}

// GenerateChannelName - генерация имени канала для процесса
func GenerateChannelName(processID string) string {
	return "channel_" + processID
}

// GenerateUniqueID - генерация уникального ID с временем
func GenerateUniqueID(prefix string) string {
	g := GetIDGenerator()
	g.mu.Lock()
	defer g.mu.Unlock()

	g.sequence++
	timestamp := time.Now().UnixNano()
	return prefix + "_" + time.Now().Format("20060102_150405") + "_" + string(rune('A'+int(g.sequence%26)))
}

// GenerateBatchID - генерация ID для пакета данных
func GenerateBatchID() string {
	return "batch_" + time.Now().Format("20060102_150405_000")
}

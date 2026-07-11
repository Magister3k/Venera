// data/parser.go - Модуль парсинга данных
// Добавлено в data/select.go для функции ParseData

package data

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseData - разбор данных в формате ключ:значение:время
func ParseData(record string) (string, string, int64, error) {
	// Разделение по последним двум двоеточиям
	parts := strings.Split(record, ":")
	if len(parts) < 3 {
		return "", "", 0, fmt.Errorf("недопустимый формат данных: %s", record)
	}

	// Последняя часть - время
	timestampStr := parts[len(parts)-1]
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("ошибка разбора времени: %w", err)
	}

	// Предпоследняя часть - значение
	value := parts[len(parts)-2]

	// Остальное - ключ
	key := strings.Join(parts[:len(parts)-2], ":")

	return key, value, timestamp, nil
}

// FormatData - форматирование данных
func FormatData(key, value string, timestamp int64) string {
	return fmt.Sprintf("%s:%s:%d", key, value, timestamp)
}

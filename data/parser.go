// data/parser.go - Модуль парсинга данных
//
// Этот модуль обеспечивает разбор JSON-данных, полученных от Tshark,
// с учетом допустимости использования в значениях символа ":".
//
// Основные функции:
// - Разбор JSON-данных от Tshark
// - Извлечение пар ключ-значение
// - Обработка значений, содержащих символ ":"
// - Приведение данных к формату ключ:значение:время
//
// Использование:
// import "venera/data"
// parser := data.NewParser()
// pairs := parser.ParseJSON(jsonData, timestamp)
// for _, pair := range pairs {
//     fmt.Printf("Key: %s, Value: %s\n", pair.Key, pair.Value)
// }

package data

import (
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// DataPair - пара ключ-значение
type DataPair struct {
	Key       string
	Value     string
	Timestamp int64
}

// Parser - структура парсера данных
type Parser struct {
	// Конфигурация парсера
}

// NewParser - создание нового парсера
func NewParser() *Parser {
	return &Parser{}
}

// ParseJSON - разбор JSON-данных
// Использует fastjson для эффективного разбора
func (p *Parser) ParseJSON(data []byte, timestamp int64) []DataPair {
	// Временная реализация для примера
	// В продакшене следует использовать fastjson для эффективного разбора
	var pairs []DataPair

	// Разбор JSON с использованием msgpack для примера
	// В реальном приложении используйте fastjson

	return pairs
}

// ParseLine - разбор одной строки данных
// Формат: ключ:значение (где значение может содержать ":")
func (p *Parser) ParseLine(line string, timestamp int64) *DataPair {
	// Найти первый символ ":" для разделения ключа и значения
	colonIndex := -1
	for i := 0; i < len(line); i++ {
		if line[i] == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return nil
	}

	key := line[:colonIndex]
	value := line[colonIndex+1:]

	return &DataPair{
		Key:       key,
		Value:     value,
		Timestamp: timestamp,
	}
}

// FormatData - форматирование данных в строку
// Формат: ключ:значение:время
func (p *Parser) FormatData(key, value string, timestamp int64) string {
	return key + ":" + value + ":" + string(rune(timestamp))
}

// FormatDataWithSeparator - форматирование данных с разделителем
func (p *Parser) FormatDataWithSeparator(key, value string, timestamp int64, separator string) string {
	return key + separator + value + separator + string(rune(timestamp))
}

// UnmarshalJSON - разбор JSON в структуру DataPair
func (p *Parser) UnmarshalJSON(data []byte) (*DataPair, error) {
	var pair DataPair
	if err := msgpack.Unmarshal(data, &pair); err != nil {
		return nil, err
	}
	return &pair, nil
}

// MarshalJSON - сериализация DataPair в JSON
func (p *Parser) MarshalJSON(pair *DataPair) ([]byte, error) {
	return msgpack.Marshal(pair)
}

// ParseMultipleLines - разбор нескольких строк данных
func (p *Parser) ParseMultipleLines(lines []string, timestamp int64) []DataPair {
	var pairs []DataPair

	for _, line := range lines {
		if pair := p.ParseLine(line, timestamp); pair != nil {
			pairs = append(pairs, *pair)
		}
	}

	return pairs
}

// GetCurrentTimestamp - получение текущей метки времени
func (p *Parser) GetCurrentTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// FormatTimestamp - форматирование метки времени
func (p *Parser) FormatTimestamp(timestamp int64) string {
	t := time.Unix(0, timestamp*int64(time.Millisecond))
	return t.Format("2006-01-02 15:04:05")
}

// ParseTimestamp - разбор метки времени из строки
func (p *Parser) ParseTimestamp(str string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return 0, err
	}
	return t.UnixNano() / int64(time.Millisecond), nil
}

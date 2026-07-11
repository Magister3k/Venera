 // data/parser.go - Модуль парсинга данных
//
// Этот модуль обеспечивает разбор JSON-данных, полученных от Tshark,
// с учетом допустимости использования в значениях символа ":".
// Использует стандартную библиотеку encoding/json для разбора JSON.
//
// Основные функции:
// - Разбор JSON-данных от Tshark с использованием encoding/json
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
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// DataPair - пара ключ-значение
type DataPair struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// Parser - структура парсера данных
type Parser struct {
	// Нет внешних зависимостей - используем стандартную библиотеку
}

// NewParser - создание нового парсера
func NewParser() *Parser {
	return &Parser{}
}

// ParseJSON - разбор JSON-данных
// Этот метод обрабатывает JSON в формате, который выводит Tshark
func (p *Parser) ParseJSON(data []byte, timestamp int64) []DataPair {
	var pairs []DataPair

	// Разбор JSON с использованием encoding/json
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		// Логируем ошибку парсинга, но продолжаем работу
		// В продакшене следует использовать логгер
		return pairs
	}

	// Проверка, является ли JSON массивом объектов
	switch v := jsonData.(type) {
	case []interface{}:
		// Обработка массива JSON-объектов (стандартный формат Tshark)
		for _, item := range v {
			if pair := p.parseJSONObject(item, timestamp); pair != nil {
				pairs = append(pairs, *pair)
			}
		}
	case map[string]interface{}:
		// Обработка одного JSON-объекта
		if pair := p.parseJSONObject(jsonData, timestamp); pair != nil {
			pairs = append(pairs, *pair)
		}
	}

	return pairs
}

// parseJSONObject - разбор одного JSON-объекта
func (p *Parser) parseJSONObject(obj interface{}, timestamp int64) *DataPair {
	// Приведение к map[string]interface{}
	fieldsMap, ok := obj.(map[string]interface{})
	if !ok {
		return nil
	}

	// Попытка извлечь поле "fields" или "fields.payload"
	// Tshark может выводить данные в различных форматах

	// Сначала пробуем извлечь "fields"
	if fields, hasFields := fieldsMap["fields"]; hasFields {
		if fieldsMap, ok := fields.(map[string]interface{}); ok {
			return p.extractPairsFromFields(fieldsMap, timestamp)
		}
	}

	// Если нет fields, пробуем извлечь "fields.payload"
	if fieldsPayload, hasPayload := fieldsMap["fields"]; hasPayload {
		if fieldsMap, ok := fieldsPayload.(map[string]interface{}); ok {
			if payload, hasPayloadField := fieldsMap["payload"]; hasPayloadField {
				// payload может содержать JSON или просто текст
				if payloadStr, ok := payload.(string); ok {
					// Пытаемся разобрать как JSON
					var payloadJSON interface{}
					if err := json.Unmarshal([]byte(payloadStr), &payloadJSON); err == nil {
						if payloadMap, ok := payloadJSON.(map[string]interface{}); ok {
							return p.parseJSONObject(payloadMap, timestamp)
						}
					}
					// Если не JSON, считаем это простым значением
					return &DataPair{
						Key:       "payload",
						Value:     payloadStr,
						Timestamp: timestamp,
					}
				}
			}
		}
	}

	// Если структура не распознана, возвращаем nil
	return nil
}

// extractPairsFromFields - извлечение пар ключ-значение из fields
func (p *Parser) extractPairsFromFields(fieldsMap map[string]interface{}, timestamp int64) *DataPair {
	// Создаем буфер для сбора всех полей
	var buf bytes.Buffer

	// Обходим все поля объекта
	first := true
	for key, value := range fieldsMap {
		if !first {
			buf.WriteString(",")
		}
		first = false
		buf.WriteString(key)
		buf.WriteString(":")
		// Преобразуем значение в строку
		switch v := value.(type) {
		case string:
			buf.WriteString(v)
		case float64:
			buf.WriteString(fmt.Sprintf("%g", v))
		case bool:
			buf.WriteString(fmt.Sprintf("%t", v))
		default:
			buf.WriteString(fmt.Sprintf("%v", v))
		}
	}

	return &DataPair{
		Key:       "fields",
		Value:     buf.String(),
		Timestamp: timestamp,
	}
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
	return key + ":" + value + ":" + fmt.Sprintf("%d", timestamp)
}

// FormatDataWithSeparator - форматирование данных с разделителем
func (p *Parser) FormatDataWithSeparator(key, value string, timestamp int64, separator string) string {
	return key + separator + value + separator + fmt.Sprintf("%d", timestamp)
}

// UnmarshalJSON - разбор JSON в структуру DataPair
// Использует encoding/json для совместимости
func (p *Parser) UnmarshalJSON(data []byte) (*DataPair, error) {
	var pair DataPair
	if err := json.Unmarshal(data, &pair); err != nil {
		return nil, err
	}
	return &pair, nil
}

// MarshalJSON - сериализация DataPair в JSON
// Использует encoding/json для совместимости
func (p *Parser) MarshalJSON(pair *DataPair) ([]byte, error) {
	return json.Marshal(pair)
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

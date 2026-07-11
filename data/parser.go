// data/parser.go - Модуль парсинга данных
//
// Этот модуль обеспечивает разбор JSON-данных, полученных от Tshark,
// с учетом допустимости использования в значениях символа ":".
// Использует fastjson для высокопроизводительного разбора JSON.
//
// Основные функции:
// - Разбор JSON-данных от Tshark с использованием fastjson
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
	"time"

	"github.com/segmentio/fastjson/v2"
)

// DataPair - пара ключ-значение
type DataPair struct {
	Key       string
	Value     string
	Timestamp int64
}

// Parser - структура парсера данных
type Parser struct {
	// fastjson парсер для эффективного разбора JSON
	parser *fastjson.Parser
}

// NewParser - создание нового парсера
func NewParser() *Parser {
	return &Parser{
		parser: &fastjson.Parser{},
	}
}

// ParseJSON - разбор JSON-данных
// Использует fastjson для высокопроизводительного разбора JSON
// Этот метод обрабатывает JSON в формате, который выводит Tshark
func (p *Parser) ParseJSON(data []byte, timestamp int64) []DataPair {
	var pairs []DataPair

	// Разбор JSON с использованием fastjson
	v, err := p.parser.ParseBytes(data)
	if err != nil {
		// Логируем ошибку парсинга, но продолжаем работу
		// В продакшене следует использовать логгер
		return pairs
	}

	// Проверка, является ли JSON массивом объектов
	if v.Type() == fastjson.TypeArray {
		// Обработка массива JSON-объектов (стандартный формат Tshark)
		arr, err := v.Array()
		if err != nil {
			return pairs
		}

		for _, item := range arr {
			if pair := p.parseJSONObject(item, timestamp); pair != nil {
				pairs = append(pairs, *pair)
			}
		}
	} else if v.Type() == fastjson.TypeObject {
		// Обработка одного JSON-объекта
		if pair := p.parseJSONObject(v, timestamp); pair != nil {
			pairs = append(pairs, *pair)
		}
	}

	return pairs
}

// parseJSONObject - разбор одного JSON-объекта
func (p *Parser) parseJSONObject(obj *fastjson.Value, timestamp int64) *DataPair {
	// Попытка извлечь поле "fields" или "fields.payload"
	// Tshark может выводить данные в различных форматах

	// Сначала пробуем извлечь "fields"
	if fields := obj.Get("fields"); fields.Type() == fastjson.TypeObject {
		return p.extractPairsFromFields(fields, timestamp)
	}

	// Если нет fields, пробуем извлечь "fields.payload"
	if fields := obj.Get("fields", "payload"); fields.Type() == fastjson.TypeString {
		// payload может содержать JSON или просто текст
		payloadBytes, err := fields.StringBytes()
		if err == nil {
			// Пытаемся разобрать как JSON
			v, err := p.parser.ParseBytes(payloadBytes)
			if err == nil {
				if v.Type() == fastjson.TypeObject {
					return p.parseJSONObject(v, timestamp)
				}
			}
			// Если не JSON, считаем это простым значением
			payloadStr, _ := fields.StringBytes()
			return &DataPair{
				Key:       "payload",
				Value:     string(payloadStr),
				Timestamp: timestamp,
			}
		}
	}

	// Если структура не распознана, возвращаем nil
	return nil
}

// extractPairsFromFields - извлечение пар ключ-значение из fields
func (p *Parser) extractPairsFromFields(fields *fastjson.Value, timestamp int64) *DataPair {
	// Создаем буфер для сбора всех полей
	var buf bytes.Buffer

	// Обходим все поля объекта
	fields.Object().Visit(func(key []byte, value *fastjson.Value) {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		buf.Write(key)
		buf.WriteString(":")
		buf.Write(value.StringBytes())
	})

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
	return key + ":" + value + ":" + string(rune(timestamp))
}

// FormatDataWithSeparator - форматирование данных с разделителем
func (p *Parser) FormatDataWithSeparator(key, value string, timestamp int64, separator string) string {
	return key + separator + value + separator + string(rune(timestamp))
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

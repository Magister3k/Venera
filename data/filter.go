// data/filter.go - Модуль фильтрации данных
//
// Этот модуль обеспечивает фильтрацию данных на основе белого и черного списков.
// Белый список содержит ключи, которые разрешены для обработки.
// Черный список содержит значения, которые запрещены для обработки.
//
// Основные функции:
// - Загрузка фильтров из файла generic.flt
// - Проверка ключей и значений на соответствие фильтрам
// - Многоуровневая фильтрация (ключ-значение)
//
// Формат файла фильтрации:
//   +ключ_|_название_  - белый список ключей (1-й уровень)
//   -значение_        - черный список значений (2-й уровень)
//
// Использование:
// import "venera/data"
// filter := data.NewDataFilter("settings/generic.flt")
// if filter.IsAllowed(key, value) {
//     // Обработка данных
// }

package data

import (
	"fmt"
	"os"
	"strings"
)

// DataFilter - структура фильтрации данных
type DataFilter struct {
	Whitelist map[string][]string // Белый список: ключ -> [значения]
	Blacklist map[string][]string // Черный список: ключ -> [значения]
}

// NewDataFilter - создание нового фильтра из файла
func NewDataFilter(filePath string) (*DataFilter, error) {
	filter := &DataFilter{
		Whitelist: make(map[string][]string),
		Blacklist: make(map[string][]string),
	}

	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filter, nil // Возвращаем пустой фильтр, если файл не найден
	}

	// Загрузка фильтров из файла
	if err := filter.LoadFromFile(filePath); err != nil {
		return nil, fmt.Errorf("ошибка загрузки фильтров: %w", err)
	}

	return filter, nil
}

// LoadFromFile - загрузка фильтров из файла
func (f *DataFilter) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла фильтрации: %w", err)
	}
	defer file.Close()

	// Чтение файла построчно
	var currentKey string
	buf := make([]byte, 65536)

	for {
		n, err := file.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("ошибка чтения файла фильтрации: %w", err)
		}

		if n == 0 {
			break
		}

		lines := strings.Split(string(buf[:n]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Проверка первого символа
			if strings.HasPrefix(line, "+") {
				// Белый список ключей
				parts := strings.SplitN(line[1:], "|", 2)
				if len(parts) == 2 {
					currentKey = strings.TrimSpace(parts[0])
					f.Whitelist[currentKey] = []string{}
				}
			} else if strings.HasPrefix(line, "-") {
				// Черный список значений
				value := strings.TrimSpace(line[1:])
				if currentKey != "" {
					f.Whitelist[currentKey] = append(f.Whitelist[currentKey], value)
				}
			}
		}
	}

	return nil
}

// IsAllowed - проверка, разрешено ли значение для данного ключа
func (f *DataFilter) IsAllowed(key, value string) bool {
	// Если белый список пуст, разрешаем все
	if len(f.Whitelist) == 0 {
		return true
	}

	// Проверка в белом списке
	allowedValues, exists := f.Whitelist[key]
	if !exists {
		// Ключ не в белом списке - запрещаем
		return false
	}

	// Если в белом списке нет значений, разрешаем все для этого ключа
	if len(allowedValues) == 0 {
		return true
	}

	// Проверка в черном списке
	for _, blackValue := range f.Blacklist[key] {
		if value == blackValue {
			return false // Значение в черном списке
		}
	}

	// Значение не в черном списке и ключ в белом - разрешаем
	return true
}

// IsKeyAllowed - проверка, разрешен ли ключ
func (f *DataFilter) IsKeyAllowed(key string) bool {
	// Если белый список пуст, разрешаем все
	if len(f.Whitelist) == 0 {
		return true
	}

	// Проверка в белом списке
	_, exists := f.Whitelist[key]
	return exists
}

// GetWhitelist - получение белого списка
func (f *DataFilter) GetWhitelist() map[string][]string {
	return f.Whitelist
}

// GetBlacklist - получение черного списка
func (f *DataFilter) GetBlacklist() map[string][]string {
	return f.Blacklist
}

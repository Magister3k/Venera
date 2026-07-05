// select.go - Работа с данными из DragonflyDB
//
// Этот модуль обеспечивает операции с данными в DragonflyDB,
// включая выборку, агрегацию и обработку данных из структур list и sorted set.
//
// Основные функции:
// - Работа с данными из структуры list
// - Работа с данными из структуры sorted set
// - Операции выборки и агрегации
// - Интеграция с PostgreSQL для финального хранения
//
// Использование:
// import "venera/data"
// selector := NewDataSelector(dragonflyDB, postgresDB)
// selector.ProcessData("queue1", "sorted1")

package data

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// DataSelector - структура выборки данных
type DataSelector struct {
	dragonflyDB *DragonflyDB
	postgresDB  *PostgreSQL
	filter      *DataFilter
	controls    map[string]string
	log         *logrus.Logger
}

// NewDataSelector - создание нового селектора данных
func NewDataSelector(dragonflyDB *DragonflyDB, postgresDB *PostgreSQL, filter *DataFilter, controls map[string]string) *DataSelector {
	return &DataSelector{
		dragonflyDB: dragonflyDB,
		postgresDB:  postgresDB,
		filter:      filter,
		controls:    controls,
		log:         logrus.WithField("module", "data_selector"),
	}
}

// ProcessData - обработка данных из DragonflyDB
func (s *DataSelector) ProcessData(listKey, sortedSetKey string) error {
	s.log.Info("Начало обработки данных")

	// Получение данных из list
	data, err := s.dragonflyDB.GetAndRemoveFromList(listKey, s.dragonflyDB.cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("ошибка получения данных: %w", err)
	}

	if len(data) == 0 {
		s.log.Info("Нет данных для обработки")
		return nil
	}

	s.log.Infof("Получено %d записей из list", len(data))

	// Разделение записей и проверка фильтрации
	var selectedData []string
	for _, record := range data {
		key, value, timestamp, err := ParseData(record)
		if err != nil {
			s.log.Warnf("Ошибка разбора данных: %v", err)
			continue
		}

		// Проверка фильтрации ключей
		if !s.filter.IsAllowed(key, value) {
			s.log.Debugf("Запись отфильтрована: %s=%s", key, value)
			continue
		}

		// Добавление в отобранные данные
		selectedData = append(selectedData, record)
	}

	s.log.Infof("Отобрано %d записей после фильтрации", len(selectedData))

	// Добавление в sorted set
	for _, record := range selectedData {
		key, value, timestamp, err := ParseData(record)
		if err != nil {
			s.log.Warnf("Ошибка разбора данных при добавлении в sorted set: %v", err)
			continue
		}

		if err := s.dragonflyDB.AddToSortedSet(sortedSetKey, fmt.Sprintf("%s:%s", key, value), timestamp); err != nil {
			s.log.Errorf("Ошибка добавления в sorted set: %v", err)
		}
	}

	// Перемещение данных в PostgreSQL
	if err := s.moveDataToPostgreSQL(sortedSetKey, len(selectedData)); err != nil {
		return fmt.Errorf("ошибка перемещения данных в PostgreSQL: %w", err)
	}

	return nil
}

// moveDataToPostgreSQL - перемещение данных в PostgreSQL
func (s *DataSelector) moveDataToPostgreSQL(sortedSetKey string, count int) error {
	s.log.Infof("Перемещение данных в PostgreSQL (количество: %d)", count)

	// Получение данных из sorted set
	data, err := s.dragonflyDB.GetSortedSetRangeWithScores(sortedSetKey, 0, int64(count-1))
	if err != nil {
		return fmt.Errorf("ошибка получения данных из sorted set: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	// Подготовка данных для пакетной вставки
	var batchData []BatchData
	for _, item := range data {
		// Разбор данных
		parts := fmt.Sprintf("%v", item.Member)
		colonIndex := -1
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == ':' {
				colonIndex = i
				break
			}
		}

		if colonIndex == -1 {
			s.log.Warnf("Недопустимый формат данных: %s", parts)
			continue
		}

		key := parts[:colonIndex]
		timestamp := int64(item.Score)

		// Разделение по последнему двоеточию
		lastColon := -1
		for i := len(key) - 1; i >= 0; i-- {
			if key[i] == ':' {
				lastColon = i
				break
			}
		}

		if lastColon == -1 {
			s.log.Warnf("Недопустимый формат данных: %s", key)
			continue
		}

		source := "default" // По умолчанию
		keyPart := key[lastColon+1:]

		// Использование controls для получения источника
		if sourceValue, exists := s.controls[key]; exists {
			source = sourceValue
		}

		batchData = append(batchData, BatchData{
			Source:    source,
			Key:       keyPart,
			Value:     key[:lastColon],
			Timestamp: timestamp,
		})
	}

	if len(batchData) == 0 {
		return nil
	}

	// Пакетная вставка в PostgreSQL
	if err := s.postgresDB.InsertBatchData(batchData); err != nil {
		return fmt.Errorf("ошибка пакетной вставки: %w", err)
	}

	s.log.Infof("Успешно перемещено %d записей в PostgreSQL", len(batchData))

	// Очистка sorted set
	if err := s.dragonflyDB.ClearSortedSet(sortedSetKey); err != nil {
		s.log.Warnf("Ошибка очистки sorted set: %v", err)
	}

	return nil
}

// ProcessListData - обработка данных из list без sorted set
func (s *DataSelector) ProcessListData(listKey string) error {
	s.log.Info("Начало обработки данных из list")

	// Получение данных из list
	data, err := s.dragonflyDB.GetAndRemoveFromList(listKey, s.dragonflyDB.cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("ошибка получения данных: %w", err)
	}

	if len(data) == 0 {
		s.log.Info("Нет данных для обработки")
		return nil
	}

	s.log.Infof("Получено %d записей из list", len(data))

	// Подготовка данных для вставки
	var batchData []BatchData
	for _, record := range data {
		key, value, timestamp, err := ParseData(record)
		if err != nil {
			s.log.Warnf("Ошибка разбора данных: %v", err)
			continue
		}

		// Проверка фильтрации ключей
		if !s.filter.IsAllowed(key, value) {
			s.log.Debugf("Запись отфильтрована: %s=%s", key, value)
			continue
		}

		// Использование controls для получения источника
		source := "default"
		if sourceValue, exists := s.controls[key]; exists {
			source = sourceValue
		}

		batchData = append(batchData, BatchData{
			Source:    source,
			Key:       key,
			Value:     value,
			Timestamp: timestamp,
		})
	}

	if len(batchData) == 0 {
		return nil
	}

	// Пакетная вставка в PostgreSQL
	if err := s.postgresDB.InsertBatchData(batchData); err != nil {
		return fmt.Errorf("ошибка пакетной вставки: %w", err)
	}

	s.log.Infof("Успешно перемещено %d записей в PostgreSQL", len(batchData))

	return nil
}

// GetListStatistics - получение статистики list
func (s *DataSelector) GetListStatistics(listKey string) (int64, error) {
	return s.dragonflyDB.GetListLength(listKey)
}

// GetSortedSetStatistics - получение статистики sorted set
func (s *DataSelector) GetSortedSetStatistics(sortedSetKey string) (int64, error) {
	return s.dragonflyDB.GetSortedSetLength(sortedSetKey)
}

// ProcessWithTimeout - обработка данных с таймаутом
func (s *DataSelector) ProcessWithTimeout(listKey, sortedSetKey string, timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- s.ProcessData(listKey, sortedSetKey)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("превышено время ожидания обработки данных")
	}
}

// GetPostgresStatistics - получение статистики из PostgreSQL
func (s *DataSelector) GetPostgresStatistics() (Statistics, error) {
	return s.postgresDB.GetStatistics()
}

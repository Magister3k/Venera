// dragonfly.go - Интеграция с DragonflyDB для кэширования данных
//
// Этот модуль обеспечивает взаимодействие с DragonflyDB для хранения и обработки
// промежуточных данных в процессе сбора идентификаторов.
//
// Основные функции:
// - Подключение к DragonflyDB с экспоненциальной задержкой при ошибке
// - Работа со структурами list для очередей данных
// - Работа со структурами sorted set для сортированных данных
// - Батчевая обработка данных
// - Настройка параметров подключения из конфигурации
//
// Использование:
// import "venera/data"
// db, err := NewDragonflyDB(cfg.DragonflyDB)
// if err != nil { log.Fatal(err) }
// defer db.Close()
// db.AddToList("queue1", "data1")
// db.AddToSortedSet("sorted1", "data1", 1234567890)

package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// DragonflyDB - структура подключения к DragonflyDB
type DragonflyDB struct {
	client *redis.Client
	cfg    DragonflyDBConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// DragonflyDBConfig - параметры подключения к DragonflyDB
type DragonflyDBConfig struct {
	Host        string
	Port        int
	Password    string
	Database    int
	BatchSize   int
	Timeout     int
}

// NewDragonflyDB - создание нового подключения к DragonflyDB
func NewDragonflyDB(cfg DragonflyDBConfig) (*DragonflyDB, error) {
	// Установка адреса
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Настройка клиента
	client := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     cfg.Password,
		DB:           cfg.Database,
		MaxRetries:   3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 1 * time.Second,
	})

	// Проверка подключения с экспоненциальной задержкой
	var db *DragonflyDB
	var err error

	for attempt := 1; attempt <= 5; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
		err = client.Ping(ctx).Err()
		cancel()

		if err == nil {
			db = &DragonflyDB{
				client: client,
				cfg:    cfg,
			}
			break
		}

		// Ждем перед следующей попыткой
		delay := time.Duration(attempt*attempt) * time.Second
		time.Sleep(delay)
	}

	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к DragonflyDB: %w", err)
	}

	return db, nil
}

// AddToList - добавление элемента в list
func (db *DragonflyDB) AddToList(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.LPush(ctx, key, value).Err()
}

// AddToSortedSet - добавление элемента в sorted set
func (db *DragonflyDB) AddToSortedSet(key, value string, score int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(score),
		Member: value,
	}).Err()
}

// GetAndRemoveFromList - получение и удаление элементов из list
func (db *DragonflyDB) GetAndRemoveFromList(key string, count int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	// Получаем элементы
	values, err := db.client.LRange(ctx, key, 0, int64(count-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения элементов: %w", err)
	}

	// Удаляем элементы
	if len(values) > 0 {
		if err := db.client.LTrim(ctx, key, int64(len(values)), -1).Err(); err != nil {
			return nil, fmt.Errorf("ошибка удаления элементов: %w", err)
		}
	}

	return values, nil
}

// GetSortedSetRange - получение диапазона из sorted set
func (db *DragonflyDB) GetSortedSetRange(key string, start, stop int64) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.ZRange(ctx, key, start, stop).Result()
}

// GetSortedSetRangeWithScores - получение диапазона из sorted set с оценками
func (db *DragonflyDB) GetSortedSetRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// RemoveFromSortedSet - удаление элементов из sorted set
func (db *DragonflyDB) RemoveFromSortedSet(key string, members []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.ZRem(ctx, key, members).Err()
}

// GetListLength - получение длины list
func (db *DragonflyDB) GetListLength(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.LLen(ctx, key).Result()
}

// GetSortedSetLength - получение длины sorted set
func (db *DragonflyDB) GetSortedSetLength(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.ZCard(ctx, key).Result()
}

// Exists - проверка существования ключа
func (db *DragonflyDB) Exists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	count, err := db.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Delete - удаление ключа
func (db *DragonflyDB) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.Del(ctx, key).Err()
}

// ClearList - очистка list
func (db *DragonflyDB) ClearList(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.Del(ctx, key).Err()
}

// ClearSortedSet - очистка sorted set
func (db *DragonflyDB) ClearSortedSet(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.cfg.Timeout)*time.Second)
	defer cancel()

	return db.client.Del(ctx, key).Err()
}

// Close - закрытие подключения
func (db *DragonflyDB) Close() error {
	if db.client != nil {
		return db.client.Close()
	}
	return nil
}

// GetClient - получение клиентского объекта (для расширенного использования)
func (db *DragonflyDB) GetClient() *redis.Client {
	return db.client
}

// GetConfig - получение конфигурации подключения
func (db *DragonflyDB) GetConfig() DragonflyDBConfig {
	return db.cfg
}

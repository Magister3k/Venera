// postgres.go - Интеграция с PostgreSQL для хранения данных
//
// Этот модуль обеспечивает взаимодействие с PostgreSQL для хранения собранных
// идентификаторов и связанной информации.
//
// Основные функции:
// - Подключение к PostgreSQL с экспоненциальной задержкой при ошибке
// - Создание базы данных при необходимости
// - Работа с pgx v5
// - Пакетная вставка данных
// - Настройка параметров подключения из конфигурации
//
// Использование:
// import "venera/data"
// db, err := NewPostgreSQL(cfg.PostgreSQL)
// if err != nil { log.Fatal(err) }
// defer db.Close()
// db.InsertData(source, key, value, timestamp)

package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL - структура подключения к PostgreSQL
type PostgreSQL struct {
	pool *pgxpool.Pool
	cfg  PostgreSQLConfig
}

// PostgreSQLConfig - параметры подключения к PostgreSQL
type PostgreSQLConfig struct {
	Host         string
	Port         int
	Database     string
	User         string
	Password     string
	MaxConnections int
}

// NewPostgreSQL - создание нового подключения к PostgreSQL
func NewPostgreSQL(cfg PostgreSQLConfig) (*PostgreSQL, error) {
	// Формирование строки подключения
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password)

	// Настройка конфигурации пула соединений
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга строки подключения: %w", err)
	}

	config.MaxConns = int32(cfg.MaxConnections)
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 10 * time.Second

	// Создание пула соединений с экспоненциальной задержкой
	var pool *pgxpool.Pool
	for attempt := 1; attempt <= 5; attempt++ {
		pool, err = pgxpool.ConnectConfig(context.Background(), config)
		if err == nil {
			break
		}

		// Ждем перед следующей попыткой
		delay := time.Duration(attempt*attempt) * time.Second
		time.Sleep(delay)
	}

	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к PostgreSQL: %w", err)
	}

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка проверки подключения: %w", err)
	}

	return &PostgreSQL{
		pool: pool,
		cfg:  cfg,
	}, nil
}

// InsertData - вставка данных в базу
func (db *PostgreSQL) InsertData(source, key, value string, timestamp int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		INSERT INTO data (source, key, value, date_first, date_last)
		VALUES ($1, $2, $3, to_timestamp($4), to_timestamp($4))
		ON CONFLICT (source, key, value) 
		DO UPDATE SET date_last = to_timestamp($4)
	`

	_, err := db.pool.Exec(ctx, query, source, key, value, timestamp)
	if err != nil {
		return fmt.Errorf("ошибка вставки данных: %w", err)
	}

	return nil
}

// InsertBatchData - пакетная вставка данных
func (db *PostgreSQL) InsertBatchData(data []BatchData) error {
	if len(data) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Создание транзакции
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Подготовка statement
	stmt, err := tx.Prepare(ctx, "insert_data",
		"INSERT INTO data (source, key, value, date_first, date_last) VALUES ($1, $2, $3, to_timestamp($4), to_timestamp($4)) ON CONFLICT (source, key, value) DO UPDATE SET date_last = to_timestamp($4)")
	if err != nil {
		return fmt.Errorf("ошибка подготовки statement: %w", err)
	}

	defer stmt.Close()

	// Вставка данных
	for _, item := range data {
		_, err = stmt.Exec(ctx, item.Source, item.Key, item.Value, item.Timestamp)
		if err != nil {
			return fmt.Errorf("ошибка вставки данных: %w", err)
		}
	}

	// Фиксация транзакции
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}

// GetDataBySource - получение данных по источнику
func (db *PostgreSQL) GetDataBySource(source string, limit, offset int) ([]DataRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT id, source, key, value, 
		       EXTRACT(EPOCH FROM date_first)::bigint as date_first,
		       EXTRACT(EPOCH FROM date_last)::bigint as date_last
		FROM data 
		WHERE source = $1 
		ORDER BY date_last DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := db.pool.Query(ctx, query, source, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса данных: %w", err)
	}
	defer rows.Close()

	var records []DataRecord
	for rows.Next() {
		var record DataRecord
		err := rows.Scan(&record.ID, &record.Source, &record.Key, &record.Value,
			&record.DateFirst, &record.DateLast)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// GetDataByFilter - получение данных по фильтру
func (db *PostgreSQL) GetDataByFilter(source, key, value string, startDate, endDate int64, limit, offset int) ([]DataRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT id, source, key, value, 
		       EXTRACT(EPOCH FROM date_first)::bigint as date_first,
		       EXTRACT(EPOCH FROM date_last)::bigint as date_last
		FROM data 
		WHERE ($1 = '' OR source = $1)
		  AND ($2 = '' OR key = $2)
		  AND ($3 = '' OR value = $3)
		  AND date_first >= to_timestamp($4)
		  AND date_last <= to_timestamp($5)
		ORDER BY date_last DESC 
		LIMIT $6 OFFSET $7
	`

	rows, err := db.pool.Query(ctx, query, source, key, value, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса данных: %w", err)
	}
	defer rows.Close()

	var records []DataRecord
	for rows.Next() {
		var record DataRecord
		err := rows.Scan(&record.ID, &record.Source, &record.Key, &record.Value,
			&record.DateFirst, &record.DateLast)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// GetStatistics - получение статистики данных
func (db *PostgreSQL) GetStatistics() (Statistics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var stats Statistics

	// Общее количество записей
	err := db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM data").Scan(&stats.TotalCount)
	if err != nil {
		return stats, fmt.Errorf("ошибка получения общего количества: %w", err)
	}

	// Количество уникальных источников
	err = db.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT source) FROM data").Scan(&stats.SourceCount)
	if err != nil {
		return stats, fmt.Errorf("ошибка получения количества источников: %w", err)
	}

	// Количество уникальных ключей
	err = db.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT key) FROM data").Scan(&stats.KeyCount)
	if err != nil {
		return stats, fmt.Errorf("ошибка получения количества ключей: %w", err)
	}

	// Количество уникальных значений
	err = db.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT value) FROM data").Scan(&stats.ValueCount)
	if err != nil {
		return stats, fmt.Errorf("ошибка получения количества значений: %w", err)
	}

	return stats, nil
}

// GetDatabaseSize - получение размера базы данных
func (db *PostgreSQL) GetDatabaseSize() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var size int64
	err := db.pool.QueryRow(ctx, "SELECT pg_database_size($1)", db.cfg.Database).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения размера базы: %w", err)
	}

	return size, nil
}

// CreateDatabaseIfNotExists - создание базы данных, если она не существует
func (db *PostgreSQL) CreateDatabaseIfNotExists(databaseName string) error {
	// Отдельное подключение для создания базы данных
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s",
		db.cfg.Host, db.cfg.Port, db.cfg.User, db.cfg.Password)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("ошибка парсинга строки подключения: %w", err)
	}

	config.Database = "postgres" // Подключаемся к стандартной базе для создания новой
	config.MaxConns = 1

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}
	defer pool.Close()

	// Проверка существования базы данных
	var exists bool
	err = pool.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", databaseName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования базы: %w", err)
	}

	// Создание базы данных, если она не существует
	if !exists {
		_, err = pool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", databaseName))
		if err != nil {
			return fmt.Errorf("ошибка создания базы данных: %w", err)
		}
	}

	return nil
}

// Close - закрытие подключения
func (db *PostgreSQL) Close() error {
	if db.pool != nil {
		db.pool.Close()
	}
	return nil
}"}

// GetPool - получение пула соединений
func (db *PostgreSQL) GetPool() *pgxpool.Pool {
	return db.pool
}

// GetConfig - получение конфигурации подключения
func (db *PostgreSQL) GetConfig() PostgreSQLConfig {
	return db.cfg
}

// DataRecord - запись данных
type DataRecord struct {
	ID         int64  `json:"id"`
	Source     string `json:"source"`
	Key        string `json:"key"`
	Value      string `json:"value"`
	DateFirst  int64  `json:"date_first"`
	DateLast   int64  `json:"date_last"`
}

// BatchData - пакет данных для вставки
type BatchData struct {
	Source    string
	Key       string
	Value     string
	Timestamp int64
}

// Statistics - статистика данных
type Statistics struct {
	TotalCount int64 `json:"total_count"`
	SourceCount int64 `json:"source_count"`
	KeyCount   int64 `json:"key_count"`
	ValueCount int64 `json:"value_count"`
}

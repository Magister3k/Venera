-- create_pg_db.sql - SQL-скрипт для создания базы данных Venera в PostgreSQL
--
-- Этот скрипт создает базу данных 'Venera' и необходимые таблицы
-- для хранения идентификаторов в потоке пакетных данных.
--
-- Основные таблицы:
-- - data - основная таблица для хранения идентификаторов
-- - indexes - индексы для ускорения поиска
-- - statistics - статистика работы системы
--
-- Использование:
-- psql -U postgres -f create_pg_db.sql
-- psql -U postgres -d Venera -f create_pg_db.sql

-- Создание базы данных (выполняется от имени postgres)
-- CREATE DATABASE Venera;

-- Подключение к базе данных
-- \c Venera

-- Создание расширения для работы с UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создание таблицы для хранения идентификаторов
CREATE TABLE IF NOT EXISTS data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    date_first TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    date_last TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_data_source ON data(source);
CREATE INDEX IF NOT EXISTS idx_data_key ON data(key);
CREATE INDEX IF NOT EXISTS idx_data_value ON data(value);
CREATE INDEX IF NOT EXISTS idx_data_date_first ON data(date_first);
CREATE INDEX IF NOT EXISTS idx_data_date_last ON data(date_last);
CREATE INDEX IF NOT EXISTS idx_data_created_at ON data(created_at);

-- Индекс для уникальности пары ключ-значение
CREATE UNIQUE INDEX IF NOT EXISTS idx_data_key_value ON data(key, value);

-- Создание таблицы для хранения статистики
CREATE TABLE IF NOT EXISTS statistics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    process_id VARCHAR(50) NOT NULL,
    process_name VARCHAR(255),
    metric_name VARCHAR(100) NOT NULL,
    metric_value NUMERIC(18, 4) NOT NULL,
    metric_type VARCHAR(20) NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для таблицы статистики
CREATE INDEX IF NOT EXISTS idx_statistics_process_id ON statistics(process_id);
CREATE INDEX IF NOT EXISTS idx_statistics_metric_name ON statistics(metric_name);
CREATE INDEX IF NOT EXISTS idx_statistics_timestamp ON statistics(timestamp);

-- Создание таблицы для хранения метаданных
CREATE TABLE IF NOT EXISTS metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    value TEXT,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для хранения логов
CREATE TABLE IF NOT EXISTS logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    source VARCHAR(100),
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- Индексы для таблицы логов
CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);

-- Создание таблицы для хранения конфигурации
CREATE TABLE IF NOT EXISTS config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    section VARCHAR(100) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT,
    description TEXT,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для уникальности секция-ключ
CREATE UNIQUE INDEX IF NOT EXISTS idx_config_section_key ON config(section, key);

-- Создание таблицы для хранения алертов
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_id VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    condition TEXT,
    action TEXT,
    enabled BOOLEAN DEFAULT true,
    triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для таблицы алертов
CREATE INDEX IF NOT EXISTS idx_alerts_alert_id ON alerts(alert_id);
CREATE INDEX IF NOT EXISTS idx_alerts_enabled ON alerts(enabled);

-- Создание функции для обновления timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создание триггеров для обновления timestamp
CREATE TRIGGER update_data_updated_at
    BEFORE UPDATE ON data
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_metadata_updated_at
    BEFORE UPDATE ON metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_config_updated_at
    BEFORE UPDATE ON config
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Создание представления для статистики по ключам
CREATE OR REPLACE VIEW v_key_statistics AS
SELECT 
    key,
    COUNT(*) AS total_count,
    MIN(date_first) AS first_seen,
    MAX(date_last) AS last_seen,
    COUNT(DISTINCT source) AS sources_count
FROM data
GROUP BY key;

-- Создание представления для статистики по источникам
CREATE OR REPLACE VIEW v_source_statistics AS
SELECT 
    source,
    COUNT(*) AS total_count,
    MIN(date_first) AS first_seen,
    MAX(date_last) AS last_seen,
    COUNT(DISTINCT key) AS keys_count
FROM data
GROUP BY source;

-- Создание представления для последних записей
CREATE OR REPLACE VIEW v_latest_records AS
SELECT 
    id,
    source,
    key,
    value,
    date_last,
    created_at
FROM data
WHERE date_last > CURRENT_TIMESTAMP - INTERVAL '1 hour'
ORDER BY date_last DESC
LIMIT 100;

-- Создание функции для добавления записи с обновлением date_last
CREATE OR REPLACE FUNCTION add_data(
    p_source VARCHAR,
    p_key VARCHAR,
    p_value TEXT
)
RETURNS UUID AS $$
DECLARE
    v_id UUID;
    v_existing_id UUID;
BEGIN
    -- Проверяем наличие существующей записи
    SELECT id INTO v_existing_id
    FROM data
    WHERE key = p_key AND value = p_value
    LIMIT 1;

    IF v_existing_id IS NOT NULL THEN
        -- Обновляем существующую запись
        UPDATE data
        SET date_last = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = v_existing_id;
        RETURN v_existing_id;
    ELSE
        -- Создаем новую запись
        INSERT INTO data (source, key, value, date_first, date_last)
        VALUES (p_source, p_key, p_value, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id INTO v_id;
        RETURN v_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Создание функции для массового добавления записей
CREATE OR REPLACE FUNCTION add_data_batch(
    p_source VARCHAR,
    p_data JSONB
)
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER := 0;
    v_item JSONB;
    v_key VARCHAR;
    v_value TEXT;
BEGIN
    FOR v_item IN SELECT * FROM jsonb_array_elements(p_data)
    LOOP
        v_key := v_item->>'key';
        v_value := v_item->>'value';
        
        IF v_key IS NOT NULL AND v_value IS NOT NULL THEN
            PERFORM add_data(p_source, v_key, v_value);
            v_count := v_count + 1;
        END IF;
    END LOOP;
    
    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- Гранты на таблицы
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO venera;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO venera;

-- Примеры использования:
-- SELECT add_data('source1', 'key1', 'value1');
-- SELECT add_data_batch('source1', '[{"key":"k1","value":"v1"},{"key":"k2","value":"v2"}]');
-- SELECT * FROM v_key_statistics;
-- SELECT * FROM v_source_statistics;
-- SELECT * FROM v_latest_records;

// cfg_loader.go - Загрузка и управление конфигурацией
//
// Этот модуль обеспечивает загрузку, валидацию и управление конфигурацией приложения Venera.
// Поддерживает загрузку из config.toml и processes.toml.
//
// Основные функции:
// - Загрузка конфигурации из TOML файлов
// - Валидация обязательных параметров
// - Установка значений по умолчанию
// - Поддержка всех разделов конфигурации из технического задания
//
// Использование:
// import "venera/config"
// cfg, err := LoadConfig("config.toml")
// if err != nil { log.Fatal(err) }
// fmt.Println(cfg.Generic.Mode)

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config - основная структура конфигурации
type Config struct {
	Generic     GenericConfig     `toml:"Generic"`
	Paths       PathsConfig       `toml:"Paths"`
	DragonflyDB DragonflyDBConfig `toml:"DragonflyDB"`
	PostgreSQL  PostgreSQLConfig  `toml:"PostgreSQL"`
	Logging     LoggingConfig     `toml:"Logging"`
}

// GenericConfig - общий раздел конфигурации
type GenericConfig struct {
	Mode               string `toml:"mode"`
	AutoStartProcesses bool   `toml:"auto_start_processes"`
	MaxProcesses       int    `toml:"max_processes"`
	WebServerPort      int    `toml:"web_server_port"`
	DiskWarningPercent int    `toml:"disk_warning_percent"`
	RAMWarningPercent  int    `toml:"ram_warning_percent"`
	MaxQueueSize       int    `toml:"max_queue_size"`
	ProcessingInterval int    `toml:"processing_interval"`
}

// PathsConfig - раздел путей
type PathsConfig struct {
	PodmanPath         string `toml:"podman_path"`
	TsharkPath         string `toml:"tshark_path"`
	FilterFile         string `toml:"filter_file"`
	ControlFile        string `toml:"control_file"`
	AlertsFile         string `toml:"alerts_file"`
	DragonflyImage     string `toml:"dragonfly_image"`
	DragonflyBackupPath string `toml:"dragonfly_backup_path"`
}

// DragonflyDBConfig - параметры DragonflyDB
type DragonflyDBConfig struct {
	Host        string `toml:"host"`
	Port        int    `toml:"port"`
	Password    string `toml:"password"`
	Database    int    `toml:"database"`
	BatchSize   int    `toml:"batch_size"`
	Timeout     int    `toml:"timeout"`
}

// PostgreSQLConfig - параметры PostgreSQL
type PostgreSQLConfig struct {
	Host            string `toml:"host"`
	Port            int    `toml:"port"`
	Database        string `toml:"database"`
	User            string `toml:"user"`
	Password        string `toml:"password"`
	MaxConnections  int    `toml:"max_connections"`
}

// LoggingConfig - параметры логирования
type LoggingConfig struct {
	Level           string `toml:"level"`
	Directory       string `toml:"directory"`
	MaxAgeDays      int    `toml:"max_age_days"`
	MaxFiles        int    `toml:"max_files"`
	CompressOldLogs bool   `toml:"compress_old_logs"`
}

// ProcessConfig - конфигурация процесса
type ProcessConfig struct {
	ID                 string `toml:"id"`
	Type               string `toml:"type"`
	Name               string `toml:"name"`
	IP                 string `toml:"ip"`
	Port               int    `toml:"port"`
	Path               string `toml:"path"`
	ScanSubfolders     bool   `toml:"scan_subfolders"`
	MonitorNewFiles    bool   `toml:"monitor_new_files"`
}

// ProcessConfigs - коллекция конфигураций процессов
type ProcessConfigs struct {
	Processes map[string]ProcessConfig `toml:"process"`
}

// LoadConfig - загрузка основной конфигурации
func LoadConfig(configPath string) (*Config, error) {
	// Проверка существования файла
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл конфигурации не найден: %s", configPath)
	}

	// Загрузка конфигурации
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	// Валидация и установка значений по умолчанию
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadProcesses - загрузка конфигурации процессов
func LoadProcesses(processesPath string) (*ProcessConfigs, error) {
	// Проверка существования файла
	if _, err := os.Stat(processesPath); os.IsNotExist(err) {
		return &ProcessConfigs{Processes: make(map[string]ProcessConfig)}, nil
	}

	// Загрузка конфигурации
	var cfg ProcessConfigs
	if _, err := toml.DecodeFile(processesPath, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка загрузки конфигурации процессов: %w", err)
	}

	return &cfg, nil
}

// Validate - валидация конфигурации
func (c *Config) Validate() error {
	// Валидация Generic
	if c.Generic.Mode != "tray" && c.Generic.Mode != "service" {
		return fmt.Errorf("недопустимый режим работы: %s (должен быть 'tray' или 'service')", c.Generic.Mode)
	}

	if c.Generic.MaxProcesses <= 0 || c.Generic.MaxProcesses > 20 {
		return fmt.Errorf("недопустимое максимальное количество процессов: %d (должно быть от 1 до 20)", c.Generic.MaxProcesses)
	}

	if c.Generic.WebServerPort <= 0 || c.Generic.WebServerPort > 65535 {
		return fmt.Errorf("недопустимый номер порта веб-сервера: %d", c.Generic.WebServerPort)
	}

	if c.Generic.DiskWarningPercent < 0 || c.Generic.DiskWarningPercent > 100 {
		return fmt.Errorf("недопустимый процент предупреждения о диске: %d", c.Generic.DiskWarningPercent)
	}

	if c.Generic.RAMWarningPercent < 0 || c.Generic.RAMWarningPercent > 100 {
		return fmt.Errorf("недопустимый процент предупреждения о RAM: %d", c.Generic.RAMWarningPercent)
	}

	// Валидация Paths
	if c.Paths.PodmanPath == "" {
		return fmt.Errorf("не указан путь к podman.exe")
	}

	if c.Paths.TsharkPath == "" {
		return fmt.Errorf("не указан путь к tshark.exe")
	}

	if c.Paths.FilterFile == "" {
		c.Paths.FilterFile = "settings\\generic.flt"
	}

	if c.Paths.ControlFile == "" {
		c.Paths.ControlFile = "settings\\generic.ctr"
	}

	if c.Paths.AlertsFile == "" {
		c.Paths.AlertsFile = "settings\\generic.alr"
	}

	if c.Paths.DragonflyImage == "" {
		c.Paths.DragonflyImage = "docker.io/dragonflydb/dragonfly:latest"
	}

	if c.Paths.DragonflyBackupPath == "" {
		c.Paths.DragonflyBackupPath = "C:\\Venera\\dragonfly_backup"
	}

	// Валидация DragonflyDB
	if c.DragonflyDB.Host == "" {
		c.DragonflyDB.Host = "localhost"
	}

	if c.DragonflyDB.Port == 0 {
		c.DragonflyDB.Port = 6379
	}

	if c.DragonflyDB.Database < 0 || c.DragonflyDB.Database > 15 {
		c.DragonflyDB.Database = 0
	}

	if c.DragonflyDB.BatchSize <= 0 {
		c.DragonflyDB.BatchSize = 1000
	}

	if c.DragonflyDB.Timeout <= 0 {
		c.DragonflyDB.Timeout = 30
	}

	// Валидация PostgreSQL
	if c.PostgreSQL.Host == "" {
		c.PostgreSQL.Host = "localhost"
	}

	if c.PostgreSQL.Port == 0 {
		c.PostgreSQL.Port = 5432
	}

	if c.PostgreSQL.Database == "" {
		c.PostgreSQL.Database = "Venera"
	}

	if c.PostgreSQL.User == "" {
		c.PostgreSQL.User = "venera"
	}

	if c.PostgreSQL.MaxConnections <= 0 {
		c.PostgreSQL.MaxConnections = 20
	}

	// Валидация Logging
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}

	if c.Logging.Directory == "" {
		c.Logging.Directory = "Logs"
	}

	if c.Logging.MaxAgeDays <= 0 {
		c.Logging.MaxAgeDays = 30
	}

	if c.Logging.MaxFiles <= 0 {
		c.Logging.MaxFiles = 10
	}

	return nil
}

// GetAbsolutePath - получение абсолютного пути относительно директории конфигурации
func (c *Config) GetAbsolutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	// Получение директории конфигурации
	configDir := "."
	if cfgPath := os.Getenv("VENERA_CONFIG_PATH"); cfgPath != "" {
		configDir = filepath.Dir(cfgPath)
	}

	return filepath.Join(configDir, path)
}

// GetFilterFilePath - получение абсолютного пути к файлу фильтрации
func (c *Config) GetFilterFilePath() string {
	return c.GetAbsolutePath(c.Paths.FilterFile)
}

// GetControlFilePath - получение абсолютного пути к файлу контроля
func (c *Config) GetControlFilePath() string {
	return c.GetAbsolutePath(c.Paths.ControlFile)
}

// GetAlertsFilePath - получение абсолютного пути к файлу алертов
func (c *Config) GetAlertsFilePath() string {
	return c.GetAbsolutePath(c.Paths.AlertsFile)
}

// GetProcessesPath - получение пути к файлу процессов
func (c *Config) GetProcessesPath() string {
	return "processes.toml"
}

// SaveConfig - сохранение конфигурации в файл
func (c *Config) SaveConfig(configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла конфигурации: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("ошибка записи конфигурации: %w", err)
	}

	return nil
}

// SaveProcesses - сохранение конфигурации процессов в файл
func (c *ProcessConfigs) SaveProcesses(processesPath string) error {
	file, err := os.Create(processesPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла процессов: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("ошибка записи конфигурации процессов: %w", err)
	}

	return nil
}

// LoadFilterFile - загрузка файла фильтрации
func LoadFilterFile(filePath string) (map[string][]string, error) {
	// Открытие файла
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла фильтрации: %w", err)
	}
	defer file.Close()

	// Карта для хранения фильтров: ключ -> [значения]
	filters := make(map[string][]string)
	currentKey := ""

	// Чтение файла построчно
	buf := make([]byte, 65536)
	for {
		n, err := file.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("ошибка чтения файла фильтрации: %w", err)
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
					filters[currentKey] = []string{}
				}
			} else if strings.HasPrefix(line, "-") {
				// Черный список значений
				if currentKey != "" {
					value := strings.TrimSpace(line[1:])
					filters[currentKey] = append(filters[currentKey], value)
				}
			}
		}
	}

	return filters, nil
}

// LoadControlFile - загрузка файла контроля (CSV)
func LoadControlFile(filePath string) (map[string]string, error) {
	// Открытие файла
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла контроля: %w", err)
	}
	defer file.Close()

	// Карта для хранения контрольных значений
	controls := make(map[string]string)

	// Чтение файла построчно
	buf := make([]byte, 65536)
	for {
		n, err := file.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("ошибка чтения файла контроля: %w", err)
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

			parts := strings.SplitN(line, "|", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				controls[key] = value
			}
		}
	}

	return controls, nil
}

// LoadAlertsFile - загрузка файла алертов (CEL)
func LoadAlertsFile(filePath string) (map[string]map[string]interface{}, error) {
	// Открытие файла
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла алертов: %w", err)
	}
	defer file.Close()

	// Карта для хранения алертов
	alerts := make(map[string]map[string]interface{})
	currentAlertID := ""

	// Чтение файла построчно
	buf := make([]byte, 65536)
	for {
		n, err := file.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("ошибка чтения файла алертов: %w", err)
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

			// Проверка на начало нового алерта
			if strings.HasPrefix(line, "alert_id") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					currentAlertID = strings.TrimSpace(parts[1])
					alerts[currentAlertID] = make(map[string]interface{})
				}
			} else if currentAlertID != "" && strings.Contains(line, "=") {
				// Параметр алерта
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					alerts[currentAlertID][key] = value
				}
			}
		}
	}

	return alerts, nil
}

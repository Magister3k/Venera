// procs_loader.go - Модуль загрузки конфигурации процессов
//
// Этот модуль обеспечивает загрузку и управление конфигурацией процессов
// из файла processes.toml. Поддерживает загрузку параметров для каждого
// процесса: тип источника, название, IP-адрес, UDP-порт, путь к файлам,
// режимы сканирования подпапок и мониторинга новых файлов.
//
// Основные функции:
// - Загрузка конфигурации процессов из TOML файла
// - Валидация параметров процессов
// - Установка значений по умолчанию
// - Поддержка всех типов источников (сетевой, папка, файл)
// - Интеграция с менеджером процессов
//
// Использование:
// import "venera/processes"
//
// loader := processes.NewProcLoader()
// configs, err := loader.LoadProcesses("processes.toml")
// if err != nil { log.Fatal(err) }
// for id, config := range configs.Processes {
//     fmt.Printf("Process %s: %s\n", id, config.Name)
// }

package processes

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// ProcLoader - загрузчик конфигурации процессов
type ProcLoader struct {
	defaultConfig ProcessConfig
}

// NewProcLoader - создание нового загрузчика процессов
func NewProcLoader() *ProcLoader {
	return &ProcLoader{
		defaultConfig: ProcessConfig{
			ScanSubfolders: true,
			MonitorNewFiles: true,
			BatchSize:      1000,
			Timeout:        30,
		},
	}
}

// LoadProcesses - загрузка конфигурации процессов из файла
// filePath - путь к файлу processes.toml
func (l *ProcLoader) LoadProcesses(filePath string) (*ProcessConfigs, error) {
	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Файл не найден, возвращаем пустую конфигурацию
		return &ProcessConfigs{
			Processes: make(map[string]ProcessConfig),
		}, nil
	}

	// Загрузка конфигурации
	var cfg ProcessConfigs
	if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка загрузки конфигурации процессов: %w", err)
	}

	// Валидация и установка значений по умолчанию
	if err := l.validateAndSetDefaults(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateAndSetDefaults - валидация конфигурации и установка значений по умолчанию
func (l *ProcLoader) validateAndSetDefaults(cfg *ProcessConfigs) error {
	// Проход по всем процессам
	for id, proc := range cfg.Processes {
		// Установка ID процесса, если он пустой
		if proc.ID == "" {
			proc.ID = id
		}

		// Валидация типа источника
		if proc.Type != "network" && proc.Type != "folder" && proc.Type != "file" {
			return fmt.Errorf("недопустимый тип источника для процесса %s: %s", id, proc.Type)
		}

		// Установка значений по умолчанию для сетевого источника
		if proc.Type == "network" {
			if proc.IP == "" {
				return fmt.Errorf("не указан IP-адрес для сетевого процесса %s", id)
			}
			if proc.Port <= 0 || proc.Port > 65535 {
				return fmt.Errorf("недопустимый порт для сетевого процесса %s: %d", id, proc.Port)
			}
			if proc.Name == "" {
				proc.Name = fmt.Sprintf("Network Stream %s", id)
			}
		}

		// Установка значений по умолчанию для папки
		if proc.Type == "folder" {
			if proc.Path == "" {
				return fmt.Errorf("не указан путь к папке для процесса %s", id)
			}
			if proc.Name == "" {
				proc.Name = fmt.Sprintf("Folder Monitor %s", id)
			}
			// Установка значений по умолчанию для режимов
			if !proc.ScanSubfolders {
				proc.ScanSubfolders = l.defaultConfig.ScanSubfolders
			}
			if !proc.MonitorNewFiles {
				proc.MonitorNewFiles = l.defaultConfig.MonitorNewFiles
			}
		}

		// Установка значений по умолчанию для отдельного файла
		if proc.Type == "file" {
			if proc.Path == "" {
				return fmt.Errorf("не указан путь к файлу для процесса %s", id)
			}
			if proc.Name == "" {
				proc.Name = fmt.Sprintf("File Monitor %s", id)
			}
		}

		// Установка значений по умолчанию для batch size и timeout
		if proc.BatchSize <= 0 {
			proc.BatchSize = l.defaultConfig.BatchSize
		}
		if proc.Timeout <= 0 {
			proc.Timeout = l.defaultConfig.Timeout
		}

		// Сохранение обновленного процесса
		cfg.Processes[id] = proc
	}

	return nil
}

// SaveProcesses - сохранение конфигурации процессов в файл
// filePath - путь к файлу processes.toml
func (l *ProcLoader) SaveProcesses(cfg *ProcessConfigs, filePath string) error {
	// Открытие файла для записи
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла процессов: %w", err)
	}
	defer file.Close()

	// Запись конфигурации
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("ошибка записи конфигурации процессов: %w", err)
	}

	return nil
}

// AddProcess - добавление нового процесса в конфигурацию
// id - уникальный ID процесса
// config - конфигурация процесса
func (l *ProcLoader) AddProcess(cfg *ProcessConfigs, id string, config ProcessConfig) error {
	// Проверка существования ID
	if _, exists := cfg.Processes[id]; exists {
		return fmt.Errorf("процесс с ID '%s' уже существует", id)
	}

	// Валидация процесса
	if err := l.validateProcess(id, config); err != nil {
		return err
	}

	// Добавление процесса
	cfg.Processes[id] = config

	return nil
}

// validateProcess - валидация конфигурации процесса
func (l *ProcLoader) validateProcess(id string, config ProcessConfig) error {
	// Валидация типа источника
	if config.Type != "network" && config.Type != "folder" && config.Type != "file" {
		return fmt.Errorf("недопустимый тип источника для процесса %s: %s", id, config.Type)
	}

	// Установка ID процесса
	if config.ID == "" {
		config.ID = id
	}

	// Валидация для сетевого источника
	if config.Type == "network" {
		if config.IP == "" {
			return fmt.Errorf("не указан IP-адрес для сетевого процесса %s", id)
		}
		if config.Port <= 0 || config.Port > 65535 {
			return fmt.Errorf("недопустимый порт для сетевого процесса %s: %d", id, config.Port)
		}
	}

	// Валидация для папки и файла
	if config.Type == "folder" || config.Type == "file" {
		if config.Path == "" {
			return fmt.Errorf("не указан путь для процесса %s", id)
		}
	}

	return nil
}

// RemoveProcess - удаление процесса из конфигурации
// id - ID процесса для удаления
func (l *ProcLoader) RemoveProcess(cfg *ProcessConfigs, id string) error {
	if _, exists := cfg.Processes[id]; !exists {
		return fmt.Errorf("процесс с ID '%s' не найден", id)
	}

	delete(cfg.Processes, id)

	return nil
}

// UpdateProcess - обновление конфигурации процесса
// id - ID процесса для обновления
// config - новые параметры процесса
func (l *ProcLoader) UpdateProcess(cfg *ProcessConfigs, id string, config ProcessConfig) error {
	if _, exists := cfg.Processes[id]; !exists {
		return fmt.Errorf("процесс с ID '%s' не найден", id)
	}

	// Валидация процесса
	if err := l.validateProcess(id, config); err != nil {
		return err
	}

	// Обновление процесса
	cfg.Processes[id] = config

	return nil
}

// GetProcess - получение конфигурации процесса по ID
// id - ID процесса
func (l *ProcLoader) GetProcess(cfg *ProcessConfigs, id string) (*ProcessConfig, error) {
	if proc, exists := cfg.Processes[id]; exists {
		return &proc, nil
	}
	return nil, fmt.Errorf("процесс с ID '%s' не найден", id)
}

// GetProcessCount - получение количества процессов
func (l *ProcLoader) GetProcessCount(cfg *ProcessConfigs) int {
	return len(cfg.Processes)
}

// GetProcessIDs - получение списка ID всех процессов
func (l *ProcLoader) GetProcessIDs(cfg *ProcessConfigs) []string {
	ids := make([]string, 0, len(cfg.Processes))
	for id := range cfg.Processes {
		ids = append(ids, id)
	}
	return ids
}

// GetNetworkProcesses - получение всех сетевых процессов
func (l *ProcLoader) GetNetworkProcesses(cfg *ProcessConfigs) map[string]ProcessConfig {
	networkProcs := make(map[string]ProcessConfig)
	for id, proc := range cfg.Processes {
		if proc.Type == "network" {
			networkProcs[id] = proc
		}
	}
	return networkProcs
}

// GetFolderProcesses - получение всех процессов для папок
func (l *ProcLoader) GetFolderProcesses(cfg *ProcessConfigs) map[string]ProcessConfig {
	folderProcs := make(map[string]ProcessConfig)
	for id, proc := range cfg.Processes {
		if proc.Type == "folder" {
			folderProcs[id] = proc
		}
	}
	return folderProcs
}

// GetFileProcesses - получение всех процессов для файлов
func (l *ProcLoader) GetFileProcesses(cfg *ProcessConfigs) map[string]ProcessConfig {
	fileProcs := make(map[string]ProcessConfig)
	for id, proc := range cfg.Processes {
		if proc.Type == "file" {
			fileProcs[id] = proc
		}
	}
	return fileProcs
}

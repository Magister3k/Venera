// file_logger.go - Модуль логирования в файлы с ротацией
//
// Этот модуль обеспечивает логирование в файлы с автоматической ротацией,
// сжатием старых логов и управлением именованием файлов.
//
// Основные функции:
// - Автоматическая ротация лог-файлов
// - Сжатие старых логов в формат gzip
// - Управление именами файлов по формату YYYY-MM-DD_HH-MM-SS_N.log
// - Настройка параметров ротации из конфигурации
//
// Использование:
// import "venera/logging"
// logger := logging.NewFileLogger("logs", config.Logging)
// logger.Write(logrus.InfoLevel, "Сообщение")

package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// FileLogger - структура файлового логгера
type FileLogger struct {
	logDir        string
	level         logrus.Level
	maxAgeDays    int
	maxFiles      int
	currentFile   *os.File
	currentPath   string
	mu            sync.Mutex
	logPattern    *regexp.Regexp
}

// LoggingConfig - конфигурация логирования
type LoggingConfig struct {
	Level           string `toml:"level"`
	Directory       string `toml:"directory"`
	MaxAgeDays      int    `toml:"max_age_days"`
	MaxFiles        int    `toml:"max_files"`
	CompressOldLogs bool   `toml:"compress_old_logs"`
}

// NewFileLogger - создание нового файлового логгера
func NewFileLogger(logDir string, config LoggingConfig) (*FileLogger, error) {
	// Создание директории, если она не существует
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка создания директории логов: %w", err)
	}

	// Создание логгера
	logger := &FileLogger{
		logDir:      logDir,
		level:       logrus.InfoLevel,
		maxAgeDays:  config.MaxAgeDays,
		maxFiles:    config.MaxFiles,
		logPattern:  regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}_\d+\.log$`),
	}

	// Открытие текущего лог-файла
	if err := logger.openNewFile(); err != nil {
		return nil, fmt.Errorf("ошибка открытия лог-файла: %w", err)
	}

	return logger, nil
}

// openNewFile - открытие нового лог-файла
func (l *FileLogger) openNewFile() error {
	// Закрытие предыдущего файла
	if l.currentFile != nil {
		l.currentFile.Close()
	}

	// Генерация имени файла
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s_%d.log", timestamp, time.Now().UnixNano()%1000)
	filePath := filepath.Join(l.logDir, fileName)

	// Открытие файла
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}

	l.currentFile = file
	l.currentPath = filePath

	return nil
}

// Write - запись сообщения в лог
func (l *FileLogger) Write(level logrus.Level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Проверка уровня логирования
	if level < l.level {
		return
	}

	// Формирование строки
	timestamp := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] %s: %s\n", timestamp, level.String(), msg)

	// Запись в файл
	if l.currentFile != nil {
		l.currentFile.WriteString(line)
		l.currentFile.Sync()
	}

	// Проверка необходимости ротации
	if err := l.checkRotation(); err != nil {
		// Логируем ошибку, но не прерываем работу
		fmt.Fprintf(os.Stderr, "Ошибка ротации логов: %v\n", err)
	}
}

// checkRotation - проверка необходимости ротации
func (l *FileLogger) checkRotation() error {
	// Ограничение количества файлов
	files, err := l.getLogFiles()
	if err != nil {
		return err
	}

	if len(files) > l.maxFiles {
		// Удаление старых файлов
		for i := 0; i < len(files)-l.maxFiles; i++ {
			if err := os.Remove(files[i]); err != nil {
				// Пытаемся удалить сжатый файл
				gzFile := files[i] + ".gz"
				if err := os.Remove(gzFile); err != nil {
					return fmt.Errorf("ошибка удаления файла: %w", err)
				}
			}
		}
	}

	// Проверка возраста файлов
	if l.maxAgeDays > 0 {
		now := time.Now()
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}

			if now.Sub(info.ModTime()).Hours() > float64(l.maxAgeDays*24) {
				// Удаление старого файла
				if err := os.Remove(file); err != nil {
					// Пытаемся удалить сжатый файл
					gzFile := file + ".gz"
					if err := os.Remove(gzFile); err != nil {
						return fmt.Errorf("ошибка удаления старого файла: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// getLogFiles - получение списка лог-файлов
func (l *FileLogger) getLogFiles() ([]string, error) {
	files, err := os.ReadDir(l.logDir)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения директории логов: %w", err)
	}

	var logFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		// Проверка соответствия паттерну
		if l.logPattern.MatchString(name) || strings.HasSuffix(name, ".log.gz") {
			logFiles = append(logFiles, filepath.Join(l.logDir, name))
		}
	}

	// Сортировка по времени модификации
	sort.Slice(logFiles, func(i, j int) bool {
		infoI, _ := os.Stat(logFiles[i])
		infoJ, _ := os.Stat(logFiles[j])
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	return logFiles, nil
}

// Close - закрытие логгера
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentFile != nil {
		if err := l.currentFile.Close(); err != nil {
			return fmt.Errorf("ошибка закрытия файла: %w", err)
		}
		l.currentFile = nil
	}

	return nil
}

// GetCurrentFilePath - получение пути к текущему файлу
func (l *FileLogger) GetCurrentFilePath() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.currentPath
}

// CompressOldLogs - сжатие старых логов
func (l *FileLogger) CompressOldLogs() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	files, err := l.getLogFiles()
	if err != nil {
		return err
	}

	// Сжатие файлов старше одного дня
	now := time.Now()
	for _, file := range files {
		if strings.HasSuffix(file, ".gz") {
			continue
		}

		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()).Hours() > 24 {
			// Сжатие файла
			if err := l.compressFile(file); err != nil {
				return fmt.Errorf("ошибка сжатия файла %s: %w", file, err)
			}
		}
	}

	return nil
}

// compressFile - сжатие одного файла
func (l *FileLogger) compressFile(filePath string) error {
	// Открытие исходного файла
	srcFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer srcFile.Close()

	// Создание сжатого файла
	gzFilePath := filePath + ".gz"
	dstFile, err := os.Create(gzFilePath)
	if err != nil {
		return fmt.Errorf("ошибка создания сжатого файла: %w", err)
	}
	defer dstFile.Close()

	// Создание gzip writer
	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	// Копирование данных
	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		return fmt.Errorf("ошибка копирования данных: %w", err)
	}

	// Удаление исходного файла
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("ошибка удаления исходного файла: %w", err)
	}

	return nil
}

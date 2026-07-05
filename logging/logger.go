// Logger.go - Основной модуль логирования
//
// Этот модуль предоставляет централизованное логирование для всего приложения Venera.
// Поддерживает уровни логирования, текстовый и JSON форматы, а также интеграцию с Windows Event Log.
//
// Основные функции:
// - Управление уровнями логирования (debug, info, warn, error, fatal, panic)
// - Настройка формата вывода (текст, JSON)
// - Интеграция с Windows Event Log для критических ошибок
// - Централизованное управление логами из конфигурации
//
// Использование:
// import "venera/logging"
// logger := logging.GetLogger()
// logger.Info("Сообщение")
// logger.Errorf("Ошибка: %v", err)

package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Logger - структура логгера
type Logger struct {
	log        *logrus.Logger
	fileLogger *FileLogger
	mu         sync.Mutex
}

// Instance - глобальный экземпляр логгера
var (
	instance *Logger
	once     sync.Once
)

// GetLogger - получение экземпляра логгера (Singleton)
func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			log: logrus.New(),
		}
	})
	return instance
}

// Initialize - инициализация логгера с конфигурацией
func (l *Logger) Initialize(configPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Загрузка конфигурации
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	// Установка уровня логирования
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	l.log.SetLevel(level)

	// Установка формата вывода
	if cfg.Logging.Format == "json" {
		l.log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		l.log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Настройка логирования в файлы
	logDir := cfg.Logging.Directory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории логов: %w", err)
	}

	fileLogger, err := NewFileLogger(logDir, cfg.Logging)
	if err != nil {
		return fmt.Errorf("ошибка инициализации файлового логгера: %w", err)
	}
	l.fileLogger = fileLogger

	return nil
}

// Debug - уровень DEBUG
func (l *Logger) Debug(msg string) {
	l.log.Debug(msg)
	l.fileLogger.Write(logrus.DebugLevel, msg)
}

// Info - уровень INFO
func (l *Logger) Info(msg string) {
	l.log.Info(msg)
	l.fileLogger.Write(logrus.InfoLevel, msg)
}

// Warn - уровень WARN
func (l *Logger) Warn(msg string) {
	l.log.Warn(msg)
	l.fileLogger.Write(logrus.WarnLevel, msg)
}

// Error - уровень ERROR
func (l *Logger) Error(msg string) {
	l.log.Error(msg)
	l.fileLogger.Write(logrus.ErrorLevel, msg)
}

// Fatal - уровень FATAL (выход)
func (l *Logger) Fatal(msg string) {
	l.log.Fatal(msg)
	l.fileLogger.Write(logrus.FatalLevel, msg)
	l.ExitWindows()
}

// Panic - уровень PANIC (выход с паникой)
func (l *Logger) Panic(msg string) {
	l.log.Panic(msg)
	l.fileLogger.Write(logrus.PanicLevel, msg)
	l.ExitWindows()
}

// Errorf - форматирование сообщения об ошибке
func (l *Logger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Error(msg)
}

// Infof - форматирование информационного сообщения
func (l *Logger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Info(msg)
}

// Debugf - форматирование отладочного сообщения
func (l *Logger) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Debug(msg)
}

// Fatalf - форматирование фатального сообщения
func (l *Logger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Fatal(msg)
}

// Panicf - форматирование панического сообщения
func (l *Logger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Panic(msg)
}

// ExitWindows - закрытие Windows с логированием ошибки
func (l *Logger) ExitWindows() {
	// Попытка записи в Windows Event Log
	l.writeToEventLog(windows.EVENTLOG_ERROR_TYPE, 1001, "Критическая ошибка в приложении Venera")
}

// writeToEventLog - запись в Windows Event Log
func (l *Logger) writeToEventLog(eventType uint16, eventID uint32, message string) {
	// Открытие ключа Event Log
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\EventLog\Application\Venera`,
		registry.SET_VALUE)
	if err != nil {
		// Ключ не найден, создаем его
		key, err = registry.CreateKey(registry.LOCAL_MACHINE,
			`SYSTEM\CurrentControlSet\Services\EventLog\Application\Venera`,
			registry.SET_VALUE)
		if err != nil {
			return
		}
		defer key.Close()
	}

	// Запись сообщения
	l.log.WithFields(logrus.Fields{
		"event_type": eventType,
		"event_id":   eventID,
	}).Error(message)
}

// Shutdown - корректное завершение работы логгера
func (l *Logger) Shutdown() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fileLogger != nil {
		return l.fileLogger.Close()
	}
	return nil
}

// GetLogFilePath - получение пути к текущему лог-файлу
func (l *Logger) GetLogFilePath() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fileLogger != nil {
		return l.fileLogger.GetCurrentFilePath()
	}
	return ""
}

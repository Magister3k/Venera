// scanner.go - Утилиты сканирования для Venera
//
// Этот модуль обеспечивает функции сканирования файлов и папок
// для систем сбора идентификаторов.
//
// Основные функции:
// - Рекурсивное сканирование папок
// - Мониторинг изменений в папках
// - Поиск файлов по шаблону
//
// Использование:
// import "venera/utils"
// files := utils.ScanDirectory("C:\\Data", true, true)
// fmt.Println(files)

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScanDirectory - сканирование директории
func ScanDirectory(dirPath string, scanSubfolders, monitorNewFiles bool) ([]string, error) {
	var files []string

	// Проверка существования директории
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("директория не найдена: %s", dirPath)
	}

	// Рекурсивное сканирование
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропуск директорий, если не нужно сканировать подпапки
		if info.IsDir() {
			if !scanSubfolders {
				return filepath.SkipDir
			}
			return nil
		}

		// Добавление файлов
		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка сканирования: %w", err)
	}

	return files, nil
}

// FindFilesByPattern - поиск файлов по шаблону
func FindFilesByPattern(dirPath, pattern string) ([]string, error) {
	var files []string

	// Рекурсивное сканирование
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && matchesPattern(info.Name(), pattern) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка поиска файлов: %w", err)
	}

	return files, nil
}

// matchesPattern - проверка соответствия имени файла шаблону
func matchesPattern(name, pattern string) bool {
	// Поддержка простых шаблонов
	switch pattern {
	case "*":
		return true
	case "*.log":
		return strings.HasSuffix(name, ".log")
	case "*.json":
		return strings.HasSuffix(name, ".json")
	default:
		return name == pattern
	}
}

// MonitorDirectory - мониторинг директории
func MonitorDirectory(dirPath string, onChange func(string), interval int) error {
	// TODO: Реализовать мониторинг директории
	return nil
}

// GetDirectorySize - получение размера директории
func GetDirectorySize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("ошибка получения размера: %w", err)
	}

	return size, nil
}

// GetFileExtension - получение расширения файла
func GetFileExtension(filePath string) string {
	return filepath.Ext(filePath)
}

// GetFileNameWithoutExtension - получение имени файла без расширения
func GetFileNameWithoutExtension(filePath string) string {
	return strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
}

// EnsureDirectoryExists - создание директории, если она не существует
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// GetFreeDiskSpace - получение свободного места на диске
func GetFreeDiskSpace(path string) (int64, error) {
	// TODO: Реализовать получение свободного места на диске
	return 0, nil
}

// scanner.go - Сканирование файлов и папок
//
// Этот модуль обеспечивает сканирование файловой системы для
// обнаружения новых файлов и изменений в папках.
//
// Основные функции:
// - Рекурсивное сканирование папок
// - Мониторинг новых файлов
// - Получение списка файлов с определенным расширением
//
// Использование:
// import "venera/utils"
// files := ScanFolder("C:\\Logs", true, true)

package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ScanOptions - опции сканирования
type ScanOptions struct {
	ScanSubfolders  bool   // Рекурсивное сканирование подпапок
	MonitorNewFiles bool   // Мониторинг новых файлов
	Extensions      []string // Фильтр по расширениям файлов
}

// ScanFolder - сканирование папки на наличие файлов
func ScanFolder(folderPath string, options *ScanOptions) ([]string, error) {
	var files []string

	// Проверка существования папки
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, err
	}

	// Функция для проверки расширения файла
	shouldInclude := func(filename string) bool {
		if options == nil || len(options.Extensions) == 0 {
			return true
		}

		ext := filepath.Ext(filename)
		for _, allowedExt := range options.Extensions {
			if strings.EqualFold(ext, allowedExt) {
				return true
			}
		}
		return false
	}

	// Рекурсивное обход папки
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропуск папок, если не включен сканирование подпапок
		if info.IsDir() && !options.ScanSubfolders && path != folderPath {
			return filepath.SkipDir
		}

		// Добавление файлов
		if !info.IsDir() && shouldInclude(info.Name()) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// MonitorFolder - мониторинг папки на наличие новых файлов (безопасная версия)
func MonitorFolder(folderPath string, callback func(string)) error {
	// Проверка существования папки
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return err
	}

	// Запуск мониторинга в отдельной горутине
	go func() {
		// Инициализация состояния папки
		seenFiles := make(map[string]bool)

		// Первоначальное сканирование
		files, err := ScanFolder(folderPath, &ScanOptions{
			ScanSubfolders:  true,
			MonitorNewFiles: true,
		})
		if err != nil {
			return
		}

		for _, file := range files {
			seenFiles[file] = true
		}

		// Цикл мониторинга (упрощенная версия)
		for {
			// Повторное сканирование
			currentFiles, err := ScanFolder(folderPath, &ScanOptions{
				ScanSubfolders:  true,
				MonitorNewFiles: true,
			})
			if err != nil {
				continue
			}

			// Поиск новых файлов
			for _, file := range currentFiles {
				if !seenFiles[file] {
					seenFiles[file] = true
					callback(file)
				}
			}

			// Пауза перед следующим сканированием
			// В реальной реализации использовать inotify или аналоги
		}
	}()

	return nil
}

// GetFileExtension - получить расширение файла
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// GetFileBaseName - получить имя файла без расширения
func GetFileBaseName(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// GetFileName - получить имя файла без пути
func GetFileName(fullPath string) string {
	return filepath.Base(fullPath)
}

// GetFolderPath - получить путь к папке
func GetFolderPath(fullPath string) string {
	return filepath.Dir(fullPath)
}

// FileExists - проверка существования файла
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// FolderExists - проверка существования папки
func FolderExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileModificationTime - получить время последней модификации файла
func GetFileModificationTime(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}

// CreateFolderIfNotExists - создать папку, если она не существует
func CreateFolderIfNotExists(path string) error {
	if !FolderExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

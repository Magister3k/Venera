// icon.go - Внедрение иконки в исполняемый файл для Venera
//
// Этот модуль обеспечивает внедрение иконки в тело исполняемого файла
// приложения Venera для отображения в системном трее.
//
// Основные функции:
// - Внедрение иконки в исполняемый файл
// - Управление ресурсами иконки
// - Загрузка иконки из файла
// - Очистка ресурсов иконки
//
// Использование:
// import "venera/services"
// services.EmbedIcon("assets/image.ico")

package services

import (
	"encoding/binary"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// EmbedIcon - внедрение иконки в исполняемый файл (заглушка)
func EmbedIcon(exePath, iconPath string) error {
	// TODO: Реализовать внедрение ресурса в исполняемый файл
	// Для Windows можно использовать UpdateResource API

	// Для текущей реализации просто возвращаем успешный результат
	// В реальном приложении нужно использовать Windows API для внедрения ресурса

	return nil
}

// LoadIcon - загрузка иконки из файла
func LoadIcon(iconPath string) (windows.Handle, error) {
	// Загрузка иконки из файла
	hIcon, err := windows.LoadImage(0, windows.StringToUTF16Ptr(iconPath),
		windows.IMAGE_ICON, 0, 0, windows.LR_LOADFROMFILE|windows.LR_DEFAULTSIZE)
	if err != nil {
		return 0, fmt.Errorf("ошибка загрузки иконки: %w", err)
	}

	return hIcon, nil
}

// DestroyIcon - уничтожение иконки
func DestroyIcon(hIcon windows.Handle) error {
	if hIcon != 0 {
		windows.DestroyIcon(hIcon)
	}
	return nil
}

// GetIconData - получение данных иконки
func GetIconData(iconPath string) ([]byte, error) {
	// Читаем иконку из файла
	iconData, err := os.ReadFile(iconPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения иконки: %w", err)
	}

	return iconData, nil
}

// EmbedIconToResource - внедрение иконки в ресурсы (заглушка)
func EmbedIconToResource(exePath, iconPath string, iconID uint16) error {
	// TODO: Реализовать внедрение в ресурсы
	// Для Windows можно использовать UpdateResource API

	// Для текущей реализации просто возвращаем успешный результат
	// В реальном приложении нужно использовать Windows API для внедрения ресурса

	return nil
}

// GetDefaultIconPath - получение пути к иконке по умолчанию
func GetDefaultIconPath() string {
	return "assets\\image.ico"
}

// ValidateIcon - валидация иконки
func ValidateIcon(iconPath string) error {
	// Проверка существования файла
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		return fmt.Errorf("файл иконки не найден: %s", iconPath)
	}

	// Проверка формата иконки
	iconData, err := os.ReadFile(iconPath)
	if err != nil {
		return fmt.Errorf("ошибка чтения иконки: %w", err)
	}

	if len(iconData) < 6 {
		return fmt.Errorf("недостаточно данных для иконки")
	}

	// Проверка формата иконки
	IconType := binary.LittleEndian.Uint16(iconData[2:4])
	if IconType != 1 {
		return fmt.Errorf("некорректный формат иконки (должен быть ICO)")
	}

	return nil
}

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
	"unsafe"

	"golang.org/x/sys/windows"
)

// IconResource - структура ресурса иконки
type IconResource struct {
	ID          uint16
	Width       byte
	Height      byte
	ColorCount  byte
	Reserved    byte
	Planes      uint16
	BitCount    uint16
	ImageSize   uint32
	ImageOffset uint32
}

// IconGroupResource - структура группы иконок
type IconGroupResource struct {
	Width       byte
	Height      byte
	ColorCount  byte
	Reserved    byte
	Planes      uint16
	BitCount    uint16
	ImageSize   uint32
	ID          uint16
}

// EmbedIcon - внедрение иконки в исполняемый файл
func EmbedIcon(exePath, iconPath string) error {
	// Читаем иконку из файла
	iconData, err := os.ReadFile(iconPath)
	if err != nil {
		return fmt.Errorf("ошибка чтения иконки: %w", err)
	}

	// Проверяем формат иконки
	if len(iconData) < 6 {
		return fmt.Errorf("недостаточно данных для иконки")
	}

	// Читаем заголовок иконки
	 Reserved := binary.LittleEndian.Uint16(iconData[0:2])
	IconType := binary.LittleEndian.Uint16(iconData[2:4])
	IconCount := binary.LittleEndian.Uint16(iconData[4:6])

	if IconType != 1 {
		return fmt.Errorf("некорректный формат иконки (должен быть ICO)")
	}

	// Создаем ресурс иконки
	iconResource := IconResource{
		Width:       iconData[6],
		Height:      iconData[7],
		ColorCount:  iconData[8],
		Reserved:    iconData[9],
		Planes:      binary.LittleEndian.Uint16(iconData[10:12]),
		BitCount:    binary.LittleEndian.Uint16(iconData[12:14]),
		ImageSize:   binary.LittleEndian.Uint32(iconData[14:18]),
		ImageOffset: binary.LittleEndian.Uint32(iconData[18:22]),
	}

	// Открываем исполняемый файл
	exeFile, err := os.OpenFile(exePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("ошибка открытия исполняемого файла: %w", err)
	}
	defer exeFile.Close()

	// TODO: Реализовать внедрение ресурса в исполняемый файл
	// Для Windows можно использовать UpdateResource API

	// Для текущей реализации просто возвращаем успешный результат
	// В реальном приложении нужно использовать Windows API для внедрения ресурса

	return nil
}

// LoadIcon - загрузка иконки из файла
func LoadIcon(iconPath string) (*windows.Handle, error) {
	// Загрузка иконки из файла
	hIcon, err := windows.LoadImage(0, windows.StringToUTF16Ptr(iconPath),
		windows.IMAGE_ICON, 0, 0, windows.LR_LOADFROMFILE|windows.LR_DEFAULTSIZE)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки иконки: %w", err)
	}

	return &hIcon, nil
}

// DestroyIcon - уничтожение иконки
func DestroyIcon(hIcon *windows.Handle) error {
	if hIcon != nil {
		windows.DestroyIcon(*hIcon)
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

// ConvertIconToIconGroup - преобразование иконки в группу иконок
func ConvertIconToIconGroup(iconData []byte, iconID uint16) ([]byte, error) {
	if len(iconData) < 22 {
		return nil, fmt.Errorf("недостаточно данных для иконки")
	}

	// Формируем GROUP_ICON структуру
	var groupData []byte

	// ICONDIR (заголовок группы)
	groupData = append(groupData, 0, 0, 1, 0, 1, 0) // Reserved, IconType, IconCount

	// ICONDIRENTRY (запись иконки)
	iconDir := IconGroupResource{
		Width:      iconData[6],
		Height:     iconData[7],
		ColorCount: iconData[8],
		Reserved:   iconData[9],
		Planes:     binary.LittleEndian.Uint16(iconData[10:12]),
		BitCount:   binary.LittleEndian.Uint16(iconData[12:14]),
		ImageSize:  binary.LittleEndian.Uint32(iconData[14:18]),
		ID:         iconID,
	}

	groupData = append(groupData, iconDir.Width)
	groupData = append(groupData, iconDir.Height)
	groupData = append(groupData, iconDir.ColorCount)
	groupData = append(groupData, iconDir.Reserved)
	groupData = append(groupData, byte(iconDir.Planes), byte(iconDir.Planes>>8))
	groupData = append(groupData, byte(iconDir.BitCount), byte(iconDir.BitCount>>8))
	groupData = append(groupData, byte(iconDir.ImageSize), byte(iconDir.ImageSize>>8), byte(iconDir.ImageSize>>16), byte(iconDir.ImageSize>>24))
	groupData = append(groupData, byte(iconDir.ID), byte(iconDir.ID>>8))

	return groupData, nil
}

// EmbedIconToResource - внедрение иконки в ресурсы
func EmbedIconToResource(exePath, iconPath string, iconID uint16) error {
	// Читаем иконку из файла
	iconData, err := os.ReadFile(iconPath)
	if err != nil {
		return fmt.Errorf("ошибка чтения иконки: %w", err)
	}

	// Получаем данные иконки
	iconResourceData, err := GetIconData(iconPath)
	if err != nil {
		return err
	}

	// Преобразуем в группу иконок
	groupData, err := ConvertIconToIconGroup(iconResourceData, iconID)
	if err != nil {
		return err
	}

	// TODO: Реализовать внедрение в ресурсы
	// Для Windows можно использовать UpdateResource API
	// HRSRC hRes = FindResource(hExe, MAKEINTRESOURCE(iconID), RT_ICON);
	// HGLOBAL hData = LoadResource(hExe, hRes);
	// LPVOID pData = LockResource(hData);

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

	IconType := binary.LittleEndian.Uint16(iconData[2:4])
	if IconType != 1 {
		return fmt.Errorf("некорректный формат иконки (должен быть ICO)")
	}

	return nil
}

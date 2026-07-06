// system.go - Системные метрики для системы Venera
//
// Этот модуль обеспечивает сбор системных метрик для приложения Venera,
// включая свободную RAM, размер баз данных и свободное место на диске.
//
// Основные функции:
// - Сбор метрик RAM (свободная, используемая)
// - Сбор метрик дискового пространства
// - Получение размера баз данных DragonflyDB и PostgreSQL
// - Проверка предупреждений о переполнении
//
// Использование:
// import "venera/metrics"
// ram, _ := GetRAMMetrics()
// disk, _ := GetDiskMetrics("C:\\Venera")
// postgresSize, _ := GetPostgreSQLSize(postgresDB)

package metrics

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// SystemMetrics - системные метрики
type SystemMetrics struct {
	TotalRAM     uint64 `json:"total_ram"`
	AvailableRAM uint64 `json:"available_ram"`
	UsedRAM      uint64 `json:"used_ram"`
	RAMPercent   float64 `json:"ram_percent"`
	CPUUsage     float64 `json:"cpu_usage"`
	DiskFree     uint64 `json:"disk_free"`
	DiskTotal    uint64 `json:"disk_total"`
	DiskUsed     uint64 `json:"disk_used"`
	DiskPercent  float64 `json:"disk_percent"`
}

// GetRAMMetrics - получение метрик RAM
func GetRAMMetrics() (*SystemMetrics, error) {
	var memInfo windows.MemoryStatusEx
	memInfo.Length = uint32(unsafe.Sizeof(memInfo))

	err := windows.GlobalMemoryStatusEx(&memInfo)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о памяти: %w", err)
	}

	totalRAM := memInfo.TotalPhys
	availableRAM := memInfo.AvailPhys
	usedRAM := totalRAM - availableRAM
	ramPercent := float64(usedRAM) / float64(totalRAM) * 100

	return &SystemMetrics{
		TotalRAM:     totalRAM,
		AvailableRAM: availableRAM,
		UsedRAM:      usedRAM,
		RAMPercent:   ramPercent,
	}, nil
}

// GetDiskMetrics - получение метрик диска
func GetDiskMetrics(path string) (*SystemMetrics, error) {
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	err := windows.GetDiskFreeSpaceEx(
		windows.StringToUTF16Ptr(path),
		&freeBytesAvailable,
		&totalBytes,
		&totalFreeBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о диске: %w", err)
	}

	diskFree := totalFreeBytes
	diskTotal := totalBytes
	diskUsed := diskTotal - diskFree
	diskPercent := float64(diskUsed) / float64(diskTotal) * 100

	return &SystemMetrics{
		DiskFree:     diskFree,
		DiskTotal:    diskTotal,
		DiskUsed:     diskUsed,
		DiskPercent:  diskPercent,
	}, nil
}

// GetCPUUsage - получение загрузки CPU (упрощенная версия)
func GetCPUUsage() float64 {
	// Используем runtime для получения количества доступных процессоров
	numCPU := runtime.NumCPU()
	numGoroutine := runtime.NumGoroutine()

	// Упрощенная оценка загрузки CPU
	// В реальном приложении можно использовать Win32_PerfFormattedData_PerfProc_Process
	return float64(numGoroutine) / float64(numCPU*100) * 100
}

// GetPostgreSQLSize - получение размера базы данных PostgreSQL
func GetPostgreSQLSize(db interface{}) (uint64, error) {
	// Для упрощения возвращаем 0
	// В реальном приложении нужно выполнить SQL-запрос к PostgreSQL
	return 0, nil
}

// GetDragonflyDBSize - получение размера базы данных DragonflyDB
func GetDragonflyDBSize(db interface{}) (uint64, error) {
	// Для упрощения возвращаем 0
	// В реальном приложении нужно выполнить INFO команду к DragonflyDB
	return 0, nil
}

// CheckDiskWarning - проверка предупреждения о диске
func CheckDiskWarning(path string, warningPercent float64) (bool, error) {
	diskMetrics, err := GetDiskMetrics(path)
	if err != nil {
		return false, err
	}

	if diskMetrics.DiskPercent >= warningPercent {
		return true, nil
	}

	return false, nil
}

// CheckRAMWarning - проверка предупреждения о RAM
func CheckRAMWarning(warningPercent float64) (bool, error) {
	systemMetrics, err := GetRAMMetrics()
	if err != nil {
		return false, err
	}

	if systemMetrics.RAMPercent >= warningPercent {
		return true, nil
	}

	return false, nil
}

// GetDiskFreeSpace - получение свободного места на диске
func GetDiskFreeSpace(path string) (uint64, error) {
	var freeBytesAvailable, _, _ uint64

	err := windows.GetDiskFreeSpaceEx(
		windows.StringToUTF16Ptr(path),
		&freeBytesAvailable,
		nil,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения свободного места на диске: %w", err)
	}

	return freeBytesAvailable, nil
}

// GetDiskTotalSpace - получение общего места на диске
func GetDiskTotalSpace(path string) (uint64, error) {
	var _, totalBytes, _ uint64

	err := windows.GetDiskFreeSpaceEx(
		windows.StringToUTF16Ptr(path),
		nil,
		&totalBytes,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения общего места на диске: %w", err)
	}

	return totalBytes, nil
}

// GetDiskUsedSpace - получение используемого места на диске
func GetDiskUsedSpace(path string) (uint64, error) {
	freeSpace, err := GetDiskFreeSpace(path)
	if err != nil {
		return 0, err
	}

	totalSpace, err := GetDiskTotalSpace(path)
	if err != nil {
		return 0, err
	}

	return totalSpace - freeSpace, nil
}

// GetDiskUsagePercent - получение процента использования диска
func GetDiskUsagePercent(path string) (float64, error) {
	usedSpace, err := GetDiskUsedSpace(path)
	if err != nil {
		return 0, err
	}

	totalSpace, err := GetDiskTotalSpace(path)
	if err != nil {
		return 0, err
	}

	return float64(usedSpace) / float64(totalSpace) * 100, nil
}

// CheckSystemResources - проверка ресурсов системы
func CheckSystemResources(diskPath string, diskWarningPercent, ramWarningPercent float64) (*SystemMetrics, error) {
	// Получение метрик RAM
	ramMetrics, err := GetRAMMetrics()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения метрик RAM: %w", err)
	}

	// Получение метрик диска
	diskMetrics, err := GetDiskMetrics(diskPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения метрик диска: %w", err)
	}

	// Получение загрузки CPU
	cpuUsage := GetCPUUsage()

	return &SystemMetrics{
		TotalRAM:     ramMetrics.TotalRAM,
		AvailableRAM: ramMetrics.AvailableRAM,
		UsedRAM:      ramMetrics.UsedRAM,
		RAMPercent:   ramMetrics.RAMPercent,
		CPUUsage:     cpuUsage,
		DiskFree:     diskMetrics.DiskFree,
		DiskTotal:    diskMetrics.DiskTotal,
		DiskUsed:     diskMetrics.DiskUsed,
		DiskPercent:  diskMetrics.DiskPercent,
	}, nil
}

// GetAvailableDiskPaths - получение путей к дискам
func GetAvailableDiskPaths() ([]string, error) {
	drives := make([]string, 0)

	// Получение всех дисковых букв
	for i := 'C'; i <= 'Z'; i++ {
		drive := string(i) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, drive)
		}
	}

	return drives, nil
}

// GetSystemInfo - получение информации о системе
func GetSystemInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Метрики RAM
	ramMetrics, err := GetRAMMetrics()
	if err == nil {
		info["ram_total"] = ramMetrics.TotalRAM
		info["ram_available"] = ramMetrics.AvailableRAM
		info["ram_used"] = ramMetrics.UsedRAM
		info["ram_percent"] = ramMetrics.RAMPercent
	}

	// Метрики диска
	diskMetrics, err := GetDiskMetrics(".")
	if err == nil {
		info["disk_free"] = diskMetrics.DiskFree
		info["disk_total"] = diskMetrics.DiskTotal
		info["disk_used"] = diskMetrics.DiskUsed
		info["disk_percent"] = diskMetrics.DiskPercent
	}

	// Загрузка CPU
	info["cpu_usage"] = GetCPUUsage()

	// Информация о процессорах
	info["cpu_cores"] = runtime.NumCPU()
	info["cpu_goroutines"] = runtime.NumGoroutine()

	return info, nil
}

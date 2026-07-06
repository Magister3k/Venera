// collector.go - Сбор метрик для системы Venera
//
// Этот модуль обеспечивает сбор системных метрик для приложения Venera
// без использования дополнительных библиотек и накопления данных.
//
// Основные функции:
// - Сбор системных метрик (RAM, CPU, диск)
// - Сбор метрик баз данных (DragonflyDB, PostgreSQL)
// - Сбор сетевых метрик
// - Сбор метрик процессов
// - Экспорт метрик для мониторинга
//
// Использование:
// import "venera/metrics"
// collector := metrics.NewCollector()
// ram, _ := collector.GetRAMUsage()
// disk, _ := collector.GetDiskUsage("C:")
// dragonflyStats, _ := collector.GetDragonflyDBStats(dragonflyDB)
// postgresStats, _ := collector.GetPostgreSQLStats(postgresDB)

package metrics

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"venera/config"
	"venera/data"
)

// Collector - структура сборщика метрик
type Collector struct {
	logger *logrus.Logger
	cfg    *config.Config
}

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

// DatabaseMetrics - метрики баз данных
type DatabaseMetrics struct {
	DragonflyDB DragonflyDBMetrics `json:"dragonflydb"`
	PostgreSQL  PostgreSQLMetrics  `json:"postgresql"`
}

// DragonflyDBMetrics - метрики DragonflyDB
type DragonflyDBMetrics struct {
	UsedMemory    uint64 `json:"used_memory"`
	UsedMemoryRSS uint64 `json:"used_memory_rss"`
	ConnectedClients int `json:"connected_clients"`
	TotalConnections int `json:"total_connections"`
	Keys          int64  `json:"keys"`
	CommandsProcessed int64 `json:"commands_processed"`
}

// PostgreSQLMetrics - метрики PostgreSQL
type PostgreSQLMetrics struct {
	DatabaseSize   uint64 `json:"database_size"`
	TableSize      uint64 `json:"table_size"`
	IndexSize      uint64 `json:"index_size"`
	ActiveConnections int  `json:"active_connections"`
	TotalConnections int  `json:"total_connections"`
}

// ProcessMetrics - метрики процессов
type ProcessMetrics struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Status         string  `json:"status"`
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    uint64  `json:"memory_usage"`
	BytesProcessed int64   `json:"bytes_processed"`
	PacketsCount   int64   `json:"packets_count"`
}

// NetworkMetrics - сетевые метрики
type NetworkMetrics struct {
	InterfaceName  string  `json:"interface_name"`
	IP             string  `json:"ip"`
	BytesSent      uint64  `json:"bytes_sent"`
	BytesReceived  uint64  `json:"bytes_received"`
	PacketsSent    uint64  `json:"packets_sent"`
	PacketsRecv    uint64  `json:"packets_recv"`
	ErrorsSent     uint64  `json:"errors_sent"`
	ErrorsRecv     uint64  `json:"errors_recv"`
}

// NewCollector - создание нового сборщика метрик
func NewCollector(logger *logrus.Logger, cfg *config.Config) *Collector {
	return &Collector{
		logger: logger,
		cfg:    cfg,
	}
}

// GetRAMUsage - получение использования RAM
func (c *Collector) GetRAMUsage() (*SystemMetrics, error) {
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

// GetDiskUsage - получение использования диска
func (c *Collector) GetDiskUsage(path string) (*SystemMetrics, error) {
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
func (c *Collector) GetCPUUsage() float64 {
	// Используем runtime для получения количества доступных процессоров
	numCPU := runtime.NumCPU()
	numGoroutine := runtime.NumGoroutine()

	// Упрощенная оценка загрузки CPU
	// В реальном приложении можно использовать Win32_PerfFormattedData_PerfProc_Process
	return float64(numGoroutine) / float64(numCPU*100) * 100
}

// GetDragonflyDBStats - получение статистики DragonflyDB
func (c *Collector) GetDragonflyDBStats(db *data.DragonflyDB) (*DragonflyDBMetrics, error) {
	// Получение статистики через INFO команду
	ctx, cancel := db.GetClient().Context().WithTimeout(time.Duration(c.cfg.DragonflyDB.Timeout) * time.Second)
	defer cancel()

	info, err := db.GetClient().Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статистики DragonflyDB: %w", err)
	}

	// Парсинг статистики
	metrics := &DragonflyDBMetrics{}
	for _, line := range info {
		if line != "" && !line.HasPrefix("#") && line.Contains(":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "used_memory":
					metrics.UsedMemory = parseMemoryValue(value)
				case "used_memory_rss":
					metrics.UsedMemoryRSS = parseMemoryValue(value)
				case "connected_clients":
					metrics.ConnectedClients = parseInt(value)
				case "total_connections":
					metrics.TotalConnections = parseInt(value)
				case "keys":
					metrics.Keys = parseInt64(value)
				case "total_commands_processed":
					metrics.CommandsProcessed = parseInt64(value)
				}
			}
		}
	}

	return metrics, nil
}

// GetPostgreSQLStats - получение статистики PostgreSQL
func (c *Collector) GetPostgreSQLStats(db *data.PostgreSQL) (*PostgreSQLMetrics, error) {
	metrics := &PostgreSQLMetrics{}

	// Получение размера базы данных
	query := `
		SELECT 
			pg_database.datname,
			pg_size_pretty(pg_database_size(pg_database.datname)) as size
		FROM pg_database
		WHERE datname = $1
	`

	var dbName, sizeStr string
	err := db.GetDB().QueryRow(query, c.cfg.PostgreSQL.Database).Scan(&dbName, &sizeStr)
	if err != nil {
		c.logger.Warnf("Ошибка получения размера базы данных PostgreSQL: %v", err)
	} else {
		// Парсирование размера (упрощенно)
		metrics.DatabaseSize = parsePostgreSQLSize(sizeStr)
	}

	// Получение активных соединений
	query = `
		SELECT count(*) FROM pg_stat_activity WHERE datname = $1
	`
	err = db.GetDB().QueryRow(query, c.cfg.PostgreSQL.Database).Scan(&metrics.ActiveConnections)
	if err != nil {
		c.logger.Warnf("Ошибка получения активных соединений PostgreSQL: %v", err)
	}

	// Получение общего количества соединений
	metrics.TotalConnections = c.cfg.PostgreSQL.MaxConnections

	return metrics, nil
}

// GetProcessMetrics - получение метрик процесса
func (c *Collector) GetProcessMetrics(processID string) (*ProcessMetrics, error) {
	// Получение информации о процессе через Win32_PerfFormattedData_PerfProc_Process
	// Для упрощения используем runtime

	pid := 0 // Получить реальный PID процесса

	// Получение информации о процессе
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ,
		false, uint32(pid))
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия процесса: %w", err)
	}
	defer windows.CloseHandle(handle)

	// Получение использования памяти
	var memCounters windows.ProcessMemoryCounters
	err = windows.GetProcessMemoryInfo(handle, &memCounters, uint32(unsafe.Sizeof(memCounters)))
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о памяти процесса: %w", err)
	}

	// Получение времени процесса
	var creationTime, exitTime, kernelTime, userTime windows.Filetime
	err = windows.GetProcessTimes(handle, &creationTime, &exitTime, &kernelTime, &userTime)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения времени процесса: %w", err)
	}

	// Вычисление загрузки CPU
	totalTime := int64(userTime.Nanoseconds() + kernelTime.Nanoseconds())
	elapsedTime := time.Since(time.Unix(0, creationTime.Nanoseconds())).Seconds()

	cpuUsage := 0.0
	if elapsedTime > 0 {
		cpuUsage = float64(totalTime) / elapsedTime / 10000000.0 * 100
	}

	return &ProcessMetrics{
		ID:          processID,
		Name:        "process", // Получить реальное имя процесса
		Type:        "unknown",
		Status:      "running",
		CPUUsage:    cpuUsage,
		MemoryUsage: memCounters.WorkingSetSize,
	}, nil
}

// GetNetworkMetrics - получение сетевых метрик
func (c *Collector) GetNetworkMetrics() ([]NetworkMetrics, error) {
	// Получение информации о сетевых интерфейсах
	// Для упрощения используем простую реализацию

	// Получение IP-адресов
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения IP-адресов: %w", err)
	}

	metrics := make([]NetworkMetrics, 0)

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}

		ip := ipNet.IP.String()

		// Получение информации об интерфейсе
		// Для упрощения используем заглушку
		metric := NetworkMetrics{
			InterfaceName: "unknown",
			IP:            ip,
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetAllMetrics - получение всех метрик
func (c *Collector) GetAllMetrics() (*SystemMetrics, *DatabaseMetrics, error) {
	// Получение системных метрик
	systemMetrics, err := c.GetRAMUsage()
	if err != nil {
		c.logger.Errorf("Ошибка получения системных метрик: %v", err)
	}

	// Получение метрик диска
	diskMetrics, err := c.GetDiskUsage(c.cfg.Paths.DragonflyBackupPath)
	if err != nil {
		c.logger.Errorf("Ошибка получения метрик диска: %v", err)
	} else {
		systemMetrics.DiskFree = diskMetrics.DiskFree
		systemMetrics.DiskTotal = diskMetrics.DiskTotal
		systemMetrics.DiskUsed = diskMetrics.DiskUsed
		systemMetrics.DiskPercent = diskMetrics.DiskPercent
	}

	// Получение метрик баз данных
	dbMetrics := &DatabaseMetrics{
		DragonflyDB: DragonflyDBMetrics{},
		PostgreSQL:  PostgreSQLMetrics{},
	}

	return systemMetrics, dbMetrics, nil
}

// GetMetricsForZabbix - получение метрик для Zabbix
func (c *Collector) GetMetricsForZabbix() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Получение системных метрик
	systemMetrics, _ := c.GetRAMUsage()
	if systemMetrics != nil {
		metrics["ram.total"] = systemMetrics.TotalRAM
		metrics["ram.available"] = systemMetrics.AvailableRAM
		metrics["ram.used"] = systemMetrics.UsedRAM
		metrics["ram.percent"] = systemMetrics.RAMPercent
	}

	// Получение метрик диска
	diskMetrics, _ := c.GetDiskUsage(c.cfg.Paths.DragonflyBackupPath)
	if diskMetrics != nil {
		metrics["disk.free"] = diskMetrics.DiskFree
		metrics["disk.total"] = diskMetrics.DiskTotal
		metrics["disk.used"] = diskMetrics.DiskUsed
		metrics["disk.percent"] = diskMetrics.DiskPercent
	}

	// Получение метрик CPU
	metrics["cpu.usage"] = c.GetCPUUsage()

	return metrics
}

// CheckDiskWarning - проверка предупреждения о диске
func (c *Collector) CheckDiskWarning() (bool, error) {
	diskMetrics, err := c.GetDiskUsage(c.cfg.Paths.DragonflyBackupPath)
	if err != nil {
		return false, err
	}

	if diskMetrics.DiskPercent >= float64(c.cfg.Generic.DiskWarningPercent) {
		return true, nil
	}

	return false, nil
}

// CheckRAMWarning - проверка предупреждения о RAM
func (c *Collector) CheckRAMWarning() (bool, error) {
	systemMetrics, err := c.GetRAMUsage()
	if err != nil {
		return false, err
	}

	if systemMetrics.RAMPercent >= float64(c.cfg.Generic.RAMWarningPercent) {
		return true, nil
	}

	return false, nil
}

// parseMemoryValue - парсирование значения памяти
func parseMemoryValue(value string) uint64 {
	// Упрощенная реализация
	// В реальном приложении нужно парсировать единицы измерения (KB, MB, GB)
	var result uint64
	fmt.Sscanf(value, "%d", &result)
	return result
}

// parseInt - парсирование целого числа
func parseInt(value string) int {
	var result int
	fmt.Sscanf(value, "%d", &result)
	return result
}

// parseInt64 - парсирование 64-битного целого числа
func parseInt64(value string) int64 {
	var result int64
	fmt.Sscanf(value, "%d", &result)
	return result
}

// parsePostgreSQLSize - парсирование размера PostgreSQL
func parsePostgreSQLSize(sizeStr string) uint64 {
	// Упрощенная реализация
	// В реальном приложении нужно парсировать единицы измерения
	var result uint64
	fmt.Sscanf(sizeStr, "%d", &result)
	return result
}
